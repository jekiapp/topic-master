package nonlogin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	nhooyrws "nhooyr.io/websocket"
)

var topicMasterHost = "http://localhost:4181"

type Topic struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EventTrigger string `json:"event_trigger"`
	GroupOwner   string `json:"group_owner"`
	Bookmarked   bool   `json:"bookmarked"`
	NsqdHosts    []struct {
		Address string `json:"address"`
	} `json:"nsqd_hosts"`
}

func getAllTopics(t *testing.T) []Topic {
	resp, err := http.Get(topicMasterHost + "/api/topic/list-all-topics")
	if err != nil {
		t.Fatalf("failed to GET /api/topic/list-all-topics: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(
		t,
		http.StatusOK,
		resp.StatusCode,
		"unexpected status code",
	)

	var result struct {
		Data struct {
			Topics []Topic `json:"topics"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	assert.NotNil(
		t,
		result.Data.Topics,
		"topics should not be nil",
	)

	for _, topic := range result.Data.Topics {
		assert.NotEmpty(t, topic.ID, "topic ID should not be empty")
		assert.NotEmpty(t, topic.Name, "topic name should not be empty")
		assert.False(t, topic.Bookmarked, "topic bookmarked should be false")
	}

	return result.Data.Topics
}

func checkTopicDetail(t *testing.T, topic Topic) Topic {
	url := fmt.Sprintf("%s/api/topic/detail?topic=%s", topicMasterHost, topic.ID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET /api/topic/detail: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(
		t,
		http.StatusOK,
		resp.StatusCode,
		"unexpected status code",
	)

	var detailResp struct {
		Data Topic `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&detailResp)
	if err != nil {
		t.Fatalf("failed to decode detail response: %v", err)
	}

	assert.Equal(t, topic.ID, detailResp.Data.ID, "detail ID should match topic ID")
	assert.Equal(t, topic.Name, detailResp.Data.Name, "detail name should match topic name")
	// Optionally check other fields if needed
	return detailResp.Data
}

func editEventTrigger(t *testing.T, topic Topic) {
	newEventTrigger := "integration test event trigger"
	updateReq := map[string]interface{}{
		"entity_id":   topic.ID,
		"description": newEventTrigger,
	}
	updateBody, _ := json.Marshal(updateReq)
	updateResp, err := http.Post(
		fmt.Sprintf("%s/api/entity/update-description", topicMasterHost),
		"application/json",
		bytes.NewReader(updateBody),
	)
	if err != nil {
		t.Fatalf("failed to POST update-description: %v", err)
	}
	defer updateResp.Body.Close()
	assert.Equal(t, http.StatusOK, updateResp.StatusCode, "unexpected status code for update-description")

	verifyResp, err := http.Get(fmt.Sprintf("%s/api/topic/detail?topic=%s", topicMasterHost, topic.ID))
	if err != nil {
		t.Fatalf("failed to GET detail after update: %v", err)
	}
	defer verifyResp.Body.Close()
	var verifyDetail struct {
		Data struct {
			EventTrigger string `json:"event_trigger"`
		} `json:"data"`
	}
	_ = json.NewDecoder(verifyResp.Body).Decode(&verifyDetail)
	assert.Equal(t, newEventTrigger, verifyDetail.Data.EventTrigger, "event trigger should be updated")
}

func checkTopicStats(t *testing.T, topic Topic) {

	if len(topic.NsqdHosts) == 0 {
		t.Skip("no nsqd hosts for topic stats")
	}
	hosts := ""
	for i, h := range topic.NsqdHosts {
		hosts += h.Address
		if i < len(topic.NsqdHosts)-1 {
			hosts += ","
		}
	}
	statsURL := fmt.Sprintf("%s/api/topic/stats?topic=%s&hosts=%s", topicMasterHost, topic.Name, hosts)
	statsResp, err := http.Get(statsURL)
	if err != nil {
		t.Fatalf("failed to GET topic stats: %v", err)
	}
	defer statsResp.Body.Close()
	assert.Equal(t, http.StatusOK, statsResp.StatusCode, "unexpected status code for topic stats")
	var stats struct {
		Data struct {
			Depth        int         `json:"depth"`
			Messages     int         `json:"messages"`
			ChannelStats interface{} `json:"channel_stats"`
		} `json:"data"`
	}
	err = json.NewDecoder(statsResp.Body).Decode(&stats)
	if err != nil {
		t.Fatalf("failed to decode topic stats: %v", err)
	}
	assert.GreaterOrEqual(t, stats.Data.Depth, 0, "depth should be >= 0")
	assert.GreaterOrEqual(t, stats.Data.Messages, 0, "messages should be >= 0")
}

func publishTopic(t *testing.T, topic Topic) {

	if len(topic.NsqdHosts) == 0 {
		t.Skip("no nsqd hosts for publish topic")
	}
	publishReq := map[string]interface{}{
		"topic":      topic.Name,
		"message":    "integration test message",
		"nsqd_hosts": []string{topic.NsqdHosts[0].Address},
	}
	publishBody, _ := json.Marshal(publishReq)
	publishURL := fmt.Sprintf("%s/api/topic/publish?entity_id=%s", topicMasterHost, topic.ID)
	publishResp, err := http.Post(
		publishURL,
		"application/json",
		bytes.NewReader(publishBody),
	)
	if err != nil {
		t.Fatalf("failed to POST publish: %v", err)
	}
	defer publishResp.Body.Close()
	assert.Equal(t, http.StatusOK, publishResp.StatusCode, "unexpected status code for publish")
	var pubResp struct {
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	err = json.NewDecoder(publishResp.Body).Decode(&pubResp)
	if err != nil {
		t.Fatalf("failed to decode publish response: %v", err)
	}
	assert.Contains(t, pubResp.Data.Message, "published", "publish response should indicate success")
}

func claimShouldFail(t *testing.T, topic Topic) {
	claimReq := map[string]interface{}{
		"entity_id": topic.ID,
	}
	claimBody, _ := json.Marshal(claimReq)
	claimResp, err := http.Post(
		fmt.Sprintf("%s/api/entity/claim", topicMasterHost),
		"application/json",
		bytes.NewReader(claimBody),
	)
	if err != nil {
		t.Fatalf("failed to POST claim: %v", err)
	}
	defer claimResp.Body.Close()
	body, _ := io.ReadAll(claimResp.Body)
	t.Logf("claimShouldFail response body: %s", string(body))
	assert.NotEqual(t, http.StatusOK, claimResp.StatusCode, "claim should fail")
}

func bookmarkShouldFail(t *testing.T, topic Topic) {
	bookmarkReq := map[string]interface{}{
		"entity_id": topic.ID,
	}
	bookmarkBody, _ := json.Marshal(bookmarkReq)
	bookmarkResp, err := http.Post(
		fmt.Sprintf("%s/api/entity/toggle-bookmark", topicMasterHost),
		"application/json",
		bytes.NewReader(bookmarkBody),
	)
	if err != nil {
		t.Fatalf("failed to POST bookmark: %v", err)
	}
	defer bookmarkResp.Body.Close()
	body, _ := io.ReadAll(bookmarkResp.Body)
	t.Logf("bookmarkShouldFail response body: %s", string(body))
	assert.NotEqual(t, http.StatusOK, bookmarkResp.StatusCode, "bookmark should fail")
}

func pauseShouldSucceed(t *testing.T, topic Topic) {
	url := fmt.Sprintf("%s/api/topic/nsq/pause?id=%s&entity_id=%s", topicMasterHost, topic.ID, topic.ID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET pause: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("pauseShouldSucceed response body: %s", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "pause should succeed")
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if assert.NoError(t, err, "pause response should be valid JSON") {
		assert.Equal(t, "success", result["status"], "pause response status should be success")
	}
}

func resumeShouldSucceed(t *testing.T, topic Topic) {
	url := fmt.Sprintf("%s/api/topic/nsq/resume?id=%s&entity_id=%s", topicMasterHost, topic.ID, topic.ID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET resume: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("resumeShouldSucceed response body: %s", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "resume should succeed")
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if assert.NoError(t, err, "resume response should be valid JSON") {
		assert.Equal(t, "success", result["status"], "resume response status should be success")
	}
}

func emptyShouldSucceed(t *testing.T, topic Topic) {
	url := fmt.Sprintf("%s/api/topic/nsq/empty?id=%s&entity_id=%s", topicMasterHost, topic.ID, topic.ID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET empty: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("emptyShouldSucceed response body: %s", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "empty should succeed")
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if assert.NoError(t, err, "empty response should be valid JSON") {
		assert.Equal(t, "success", result["status"], "empty response status should be success")
	}
}

func tailTopic(t *testing.T, topic Topic, messageCh chan<- string, errCh chan<- error) {
	if len(topic.NsqdHosts) == 0 {
		errCh <- nil // skip if no hosts
		return
	}
	hosts := ""
	for i, h := range topic.NsqdHosts {
		hosts += h.Address
		if i < len(topic.NsqdHosts)-1 {
			hosts += ","
		}
	}

	// Build ws URL with all required params
	wsBase := "ws://" + topicMasterHost[len("http://"):]
	wsTailURL := fmt.Sprintf("%s/api/topic/tail?topic=%s&limit_msg=1&nsqd_hosts=%s&entity_id=%s",
		wsBase,
		topic.Name,
		url.QueryEscape(hosts),
		topic.ID,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := nhooyrws.Dial(ctx, wsTailURL, nil)
	if err != nil {
		errCh <- err
		return
	}
	defer conn.Close(nhooyrws.StatusNormalClosure, "done")

	_, data, err := conn.Read(ctx)
	if err != nil {
		errCh <- err
		return
	}

	t.Logf("raw ws message: %q", data)

	// Convert to string, strip trailing \x1e, and send
	msg := string(bytes.Trim(data, "\x1e"))
	messageCh <- msg
}

func TestTopicIntegrationFlow(t *testing.T) {
	if envHost := os.Getenv("TOPIC_MASTER_HOST"); envHost != "" {
		topicMasterHost = envHost
	}
	var topics []Topic
	t.Run("getAllTopics", func(t *testing.T) {
		alltopics := getAllTopics(t)
		if len(alltopics) == 0 {
			t.Fatalf("no topics to test detail API")
		}
		topics = alltopics
	})

	var topicDetail Topic
	t.Run("checkTopicDetail", func(t *testing.T) {
		topicDetail = checkTopicDetail(t, topics[0])
	})

	t.Run("editEventTrigger", func(t *testing.T) {
		editEventTrigger(t, topicDetail)
	})

	t.Run("checkTopicStats", func(t *testing.T) {
		checkTopicStats(t, topicDetail)
	})

	t.Run("tail and publish", func(t *testing.T) {
		messageCh := make(chan string, 1)
		errCh := make(chan error, 1)
		go tailTopic(t, topicDetail, messageCh, errCh)
		// Wait a moment to ensure tail is listening before publish
		time.Sleep(300 * time.Millisecond)
		// Now publish
		publishTopic(t, topicDetail)
		select {
		case msg := <-messageCh:
			assert.Contains(t, msg, "integration test message", "tail should receive published message")
		case err := <-errCh:
			if err != nil {
				t.Fatalf("tailTopic error: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for tail message")
		}
	})

	t.Run("claimShouldFail", func(t *testing.T) {
		claimShouldFail(t, topicDetail)
	})

	t.Run("bookmarkShouldFail", func(t *testing.T) {
		bookmarkShouldFail(t, topicDetail)
	})

	t.Run("pauseShouldSucceed", func(t *testing.T) {
		pauseShouldSucceed(t, topicDetail)
	})

	t.Run("resumeShouldSucceed", func(t *testing.T) {
		resumeShouldSucceed(t, topicDetail)
	})

	t.Run("emptyShouldSucceed", func(t *testing.T) {
		emptyShouldSucceed(t, topicDetail)
	})

}
