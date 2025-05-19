package lookupd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jekiapp/nsqper/internal/config"
)

var lookupdAddr string

// Init sets the lookupd address from config
func Init(cfg *config.Config) {
	lookupdAddr = cfg.NSQLookupdAddr
}

// GetAllTopics fetches all topics from lookupd
func GetAllTopics() ([]string, error) {
	if lookupdAddr == "" {
		return nil, fmt.Errorf("lookupd address not initialized")
	}
	url := fmt.Sprintf("%s/topics", lookupdAddr)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lookupd returned status %d", resp.StatusCode)
	}
	var result struct {
		Topics []string `json:"topics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Topics, nil
}
