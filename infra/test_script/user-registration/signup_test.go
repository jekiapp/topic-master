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

func TestSignupAndViewApplication(t *testing.T) {
	client := &http.Client{}
	accessToken := helpers.LoginAsRoot(t, client, helpers.GetHost())
	suffix := "8687"
	group := helpers.CreateGroup(
		t,
		client,
		helpers.GetHost(),
		accessToken,
		"engineering-signup-"+suffix,
		"Engineering Team for signup test",
	)

	var applicationID string

	t.Run("signup", func(t *testing.T) {
		signupReq := map[string]interface{}{
			"username":         "alice-" + suffix,
			"name":             "Alice Smith",
			"password":         "alicepass",
			"confirm_password": "alicepass",
			"reason":           "I want to join engineering",
			"group_id":         group.ID,
			"group_name":       group.Name,
			"group_role":       "member",
		}
		body, _ := json.Marshal(signupReq)
		resp, err := client.Post(helpers.GetHost()+"/api/signup", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var signupResp struct {
			Data struct {
				ApplicationID string `json:"application_id"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&signupResp)
		require.NoError(t, err)
		require.NotEmpty(t, signupResp.Data.ApplicationID, "application_id should be present")
		applicationID = signupResp.Data.ApplicationID
	})

	aliceID := ""
	t.Run("check detail", func(t *testing.T) {
		require.NotEmpty(t, applicationID, "application_id should be set from signup subtest")
		appDetailReq, _ := http.NewRequest("GET", helpers.GetHost()+"/api/signup/app?id="+applicationID, nil)
		appDetailResp, err := client.Do(appDetailReq)
		require.NoError(t, err)
		defer appDetailResp.Body.Close()
		require.Equal(t, http.StatusOK, appDetailResp.StatusCode)
		var appDetail struct {
			Data struct {
				Application struct {
					ID     string `json:"id"`
					UserID string `json:"user_id"`
					Status string `json:"status"`
					Reason string `json:"reason"`
					Type   string `json:"type"`
				} `json:"application"`
				User struct {
					ID       string `json:"id"`
					Username string `json:"username"`
					Name     string `json:"name"`
					Status   string `json:"status"`
				} `json:"user"`
			} `json:"data"`
		}
		bodyBytes, _ := io.ReadAll(appDetailResp.Body)
		err = json.Unmarshal(bodyBytes, &appDetail)
		require.NoError(t, err, "failed to decode signup app detail: %s", string(bodyBytes))
		require.Equal(t, applicationID, appDetail.Data.Application.ID)
		require.Equal(t, "alice-"+suffix, appDetail.Data.User.Username)
		require.Equal(t, "Alice Smith", appDetail.Data.User.Name)
		aliceID = appDetail.Data.User.ID
	})

	t.Run("root assignment list", func(t *testing.T) {
		req, _ := http.NewRequest("GET", helpers.GetHost()+"/api/tickets/list-my-assignment", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var assignmentResp struct {
			Data struct {
				Applications []struct {
					ID            string `json:"id"`
					ApplicantName string `json:"applicant_name"`
				} `json:"applications"`
				HasNext bool `json:"has_next"`
			} `json:"data"`
		}
		body, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &assignmentResp)
		require.NoError(t, err, "failed to decode assignment list: %s", string(body))
		found := false
		for _, app := range assignmentResp.Data.Applications {
			if app.ID == applicationID {
				found = true
				break
			}
		}
		require.True(t, found, "application_id %s should be present in root's assignment list", applicationID)
	})

	t.Run("root open application detail", func(t *testing.T) {
		req, _ := http.NewRequest("GET", helpers.GetHost()+"/api/tickets/detail?id="+applicationID, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var detailResp struct {
			Data struct {
				Ticket struct {
					ID string `json:"id"`
				} `json:"ticket"`
				Applicant struct {
					Username string `json:"username"`
					Name     string `json:"name"`
				} `json:"applicant"`
			} `json:"data"`
		}
		body, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &detailResp)
		require.NoError(t, err, "failed to decode ticket detail: %s", string(body))
		require.Equal(t, applicationID, detailResp.Data.Ticket.ID)
		require.Equal(t, "alice-"+suffix, detailResp.Data.Applicant.Username)
		require.Equal(t, "Alice Smith", detailResp.Data.Applicant.Name)
	})

	t.Run("root approve application", func(t *testing.T) {
		approveReq := map[string]interface{}{
			"action":         "approve",
			"application_id": applicationID,
		}
		body, _ := json.Marshal(approveReq)
		req, _ := http.NewRequest("POST", helpers.GetHost()+"/api/tickets/action", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var approveResp struct {
			Data struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			} `json:"data"`
		}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &approveResp)
		require.NoError(t, err, "failed to decode approve response: %s", string(respBody))
		require.Equal(t, "success", approveResp.Data.Status)
	})

	t.Run("user can login after approval", func(t *testing.T) {
		loginResp, _ := helpers.LoginUser(t, client, "alice-"+suffix, "alicepass")
		defer loginResp.Body.Close()
		if !assert.Equal(t, http.StatusOK, loginResp.StatusCode) {
			body, _ := io.ReadAll(loginResp.Body)
			fmt.Println(string(body))
		}
	})

	t.Run("cleanup", func(t *testing.T) {
		helpers.DeleteUser(t, client, accessToken, aliceID)
		helpers.DeleteGroup(t, client, accessToken, group.ID)
	})
}
