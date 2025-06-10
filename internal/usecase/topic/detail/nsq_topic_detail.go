// topic detail usecase

package detail

import (
	"context"
	"fmt"

	"github.com/jekiapp/topic-master/internal/config"
	"github.com/jekiapp/topic-master/internal/model/entity"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type NsqTopicDetailResponse struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	EventTrigger   string         `json:"event_trigger"`
	GroupOwner     string         `json:"group_owner"`
	Bookmarked     bool           `json:"bookmarked"`
	Permission     Permission     `json:"permission"`
	NsqdHosts      []string       `json:"nsqd_hosts"`
	PlatformStatus PlatformStatus `json:"platform_status"`
}

type PlatformStatus struct {
	IsPaused bool `json:"is_paused"`
}

type Permission struct {
	CanPause              bool `json:"can_pause"`
	CanPublish            bool `json:"can_publish"`
	CanTail               bool `json:"can_tail"`
	CanDelete             bool `json:"can_delete"`
	CanEmptyQueue         bool `json:"can_empty_queue"`
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
	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, entityID)
	if err != nil {
		nsqdHosts = nil // or log error, but don't fail the whole response
	}

	// detect if the hosts is docker network
	// this is only happens in local development
	nsqdHosts = util.ReplaceDockerHostWithLocalhost(nsqdHosts)

	permission := Permission{}
	if ent.GroupOwner == entity.GroupNone {
		permission = Permission{
			CanPause:              true,
			CanPublish:            true,
			CanTail:               true,
			CanDelete:             true,
			CanEmptyQueue:         true,
			CanUpdateEventTrigger: true,
		}
	}

	// --- Fill Bookmarked ---
	bookmarked := false
	user := util.GetUserInfo(ctx)
	if user != nil && user.ID != "" {
		isB, err := entityrepo.IsBookmarked(uc.repo.(*nsqTopicDetailRepo).db, ent.ID, user.ID)
		if err == nil {
			bookmarked = isB
		}
	}

	resp := NsqTopicDetailResponse{
		ID:           ent.ID,
		Name:         ent.Name,
		EventTrigger: ent.Description,
		GroupOwner:   ent.GroupOwner,
		Bookmarked:   bookmarked,
		Permission:   permission,
		NsqdHosts:    nsqdHosts,
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
	GetEntityByID(id string) (entity.Entity, error)
	GetNsqdHosts(lookupdURL, topic string) ([]string, error)
}

type nsqTopicDetailRepo struct {
	db *buntdb.DB
}

func (r *nsqTopicDetailRepo) GetEntityByID(topic string) (entity.Entity, error) {
	return entityrepo.GetEntityByID(r.db, topic)
}

func (r *nsqTopicDetailRepo) GetNsqdHosts(lookupdURL, topicID string) ([]string, error) {
	entity, err := r.GetEntityByID(topicID)
	if err != nil {
		return nil, fmt.Errorf("error getting entity by id: %v", err)
	}

	nsqds, err := nsqrepo.GetNsqdsForTopic(lookupdURL, entity.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting nsqds for topic: %v", err)
	}
	hosts := make([]string, 0, len(nsqds))
	for _, n := range nsqds {
		hosts = append(hosts, fmt.Sprintf("%s:%d", n.BroadcastAddress, n.HTTPPort))
	}
	return hosts, nil
}
