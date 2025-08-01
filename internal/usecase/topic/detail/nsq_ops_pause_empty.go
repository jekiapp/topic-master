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
	nsqmodel "github.com/jekiapp/topic-master/internal/model/nsq"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/jekiapp/topic-master/pkg/util"
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
	GetNsqdHosts(lookupdURL, topicName string) ([]nsqmodel.SimpleNsqd, error)
	PauseTopicOnNsqd(host, topic string) error
	EmptyTopicOnNsqd(host, topic string) error
	IsTopicPausedOnNsqd(host, topic string) (bool, error)
	ResumeTopicOnNsqd(host, topic string) error
	GetStats(nsqdHosts []string, topic, channel string) ([]nsqmodel.Stats, error)
}

type nsqOpsPauseEmptyRepo struct {
	db *buntdb.DB
}

func (r *nsqOpsPauseEmptyRepo) GetEntityByID(id string) (entity.Entity, error) {
	return entityrepo.GetEntityByID(r.db, id)
}

func (r *nsqOpsPauseEmptyRepo) GetNsqdHosts(lookupdURL, topicName string) ([]nsqmodel.SimpleNsqd, error) {
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

func (r *nsqOpsPauseEmptyRepo) ResumeTopicOnNsqd(host, topic string) error {
	return nsqrepo.ResumeTopicOnNsqd(host, topic)
}

func (r *nsqOpsPauseEmptyRepo) GetStats(nsqdHosts []string, topic, channel string) ([]nsqmodel.Stats, error) {
	return nsqrepo.GetStats(nsqdHosts, topic, channel)
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

	hosts := make([]string, 0, len(nsqdHosts))
	for _, h := range nsqdHosts {
		hosts = append(hosts, h.Address)
	}

	stats, err := uc.repo.GetStats(hosts, ent.Name, "")
	if err != nil {
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to get stats: %w", err)
	}

	pausedHosts := 0
	for _, stat := range stats {
		if stat.Paused {
			pausedHosts++
		}
	}
	if pausedHosts == len(nsqdHosts) {
		return NsqOpsPauseEmptyResponse{Message: "Topic is already paused on all hosts"}, nil
	}
	errs := util.ParallelForEachHost(hosts, ent.Name, "", func(host, topic, _ string) error {
		return uc.repo.PauseTopicOnNsqd(host, topic)
	})
	for i, e := range errs {
		if e != nil {
			return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to pause topic on nsqd host %s: %w", hosts[i], e)
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

func (uc NsqOpsPauseEmptyUsecase) HandleResume(ctx context.Context, params map[string]string) (NsqOpsPauseEmptyResponse, error) {
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

	hosts := make([]string, 0, len(nsqdHosts))
	for _, h := range nsqdHosts {
		hosts = append(hosts, h.Address)
	}
	errs := util.ParallelForEachHost(hosts, ent.Name, "", func(host, topic, _ string) error {
		return uc.repo.ResumeTopicOnNsqd(host, topic)
	})
	for i, e := range errs {
		if e != nil {
			return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to resume topic on nsqd host %s: %w", hosts[i], e)
		}
	}
	return NsqOpsPauseEmptyResponse{Message: "Topic resumed successfully"}, nil
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

	hosts := make([]string, 0, len(nsqdHosts))
	for _, h := range nsqdHosts {
		hosts = append(hosts, h.Address)
	}

	switch action {
	case "pause":
		errs := util.ParallelForEachHost(hosts, ent.Name, "", func(host, topic, _ string) error {
			return uc.repo.PauseTopicOnNsqd(host, topic)
		})
		for i, e := range errs {
			if e != nil {
				return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to pause topic on nsqd host %s: %w", hosts[i], e)
			}
		}
		return NsqOpsPauseEmptyResponse{Message: "Topic paused successfully"}, nil
	case "empty":
		errs := util.ParallelForEachHost(hosts, ent.Name, "", func(host, topic, _ string) error {
			return uc.repo.EmptyTopicOnNsqd(host, topic)
		})
		for i, e := range errs {
			if e != nil {
				return NsqOpsPauseEmptyResponse{}, fmt.Errorf("failed to empty topic on nsqd host %s: %w", hosts[i], e)
			}
		}
		return NsqOpsPauseEmptyResponse{Message: "Topic emptied successfully"}, nil
	default:
		return NsqOpsPauseEmptyResponse{}, fmt.Errorf("invalid action: %s", action)
	}
}
