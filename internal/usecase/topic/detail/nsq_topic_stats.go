package detail

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jekiapp/topic-master/internal/config"
	"github.com/jekiapp/topic-master/internal/model/nsq"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
)

// iNsqTopicStatsRepo abstracts fetching topic stats from nsqd hosts
// params: hosts []string, topic string
// returns: TopicStatsResult, error
//
//go:generate mockgen -source=nsq_topic_stats.go -destination=mock_nsq_topic_stats_repo.go -package=detail iNsqTopicStatsRepo
type iNsqTopicStatsRepo interface {
	GetTopicStatsWithChannels(hosts []string, topic string) (nsqrepo.TopicStatsResult, error)
}

type nsqTopicStatsRepo struct{}

func (r *nsqTopicStatsRepo) GetTopicStatsWithChannels(hosts []string, topic string) (nsqrepo.TopicStatsResult, error) {
	var aggregatedResult nsqrepo.TopicStatsResult
	aggregatedResult.ChannelStats = make(map[string]nsq.ChannelStats)
	var errs []error

	for _, host := range hosts {
		result, err := nsqrepo.GetTopicStatsWithChannels(host, topic)
		if err != nil {
			log.Printf("error getting topic stats for host %s: %v", host, err)
			errs = append(errs, err)
			continue
		}

		// Aggregate topic stats
		aggregatedResult.TopicDepth += result.TopicDepth
		aggregatedResult.TopicMessages += result.TopicMessages

		// Aggregate channel stats
		for channelName, channelStats := range result.ChannelStats {
			existing := aggregatedResult.ChannelStats[channelName]
			existing.Depth += channelStats.Depth
			existing.Messages += channelStats.Messages
			existing.InFlight += channelStats.InFlight
			existing.Requeued += channelStats.Requeued
			existing.Deferred += channelStats.Deferred
			existing.ConsumerCount += channelStats.ConsumerCount
			aggregatedResult.ChannelStats[channelName] = existing
		}
	}

	if len(errs) > 0 {
		return aggregatedResult, fmt.Errorf("error getting topic stats for hosts: %v", errs)
	}
	return aggregatedResult, nil
}

type NsqTopicStatsResponse struct {
	Depth        int                         `json:"depth"`
	Messages     int                         `json:"messages"`
	ChannelStats map[string]nsq.ChannelStats `json:"channel_stats"`
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
		return NsqTopicStatsResponse{}, fmt.Errorf("topic is required")
	}

	result, err := uc.repo.GetTopicStatsWithChannels(hosts, topic)
	if err != nil {
		return NsqTopicStatsResponse{}, err
	}

	return NsqTopicStatsResponse{
		Depth:        result.TopicDepth,
		Messages:     result.TopicMessages,
		ChannelStats: result.ChannelStats,
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
