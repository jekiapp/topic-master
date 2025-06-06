// usecase to sync topics
// create new public function to sync topics, use logic/topic/topic.go
// implement interface iSyncTopics in logic/topic/topic.go
// use topic repo for the actual interface implementation

// this public function can be called directly from main.go
// create also generic http query handler for it to be able to be triggered from http handler

package topic

import (
	"context"

	"github.com/jekiapp/topic-master/internal/logic/topic"
	"github.com/jekiapp/topic-master/internal/model/acl"
	entity "github.com/jekiapp/topic-master/internal/repository/entity"
	nsq "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/tidwall/buntdb"
)

type SyncTopicsResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type iSyncTopicsRepo interface {
	topic.ISyncTopics
}

type syncTopicsRepo struct {
	db *buntdb.DB
}

func (r *syncTopicsRepo) GetAllTopics() ([]string, error) {
	return nsq.GetAllTopics()
}

func (r *syncTopicsRepo) GetNsqTopicEntity(topic string) (*acl.Entity, error) {
	return entity.GetNsqTopicEntity(r.db, topic)
}

func (r *syncTopicsRepo) CreateNsqTopicEntity(topic string) (*acl.Entity, error) {
	return entity.CreateNsqTopicEntity(r.db, topic)
}

func (r *syncTopicsRepo) GetAllNsqTopicEntities() ([]acl.Entity, error) {
	return entity.GetAllNsqTopicEntities(r.db)
}

func (r *syncTopicsRepo) DeleteNsqTopicEntity(topic string) error {
	return entity.DeleteNsqTopicEntity(r.db, topic)
}

type SyncTopicsUsecase struct {
	db   *buntdb.DB
	repo iSyncTopicsRepo
}

func NewSyncTopicsUsecase(db *buntdb.DB) SyncTopicsUsecase {
	return SyncTopicsUsecase{
		db:   db,
		repo: &syncTopicsRepo{db: db},
	}
}

func (uc SyncTopicsUsecase) HandleQuery(ctx context.Context, _ map[string]string) (SyncTopicsResponse, error) {
	err := topic.SyncTopics(uc.db, uc.repo)
	if err != nil {
		return SyncTopicsResponse{Success: false, Error: err.Error()}, err
	}
	return SyncTopicsResponse{Success: true}, nil
}
