package topic

import (
	"context"

	"github.com/jekiapp/nsqper/internal/model/acl"
	entityrepo "github.com/jekiapp/nsqper/internal/repository/entity"
	"github.com/tidwall/buntdb"
)

type ListTopicsResponse struct {
	Topics []*acl.Entity `json:"topics"`
	Error  string        `json:"error,omitempty"`
}

type iListTopicsRepo interface {
	ListNsqTopicEntitiesByGroup(group string) ([]*acl.Entity, error)
	GetAllNsqTopicEntities() ([]*acl.Entity, error)
}

type listTopicsRepo struct {
	db *buntdb.DB
}

func (r *listTopicsRepo) ListNsqTopicEntitiesByGroup(group string) ([]*acl.Entity, error) {
	return entityrepo.ListNsqTopicEntitiesByGroup(r.db, group)
}

func (r *listTopicsRepo) GetAllNsqTopicEntities() ([]*acl.Entity, error) {
	return entityrepo.GetAllNsqTopicEntities(r.db)
}

type ListTopicsUsecase struct {
	db   *buntdb.DB
	repo iListTopicsRepo
}

func NewListTopicsUsecase(db *buntdb.DB) ListTopicsUsecase {
	return ListTopicsUsecase{
		db:   db,
		repo: &listTopicsRepo{db: db},
	}
}

// HandleQuery handles HTTP query for listing topics by group.
// params should contain "group" key.
func (uc ListTopicsUsecase) HandleQuery(ctx context.Context, params map[string]string) (ListTopicsResponse, error) {
	group := params["group"]
	if group == "" {
		return ListTopicsResponse{Error: "group is required"}, nil
	}
	var topics []*acl.Entity
	var err error
	if group == "root" {
		topics, err = uc.repo.GetAllNsqTopicEntities()
	} else {
		topics, err = uc.repo.ListNsqTopicEntitiesByGroup(group)
	}
	if err != nil {
		return ListTopicsResponse{Error: err.Error()}, err
	}
	return ListTopicsResponse{Topics: topics}, nil
}
