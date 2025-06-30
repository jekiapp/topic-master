package detail

import (
	"context"
	"errors"
	"fmt"
	"strings"

	topicLogic "github.com/jekiapp/topic-master/internal/logic/topic"
	"github.com/jekiapp/topic-master/internal/model/entity"
	modelnsq "github.com/jekiapp/topic-master/internal/model/nsq"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

// NsqChannelListResponse and NsqChannelResponse for JSON alias

type NsqChannelListResponse struct {
	Channels []NsqChannelResponse `json:"channels"`
}

type NsqChannelResponse struct {
	ID           string `json:"id"`
	IsBookmarked bool   `json:"is_bookmarked"`
	Name         string `json:"name"`
	GroupOwner   string `json:"group_owner"`
	Description  string `json:"description"`
	Topic        string `json:"topic"`
	IsPaused     bool   `json:"is_paused"`

	CanPause  bool `json:"can_pause"`
	CanEmpty  bool `json:"can_empty"`
	CanDelete bool `json:"can_delete"`
}

func NewNsqChannelListUsecase(db *buntdb.DB) NsqChannelListUsecase {
	return NsqChannelListUsecase{
		db:   db,
		repo: &nsqChannelListRepo{db: db},
	}
}

type NsqChannelListUsecase struct {
	db   *buntdb.DB
	repo iNsqChannelListRepo
}

// HandleQuery handles HTTP query for listing channels by topic.
// params should contain "topic" key and "hosts" key (comma-separated string).
func (uc NsqChannelListUsecase) HandleQuery(ctx context.Context, params map[string]string) (NsqChannelListResponse, error) {
	topic, ok := params["topic"]
	if !ok {
		return NsqChannelListResponse{}, errors.New("topic is required")
	}

	hostsStr, ok := params["hosts"]
	if !ok {
		return NsqChannelListResponse{}, errors.New("hosts is required")
	}

	hosts := []string{}
	for _, h := range strings.Split(hostsStr, ",") {
		if h != "" {
			hosts = append(hosts, strings.TrimSpace(h))
		}
	}

	channelsDB, err := uc.repo.GetAllNsqTopicChannels(topic)
	if err != nil && err != buntdb.ErrNotFound {
		return NsqChannelListResponse{}, err
	}

	stats, err := uc.repo.GetStats(hosts, topic, "")
	if err != nil {
		return NsqChannelListResponse{}, err
	}

	if len(stats) == 0 {
		return NsqChannelListResponse{}, errors.New("no stats found")
	}

	channelStats := make(map[string]modelnsq.ChannelStats)
	for _, stat := range stats {
		for _, channel := range stat.Channels {
			channelStats[channel.ChannelName] = modelnsq.ChannelStats{
				Depth:         channel.Depth,
				Messages:      channel.MessageCount,
				InFlight:      channel.InFlightCount,
				Requeued:      channel.RequeueCount,
				Deferred:      channel.DeferredCount,
				Paused:        channel.Paused,
				ConsumerCount: channel.ClientCount,
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
			return NsqChannelListResponse{}, err
		}
		for _, c := range channelsDB {
			channelsMap[c.Name] = c
		}
	}

	user := util.GetUserInfo(ctx)
	userID := ""
	userGroups := map[string]struct{}{}
	if user != nil {
		userID = user.ID
		for _, g := range user.Groups {
			userGroups[g.GroupName] = struct{}{}
		}
	}

	channelResponses := make([]NsqChannelResponse, 0, len(channelStats))
	for name, cstats := range channelStats {
		c, ok := channelsMap[name]
		if !ok {
			continue
		}
		isBookmarked := false
		if userID != "" {
			b, err := uc.repo.IsBookmarked(c.ID, userID)
			if err == nil {
				isBookmarked = b
			}
		}

		// Permission logic (simplified like topic_detail)
		canPause, canEmpty, canDelete := false, false, false
		if c.GroupOwner == entity.GroupNone || (user != nil && userGroups[c.GroupOwner] == struct{}{}) {
			canPause, canEmpty, canDelete = true, true, true
		}

		channelResponses = append(channelResponses, NsqChannelResponse{
			ID:           c.ID,
			Name:         c.Name,
			GroupOwner:   c.GroupOwner,
			Description:  c.Description,
			Topic:        c.Metadata["topic"],
			IsBookmarked: isBookmarked,
			IsPaused:     cstats.Paused,
			CanPause:     canPause,
			CanEmpty:     canEmpty,
			CanDelete:    canDelete,
		})
	}

	return NsqChannelListResponse{Channels: channelResponses}, nil
}

//go:generate mockgen -source=nsq_channel_list.go -destination=mock_nsq_channel_list_repo.go -package=detail iNsqChannelListRepo
type iNsqChannelListRepo interface {
	topicLogic.ICreateChannel
	modelnsq.IStatsGetter
	GetAllNsqTopicChannels(topic string) ([]entity.Entity, error)
	DeleteChannel(topic, channel string) error
	IsBookmarked(id, userID string) (bool, error)
}

type nsqChannelListRepo struct {
	db *buntdb.DB
}

func (r *nsqChannelListRepo) GetAllNsqTopicChannels(topic string) ([]entity.Entity, error) {
	return nsqrepo.GetAllNsqTopicChannels(r.db, topic)
}

func (r *nsqChannelListRepo) DeleteChannel(topic, channel string) error {
	return nsqrepo.DeleteNsqChannelEntity(r.db, topic, channel)
}

func (r *nsqChannelListRepo) GetAllNsqChannelByTopic(topic string) ([]entity.Entity, error) {
	return nsqrepo.GetAllNsqTopicChannels(r.db, topic)
}

func (r *nsqChannelListRepo) CreateNsqChannelEntity(topic, channel string) (*entity.Entity, error) {
	return nsqrepo.CreateNsqChannelEntity(r.db, topic, channel)
}

func (r *nsqChannelListRepo) GetStats(nsqdHosts []string, topic, channel string) ([]modelnsq.Stats, error) {
	return nsqrepo.GetStats(nsqdHosts, topic, channel)
}

func (r *nsqChannelListRepo) IsBookmarked(id, userID string) (bool, error) {
	return entityrepo.IsBookmarked(r.db, id, userID)
}
