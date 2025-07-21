package nonlogin

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAllTopicsIntegration(t *testing.T) {
	resp, err := http.Get("http://topic-master:4181/api/topic/list-all-topics")
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

	type Topic struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		EventTrigger string `json:"event_trigger"`
		GroupOwner   string `json:"group_owner"`
		Bookmarked   bool   `json:"bookmarked"`
	}

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
}
