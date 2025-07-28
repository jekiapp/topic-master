package root

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/jekiapp/topic-master/infra/test_script/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func editUserName(t *testing.T, client *http.Client, accessToken string, user helpers.User, newName string) {
	editReq := map[string]interface{}{
		"username": user.Username,
		"groups":   user.GroupDetails,
		"name":     newName,
	}
	body, _ := json.Marshal(editReq)
	req, _ := http.NewRequest("POST", helpers.GetHost()+"/api/user/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	}
}

func addUserToGroup(t *testing.T, client *http.Client, accessToken string, user helpers.User, groups []helpers.GroupsReq) {
	editReq := map[string]interface{}{
		"username": user.Username,
		"name":     user.Name,
		"groups":   groups,
	}
	body, _ := json.Marshal(editReq)
	req, _ := http.NewRequest("POST", helpers.GetHost()+"/api/user/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	}
}

func TestRootUserUserIntegration(t *testing.T) {
	const randomSuffix = "8382"

	client := &http.Client{}
	accessToken := helpers.LoginAsRoot(t, client, helpers.GetHost())

	var createdGroups []helpers.TestGroup
	var createdUsers []helpers.User

	t.Run("user list should only have root", func(t *testing.T) {
		users, err := helpers.GetAllUsers(client, accessToken)
		require.NoError(t, err)
		var rootCount int
		for _, u := range users {
			if u.Username == "root" {
				rootCount++
			}
		}
		require.Equal(t, 1, rootCount, "should only have one user (root)")
	})

	t.Run("create 2 groups", func(t *testing.T) {
		groupNames := []string{"engineering-user-" + randomSuffix, "marketing-user-" + randomSuffix}
		descriptions := []string{"Engineering Team for user test", "Marketing Team for user test"}
		for i := 0; i < 2; i++ {
			g := helpers.CreateGroup(t, client, helpers.GetHost(), accessToken, groupNames[i], descriptions[i])
			createdGroups = append(createdGroups, g)
		}
		require.Len(t, createdGroups, 2, "should have 2 created groups")
	})

	t.Run("get group list", func(t *testing.T) {
		groups, err := helpers.GetAllGroups(t, client, helpers.GetHost(), accessToken)
		require.NoError(t, err)
		var found int
		for _, g := range groups {
			if g.Name == "engineering-user-"+randomSuffix || g.Name == "marketing-user-"+randomSuffix {
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
			Group    helpers.TestGroup
			Role     string
		}{
			{"alice-" + randomSuffix, "Alice Smith", "alicepass", createdGroups[0], "admin"},
			{"bob-" + randomSuffix, "Bob Jones", "bobpass", createdGroups[0], "member"},
			{"carol-" + randomSuffix, "Carol White", "carolpass", createdGroups[1], "admin"},
			{"dave-" + randomSuffix, "Dave Black", "davepass", createdGroups[1], "member"},
		}
		for _, input := range userInputs {
			u, err := helpers.CreateUser(client, accessToken, input.Username, input.Name, input.Password, []helpers.GroupsReq{{GroupID: input.Group.ID, Role: input.Role}})
			require.NoError(t, err)
			createdUsers = append(createdUsers, u)
		}
		require.Len(t, createdUsers, 4, "should have 4 created users")
	})

	t.Run("edit 1st user, change the name", func(t *testing.T) {
		require.NotEmpty(t, createdUsers)
		firstUser := createdUsers[0]
		newName := "Alice Cooper"
		editUserName(t, client, accessToken, firstUser, newName)
		// Verify
		users, err := helpers.GetAllUsers(client, accessToken)
		require.NoError(t, err)
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
		addUserToGroup(t, client, accessToken, secondUser, []helpers.GroupsReq{{GroupID: createdGroups[1].ID, Role: "member"}})
		// Verify
		users, err := helpers.GetAllUsers(client, accessToken)
		require.NoError(t, err)
		var found bool
		for _, u := range users {
			if u.ID == secondUser.ID {
				found = true
				var inGroup bool
				for _, g := range u.GroupDetails {
					if g.GroupID == createdGroups[1].ID {
						inGroup = true
					}
				}
				require.True(t, inGroup, "user should be in the 2nd group")
			}
		}
		require.True(t, found, "edited user should be found in list")
	})

	t.Cleanup(func() {
		// Clean up created users (except root)
		for _, u := range createdUsers {
			if u.Username != "root" {
				helpers.DeleteUser(t, client, accessToken, u.ID)
			}
		}
		// Clean up created groups (except root)
		for _, g := range createdGroups {
			if g.Name != "root" {
				helpers.DeleteGroup(t, client, accessToken, g.ID)
			}
		}
	})
}
