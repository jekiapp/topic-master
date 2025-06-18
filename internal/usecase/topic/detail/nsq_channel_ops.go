// this usecase is for pausing, emptying, and resuming a channel
// it will receive param id, channel, and action
// logic:
// 1. get entity by id
// 2. if the entity resource is not NSQ then error
// 3. if the action is pause then pause channel
// 4. if the action is empty then empty_queue channel
// 5. if the action is resume then resume channel

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

// Input struct for pausing, emptying, or resuming a channel
// json alias for API compatibility
// Example: {"id": "entity_id", "channel": "channel_name", "action": "pause"}
type NsqChannelOpsInput struct {
	ID      string `json:"id"`
	Channel string `json:"channel"`
	Action  string `json:"action"`
}

type NsqChannelOpsResponse struct {
	Message string `json:"message"`
}

type NsqChannelOpsUsecase struct {
	cfg  *config.Config
	repo iNsqChannelOpsRepo
}

type iNsqChannelOpsRepo interface {
	GetEntityByID(id string) (entity.Entity, error)
	GetNsqdHosts(lookupdURL, topicName string) ([]nsqmodel.SimpleNsqd, error)
	PauseChannelOnNsqd(host, topic, channel string) error
	EmptyChannelOnNsqd(host, topic, channel string) error
	ResumeChannelOnNsqd(host, topic, channel string) error
	GetStats(nsqdHosts []string, topic, channel string) ([]nsqmodel.Stats, error)
}

type nsqChannelOpsRepo struct {
	db *buntdb.DB
}

func (r *nsqChannelOpsRepo) GetEntityByID(id string) (entity.Entity, error) {
	return entityrepo.GetEntityByID(r.db, id)
}

func (r *nsqChannelOpsRepo) GetNsqdHosts(lookupdURL, topicName string) ([]nsqmodel.SimpleNsqd, error) {
	return nsqlogic.GetNsqdHosts(lookupdURL, topicName)
}

func (r *nsqChannelOpsRepo) PauseChannelOnNsqd(host, topic, channel string) error {
	return nsqrepo.PauseChannelOnNsqd(host, topic, channel)
}

func (r *nsqChannelOpsRepo) EmptyChannelOnNsqd(host, topic, channel string) error {
	return nsqrepo.EmptyChannelOnNsqd(host, topic, channel)
}

func (r *nsqChannelOpsRepo) ResumeChannelOnNsqd(host, topic, channel string) error {
	return nsqrepo.ResumeChannelOnNsqd(host, topic, channel)
}

func (r *nsqChannelOpsRepo) GetStats(nsqdHosts []string, topic, channel string) ([]nsqmodel.Stats, error) {
	return nsqrepo.GetStats(nsqdHosts, topic, channel)
}

func NewNsqChannelOpsUsecase(cfg *config.Config, db *buntdb.DB) NsqChannelOpsUsecase {
	return NsqChannelOpsUsecase{
		cfg:  cfg,
		repo: &nsqChannelOpsRepo{db: db},
	}
}

func (uc NsqChannelOpsUsecase) HandlePause(ctx context.Context, params map[string]string) (NsqChannelOpsResponse, error) {
	id, ok := params["id"]
	if !ok {
		return NsqChannelOpsResponse{}, fmt.Errorf("id is required")
	}
	channel, ok := params["channel"]
	if !ok {
		return NsqChannelOpsResponse{}, fmt.Errorf("channel is required")
	}
	ent, err := uc.repo.GetEntityByID(id)
	if err != nil {
		return NsqChannelOpsResponse{}, fmt.Errorf("entity not found: %w", err)
	}

	if ent.Resource != "NSQ" {
		return NsqChannelOpsResponse{}, fmt.Errorf("entity resource is not NSQ")
	}

	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, ent.Name)
	if err != nil {
		return NsqChannelOpsResponse{}, fmt.Errorf("failed to get nsqd hosts: %w", err)
	}

	hosts := make([]string, 0, len(nsqdHosts))
	for _, h := range nsqdHosts {
		hosts = append(hosts, h.Address)
	}

	stats, err := uc.repo.GetStats(hosts, ent.Name, channel)
	if err != nil {
		return NsqChannelOpsResponse{}, fmt.Errorf("failed to get stats: %w", err)
	}

	pausedHosts := 0
	for _, stat := range stats {
		if stat.Paused {
			pausedHosts++
		}
	}
	if pausedHosts == len(nsqdHosts) {
		return NsqChannelOpsResponse{Message: "Channel is already paused on all hosts"}, nil
	}
	errs := util.ParallelForEachHost(hosts, ent.Name, channel, func(host, topic, channel string) error {
		return uc.repo.PauseChannelOnNsqd(host, topic, channel)
	})
	for i, e := range errs {
		if e != nil {
			return NsqChannelOpsResponse{}, fmt.Errorf("failed to pause channel on nsqd host %s: %w", hosts[i], e)
		}
	}
	return NsqChannelOpsResponse{Message: "Channel paused successfully"}, nil
}

