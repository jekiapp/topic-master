// topic detail usecase

package detail

import (
	"context"
	"fmt"

	"github.com/jekiapp/topic-master/internal/config"
	"github.com/jekiapp/topic-master/internal/model/acl"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/tidwall/buntdb"
)

type NsqTopicDetailResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	EventTrigger string     `json:"event_trigger"`
	GroupOwner   string     `json:"group_owner"`
	Bookmarked   bool       `json:"bookmarked"`
	Permission   permission `json:"permission"`
	NsqdHosts    []string   `json:"nsqd_hosts"`
}

type permission struct {
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

func NewNsqTopicDetailUsecase(db *buntdb.DB) NsqTopicDetailUsecase {
	return NsqTopicDetailUsecase{
		repo: &nsqTopicDetailRepo{db: db},
	}
}

// params should contain "topic" and "lookupd_url" keys
func (uc NsqTopicDetailUsecase) HandleQuery(ctx context.Context, params map[string]string) (NsqTopicDetailResponse, error) {
	topic, ok := params["topic"]
	if !ok {
		return NsqTopicDetailResponse{}, nil // or return error
	}

	entity, err := uc.repo.GetNsqTopicEntity(topic)
	if err != nil {
		return NsqTopicDetailResponse{}, err
	}
	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, topic)
	if err != nil {
		nsqdHosts = nil // or log error, but don't fail the whole response
	}
	resp := NsqTopicDetailResponse{
		ID:           entity.ID,
		Name:         entity.Name,
		EventTrigger: entity.Description, // TODO: map correctly if needed
		GroupOwner:   entity.GroupOwner,
		Bookmarked:   false,        // TODO: fill later
		Permission:   permission{}, // TODO: fill later
		NsqdHosts:    nsqdHosts,
	}
	return resp, nil
}

type iNsqTopicDetailRepo interface {
	GetNsqTopicEntity(topic string) (*acl.Entity, error)
	GetNsqdHosts(lookupdURL, topic string) ([]string, error)
}

type nsqTopicDetailRepo struct {
	db *buntdb.DB
}

func (r *nsqTopicDetailRepo) GetNsqTopicEntity(topic string) (*acl.Entity, error) {
	return entityrepo.GetNsqTopicEntity(r.db, topic)
}

func (r *nsqTopicDetailRepo) GetNsqdHosts(lookupdURL, topic string) ([]string, error) {
	nsqds, err := nsqrepo.GetNsqdsForTopic(lookupdURL, topic)
	if err != nil {
		return nil, err
	}
	hosts := make([]string, 0, len(nsqds))
	for _, n := range nsqds {
		hosts = append(hosts, fmt.Sprintf("%s:%d", n.BroadcastAddress, n.HTTPPort))
	}
	return hosts, nil
}
