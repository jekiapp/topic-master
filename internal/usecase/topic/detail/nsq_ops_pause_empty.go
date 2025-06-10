// this usecase is for pausing and emptying topic
// it will receive param id and action
// logic:
// 1. get entity by id
// 2. if the entity resource is not NSQ then error
// 3. if the action is pause then pause topic
// 4. if the action is empty then empty_queue topic

package detail

import (
	"context"
	"fmt"

	"github.com/jekiapp/topic-master/internal/config"
	nsqlogic "github.com/jekiapp/topic-master/internal/logic/nsq"
	"github.com/jekiapp/topic-master/internal/model/entity"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/tidwall/buntdb"
)

// Input struct for pausing or emptying a topic
// json alias for API compatibility
// Example: {"id": "entity_id", "action": "pause"}
type NsqOpsPauseEmptyInput struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type NsqOpsPauseEmptyResponse struct {
	Message string `json:"message"`
}

type NsqOpsPauseEmptyUsecase struct {
	cfg  *config.Config
	repo iNsqOpsPauseEmptyRepo
}

type iNsqOpsPauseEmptyRepo interface {
	GetEntityByID(id string) (entity.Entity, error)
	GetNsqdHosts(lookupdURL, topicName string) ([]string, error)
	PauseTopicOnNsqd(host, topic string) error
	EmptyTopicOnNsqd(host, topic string) error
	IsTopicPausedOnNsqd(host, topic string) (bool, error)
}

type nsqOpsPauseEmptyRepo struct {
	db *buntdb.DB
}

func (r *nsqOpsPauseEmptyRepo) GetEntityByID(id string) (entity.Entity, error) {
	return entityrepo.GetEntityByID(r.db, id)
}

func (r *nsqOpsPauseEmptyRepo) GetNsqdHosts(lookupdURL, topicName string) ([]string, error) {
	return nsqlogic.GetNsqdHosts(lookupdURL, topicName)
}

func (r *nsqOpsPauseEmptyRepo) PauseTopicOnNsqd(host, topic string) error {
	return nsqrepo.PauseTopicOnNsqd(host, topic)
}

func (r *nsqOpsPauseEmptyRepo) EmptyTopicOnNsqd(host, topic string) error {
	return nsqrepo.EmptyTopicOnNsqd(host, topic)
}

func (r *nsqOpsPauseEmptyRepo) IsTopicPausedOnNsqd(host, topic string) (bool, error) {
	return nsqrepo.IsTopicPausedOnNsqd(host, topic)
}

func NewNsqOpsPauseEmptyUsecase(cfg *config.Config, db *buntdb.DB) NsqOpsPauseEmptyUsecase {
	return NsqOpsPauseEmptyUsecase{
		cfg:  cfg,
		repo: &nsqOpsPauseEmptyRepo{db: db},
	}
}

func (uc NsqOpsPauseEmptyUsecase) HandlePause(ctx context.Context, params map[string]string) (NsqOpsPauseEmptyResponse, error) {
	id, ok := params["id"]
	if !ok {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("id is required")
	}
	ent, err := uc.repo.GetEntityByID(id)
	if err != nil {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("entity not found: %w", err)
	}

	if ent.Resource != "NSQ" {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("entity resource is not NSQ")
	}

	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, ent.Name)
	if err != nil {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to get nsqd hosts: %w", err)
	}
	pausedHosts := 0
	for _, host := range nsqdHosts {
		paused, err := uc.repo.IsTopicPausedOnNsqd(host, ent.Name)
		if err != nil {
			return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to check paused status on nsqd host %s: %w", host, err)
		}
		if paused {
			pausedHosts++
		}
	}
	if pausedHosts == len(nsqdHosts) {
		return NsqOpsPauseEmptyResponse{Message: "Topic is already paused on all hosts"}, nil
	}
	for _, host := range nsqdHosts {
		if err := uc.repo.PauseTopicOnNsqd(host, ent.Name); err != nil {
			return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to pause topic on nsqd host %s: %w", host, err)
		}
	}
	return NsqOpsPauseEmptyResponse{Message: "Topic paused successfully"}, nil
}

func (uc NsqOpsPauseEmptyUsecase) HandleEmpty(ctx context.Context, params map[string]string) (NsqOpsPauseEmptyResponse, error) {
	id, ok := params["id"]
	if !ok {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("id is required")
	}
	return uc.doNsqOps(ctx, id, "empty")
}

func (uc NsqOpsPauseEmptyUsecase) doNsqOps(ctx context.Context, id, action string) (NsqOpsPauseEmptyResponse, error) {
	ent, err := uc.repo.GetEntityByID(id)
	if err != nil {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("entity not found: %w", err)
	}

	if ent.Resource != "NSQ" {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("entity resource is not NSQ")
	}

	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, ent.Name)
	if err != nil {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to get nsqd hosts: %w", err)
	}

	switch action {
	case "pause":
		for _, host := range nsqdHosts {
			if err := uc.repo.PauseTopicOnNsqd(host, ent.Name); err != nil {
				return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to pause topic on nsqd host %s: %w", host, err)
			}
		}
		return NsqOpsPauseEmptyResponse{Message: "Topic paused successfully"}, nil
	case "empty":
		for _, host := range nsqdHosts {
			if err := uc.repo.EmptyTopicOnNsqd(host, ent.Name); err != nil {
				return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to empty topic on nsqd host %s: %w", host, err)
			}
		}
		return NsqOpsPauseEmptyResponse{Message: "Topic emptied successfully"}, nil
	default:
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("invalid action: %s", action)
	}
}
