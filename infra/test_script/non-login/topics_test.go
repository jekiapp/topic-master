package nonlogin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

	t.Run("publishTopic", func(t *testing.T) {
		publishTopic(t, topicDetail)
	})
}
