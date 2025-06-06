package nsq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	modelnsq "github.com/jekiapp/topic-master/internal/model/nsq"
)

// GetNsqdsForTopic fetches all nsqd nodes for a topic from the given lookupd URL
func GetNsqdsForTopic(lookupdURL, topic string) ([]modelnsq.Nsqd, error) {
	url := fmt.Sprintf("%s/lookup?topic=%s", lookupdURL, topic)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lookupd returned status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Producers []modelnsq.Nsqd `json:"producers"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return parsed.Producers, nil
}

// GetTopicStats fetches stats for a given topic from a given nsqd host
func GetTopicStats(nsqdHost, topic string) (depth int, messages int, err error) {
	url := fmt.Sprintf("http://%s/stats?format=json", nsqdHost)
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("nsqd returned status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}
	var parsed struct {
		Topics []struct {
			TopicName    string `json:"topic_name"`
			Depth        int    `json:"depth"`
			MessageCount int    `json:"message_count"`
		} `json:"topics"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, 0, err
	}
	for _, t := range parsed.Topics {
		if t.TopicName == topic {
			return t.Depth, t.MessageCount, nil
		}
	}
	return 0, 0, fmt.Errorf("topic %s not found in nsqd stats", topic)
}
