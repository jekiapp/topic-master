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

// DeleteTopicFromNsqd deletes a topic from the given nsqd host
func DeleteTopicFromNsqd(host, topic string) error {
	url := fmt.Sprintf("http://%s/topic/delete?topic=%s", host, topic)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to delete topic from nsqd %s: %w", host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("nsqd returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PauseTopicOnNsqd pauses a topic on the given nsqd host
func PauseTopicOnNsqd(host, topic string) error {
	url := fmt.Sprintf("http://%s/topic/pause?topic=%s", host, topic)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to pause topic on nsqd %s: %w", host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("nsqd returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// EmptyTopicOnNsqd empties a topic on the given nsqd host
func EmptyTopicOnNsqd(host, topic string) error {
	url := fmt.Sprintf("http://%s/topic/empty?topic=%s", host, topic)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to empty topic on nsqd %s: %w", host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("nsqd returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// IsTopicPausedOnNsqd checks if a topic is paused on the given nsqd host
func IsTopicPausedOnNsqd(host, topic string) (bool, error) {
	url := fmt.Sprintf("http://%s/stats?format=json", host)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("nsqd returned status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var parsed struct {
		Topics []struct {
			TopicName string `json:"topic_name"`
			Paused    bool   `json:"paused"`
		} `json:"topics"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false, err
	}
	for _, t := range parsed.Topics {
		if t.TopicName == topic {
			return t.Paused, nil
		}
	}
	return false, fmt.Errorf("topic %s not found in nsqd stats", topic)
}
