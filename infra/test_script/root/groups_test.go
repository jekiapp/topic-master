package root

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var groupTestHost = "http://localhost:4181"

func TestRootUserGroupListIntegration(t *testing.T) {
	if envHost := os.Getenv("TOPIC_MASTER_HOST"); envHost != "" {
		groupTestHost = envHost
	}

	client := &http.Client{}
	accessToken := LoginAsRoot(t, client, groupTestHost)

	var createdGroups []TestGroup

	t.Run("group list should only have root", func(t *testing.T) {
		groups := GetAllGroups(t, client, groupTestHost, accessToken)
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
			createdGroup := CreateGroup(t, client, groupTestHost, accessToken, groupNames[i], descriptions[i])
			createdGroups = append(createdGroups, createdGroup)
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
		groups := GetAllGroups(t, client, groupTestHost, accessToken)
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
		groups := GetAllGroups(t, client, groupTestHost, accessToken)
		for _, g := range groups {
			require.NotEqual(t, secondGroup.ID, g.ID, "deleted group should not be in the list")
		}
	})

	t.Run("deleting root group should be failed", func(t *testing.T) {
		groups := GetAllGroups(t, client, groupTestHost, accessToken)
		var rootGroup *TestGroup
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
