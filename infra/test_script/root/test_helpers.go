package root

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestGroup struct {
	ID          string
	Name        string
	Description string
	Members     string
}

func LoginAsRoot(t *testing.T, client *http.Client, host string) string {
	loginPayload := map[string]string{
		"username": "root",
		"password": "rootroot",
	}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", host+"/api/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := client.Do(loginReq)
	require.NoError(t, err)
	defer loginResp.Body.Close()
	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var accessToken string
	for _, c := range loginResp.Cookies() {
		if c.Name == "access_token" {
			accessToken = c.Value
		}
	}
	require.NotEmpty(t, accessToken, "access_token cookie should be set after login")
	return accessToken
}

func GetAllGroups(t *testing.T, client *http.Client, host, accessToken string) []TestGroup {
	groupListReq, _ := http.NewRequest("POST", host+"/api/group/list", bytes.NewReader([]byte(`{}`)))
	groupListReq.Header.Set("Content-Type", "application/json")
	groupListReq.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	groupListResp, err := client.Do(groupListReq)
	require.NoError(t, err)
	defer groupListResp.Body.Close()
	body, _ := io.ReadAll(groupListResp.Body)
	var groupList struct {
		Data struct {
			Groups []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Members     string `json:"members"`
			} `json:"groups"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &groupList)
	require.NoError(t, err)
	var result []TestGroup
	for _, g := range groupList.Data.Groups {
		result = append(result, TestGroup{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Members:     g.Members,
		})
	}
	return result
}

func CreateGroup(t *testing.T, client *http.Client, host, accessToken, name, description string) TestGroup {
	createReq := map[string]string{
		"name":        name,
		"description": description,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", host+"/api/group/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	// Fetch group list to get the new group
	groups := GetAllGroups(t, client, host, accessToken)
	for _, g := range groups {
		if g.Name == name {
			return g
		}
	}
	t.Fatalf("group %s not found after creation", name)
	return TestGroup{}
}
