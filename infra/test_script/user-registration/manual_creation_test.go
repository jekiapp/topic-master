package userregistration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	helpers "github.com/jekiapp/topic-master/infra/test_script/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManualUserCreationAndPasswordChange(t *testing.T) {
	client := &http.Client{}
	accessToken := helpers.LoginAsRoot(t, client, helpers.GetHost())
	group := helpers.CreateGroup(
		t,
		client,
		helpers.GetHost(),
		accessToken,
		"engineering-user",
		"Engineering Team for manual creation test",
	)

	// Step 2: Create a user with a generated password
	username := "bob"
	name := "Bob Marley"
	initialPassword := "bobpass"
	role := "member"
	u, err := helpers.CreateUser(
		client,
		accessToken,
		username,
		name,
		initialPassword,
		[]helpers.GroupsReq{{GroupID: group.ID, Role: role}},
	)
	require.NoError(t, err)
	require.Equal(t, username, u.Username)

	// Step 3: Try to login with the generated password (should require password change)
	loginResp, _ := helpers.LoginUser(t, client, username, initialPassword)
	defer loginResp.Body.Close()
	body, _ := io.ReadAll(loginResp.Body)
	var loginResult map[string]interface{}
	_ = json.Unmarshal(body, &loginResult)

	var token string
	if redirect, ok := loginResult["redirect"].(string); ok {
		// Extract token from redirect URL
		const prefix = "/reset-password?token="
		if idx := bytes.Index([]byte(redirect), []byte(prefix)); idx != -1 {
			token = redirect[idx+len(prefix):]
		}
	}
	if token == "" {
		t.Fatalf("expected redirect with token in login response, got: %v", loginResult)
	}

	// Step 4: GET to /api/user/reset-password with token to validate and get username
	getURL := helpers.GetHost() + "/api/user/reset-password?token=" + token
	getResp, err := client.Get(getURL)
	require.NoError(t, err)
	defer getResp.Body.Close()
	getRespBody, _ := io.ReadAll(getResp.Body)
	if !assert.Equal(t, http.StatusOK, getResp.StatusCode) {
		fmt.Println(string(getRespBody))
	}
	var getResult struct {
		Data struct {
			Username string `json:"username"`
		} `json:"data"`
	}
	_ = json.Unmarshal(getRespBody, &getResult)
	if getResult.Data.Username != username {
		t.Fatalf("expected username in reset GET, got: %v", getResult)
	}

	// Step 5: Reset the password using the reset-password endpoint
	newPassword := "bobnewpass"
	resetReq := map[string]string{
		"token":            token,
		"new_password":     newPassword,
		"confirm_password": newPassword,
	}
	resetBody, _ := json.Marshal(resetReq)
	resetReqObj, _ := http.NewRequest("POST", helpers.GetHost()+"/api/user/reset-password", bytes.NewReader(resetBody))
	resetReqObj.Header.Set("Content-Type", "application/json")
	resetResp, err := client.Do(resetReqObj)
	require.NoError(t, err)
	defer resetResp.Body.Close()
	resetRespBody, _ := io.ReadAll(resetResp.Body)
	if !assert.Equal(t, http.StatusOK, resetResp.StatusCode) {
		fmt.Println(string(resetRespBody))
	}
	var resetResult struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(resetRespBody, &resetResult)
	if resetResult.Status != "success" {
		t.Fatalf("expected password reset success, got: %v body: %s", resetResult, string(resetRespBody))
	}

	// Step 6: Login with the new password (should succeed)
	loginResp2, _ := helpers.LoginUser(t, client, username, newPassword)
	defer loginResp2.Body.Close()
	require.Equal(t, http.StatusOK, loginResp2.StatusCode)
}
