package topic

import (
	"context"

	"github.com/jekiapp/topic-master/internal/model/entity"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	dbPkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type ListTopicsResponse struct {
	Topics []TopicResponse `json:"topics"`
}

type TopicResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EventTrigger string `json:"event_trigger"`
	GroupOwner   string `json:"group_owner"`
	Bookmarked   bool   `json:"bookmarked"`
}

func NewListAllTopicsUsecase(db *buntdb.DB) ListAllTopicsUsecase {
	return ListAllTopicsUsecase{
		db:   db,
		repo: &listTopicsRepo{db: db},
	}
}

// HandleQuery handles HTTP query for listing topics by group.
// params should contain "group" key.
func (uc ListAllTopicsUsecase) HandleQuery(ctx context.Context, params map[string]string) (ListTopicsResponse, error) {
	var topicEntities []entity.Entity
	var err error
	topicEntities, err = uc.repo.GetAllNsqTopicEntities()
	if err != nil && err != dbPkg.ErrNotFound {
		return ListTopicsResponse{}, err
	}
	topics := make([]TopicResponse, len(topicEntities))
	for i, t := range topicEntities {
		topics[i] = TopicResponse{
			ID:           t.ID,
			Name:         t.Name,
			EventTrigger: t.Description,
			GroupOwner:   t.GroupOwner,
			Bookmarked:   i%2 == 0, // todo: get data from db later
		}
	}
	return ListTopicsResponse{Topics: topics}, nil
}

type iListTopicsRepo interface {
	ListNsqTopicEntitiesByGroup(group string) ([]entity.Entity, error)
	GetAllNsqTopicEntities() ([]entity.Entity, error)
}

type listTopicsRepo struct {
	db *buntdb.DB
}

func (r *listTopicsRepo) ListNsqTopicEntitiesByGroup(group string) ([]entity.Entity, error) {
	return entityrepo.ListNsqTopicEntitiesByGroup(r.db, group)
}

func (r *listTopicsRepo) GetAllNsqTopicEntities() ([]entity.Entity, error) {
	return entityrepo.GetAllNsqTopicEntities(r.db)
}

type ListAllTopicsUsecase struct {
	db   *buntdb.DB
	repo iListTopicsRepo
}
