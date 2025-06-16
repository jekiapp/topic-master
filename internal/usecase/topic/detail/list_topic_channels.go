package detail

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

//go:generate mockgen -source=list_topic_channels.go -destination=mock_list_topic_channels_repo.go -package=detail iListTopicChannelsRepo
type iListTopicChannelsRepo interface {
	GetAllNsqTopicChannels(topic string) ([]entity.Entity, error)
	GetChannelStats(hosts []string, topic string) (map[string]modelnsq.ChannelStats, error)
	DeleteChannel(topic, channel string) error
	CreateChannel(topic, channel string) (*entity.Entity, error)
}

type listTopicChannelsRepo struct {
	db *buntdb.DB
}

func (r *listTopicChannelsRepo) GetAllNsqTopicChannels(topic string) ([]entity.Entity, error) {
	return nsqrepo.GetAllNsqTopicChannels(r.db, topic)
}

func (r *listTopicChannelsRepo) GetChannelStats(hosts []string, topic string) (map[string]modelnsq.ChannelStats, error) {
	return nsqrepo.GetAllChannelStats(hosts, topic)
}

func (r *listTopicChannelsRepo) DeleteChannel(topic, channel string) error {
	return nsqrepo.DeleteNsqChannelEntity(r.db, topic, channel)
}

func (r *listTopicChannelsRepo) CreateChannel(topic, channel string) (*entity.Entity, error) {
	return nsqrepo.CreateNsqChannelEntity(r.db, topic, channel)
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

	channels, err := uc.repo.GetAllNsqTopicChannels(topic)
	if err != nil {
		return ListTopicChannelsResponse{}, err
	}

	channelStats, err := uc.repo.GetChannelStats(hosts, topic)
	if err != nil {
		fmt.Printf("error getting channel stats: %v\n", err)
	}

	// sync channel from db with channel from upstream
	channelMap := make(map[string]entity.Entity)
	for _, c := range channels {
		channelMap[c.Name] = c
	}

	hasChanges := false

	// no need to remove channel that doesnt exists in upstream but exists in db
	// rely on the removal from the channel list manually

	// Add channels that exist in upstream but not in DB
	for channelName := range channelStats {
		if _, exists := channelMap[channelName]; !exists {
			if _, err := uc.repo.CreateChannel(topic, channelName); err != nil {
				fmt.Printf("error creating channel %s: %v\n", channelName, err)
			} else {
				hasChanges = true
			}
		}
	}

	// Refresh channels only if there were changes
	if hasChanges {
		channels, err = uc.repo.GetAllNsqTopicChannels(topic)
		if err != nil {
			return ListTopicChannelsResponse{}, err
		}
	}

	channelResponses := make([]ChannelResponse, len(channels))
	for i, c := range channels {
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