func (uc NsqChannelOpsUsecase) HandleEmpty(ctx context.Context, params map[string]string) (NsqChannelOpsResponse, error) {
	id, ok := params["id"]
	if !ok {
		return NsqChannelOpsResponse{}, fmt.Errorf("id is required")
	}
	channel, ok := params["channel"]
	if !ok {
		return NsqChannelOpsResponse{}, fmt.Errorf("channel is required")
	}
	return uc.doNsqChannelOps(ctx, id, channel, "empty")
}

func (uc NsqChannelOpsUsecase) HandleResume(ctx context.Context, params map[string]string) (NsqChannelOpsResponse, error) {
	id, ok := params["id"]
	if !ok {
		return NsqChannelOpsResponse{}, fmt.Errorf("id is required")
	}
	channel, ok := params["channel"]
	if !ok {
		return NsqChannelOpsResponse{}, fmt.Errorf("channel is required")
	}
	ent, err := uc.repo.GetEntityByID(id)
	if err != nil {
		return NsqChannelOpsResponse{}, fmt.Errorf("entity not found: %w", err)
	}

	if ent.Resource != "NSQ" {
		return NsqChannelOpsResponse{}, fmt.Errorf("entity resource is not NSQ")
	}

	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, ent.Name)
	if err != nil {
		return NsqChannelOpsResponse{}, fmt.Errorf("failed to get nsqd hosts: %w", err)
	}

	hosts := make([]string, 0, len(nsqdHosts))
	for _, h := range nsqdHosts {
		hosts = append(hosts, h.Address)
	}
	errs := util.ParallelForEachHost(hosts, ent.Name, channel, func(host, topic, channel string) error {
		return uc.repo.ResumeChannelOnNsqd(host, topic, channel)
	})
	for i, e := range errs {
		if e != nil {
			return NsqChannelOpsResponse{}, fmt.Errorf("failed to resume channel on nsqd host %s: %w", hosts[i], e)
		}
	}
	return NsqChannelOpsResponse{Message: "Channel resumed successfully"}, nil
}

func (uc NsqChannelOpsUsecase) doNsqChannelOps(ctx context.Context, id, channel, action string) (NsqChannelOpsResponse, error) {
	ent, err := uc.repo.GetEntityByID(id)
	if err != nil {
		return NsqChannelOpsResponse{}, fmt.Errorf("entity not found: %w", err)
	}

	if ent.Resource != "NSQ" {
		return NsqChannelOpsResponse{}, fmt.Errorf("entity resource is not NSQ")
	}

	nsqdHosts, err := uc.repo.GetNsqdHosts(uc.cfg.NSQLookupdHTTPAddr, ent.Name)
	if err != nil {
		return NsqChannelOpsResponse{}, fmt.Errorf("failed to get nsqd hosts: %w", err)
	}

	hosts := make([]string, 0, len(nsqdHosts))
	for _, h := range nsqdHosts {
		hosts = append(hosts, h.Address)
	}

	switch action {
	case "pause":
		errs := util.ParallelForEachHost(hosts, ent.Name, channel, func(host, topic, channel string) error {
			return uc.repo.PauseChannelOnNsqd(host, topic, channel)
		})
		for i, e := range errs {
			if e != nil {
				return NsqChannelOpsResponse{}, fmt.Errorf("failed to pause channel on nsqd host %s: %w", hosts[i], e)
			}
		}
		return NsqChannelOpsResponse{Message: "Channel paused successfully"}, nil
	case "empty":
		errs := util.ParallelForEachHost(hosts, ent.Name, channel, func(host, topic, channel string) error {
			return uc.repo.EmptyChannelOnNsqd(host, topic, channel)
		})
		for i, e := range errs {
			if e != nil {
				return NsqChannelOpsResponse{}, fmt.Errorf("failed to empty channel on nsqd host %s: %w", hosts[i], e)
			}
		}
		return NsqChannelOpsResponse{Message: "Channel emptied successfully"}, nil
	case "resume":
		errs := util.ParallelForEachHost(hosts, ent.Name, channel, func(host, topic, channel string) error {
			return uc.repo.ResumeChannelOnNsqd(host, topic, channel)
		})
		for i, e := range errs {
			if e != nil {
				return NsqChannelOpsResponse{}, fmt.Errorf("failed to resume channel on nsqd host %s: %w", hosts[i], e)
			}
		}
		return NsqChannelOpsResponse{Message: "Channel resumed successfully"}, nil
	default:
		return NsqChannelOpsResponse{}, fmt.Errorf("invalid action: %s", action)
	}
}
