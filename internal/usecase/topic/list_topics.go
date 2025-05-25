package topic

import (
	"context"
	"errors"

	"github.com/jekiapp/nsqper/internal/model/acl"
	entityrepo "github.com/jekiapp/nsqper/internal/repository/entity"
	"github.com/jekiapp/nsqper/pkg/util"
	"github.com/tidwall/buntdb"
)

type ListTopicsResponse struct {
	Topics []acl.Entity `json:"topics"`
}

type iListTopicsRepo interface {
	ListNsqTopicEntitiesByGroup(group string) ([]acl.Entity, error)
	GetAllNsqTopicEntities() ([]acl.Entity, error)
}

type listTopicsRepo struct {
	db *buntdb.DB
}

func (r *listTopicsRepo) ListNsqTopicEntitiesByGroup(group string) ([]acl.Entity, error) {
	return entityrepo.ListNsqTopicEntitiesByGroup(r.db, group)
}

func (r *listTopicsRepo) GetAllNsqTopicEntities() ([]acl.Entity, error) {
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
	userInfo := util.GetUserInfo(ctx)
	if userInfo == nil {
		return ListTopicsResponse{}, errors.New("user is not authenticated")
	}
	usergroups := userInfo.Groups
	if len(usergroups) == 0 {
		return ListTopicsResponse{}, errors.New("user is not a member of any group")
	}
	group := usergroups[0]
	var topics []acl.Entity
	var err error
	if group.GroupName == "root" {
		topics, err = uc.repo.GetAllNsqTopicEntities()
	} else {
		for _, group := range usergroups {
			t, err := uc.repo.ListNsqTopicEntitiesByGroup(group.GroupName)
			if err != nil {
				return ListTopicsResponse{}, err
			}
			topics = append(topics, t...)
		}
	}
	if err != nil {
		return ListTopicsResponse{}, err
	}
	return ListTopicsResponse{Topics: topics}, nil
}
