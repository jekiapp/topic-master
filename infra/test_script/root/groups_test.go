package root

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var groupTestHost = "http://localhost:4181"

type Group struct {
	ID          string
	Name        string
	Description string
	Members     string
}

func loginAsRoot(t *testing.T, client *http.Client) string {
	loginPayload := map[string]string{
		"username": "root",
		"password": "rootroot",
	}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", groupTestHost+"/api/login", bytes.NewReader(loginBody))
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

func getAllGroups(t *testing.T, client *http.Client, accessToken string) []Group {
	groupListReq, _ := http.NewRequest("POST", groupTestHost+"/api/group/list", bytes.NewReader([]byte(`{}`)))
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
	var result []Group
	for _, g := range groupList.Data.Groups {
		result = append(result, Group{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Members:     g.Members,
		})
	}
	return result
}

func TestRootUserGroupListIntegration(t *testing.T) {
	if envHost := os.Getenv("TOPIC_MASTER_HOST"); envHost != "" {
		groupTestHost = envHost
	}

	client := &http.Client{}
	accessToken := loginAsRoot(t, client)

	var createdGroups []Group

	t.Run("group list should only have root", func(t *testing.T) {
		groups := getAllGroups(t, client, accessToken)
		var rootCount int
		for _, g := range groups {
			if g.Name == "root" {
				rootCount++
			}
		}
		require.Equal(t, 1, rootCount, "should only have one group (root)")
	})

	t.Run("create 3 groups", func(t *testing.T) {
		groupNames := []string{"engineering", "marketing", "support"}
		descriptions := []string{"Engineering Team", "Marketing Team", "Support Team"}
		for i := 0; i < 3; i++ {
			createReq := map[string]string{
				"name":        groupNames[i],
				"description": descriptions[i],
			}
			body, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", groupTestHost+"/api/group/create", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
		}
		// Fetch group list and store for later
		groups := getAllGroups(t, client, accessToken)
		createdGroups = nil
		for _, g := range groups {
			if g.Name != "root" {
				createdGroups = append(createdGroups, g)
			}
		}
		require.Len(t, createdGroups, 3, "should have 3 created groups (excluding root)")
	})

	t.Run("edit the first group, change the description", func(t *testing.T) {
		require.NotEmpty(t, createdGroups)
		firstGroup := createdGroups[0]
		newDesc := "Updated Engineering Description"
		editReq := map[string]string{
			"id":          firstGroup.ID,
			"description": newDesc,
		}
		body, _ := json.Marshal(editReq)
		req, _ := http.NewRequest("POST", groupTestHost+"/api/group/update-group-by-id", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		// Verify update
		groups := getAllGroups(t, client, accessToken)
		var found bool
		for _, g := range groups {
			if g.ID == firstGroup.ID {
				found = true
				require.Equal(t, newDesc, g.Description)
			}
		}
		require.True(t, found, "edited group should be found in list")
	})

	t.Run("delete the second group", func(t *testing.T) {
		require.Len(t, createdGroups, 3)
		secondGroup := createdGroups[1]
		deleteReq := map[string]string{
			"id": secondGroup.ID,
		}
		body, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest("POST", groupTestHost+"/api/group/delete-group", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		// Verify deletion
		groups := getAllGroups(t, client, accessToken)
		for _, g := range groups {
			require.NotEqual(t, secondGroup.ID, g.ID, "deleted group should not be in the list")
		}
	})

	t.Run("deleting root group should be failed", func(t *testing.T) {
		groups := getAllGroups(t, client, accessToken)
		var rootGroup *Group
		for _, g := range groups {
			if g.Name == "root" {
				rootGroup = &g
				break
			}
		}
		require.NotNil(t, rootGroup, "root group should exist")
		deleteReq := map[string]string{
			"id": rootGroup.ID,
		}
		body, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest("POST", groupTestHost+"/api/group/delete-group", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		// Should not be 200 OK for forbidden delete
		require.NotEqual(t, http.StatusOK, resp.StatusCode, "deleting root group should not be allowed")
	})
}
