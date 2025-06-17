package detail

import (
	"context"
	"errors"
	"fmt"
	"strings"

	topicLogic "github.com/jekiapp/topic-master/internal/logic/topic"
	"github.com/jekiapp/topic-master/internal/model/entity"
	modelnsq "github.com/jekiapp/topic-master/internal/model/nsq"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/tidwall/buntdb"
)

type ListTopicChannelsResponse struct {
	Channels []ChannelResponse `json:"channels"`
}

type ChannelResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Depth       int    `json:"depth"`
	Messages    int    `json:"messages"`
	InFlight    int    `json:"in_flight"`
	Requeued    int    `json:"requeued"`
	Deferred    int    `json:"deferred"`
}

func NewListTopicChannelsUsecase(db *buntdb.DB) ListTopicChannelsUsecase {
	return ListTopicChannelsUsecase{
		db:   db,
		repo: &listTopicChannelsRepo{db: db},
	}
}

type ListTopicChannelsUsecase struct {
	db   *buntdb.DB
	repo iListTopicChannelsRepo
}

// HandleQuery handles HTTP query for listing channels by topic.
// params should contain "topic" key and "hosts" key (comma-separated string).
func (uc ListTopicChannelsUsecase) HandleQuery(ctx context.Context, params map[string]string) (ListTopicChannelsResponse, error) {
	topic, ok := params["topic"]
	if !ok {
		return ListTopicChannelsResponse{}, errors.New("topic is required")
	}

	hostsStr, ok := params["hosts"]
	if !ok {
		return ListTopicChannelsResponse{}, errors.New("hosts is required")
	}

	hosts := []string{}
	for _, h := range strings.Split(hostsStr, ",") {
		if h != "" {
			hosts = append(hosts, strings.TrimSpace(h))
		}
	}

	channelsDB, err := uc.repo.GetAllNsqTopicChannels(topic)
	if err != nil && err != buntdb.ErrNotFound {
		return ListTopicChannelsResponse{}, err
	}

	stats, err := uc.repo.GetStats(hosts, topic, "")
	if err != nil {
		return ListTopicChannelsResponse{}, err
	}

	if len(stats) == 0 {
		return ListTopicChannelsResponse{}, errors.New("no stats found")
	}

	channelStats := make(map[string]modelnsq.ChannelStats)
	for _, stat := range stats {
		for _, channel := range stat.Channels {
			channelStats[channel.ChannelName] = modelnsq.ChannelStats{
				Depth:    channel.Depth,
				Messages: channel.MessageCount,
				InFlight: channel.InFlightCount,
				Requeued: channel.RequeueCount,
				Deferred: channel.DeferredCount,
			}
		}
	}

	channelsMap := make(map[string]entity.Entity)
	for _, c := range channelsDB {
		channelsMap[c.Name] = c
	}

	hasChanges := false

	// no need to remove channel that doesn't exists in upstream but exists in db
	// rely on the removal from the channel list manually

	// Add channels that exist in upstream but not in DB
	for channelName := range channelStats {
		if _, exists := channelsMap[channelName]; !exists {
			if _, err := topicLogic.CreateChannel(topic, channelName, uc.repo); err != nil {
				fmt.Printf("error creating channel %s: %v\n", channelName, err)
			} else {
				hasChanges = true
			}
		}
	}

	// Refresh channels only if there were changes
	if hasChanges {
		channelsDB, err = uc.repo.GetAllNsqTopicChannels(topic)
		if err != nil {
			return ListTopicChannelsResponse{}, err
		}
	}

	channelResponses := make([]ChannelResponse, len(channelsDB))
	for i, c := range channelsDB {
		stats := channelStats[c.Name]
		channelResponses[i] = ChannelResponse{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Topic:       c.Metadata["topic"],
			Depth:       stats.Depth,
			Messages:    stats.Messages,
			InFlight:    stats.InFlight,
			Requeued:    stats.Requeued,
			Deferred:    stats.Deferred,
		}
	}

	return ListTopicChannelsResponse{Channels: channelResponses}, nil
}

//go:generate mockgen -source=list_topic_channels.go -destination=mock_list_topic_channels_repo.go -package=detail iListTopicChannelsRepo
type iListTopicChannelsRepo interface {
	topicLogic.ICreateChannel
	modelnsq.IStatsGetter
	GetAllNsqTopicChannels(topic string) ([]entity.Entity, error)
	DeleteChannel(topic, channel string) error
}

type listTopicChannelsRepo struct {
	db *buntdb.DB
}

func (r *listTopicChannelsRepo) GetAllNsqTopicChannels(topic string) ([]entity.Entity, error) {
	return nsqrepo.GetAllNsqTopicChannels(r.db, topic)
}

func (r *listTopicChannelsRepo) DeleteChannel(topic, channel string) error {
	return nsqrepo.DeleteNsqChannelEntity(r.db, topic, channel)
}

func (r *listTopicChannelsRepo) GetAllNsqChannelByTopic(topic string) ([]entity.Entity, error) {
	return nsqrepo.GetAllNsqTopicChannels(r.db, topic)
}

func (r *listTopicChannelsRepo) CreateNsqChannelEntity(topic, channel string) (*entity.Entity, error) {
	return nsqrepo.CreateNsqChannelEntity(r.db, topic, channel)
}

func (r *listTopicChannelsRepo) GetStats(nsqdHosts []string, topic, channel string) ([]modelnsq.Stats, error) {
	return nsqrepo.GetStats(nsqdHosts, topic, channel)
}
