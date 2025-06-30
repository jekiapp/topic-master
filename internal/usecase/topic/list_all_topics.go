package topic

import (
	"context"

	"github.com/jekiapp/topic-master/internal/model/entity"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	dbPkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
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

type ListAllTopicsUsecase struct {
	db   *buntdb.DB
	repo iListTopicsRepo
}

// HandleQuery handles HTTP query for listing topics by group.
// params should contain "group" key.
func (uc ListAllTopicsUsecase) HandleQuery(ctx context.Context, params map[string]string) (ListTopicsResponse, error) {
	if params["is_bookmarked"] == "true" {
		return uc.listBookmarkedTopics(ctx)
	}
	return uc.listAllTopics(ctx)
}

func (uc ListAllTopicsUsecase) listBookmarkedTopics(ctx context.Context) (ListTopicsResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil || user.ID == "" {
		return ListTopicsResponse{Topics: nil}, nil
	}
	ids, err := uc.repo.ListBookmarkedTopicIDsByUser(user.ID)
	if err != nil {
		return ListTopicsResponse{}, err
	}
	if len(ids) == 0 {
		return ListTopicsResponse{Topics: nil}, nil
	}
	topicEntities, err := uc.repo.GetNsqTopicEntitiesByIDs(ids)
	if err != nil {
		return ListTopicsResponse{}, err
	}
	topics := make([]TopicResponse, len(topicEntities))
	for i, t := range topicEntities {
		topics[i] = TopicResponse{
			ID:           t.ID,
			Name:         t.Name,
			EventTrigger: t.Description,
			GroupOwner:   t.GroupOwner,
			Bookmarked:   true,
		}
	}
	return ListTopicsResponse{Topics: topics}, nil
}

func (uc ListAllTopicsUsecase) listAllTopics(ctx context.Context) (ListTopicsResponse, error) {
	user := util.GetUserInfo(ctx)
	userID := ""
	if user != nil {
		userID = user.ID
	}
	var topicEntities []entity.Entity
	var err error
	topicEntities, err = uc.repo.GetAllNsqTopicEntities()
	if err != nil && err != dbPkg.ErrNotFound {
		return ListTopicsResponse{}, err
	}
	topics := make([]TopicResponse, len(topicEntities))
	for i, t := range topicEntities {
		bookmarked := false
		if userID != "" {
			b, err := uc.repo.IsBookmarked(t.ID, userID)
			if err == nil {
				bookmarked = b
			}
		}
		topics[i] = TopicResponse{
			ID:           t.ID,
			Name:         t.Name,
			EventTrigger: t.Description,
			GroupOwner:   t.GroupOwner,
			Bookmarked:   bookmarked,
		}
	}
	return ListTopicsResponse{Topics: topics}, nil
}

type iListTopicsRepo interface {
	ListNsqTopicEntitiesByGroup(group string) ([]entity.Entity, error)
	GetAllNsqTopicEntities() ([]entity.Entity, error)
	IsBookmarked(entityID, userID string) (bool, error)
	ListBookmarkedTopicIDsByUser(userID string) ([]string, error)
	GetNsqTopicEntitiesByIDs(ids []string) ([]entity.Entity, error)
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

func (r *listTopicsRepo) IsBookmarked(entityID, userID string) (bool, error) {
	return entityrepo.IsBookmarked(r.db, entityID, userID)
}

func (r *listTopicsRepo) ListBookmarkedTopicIDsByUser(userID string) ([]string, error) {
	return entityrepo.ListBookmarkedTopicIDsByUser(r.db, userID, entity.EntityType_NSQTopic)
}

func (r *listTopicsRepo) GetNsqTopicEntitiesByIDs(ids []string) ([]entity.Entity, error) {
	return entityrepo.GetNsqTopicEntitiesByIDs(r.db, ids)
}
