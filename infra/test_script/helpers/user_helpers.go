package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type User struct {
	ID       string
	Username string
	Name     string
	Groups   []string
	Role     string
}

type TestGroup struct {
	ID          string
	Name        string
	Description string
	Members     string
}

// GetHost returns the test host, respecting the USER_TEST_HOST environment variable if set.
func GetHost() string {
	host := os.Getenv("USER_TEST_HOST")
	if host != "" {
		return host
	}
	return "http://localhost:4181"
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

func GetAllGroups(t *testing.T, client *http.Client, host, accessToken string) ([]TestGroup, error) {
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
	return result, nil
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
	groups, err := GetAllGroups(t, client, host, accessToken)
	require.NoError(t, err)
	for _, g := range groups {
		if g.Name == name {
			return g
		}
	}
	t.Fatalf("group %s not found after creation", name)
	return TestGroup{}
}

func GetAllUsers(client *http.Client, accessToken string) ([]User, error) {
	req, _ := http.NewRequest("POST", GetHost()+"/api/user/list", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var userList struct {
		Data struct {
			Users []struct {
				ID       string `json:"id"`
				Username string `json:"username"`
				Name     string `json:"name"`
				Groups   string `json:"groups"`
				Role     string `json:"role"`
			} `json:"users"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &userList)
	if err != nil {
		return nil, err
	}
	var result []User
	for _, u := range userList.Data.Users {
		result = append(result, User{
			ID:       u.ID,
			Username: u.Username,
			Name:     u.Name,
			Groups:   strings.Split(u.Groups, ","),
			Role:     u.Role,
		})
	}
	return result, nil
}

type GroupsReq struct {
	GroupID string `json:"group_id"`
	Role    string `json:"role"`
}

func CreateUser(client *http.Client, accessToken, username, name, password string, groups []GroupsReq) (User, error) {
	createReq := map[string]interface{}{
		"username": username,
		"name":     name,
		"password": password,
		"groups":   groups,
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", GetHost()+"/api/user/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()
	bodyresp, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return User{}, fmt.Errorf("failed to create user %s: %s, %s", username, resp.Status, string(bodyresp))
	}
	users, err := GetAllUsers(client, accessToken)
	if err != nil {
		return User{}, err
	}
	for _, u := range users {
		if u.Username == username {
			return u, nil
		}
	}
	return User{}, fmt.Errorf("user %s not found after creation", username)
}

// LoginUser logs in a user and returns the response and cookies.
func LoginUser(t *testing.T, client *http.Client, username, password string) (*http.Response, []*http.Cookie) {
	loginPayload := map[string]string{
		"username": username,
		"password": password,
	}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", GetHost()+"/api/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := client.Do(loginReq)
	require.NoError(t, err)
	return loginResp, loginResp.Cookies()
}
