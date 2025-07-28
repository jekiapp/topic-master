package nonlogin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	helpers "github.com/jekiapp/topic-master/infra/test_script/helpers"
	"github.com/stretchr/testify/assert"
)

var channelTestHost = helpers.GetHost()

type Channel struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	GroupOwner   string `json:"group_owner"`
	Description  string `json:"description"`
	Topic        string `json:"topic"`
	IsBookmarked bool   `json:"is_bookmarked"`
	IsPaused     bool   `json:"is_paused"`
	IsFreeAction bool   `json:"is_free_action"`
}

func listChannels(t *testing.T, topic Topic) []Channel {
	if len(topic.NsqdHosts) == 0 {
		t.Skip("no nsqd hosts for channel list")
	}
	hosts := ""
	for i, h := range topic.NsqdHosts {
		hosts += h.Address
		if i < len(topic.NsqdHosts)-1 {
			hosts += ","
		}
	}
	url := fmt.Sprintf("%s/api/topic/nsq/list-channels?topic=%s&hosts=%s", channelTestHost, topic.Name, hosts)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET list-channels: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code for list-channels")
	var result struct {
		Data struct {
			Channels []Channel `json:"channels"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode list-channels: %v", err)
	}
	return result.Data.Channels
}

func pauseChannelShouldSucceed(t *testing.T, channel Channel) {
	url := fmt.Sprintf("%s/api/channel/nsq/pause?id=%s&channel=%s&entity_id=%s", channelTestHost, channel.ID, channel.Name, channel.ID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET pause channel: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("pauseChannelShouldSucceed response body: %s", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "pause channel should succeed")
	var result map[string]interface{}
	_ = json.Unmarshal(body, &result)
	assert.Equal(t, "success", result["status"], "pause channel response status should be success")
}

func resumeChannelShouldSucceed(t *testing.T, channel Channel) {
	url := fmt.Sprintf("%s/api/channel/nsq/resume?id=%s&channel=%s&entity_id=%s", channelTestHost, channel.ID, channel.Name, channel.ID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET resume channel: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("resumeChannelShouldSucceed response body: %s", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "resume channel should succeed")
	var result map[string]interface{}
	_ = json.Unmarshal(body, &result)
	assert.Equal(t, "success", result["status"], "resume channel response status should be success")
}

func emptyChannelShouldSucceed(t *testing.T, channel Channel) {
	url := fmt.Sprintf("%s/api/channel/nsq/empty?id=%s&channel=%s&entity_id=%s", channelTestHost, channel.ID, channel.Name, channel.ID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET empty channel: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("emptyChannelShouldSucceed response body: %s", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "empty channel should succeed")
	var result map[string]interface{}
	_ = json.Unmarshal(body, &result)
	assert.Equal(t, "success", result["status"], "empty channel response status should be success")
}

func deleteChannelShouldSucceed(t *testing.T, channel Channel) {
	url := fmt.Sprintf("%s/api/channel/nsq/delete?id=%s&channel=%s&entity_id=%s", channelTestHost, channel.ID, channel.Name, channel.ID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET delete channel: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("deleteChannelShouldSucceed response body: %s", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "delete channel should succeed")
	var result map[string]interface{}
	_ = json.Unmarshal(body, &result)
	assert.Equal(t, "success", result["status"], "delete channel response status should be success")
}

func claimChannelShouldFail(t *testing.T, channel Channel) {
	claimReq := map[string]interface{}{
		"entity_id": channel.ID,
	}
	claimBody, _ := json.Marshal(claimReq)
	claimResp, err := http.Post(
		fmt.Sprintf("%s/api/entity/claim", channelTestHost),
		"application/json",
		bytes.NewReader(claimBody),
	)
	if err != nil {
		t.Fatalf("failed to POST claim channel: %v", err)
	}
	defer claimResp.Body.Close()
	body, _ := io.ReadAll(claimResp.Body)
	t.Logf("claimChannelShouldFail response body: %s", string(body))
	assert.NotEqual(t, http.StatusOK, claimResp.StatusCode, "claim channel should fail")
}

func bookmarkChannelShouldFail(t *testing.T, channel Channel) {
	bookmarkReq := map[string]interface{}{
		"entity_id": channel.ID,
	}
	bookmarkBody, _ := json.Marshal(bookmarkReq)
	bookmarkResp, err := http.Post(
		fmt.Sprintf("%s/api/entity/toggle-bookmark", channelTestHost),
		"application/json",
		bytes.NewReader(bookmarkBody),
	)
	if err != nil {
		t.Fatalf("failed to POST bookmark channel: %v", err)
	}
	defer bookmarkResp.Body.Close()
	body, _ := io.ReadAll(bookmarkResp.Body)
	t.Logf("bookmarkChannelShouldFail response body: %s", string(body))
	assert.NotEqual(t, http.StatusOK, bookmarkResp.StatusCode, "bookmark channel should fail")
}

func TestChannelIntegrationFlow(t *testing.T) {
	if envHost := os.Getenv("TOPIC_MASTER_HOST"); envHost != "" {
		channelTestHost = envHost
	}
	var topics []Topic
	t.Run("getAllTopicsForChannel", func(t *testing.T) {
		alltopics := getAllTopics(t)
		if len(alltopics) == 0 {
			t.Fatalf("no topics to test channel API")
		}
		topics = alltopics
	})

	var topicDetail Topic
	t.Run("checkTopicDetail", func(t *testing.T) {
		topicDetail = checkTopicDetail(t, topics[3])
	})

	var channels []Channel
	var channel Channel
	t.Run("listChannels", func(t *testing.T) {
		channels = listChannels(t, topicDetail)
		if len(channels) == 0 {
			t.Fatalf("no channels to test")
		}
		channel = channels[0]
	})

	t.Run("pauseChannelShouldSucceed", func(t *testing.T) {
		pauseChannelShouldSucceed(t, channel)
	})
	t.Run("resumeChannelShouldSucceed", func(t *testing.T) {
		resumeChannelShouldSucceed(t, channel)
	})
	t.Run("emptyChannelShouldSucceed", func(t *testing.T) {
		emptyChannelShouldSucceed(t, channel)
	})
	t.Run("claimChannelShouldFail", func(t *testing.T) {
		claimChannelShouldFail(t, channel)
	})
	t.Run("bookmarkChannelShouldFail", func(t *testing.T) {
		bookmarkChannelShouldFail(t, channel)
	})
	t.Run("deleteChannelShouldSucceed", func(t *testing.T) {
		deleteChannelShouldSucceed(t, channel)
	})
}
