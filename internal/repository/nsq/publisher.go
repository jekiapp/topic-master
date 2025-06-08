package nsq

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Publish publishes a message to the given topic on all provided nsqd hosts using go-nsq.
func Publish(topic string, message string, host string) error {
	url := fmt.Sprintf("http://%s/pub?topic=%s", host, topic)
	resp, err := http.Post(url, "application/octet-stream", strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("failed to publish to %s: %w", host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nsqd returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
