package testautomation

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAllTopicsIntegration(t *testing.T) {
	topicFile := "../../test_setup/topics.txt"
	file, err := os.Open(topicFile)
	if err != nil {
		t.Fatalf("failed to open topics.txt: %v", err)
	}
	defer file.Close()

	expectedTopics := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		topic := strings.TrimSpace(scanner.Text())
		if topic != "" {
			expectedTopics[topic] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("failed to read topics.txt: %v", err)
	}

	resp, err := http.Get("http://localhost:4181/api/topic/list-all-topics")
	if err != nil {
		t.Fatalf("failed to GET /api/topic/list-all-topics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	type Topic struct {
		Name       string `json:"name"`
		GroupOwner string `json:"group_owner"`
	}

	var result struct {
		Data struct {
			Topics []Topic `json:"topics"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	returnedTopics := make(map[string]Topic)
	for _, topic := range result.Data.Topics {
		returnedTopics[topic.Name] = topic
	}

	missing := []string{}
	for topic := range expectedTopics {
		assert.Equal(t, returnedTopics[topic].GroupOwner, "None")
		if _, ok := returnedTopics[topic]; !ok {
			missing = append(missing, topic)
		}
	}
	if len(missing) > 0 {
		t.Errorf("missing topics in response: %v", missing)
	}
}
