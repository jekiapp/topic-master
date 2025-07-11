package detail

import (
	"context"
	"fmt"

	"github.com/jekiapp/topic-master/internal/config"
	nsqlogic "github.com/jekiapp/topic-master/internal/logic/nsq"
	"github.com/jekiapp/topic-master/internal/model/entity"
	nsqmodel "github.com/jekiapp/topic-master/internal/model/nsq"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

// Input struct for deleting a topic
// json alias for API compatibility
// Only id is required
// Example: {"id": "entity_id"}
type DeleteTopicInput struct {
	ID string `json:"id"`
}

type DeleteTopicResponse struct {
	Message string `json:"message"`
}

type DeleteTopicUsecase struct {
	cfg  *config.Config
	repo iDeleteTopicRepo
}

type iDeleteTopicRepo interface {
	GetEntityByID(id string) (entity.Entity, error)
	GetNsqdHosts(lookupdURL, topic string) ([]nsqmodel.SimpleNsqd, error)
	DeleteEntityByID(id string) error
	DeleteTopicFromNsqd(host, topic string) error
	GetChannelsByTopic(topic string) ([]entity.Entity, error)
}

type deleteTopicRepo struct {
	db *buntdb.DB
}

func (r *deleteTopicRepo) GetEntityByID(id string) (entity.Entity, error) {
	return entityrepo.GetEntityByID(r.db, id)
}

func (r *deleteTopicRepo) GetNsqdHosts(lookupdURL, topicName string) ([]nsqmodel.SimpleNsqd, error) {
	return nsqlogic.GetNsqdHosts(lookupdURL, topicName)
}

func (r *deleteTopicRepo) DeleteEntityByID(id string) error {
	return dbpkg.DeleteByID[entity.Entity](r.db, id)
}

func (r *deleteTopicRepo) DeleteTopicFromNsqd(host, topic string) error {
	return nsqrepo.DeleteTopicFromNsqd(host, topic)
}

func (r *deleteTopicRepo) GetChannelsByTopic(topic string) ([]entity.Entity, error) {
	return dbpkg.SelectAll[entity.Entity](r.db, "="+topic, entity.IdxEntity_TopicChannel)
}

func NewDeleteTopicUsecase(cfg *config.Config, db *buntdb.DB) DeleteTopicUsecase {
	return DeleteTopicUsecase{
		cfg:  cfg,
		repo: &deleteTopicRepo{db: db},
	}
}

func (uc DeleteTopicUsecase) Handle(ctx context.Context, params map[string]string) (DeleteTopicResponse, error) {
	id, ok := params["id"]
	if !ok {
		return DeleteTopicResponse{}, fmt.Errorf("id is required")
	}
	ent, err := uc.repo.GetEntityByID(id)
	if err != nil {
		return DeleteTopicResponse{}, fmt.Errorf("entity not found: %w", err)
	}
	if ent.Resource == "" {
		return DeleteTopicResponse{}, fmt.Errorf("entity resource is empty")
	}

	if ent.Resource == "NSQ" {
		nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, ent.Name)
		if err != nil {
			return DeleteTopicResponse{}, fmt.Errorf("failed to get nsqd hosts: %w", err)
		}
		for _, host := range nsqdHosts {
			if err := uc.repo.DeleteTopicFromNsqd(host.Address, ent.Name); err != nil {
				return DeleteTopicResponse{}, fmt.Errorf("failed to delete topic from nsqd host %s: %w", host, err)
			}
		}
	} else {
		return DeleteTopicResponse{}, fmt.Errorf("entity %s is not supported", ent.Resource)
	}

	if err := uc.repo.DeleteEntityByID(ent.ID); err != nil {
		return DeleteTopicResponse{}, fmt.Errorf("failed to delete entity from db: %w", err)
	} else {
		fmt.Printf("[INFO] topic %s deleted from db\n", ent.Name)
	}

	channels, err := uc.repo.GetChannelsByTopic(ent.Name)
	if err != nil && err != buntdb.ErrNotFound {
		return DeleteTopicResponse{}, fmt.Errorf("failed to get channels by topic: %w", err)
	}
	for _, channel := range channels {
		if err := uc.repo.DeleteEntityByID(channel.ID); err != nil {
			fmt.Printf("[ERROR] failed to delete channel from db: %s\n", err)
		} else {
			fmt.Printf("[INFO] channel %s deleted from db\n", channel.Name)
		}
	}

	return DeleteTopicResponse{Message: "Topic deleted successfully"}, nil
}
