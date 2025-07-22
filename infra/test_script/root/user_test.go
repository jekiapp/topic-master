package root

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var userTestHost = "http://localhost:4181"

type User struct {
	ID       string
	Username string
	Name     string
	Groups   []string
	Role     string
}

func getAllUsers(t *testing.T, client *http.Client, accessToken string) []User {
	req, _ := http.NewRequest("POST", userTestHost+"/api/user/list", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	require.NoError(t, err)
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
	require.NoError(t, err)
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
	return result
}

func createUser(t *testing.T, client *http.Client, accessToken, username, name, password, groupID, role string) User {
	createReq := map[string]interface{}{
		"username": username,
		"name":     name,
		"password": password,
		"groups":   []string{groupID},
		"role":     role,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", userTestHost+"/api/user/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	// Fetch user list to get the new user
	users := getAllUsers(t, client, accessToken)
	for _, u := range users {
		if u.Username == username {
			return u
		}
	}
	t.Fatalf("user %s not found after creation", username)
	return User{}
}

func editUserName(t *testing.T, client *http.Client, accessToken, userID, newName string) {
	editReq := map[string]interface{}{
		"id":   userID,
		"name": newName,
	}
	body, _ := json.Marshal(editReq)
	req, _ := http.NewRequest("POST", userTestHost+"/api/user/update-user-by-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func addUserToGroup(t *testing.T, client *http.Client, accessToken, userID, groupID string) {
	editReq := map[string]interface{}{
		"id":     userID,
		"groups": []string{groupID},
	}
	body, _ := json.Marshal(editReq)
	req, _ := http.NewRequest("POST", userTestHost+"/api/user/update-user-by-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRootUserUserIntegration(t *testing.T) {
	if envHost := os.Getenv("TOPIC_MASTER_HOST"); envHost != "" {
		userTestHost = envHost
	}

	client := &http.Client{}
	accessToken := LoginAsRoot(t, client, userTestHost)

	var createdGroups []TestGroup
	var createdUsers []User

	t.Run("user list should only have root", func(t *testing.T) {
		users := getAllUsers(t, client, accessToken)
		var rootCount int
		for _, u := range users {
			if u.Username == "root" {
				rootCount++
			}
		}
		require.Equal(t, 1, rootCount, "should only have one user (root)")
	})

	t.Run("create 2 groups", func(t *testing.T) {
		groupNames := []string{"engineering-user", "marketing-user"}
		descriptions := []string{"Engineering Team for user test", "Marketing Team for user test"}
		for i := 0; i < 2; i++ {
			g := CreateGroup(t, client, userTestHost, accessToken, groupNames[i], descriptions[i])
			createdGroups = append(createdGroups, g)
		}
		require.Len(t, createdGroups, 2, "should have 2 created groups")
	})

	t.Run("get group list", func(t *testing.T) {
		groups := GetAllGroups(t, client, userTestHost, accessToken)
		var found int
		for _, g := range groups {
			if g.Name == "engineering-user" || g.Name == "marketing-user" {
				found++
			}
		}
		require.Equal(t, 2, found, "should have engineering-user and marketing-user groups")
	})

	t.Run("create 4 users, 2 for each group with admin and member role", func(t *testing.T) {
		userInputs := []struct {
			Username string
			Name     string
			Password string
			Group    TestGroup
			Role     string
		}{
			{"alice", "Alice Smith", "alicepass", createdGroups[0], "admin"},
			{"bob", "Bob Jones", "bobpass", createdGroups[0], "member"},
			{"carol", "Carol White", "carolpass", createdGroups[1], "admin"},
			{"dave", "Dave Black", "davepass", createdGroups[1], "member"},
		}
		for _, input := range userInputs {
			u := createUser(t, client, accessToken, input.Username, input.Name, input.Password, input.Group.ID, input.Role)
			createdUsers = append(createdUsers, u)
		}
		require.Len(t, createdUsers, 4, "should have 4 created users")
	})

	t.Run("edit 1st user, change the name", func(t *testing.T) {
		require.NotEmpty(t, createdUsers)
		firstUser := createdUsers[0]
		newName := "Alice Cooper"
		editUserName(t, client, accessToken, firstUser.ID, newName)
		// Verify
		users := getAllUsers(t, client, accessToken)
		var found bool
		for _, u := range users {
			if u.ID == firstUser.ID {
				found = true
				require.Equal(t, newName, u.Name)
			}
		}
		require.True(t, found, "edited user should be found in list")
	})

	t.Run("edit 2nd user, add to the 2nd group as member", func(t *testing.T) {
		require.Len(t, createdUsers, 4)
		secondUser := createdUsers[1]
		addUserToGroup(t, client, accessToken, secondUser.ID, createdGroups[1].ID)
		// Verify
		users := getAllUsers(t, client, accessToken)
		var found bool
		for _, u := range users {
			if u.ID == secondUser.ID {
				found = true
				var inGroup bool
				for _, gid := range u.Groups {
					if gid == createdGroups[1].ID {
						inGroup = true
					}
				}
				require.True(t, inGroup, "user should be in the 2nd group")
			}
		}
		require.True(t, found, "edited user should be found in list")
	})
}
