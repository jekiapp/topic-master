package nsq

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

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

// GetStats fetches stats for a given topic and channel from multiple nsqd hosts in parallel
func GetStats(nsqdHosts []string, topic, channel string) ([]modelnsq.Stats, error) {
	type result struct {
		stat modelnsq.Stats
		err  error
	}
	results := make(chan result)
	var wg sync.WaitGroup
	for _, host := range nsqdHosts {
		wg.Add(1)
		go func(nsqdHost string) {
			defer wg.Done()
			urlStr := fmt.Sprintf("http://%s/stats?format=json", nsqdHost)
			if topic != "" {
				urlStr += "&topic=" + url.QueryEscape(topic)
				if channel != "" {
					urlStr += "&channel=" + url.QueryEscape(channel)
				}
			}

			resp, err := http.Get(urlStr)
			if err != nil {
				results <- result{err: err}
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				results <- result{err: fmt.Errorf("nsqd %s returned status %d", nsqdHost, resp.StatusCode)}
				return
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				results <- result{err: err}
				return
			}
			var parsed struct {
				Topics []modelnsq.Stats `json:"topics"`
			}
			if err := json.Unmarshal(body, &parsed); err != nil {
				results <- result{err: err}
				return
			}
			for _, t := range parsed.Topics {
				results <- result{stat: t}
			}
		}(host)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	var stats []modelnsq.Stats
	var firstErr error
	for r := range results {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		if r.err == nil {
			stats = append(stats, r.stat)
		}
	}
	if len(stats) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return stats, nil
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

// TopicStatsResult represents the complete stats for a topic including channels
type TopicStatsResult struct {
	TopicDepth    int                              `json:"topic_depth"`
	TopicMessages int                              `json:"topic_messages"`
	ChannelStats  map[string]modelnsq.ChannelStats `json:"channel_stats"`
}

// GetTopicStatsWithChannels fetches stats for a given topic and all its channels from a given nsqd host
func GetTopicStatsWithChannels(nsqdHost, topic string) (TopicStatsResult, error) {
	url := fmt.Sprintf("http://%s/stats?format=json&topic=%s", nsqdHost, topic)
	resp, err := http.Get(url)
	if err != nil {
		return TopicStatsResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return TopicStatsResult{}, fmt.Errorf("nsqd returned status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TopicStatsResult{}, err
	}

	var parsed struct {
		Topics []struct {
			TopicName    string `json:"topic_name"`
			Depth        int    `json:"depth"`
			MessageCount int    `json:"message_count"`
			Channels     []struct {
				ChannelName   string `json:"channel_name"`
				Depth         int    `json:"depth"`
				MessageCount  int    `json:"message_count"`
				InFlightCount int    `json:"in_flight_count"`
				RequeueCount  int    `json:"requeue_count"`
				DeferredCount int    `json:"deferred_count"`
				ClientCount   int    `json:"client_count"`
			} `json:"channels"`
		} `json:"topics"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return TopicStatsResult{}, err
	}

	for _, t := range parsed.Topics {
		if t.TopicName == topic {
			result := TopicStatsResult{
				TopicDepth:    t.Depth,
				TopicMessages: t.MessageCount,
				ChannelStats:  make(map[string]modelnsq.ChannelStats),
			}

			for _, c := range t.Channels {
				result.ChannelStats[c.ChannelName] = modelnsq.ChannelStats{
					Depth:         c.Depth,
					Messages:      c.MessageCount,
					InFlight:      c.InFlightCount,
					Requeued:      c.RequeueCount,
					Deferred:      c.DeferredCount,
					ConsumerCount: c.ClientCount,
				}
			}

			return result, nil
		}
	}
	return TopicStatsResult{}, fmt.Errorf("topic %s not found in nsqd stats", topic)
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
	url := fmt.Sprintf("http://%s/stats?format=json&topic=%s", host, topic)
	// get with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	// give timeout to all
	resp, err := http.DefaultClient.Do(req)
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

// ResumeTopicOnNsqd resumes a topic on the given nsqd host
func ResumeTopicOnNsqd(host, topic string) error {
	url := fmt.Sprintf("http://%s/topic/unpause?topic=%s", host, topic)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to resume topic on nsqd %s: %w", host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("nsqd returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// PauseChannelOnNsqd pauses a channel on the given nsqd host
func PauseChannelOnNsqd(host, topic, channel string) error {
	url := fmt.Sprintf("http://%s/channel/pause?topic=%s&channel=%s", host, topic, channel)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to pause channel on nsqd %s: %w", host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("nsqd returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// EmptyChannelOnNsqd empties a channel on the given nsqd host
func EmptyChannelOnNsqd(host, topic, channel string) error {
	url := fmt.Sprintf("http://%s/channel/empty?topic=%s&channel=%s", host, topic, channel)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to empty channel on nsqd %s: %w", host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("nsqd returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ResumeChannelOnNsqd resumes a channel on the given nsqd host
func ResumeChannelOnNsqd(host, topic, channel string) error {
	url := fmt.Sprintf("http://%s/channel/unpause?topic=%s&channel=%s", host, topic, channel)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to resume channel on nsqd %s: %w", host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("nsqd returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
