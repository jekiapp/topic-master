package loggedin

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	helpers "github.com/jekiapp/topic-master/infra/test_script/helpers"
	"github.com/stretchr/testify/require"
)

func TestLoggedIn(t *testing.T) {
	client := &http.Client{}
	rootToken := helpers.LoginAsRoot(t, client, helpers.GetHost())
	groupPayment := helpers.CreateGroup(
		t,
		client,
		helpers.GetHost(),
		rootToken,
		"payment-team",
		"Payment Team for logged in test",
	)
	groupOrder := helpers.CreateGroup(
		t,
		client,
		helpers.GetHost(),
		rootToken,
		"order-team",
		"Order Team for logged in test",
	)

	aliceToken := UserSignup(t, "alice", rootToken, groupPayment)
	bobToken := UserSignup(t, "bob", rootToken, groupOrder)
	charlieToken := UserSignup(t, "charlie", rootToken, groupPayment)

	// alice can list all topics
	t.Run("alice list topics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", helpers.GetHost()+"/api/topic/list-all-topics", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to GET /api/topic/list-all-topics: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}
		var result struct {
			Data struct {
				Topics []struct {
					ID         string `json:"id"`
					Name       string `json:"name"`
					Bookmarked bool   `json:"bookmarked"`
				} `json:"topics"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(result.Data.Topics) < 3 {
			t.Fatalf("expected at least 3 topics, got %d", len(result.Data.Topics))
		}
	})

	// alice bookmark 3 top topics -> validate bookmark
	var bookmarkedTopicIDs []string
	t.Run("alice bookmark topics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", helpers.GetHost()+"/api/topic/list-all-topics", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to GET /api/topic/list-all-topics: %v", err)
		}
		defer resp.Body.Close()
		var result struct {
			Data struct {
				Topics []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"topics"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		for i := 0; i < 3; i++ {
			topic := result.Data.Topics[i]
			bookmarkReq := map[string]interface{}{
				"entity_id": topic.ID,
				"bookmark":  true,
			}
			body, _ := json.Marshal(bookmarkReq)
			bookmarkRequest, _ := http.NewRequest("POST", helpers.GetHost()+"/api/entity/toggle-bookmark", bytes.NewReader(body))
			bookmarkRequest.Header.Set("Content-Type", "application/json")
			bookmarkRequest.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
			bookmarkResp, err := client.Do(bookmarkRequest)
			if err != nil {
				t.Fatalf("failed to POST toggle-bookmark: %v", err)
			}
			defer bookmarkResp.Body.Close()
			if bookmarkResp.StatusCode != http.StatusOK {
				respBody, _ := io.ReadAll(bookmarkResp.Body)
				t.Fatalf("unexpected status code for bookmark: %d, body: %s", bookmarkResp.StatusCode, string(respBody))
			}
			bookmarkedTopicIDs = append(bookmarkedTopicIDs, topic.ID)
		}

		// Validate bookmarks using is_bookmarked=true
		validateReq, _ := http.NewRequest("GET", helpers.GetHost()+"/api/topic/list-all-topics?is_bookmarked=true", nil)
		validateReq.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		validateResp, err := client.Do(validateReq)
		if err != nil {
			t.Fatalf("failed to GET /api/topic/list-all-topics?is_bookmarked=true: %v", err)
		}
		defer validateResp.Body.Close()
		var validateResult struct {
			Data struct {
				Topics []struct {
					ID string `json:"id"`
				} `json:"topics"`
			} `json:"data"`
		}
		err = json.NewDecoder(validateResp.Body).Decode(&validateResult)
		if err != nil {
			t.Fatalf("failed to decode validate response: %v", err)
		}
		found := 0
		for _, topic := range validateResult.Data.Topics {
			for _, id := range bookmarkedTopicIDs {
				if topic.ID == id {
					found++
				}
			}
		}
		if found != 3 {
			t.Fatalf("expected 3 bookmarked topics, got %d", found)
		}
	})

	// alice can get topic detail topic[0] -> validate to be bookmarked
	var cachedNsqdHosts string
	var cachedTopicName string
	t.Run("alice topic detail bookmarked", func(t *testing.T) {
		if len(bookmarkedTopicIDs) == 0 {
			t.Skip("no bookmarked topics from previous step")
		}
		topicID := bookmarkedTopicIDs[0]
		detailReq, _ := http.NewRequest("GET", helpers.GetHost()+"/api/topic/detail?topic="+topicID, nil)
		detailReq.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		detailResp, err := client.Do(detailReq)
		if err != nil {
			t.Fatalf("failed to GET /api/topic/detail: %v", err)
		}
		defer detailResp.Body.Close()
		if detailResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(detailResp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", detailResp.StatusCode, string(body))
		}
		var detail struct {
			Data struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				Bookmarked bool   `json:"bookmarked"`
				NsqdHosts  []struct {
					Address string `json:"address"`
				} `json:"nsqd_hosts"`
			} `json:"data"`
		}
		err = json.NewDecoder(detailResp.Body).Decode(&detail)
		if err != nil {
			t.Fatalf("failed to decode detail response: %v", err)
		}
		if !detail.Data.Bookmarked {
			t.Fatalf("expected topic to be bookmarked, got false")
		}
		// Cache nsqd_hosts and topic name for later use
		hosts := ""
		for i, h := range detail.Data.NsqdHosts {
			hosts += h.Address
			if i < len(detail.Data.NsqdHosts)-1 {
				hosts += ","
			}
		}
		cachedNsqdHosts = hosts
		cachedTopicName = detail.Data.Name
	})

	// alice claim the topic to be owned by payment team
	var aliceClaimTicketID string
	var aliceChannelClaimTicketID string
	t.Run("alice claim topic", func(t *testing.T) {
		if len(bookmarkedTopicIDs) == 0 {
			t.Skip("no bookmarked topics from previous step")
		}
		topicID := bookmarkedTopicIDs[0]
		claimReq := map[string]interface{}{
			"entity_id":  topicID,
			"group_id":   groupPayment.ID,
			"group_name": groupPayment.Name,
		}
		body, _ := json.Marshal(claimReq)
		claimRequest, _ := http.NewRequest("POST", helpers.GetHost()+"/api/entity/claim", bytes.NewReader(body))
		claimRequest.Header.Set("Content-Type", "application/json")
		claimRequest.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		claimResp, err := client.Do(claimRequest)
		if err != nil {
			t.Fatalf("failed to POST claim: %v", err)
		}
		defer claimResp.Body.Close()
		if claimResp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(claimResp.Body)
			t.Fatalf("unexpected status code for claim: %d, body: %s", claimResp.StatusCode, string(respBody))
		}
		var claimResult struct {
			Data struct {
				ApplicationID string `json:"application_id"`
			} `json:"data"`
		}
		err = json.NewDecoder(claimResp.Body).Decode(&claimResult)
		if err == nil && claimResult.Data.ApplicationID != "" {
			aliceClaimTicketID = claimResult.Data.ApplicationID
		}
		require.NotEmpty(t, aliceClaimTicketID)
	})

	// alice should be able to see application detail (from signup)
	t.Run("alice application detail", func(t *testing.T) {
		// Get alice's own applications
		applicationsReq, _ := http.NewRequest("GET", helpers.GetHost()+"/api/tickets/list-my-applications", nil)
		applicationsReq.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		applicationsResp, err := client.Do(applicationsReq)
		if err != nil {
			t.Fatalf("failed to GET /api/tickets/list-my-applications: %v", err)
		}
		defer applicationsResp.Body.Close()
		if applicationsResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(applicationsResp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", applicationsResp.StatusCode, string(body))
		}
		var applicationsResult struct {
			Data struct {
				Applications []struct {
					ID string `json:"id"`
				} `json:"applications"`
			} `json:"data"`
		}
		err = json.NewDecoder(applicationsResp.Body).Decode(&applicationsResult)
		if err != nil {
			t.Fatalf("failed to decode applications response: %v", err)
		}
		if len(applicationsResult.Data.Applications) == 0 {
			t.Fatalf("expected at least one application for alice")
		}
		// Try to get detail for the first application
		appID := applicationsResult.Data.Applications[0].ID
		detailReq, _ := http.NewRequest("GET", helpers.GetHost()+"/api/signup/app?id="+appID, nil)
		detailReq.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		detailResp, err := client.Do(detailReq)
		if err != nil {
			t.Fatalf("failed to GET /api/signup/app: %v", err)
		}
		defer detailResp.Body.Close()
		if detailResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(detailResp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", detailResp.StatusCode, string(body))
		}
		var detail struct {
			Data struct {
				Application struct {
					ID string `json:"id"`
				} `json:"application"`
			} `json:"data"`
		}
		err = json.NewDecoder(detailResp.Body).Decode(&detail)
		if err != nil {
			t.Fatalf("failed to decode application detail: %v", err)
		}
		if detail.Data.Application.ID != appID {
			t.Fatalf("expected application ID %s, got %s", appID, detail.Data.Application.ID)
		}
	})

	// in the topic detail, alice also claim the channel[0] to be owned by payment team
	t.Run("alice claim channel", func(t *testing.T) {
		if len(bookmarkedTopicIDs) == 0 {
			t.Skip("no bookmarked topics from previous step")
		}
		if cachedNsqdHosts == "" || cachedTopicName == "" {
			t.Skip("no nsqd hosts or topic name cached from topic detail")
		}
		// List channels for the topic with hosts param (using topic name)
		channelsReq, _ := http.NewRequest(
			"GET",
			helpers.GetHost()+"/api/topic/nsq/list-channels?topic="+cachedTopicName+"&hosts="+cachedNsqdHosts,
			nil,
		)
		channelsReq.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		channelsResp, err := client.Do(channelsReq)
		if err != nil {
			t.Fatalf("failed to GET /api/topic/nsq/list-channels: %v", err)
		}
		defer channelsResp.Body.Close()
		if channelsResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(channelsResp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", channelsResp.StatusCode, string(body))
		}
		var channelsResult struct {
			Data struct {
				Channels []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"channels"`
			} `json:"data"`
		}
		err = json.NewDecoder(channelsResp.Body).Decode(&channelsResult)
		if err != nil {
			t.Fatalf("failed to decode channels response: %v", err)
		}
		if len(channelsResult.Data.Channels) == 0 {
			t.Skip("no channels found for topic")
		}
		channel := channelsResult.Data.Channels[0]
		claimReq := map[string]interface{}{
			"entity_id":  channel.ID,
			"group_id":   groupPayment.ID,
			"group_name": groupPayment.Name,
		}
		body, _ := json.Marshal(claimReq)
		claimRequest, _ := http.NewRequest("POST", helpers.GetHost()+"/api/entity/claim", bytes.NewReader(body))
		claimRequest.Header.Set("Content-Type", "application/json")
		claimRequest.AddCookie(&http.Cookie{Name: "access_token", Value: aliceToken})
		claimResp, err := client.Do(claimRequest)
		if err != nil {
			t.Fatalf("failed to POST claim for channel: %v", err)
		}
		defer claimResp.Body.Close()
		if claimResp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(claimResp.Body)
			t.Fatalf("unexpected status code for channel claim: %d, body: %s", claimResp.StatusCode, string(respBody))
		}
		var claimResult struct {
			Data struct {
				ApplicationID string `json:"application_id"`
			} `json:"data"`
		}
		err = json.NewDecoder(claimResp.Body).Decode(&claimResult)
		if err == nil && claimResult.Data.ApplicationID != "" {
			aliceChannelClaimTicketID = claimResult.Data.ApplicationID
		}
		require.NotEmpty(t, aliceChannelClaimTicketID)
	})

	// charlie can list tickets
	t.Run("charlie list tickets", func(t *testing.T) {
		req, _ := http.NewRequest("GET", helpers.GetHost()+"/api/tickets/list-my-assignment?page=1&limit=100", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: charlieToken})
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to GET /api/tickets/list-my-assignment: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}
		var result struct {
			Data struct {
				Tickets []struct {
					ID string `json:"id"`
				} `json:"applications"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(result.Data.Tickets) != 2 {
			t.Fatalf("expected 2 tickets for charlie (both alice's claims), got %d", len(result.Data.Tickets))
		}
		// Ensure both alice's claim tickets are present
		foundTopic := false
		foundChannel := false
		for _, ticket := range result.Data.Tickets {
			if ticket.ID == aliceClaimTicketID {
				foundTopic = true
			}
			if ticket.ID == aliceChannelClaimTicketID {
				foundChannel = true
			}
		}
		if !foundTopic || !foundChannel {
			t.Fatalf("expected to find both alice's claim tickets in charlie's assignment list")
		}
	})
	t.Run("charlie approve alice claims", func(t *testing.T) {
		for _, ticketID := range []string{aliceClaimTicketID, aliceChannelClaimTicketID} {
			if ticketID == "" {
				t.Fatalf("no alice claim ticket to approve")
			}
			approveReq := map[string]interface{}{
				"action":         "approve",
				"application_id": ticketID,
			}
			body, _ := json.Marshal(approveReq)
			approveRequest, _ := http.NewRequest("POST", helpers.GetHost()+"/api/tickets/action", bytes.NewReader(body))
			approveRequest.Header.Set("Content-Type", "application/json")
			approveRequest.AddCookie(&http.Cookie{Name: "access_token", Value: charlieToken})
			approveResp, err := client.Do(approveRequest)
			if err != nil {
				t.Fatalf("failed to POST approve ticket: %v", err)
			}
			defer approveResp.Body.Close()
			if approveResp.StatusCode != http.StatusOK {
				respBody, _ := io.ReadAll(approveResp.Body)
				t.Fatalf("unexpected status code for approve: %d, body: %s", approveResp.StatusCode, string(respBody))
			}
		}
	})

	// bob can list topics
	t.Run("bob list topics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", helpers.GetHost()+"/api/topic/list-all-topics", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: bobToken})
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to GET /api/topic/list-all-topics: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}
		var result struct {
			Data struct {
				Topics []struct {
					ID string `json:"id"`
				} `json:"topics"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(result.Data.Topics) == 0 {
			t.Fatalf("expected at least one topic for bob")
		}
	})

	// bob try to change the description(event trigger) of topic[0] -> should fail
	t.Run("bob update topic forbidden", func(t *testing.T) {
		if len(bookmarkedTopicIDs) == 0 {
			t.Skip("no bookmarked topics from previous step")
		}
		topicID := bookmarkedTopicIDs[0]
		updateReq := map[string]interface{}{
			"entity_id":   topicID,
			"description": "unauthorized update by bob",
		}
		body, _ := json.Marshal(updateReq)
		updateRequest, _ := http.NewRequest("POST", helpers.GetHost()+"/api/entity/update-description", bytes.NewReader(body))
		updateRequest.Header.Set("Content-Type", "application/json")
		updateRequest.AddCookie(&http.Cookie{Name: "access_token", Value: bobToken})
		updateResp, err := client.Do(updateRequest)
		if err != nil {
			t.Fatalf("failed to POST update-description: %v", err)
		}
		defer updateResp.Body.Close()
		if updateResp.StatusCode == http.StatusOK {
			t.Fatalf("expected forbidden, got status OK")
		}
	})

	// bob try to pause topic[0] -> should fail
	t.Run("bob pause topic forbidden", func(t *testing.T) {
		if len(bookmarkedTopicIDs) == 0 {
			t.Skip("no bookmarked topics from previous step")
		}
		topicID := bookmarkedTopicIDs[0]
		pauseReq, _ := http.NewRequest("GET", helpers.GetHost()+"/api/topic/nsq/pause?id="+topicID+"&entity_id="+topicID, nil)
		pauseReq.AddCookie(&http.Cookie{Name: "access_token", Value: bobToken})
		pauseResp, err := client.Do(pauseReq)
		if err != nil {
			t.Fatalf("failed to GET pause: %v", err)
		}
		defer pauseResp.Body.Close()
		if pauseResp.StatusCode == http.StatusOK {
			t.Fatalf("expected forbidden, got status OK")
		}
	})

	// bob try to delete channel[0] -> should fail
	t.Run("bob delete channel forbidden", func(t *testing.T) {
		if len(bookmarkedTopicIDs) == 0 {
			t.Skip("no bookmarked topics from previous step")
		}
		if cachedNsqdHosts == "" || cachedTopicName == "" {
			t.Skip("no nsqd hosts or topic name cached from topic detail")
		}
		// List channels for the topic with hosts param (using topic name)
		channelsReq, _ := http.NewRequest(
			"GET",
			helpers.GetHost()+"/api/topic/nsq/list-channels?topic="+cachedTopicName+"&hosts="+cachedNsqdHosts,
			nil,
		)
		channelsReq.AddCookie(&http.Cookie{Name: "access_token", Value: bobToken})
		channelsResp, err := client.Do(channelsReq)
		if err != nil {
			t.Fatalf("failed to GET /api/topic/nsq/list-channels: %v", err)
		}
		defer channelsResp.Body.Close()
		if channelsResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(channelsResp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", channelsResp.StatusCode, string(body))
		}
		var channelsResult struct {
			Data struct {
				Channels []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"channels"`
			} `json:"data"`
		}
		err = json.NewDecoder(channelsResp.Body).Decode(&channelsResult)
		if err != nil {
			t.Fatalf("failed to decode channels response: %v", err)
		}
		if len(channelsResult.Data.Channels) == 0 {
			t.Skip("no channels found for topic")
		}
		channel := channelsResult.Data.Channels[0]
		deleteReq, _ := http.NewRequest("GET", helpers.GetHost()+"/api/channel/nsq/delete?id="+channel.ID+"&channel="+channel.Name+"&entity_id="+channel.ID, nil)
		deleteReq.AddCookie(&http.Cookie{Name: "access_token", Value: bobToken})
		deleteResp, err := client.Do(deleteReq)
		if err != nil {
			t.Fatalf("failed to GET delete channel: %v", err)
		}
		defer deleteResp.Body.Close()
		if deleteResp.StatusCode == http.StatusOK {
			t.Fatalf("expected forbidden, got status OK")
		}
	})

	// bob try to claim the topic[0] -> should be okay
	var bobClaimTicketID string
	t.Run("bob claim topic", func(t *testing.T) {
		if len(bookmarkedTopicIDs) == 0 {
			t.Skip("no bookmarked topics from previous step")
		}
		topicID := bookmarkedTopicIDs[0]
		claimReq := map[string]interface{}{
			"entity_id":  topicID,
			"group_id":   groupOrder.ID,
			"group_name": groupOrder.Name,
		}
		body, _ := json.Marshal(claimReq)
		claimRequest, _ := http.NewRequest("POST", helpers.GetHost()+"/api/entity/claim", bytes.NewReader(body))
		claimRequest.Header.Set("Content-Type", "application/json")
		claimRequest.AddCookie(&http.Cookie{Name: "access_token", Value: bobToken})
		claimResp, err := client.Do(claimRequest)
		if err != nil {
			t.Fatalf("failed to POST claim: %v", err)
		}
		defer claimResp.Body.Close()
		if claimResp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(claimResp.Body)
			t.Fatalf("unexpected status code for claim: %d, body: %s", claimResp.StatusCode, string(respBody))
		}
		var claimResult struct {
			Data struct {
				ApplicationID string `json:"application_id"`
			} `json:"data"`
		}
		err = json.NewDecoder(claimResp.Body).Decode(&claimResult)
		if err == nil && claimResult.Data.ApplicationID != "" {
			bobClaimTicketID = claimResult.Data.ApplicationID
		}
	})

	// charlie see detail of bob's ticket
	t.Run("charlie ticket detail bob", func(t *testing.T) {
		if bobClaimTicketID == "" {
			t.Skip("no bob claim ticket to check detail")
		}
		req, _ := http.NewRequest("GET", helpers.GetHost()+"/api/tickets/detail?id="+bobClaimTicketID, nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: charlieToken})
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to GET /api/tickets/detail: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}
		var detail struct {
			Data struct {
				Ticket struct {
					ID string `json:"id"`
				} `json:"ticket"`
				Assignees []struct {
					UserID   string `json:"user_id"`
					Username string `json:"username"`
				} `json:"assignees"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&detail)
		if err != nil {
			t.Fatalf("failed to decode ticket detail: %v", err)
		}
		if detail.Data.Ticket.ID != bobClaimTicketID {
			t.Fatalf("expected ticket ID %s, got %s", bobClaimTicketID, detail.Data.Ticket.ID)
		}
		// Assignee should be charlie (by username or id, depending on API)
		if detail.Data.Assignees[0].Username != "charlie" {
			t.Fatalf("expected assignee to be charlie, got %s", detail.Data.Assignees[0].Username)
		}
	})

	// charlie should see the ticket for topic claim from bob and reject it
	t.Run("charlie reject bob claim", func(t *testing.T) {
		// Reject bob's claim ticket
		if bobClaimTicketID == "" {
			t.Skip("no bob claim ticket to reject")
		}
		rejectReq := map[string]interface{}{
			"action":         "reject",
			"application_id": bobClaimTicketID,
		}
		body, _ := json.Marshal(rejectReq)
		rejectRequest, _ := http.NewRequest("POST", helpers.GetHost()+"/api/tickets/action", bytes.NewReader(body))
		rejectRequest.Header.Set("Content-Type", "application/json")
		rejectRequest.AddCookie(&http.Cookie{Name: "access_token", Value: charlieToken})
		rejectResp, err := client.Do(rejectRequest)
		if err != nil {
			t.Fatalf("failed to POST reject ticket: %v", err)
		}
		defer rejectResp.Body.Close()
		if rejectResp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(rejectResp.Body)
			t.Fatalf("unexpected status code for reject: %d, body: %s", rejectResp.StatusCode, string(respBody))
		}
	})

}
