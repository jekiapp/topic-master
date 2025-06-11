package topic

import (
	"context"

	"github.com/jekiapp/topic-master/internal/model/entity"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ListMyBookmarkedTopicsResponse struct {
	Topics []MyBookmarkedTopicResponse `json:"topics"`
}

type MyBookmarkedTopicResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EventTrigger string `json:"event_trigger"`
	GroupOwner   string `json:"group_owner"`
	Bookmarked   bool   `json:"bookmarked"`
}

func NewListMyBookmarkedTopicsUsecase(db *buntdb.DB) ListMyBookmarkedTopicsUsecase {
	return ListMyBookmarkedTopicsUsecase{
		db:   db,
		repo: &listMyBookmarkedTopicsRepo{db: db},
	}
}

type ListMyBookmarkedTopicsUsecase struct {
	db   *buntdb.DB
	repo iListMyBookmarkedTopicsRepo
}

// HandleQuery lists only topics bookmarked by the current user.
func (uc ListMyBookmarkedTopicsUsecase) HandleQuery(ctx context.Context, _ map[string]string) (ListMyBookmarkedTopicsResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil || user.ID == "" {
		return ListMyBookmarkedTopicsResponse{Topics: nil}, nil
	}
	ids, err := uc.repo.ListBookmarkedTopicIDsByUser(user.ID)
	if err != nil {
		return ListMyBookmarkedTopicsResponse{}, err
	}
	if len(ids) == 0 {
		return ListMyBookmarkedTopicsResponse{Topics: nil}, nil
	}
	topicEntities, err := uc.repo.GetNsqTopicEntitiesByIDs(ids)
	if err != nil {
		return ListMyBookmarkedTopicsResponse{}, err
	}
	topics := make([]MyBookmarkedTopicResponse, len(topicEntities))
	for i, t := range topicEntities {
		topics[i] = MyBookmarkedTopicResponse{
			ID:           t.ID,
			Name:         t.Name,
			EventTrigger: t.Description,
			GroupOwner:   t.GroupOwner,
			Bookmarked:   true,
		}
	}
	return ListMyBookmarkedTopicsResponse{Topics: topics}, nil
}

type iListMyBookmarkedTopicsRepo interface {
	ListBookmarkedTopicIDsByUser(userID string) ([]string, error)
	GetNsqTopicEntitiesByIDs(ids []string) ([]entity.Entity, error)
}

type listMyBookmarkedTopicsRepo struct {
	db *buntdb.DB
}

func (r *listMyBookmarkedTopicsRepo) ListBookmarkedTopicIDsByUser(userID string) ([]string, error) {
	return entityrepo.ListBookmarkedTopicIDsByUser(r.db, userID, entity.EntityType_NSQTopic)
}

func (r *listMyBookmarkedTopicsRepo) GetNsqTopicEntitiesByIDs(ids []string) ([]entity.Entity, error) {
	return entityrepo.GetNsqTopicEntitiesByIDs(r.db, ids)
}
