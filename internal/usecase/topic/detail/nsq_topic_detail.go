// topic detail usecase

package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/jekiapp/topic-master/internal/config"
	nsqlogic "github.com/jekiapp/topic-master/internal/logic/nsq"
	"github.com/jekiapp/topic-master/internal/model/entity"
	nsqmodel "github.com/jekiapp/topic-master/internal/model/nsq"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type NsqTopicDetailResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	EventTrigger   string                `json:"event_trigger"`
	GroupOwner     string                `json:"group_owner"`
	Bookmarked     bool                  `json:"bookmarked"`
	Permission     Permission            `json:"permission"`
	NsqdHosts      []nsqmodel.SimpleNsqd `json:"nsqd_hosts"`
	PlatformStatus PlatformStatus        `json:"platform_status"`
}

type PlatformStatus struct {
	IsPaused bool `json:"is_paused"`
}

type Permission struct {
	CanPause              bool `json:"can_pause"`
	CanPublish            bool `json:"can_publish"`
	CanTail               bool `json:"can_tail"`
	CanDelete             bool `json:"can_delete"`
	CanEmpty              bool `json:"can_empty"`
	CanUpdateEventTrigger bool `json:"can_update_event_trigger"`
}

type NsqTopicDetailUsecase struct {
	cfg  *config.Config
	repo iNsqTopicDetailRepo
}

func NewNsqTopicDetailUsecase(cfg *config.Config, db *buntdb.DB) NsqTopicDetailUsecase {
	return NsqTopicDetailUsecase{
		cfg:  cfg,
		repo: &nsqTopicDetailRepo{db: db},
	}
}

// params should contain "topic" and "lookupd_url" keys
func (uc NsqTopicDetailUsecase) HandleQuery(ctx context.Context, params map[string]string) (NsqTopicDetailResponse, error) {
	entityID, ok := params["topic"]
	if !ok {
		return NsqTopicDetailResponse{}, nil // or return error
	}

	ent, err := uc.repo.GetEntityByID(entityID)
	if err != nil {
		return NsqTopicDetailResponse{}, fmt.Errorf("error getting topic entity: %v", err)
	}

	topicName := ent.Name
	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, topicName)
	if err != nil {
		nsqdHosts = nil // or log error, but don't fail the whole response
	}

	hosts := make([]string, 0, len(nsqdHosts))
	for _, h := range nsqdHosts {
		hosts = append(hosts, h.Address)
	}

	topicOwned := false
	user := util.GetUserInfo(ctx)
	if user != nil {
		for _, group := range user.Groups {
			if group.GroupName == ent.GroupOwner {
				topicOwned = true
				break
			}
		}
	}

	permission := Permission{}
	if ent.GroupOwner == entity.GroupNone || topicOwned {
		permission = Permission{
			CanPause:              true,
			CanPublish:            true,
			CanTail:               true,
			CanDelete:             true,
			CanEmpty:              true,
			CanUpdateEventTrigger: true,
		}
	}

	// --- Fill Bookmarked ---
	bookmarked := false
	if user != nil && user.ID != "" {
		isB, err := uc.repo.IsBookmarked(ent.ID, user.ID)
		if err == nil {
			bookmarked = isB
		}
	}

	// --- Fill PlatformStatus ---
	platformStatus := PlatformStatus{}

	stats, err := uc.repo.GetStats(hosts, topicName, "")
	if err != nil {
		log.Printf("failed to get stats: %v", err)
	}

	// check if any of the nsqd hosts is paused
	for _, stat := range stats {
		if stat.Paused {
			platformStatus.IsPaused = true
			break
		}
	}

	resp := NsqTopicDetailResponse{
		ID:             ent.ID,
		Name:           ent.Name,
		EventTrigger:   ent.Description,
		GroupOwner:     ent.GroupOwner,
		Bookmarked:     bookmarked,
		Permission:     permission,
		NsqdHosts:      nsqdHosts,
		PlatformStatus: platformStatus,
	}
	return resp, nil
}

type PublishMessageResponse struct {
	Message string `json:"message"`
}

type PublishMessageInput struct {
	Topic     string   `json:"topic"`
	Message   string   `json:"message"`
	NsqdHosts []string `json:"nsqd_hosts"`
}

func (uc NsqTopicDetailUsecase) HandlePublish(ctx context.Context, input PublishMessageInput) (PublishMessageResponse, error) {
	host := input.NsqdHosts[0]
	err := nsqrepo.Publish(input.Topic, input.Message, host)
	if err != nil {
		return PublishMessageResponse{}, fmt.Errorf("error publishing message: %v", err)
	}
	return PublishMessageResponse{
		Message: "Message published",
	}, nil
}

type iNsqTopicDetailRepo interface {
	nsqmodel.IStatsGetter
	GetEntityByID(id string) (entity.Entity, error)
	GetNsqdHosts(lookupdURL, topic string) ([]nsqmodel.SimpleNsqd, error)
	IsBookmarked(id, userID string) (bool, error)
}

type nsqTopicDetailRepo struct {
	db *buntdb.DB
}

func (r *nsqTopicDetailRepo) IsBookmarked(id, userID string) (bool, error) {
	return entityrepo.IsBookmarked(r.db, id, userID)
}

func (r *nsqTopicDetailRepo) GetEntityByID(topic string) (entity.Entity, error) {
	return entityrepo.GetEntityByID(r.db, topic)
}

func (r *nsqTopicDetailRepo) GetNsqdHosts(lookupdURL, topicID string) ([]nsqmodel.SimpleNsqd, error) {
	return nsqlogic.GetNsqdHosts(lookupdURL, topicID)
}

func (r *nsqTopicDetailRepo) GetStats(nsqdHosts []string, topic, channel string) ([]nsqmodel.Stats, error) {
	return nsqrepo.GetStats(nsqdHosts, topic, channel)
}
