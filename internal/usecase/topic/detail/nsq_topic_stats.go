// this usecase will return the stats for a given topic
// the stats will be returned in the following format:
// {
//   "depth": 100,
//   "messages": 5000
// }
// it will expect parameters of array of nsqd hosts and topic name

package detail

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jekiapp/topic-master/internal/config"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
)

// iNsqTopicStatsRepo abstracts fetching topic stats from nsqd hosts
// params: hosts []string, topic string
// returns: depth, messages, error
//
//go:generate mockgen -source=nsq_topic_stats.go -destination=mock_nsq_topic_stats_repo.go -package=detail iNsqTopicStatsRepo
type iNsqTopicStatsRepo interface {
	GetTopicStats(hosts []string, topic string) (depth int, messages int, err error)
}

type nsqTopicStatsRepo struct{}

func (r *nsqTopicStatsRepo) GetTopicStats(hosts []string, topic string) (int, int, error) {
	var totalDepth, totalMessages int
	var errs []error
	for _, host := range hosts {
		depth, messages, err := nsqrepo.GetTopicStats(host, topic)
		if err != nil {
			log.Printf("error getting topic stats for host %s: %v", host, err)
			errs = append(errs, err)
			continue
		}
		totalDepth += depth
		totalMessages += messages
	}
	if len(errs) > 0 {
		return totalDepth, totalMessages, fmt.Errorf("error getting topic stats for hosts: %v", errs)
	}
	return totalDepth, totalMessages, nil
}

type NsqTopicStatsResponse struct {
	Depth    int `json:"depth"`
	Messages int `json:"messages"`
}

type NsqTopicStatsUsecase struct {
	cfg  *config.Config
	repo iNsqTopicStatsRepo
}

func NewNsqTopicStatsUsecase(cfg *config.Config) NsqTopicStatsUsecase {
	return NsqTopicStatsUsecase{
		cfg:  cfg,
		repo: &nsqTopicStatsRepo{},
	}
}

// HandleQuery expects params: "hosts" (comma-separated string), "topic" (string)
func (uc NsqTopicStatsUsecase) HandleQuery(ctx context.Context, params map[string]string) (NsqTopicStatsResponse, error) {
	hostsStr, ok := params["hosts"]
	if !ok {
		return NsqTopicStatsResponse{}, fmt.Errorf("hosts is required")
	}
	hosts := []string{}
	for _, h := range splitAndTrim(hostsStr, ",") {
		if h != "" {
			hosts = append(hosts, h)
		}
	}
	topic, ok := params["topic"]
	if !ok {
		return NsqTopicStatsResponse{}, nil // or return error
	}
	if !ok {
		return NsqTopicStatsResponse{}, nil // or return error
	}
	depth, messages, err := uc.repo.GetTopicStats(hosts, topic)
	if err != nil {
		return NsqTopicStatsResponse{}, err
	}
	return NsqTopicStatsResponse{
		Depth:    depth,
		Messages: messages,
	}, nil
}

// splitAndTrim splits a string by sep and trims spaces from each element
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, p := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(p)
		parts = append(parts, trimmed)
	}
	return parts
}
