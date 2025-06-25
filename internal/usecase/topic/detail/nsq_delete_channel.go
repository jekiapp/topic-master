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
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

// Input struct for deleting a channel
// json alias for API compatibility
// Only id is required
// Example: {"id": "entity_id"}
type DeleteChannelInput struct {
	ID string `json:"id"`
}

type DeleteChannelResponse struct {
	Message string `json:"message"`
}

type DeleteChannelUsecase struct {
	cfg  *config.Config
	repo iDeleteChannelRepo
}

func NewDeleteChannelUsecase(cfg *config.Config, db *buntdb.DB) DeleteChannelUsecase {
	return DeleteChannelUsecase{
		cfg:  cfg,
		repo: &deleteChannelRepo{db: db},
	}
}

func (uc DeleteChannelUsecase) Handle(ctx context.Context, params map[string]string) (DeleteChannelResponse, error) {
	id, ok := params["id"]
	if !ok {
		return DeleteChannelResponse{}, fmt.Errorf("id is required")
	}
	ent, err := uc.repo.GetEntityByID(id)
	if err != nil {
		return DeleteChannelResponse{}, fmt.Errorf("entity not found: %w", err)
	}
	if ent.Resource != entity.EntityResource_NSQ || ent.TypeID != entity.EntityType_NSQChannel {
		return DeleteChannelResponse{}, fmt.Errorf("entity is not an NSQ channel")
	}
	topic, ok := ent.Metadata["topic"]
	if !ok || topic == "" {
		return DeleteChannelResponse{}, fmt.Errorf("channel entity missing topic metadata")
	}
	channel := ent.Name

	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, topic)
	if err != nil {
		return DeleteChannelResponse{}, fmt.Errorf("failed to get nsqd hosts: %w", err)
	}
	hostAddrs := make([]string, len(nsqdHosts))
	for i, h := range nsqdHosts {
		hostAddrs[i] = h.Address
	}

	errs := util.ParallelForEachHost(hostAddrs, topic, channel, func(host, topic, channel string) error {
		return uc.repo.DeleteChannelFromNsqd(host, topic, channel)
	})
	for _, e := range errs {
		if e != nil {
			log.Println("failed to delete channel from some nsqd", e)
		}
	}

	if err := uc.repo.DeleteChannelEntity(ent.ID); err != nil {
		return DeleteChannelResponse{}, fmt.Errorf("failed to delete channel entity from db: %w", err)
	}

	return DeleteChannelResponse{Message: "Channel deleted successfully"}, nil
}

type iDeleteChannelRepo interface {
	GetEntityByID(id string) (entity.Entity, error)
	GetNsqdHosts(lookupdURL, topic string) ([]nsqmodel.SimpleNsqd, error)
	DeleteChannelEntity(id string) error
	DeleteChannelFromNsqd(host, topic, channel string) error
}

type deleteChannelRepo struct {
	db *buntdb.DB
}

func (r *deleteChannelRepo) GetEntityByID(id string) (entity.Entity, error) {
	return entityrepo.GetEntityByID(r.db, id)
}

func (r *deleteChannelRepo) GetNsqdHosts(lookupdURL, topicName string) ([]nsqmodel.SimpleNsqd, error) {
	return nsqlogic.GetNsqdHosts(lookupdURL, topicName)
}

func (r *deleteChannelRepo) DeleteChannelEntity(id string) error {
	return dbpkg.DeleteByID[entity.Entity](r.db, id)
}

func (r *deleteChannelRepo) DeleteChannelFromNsqd(host, topic, channel string) error {
	return nsqrepo.DeleteChannelFromNsqd(host, topic, channel)
}
