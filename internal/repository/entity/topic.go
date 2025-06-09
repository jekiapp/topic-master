package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func CreateNsqTopicEntity(dbConn *buntdb.DB, topic string) (*entity.Entity, error) {
	entityObj := &entity.Entity{
		ID:         uuid.NewString(),
		TypeID:     entity.EntityType_NSQTopic,
		Name:       topic,
		Resource:   "NSQ",
		Status:     "active",
		GroupOwner: entity.GroupNone,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := db.Insert(dbConn, entityObj); err != nil {
		return nil, err
	}
	return entityObj, nil
}

func GetNsqTopicEntity(dbConn *buntdb.DB, topic string) (*entity.Entity, error) {
	pivot := entity.EntityType_NSQTopic + ":" + topic
	entityObj, err := db.SelectOne[entity.Entity](dbConn, pivot, entity.IdxEntity_TypeName)
	if err != nil {
		return nil, err
	}
	return &entityObj, nil
}

func GetAllNsqTopicEntities(dbConn *buntdb.DB) ([]entity.Entity, error) {
	entities, err := db.SelectAll[entity.Entity](dbConn, "="+entity.EntityType_NSQTopic, entity.IdxEntity_TypeID)
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func DeleteNsqTopicEntity(dbConn *buntdb.DB, topic string) error {
	tmp := &entity.Entity{TypeID: entity.EntityType_NSQTopic, Name: topic}
	return db.DeleteByIndex(dbConn, tmp, entity.IdxEntity_TypeName)
}

// ListNsqTopicEntitiesByGroup returns all nsq topic entities owned by the given group. If group is entity.GroupRoot, returns all topics.
func ListNsqTopicEntitiesByGroup(dbConn *buntdb.DB, group string) ([]entity.Entity, error) {
	pivot := group + ":" + entity.EntityType_NSQTopic
	entities, err := db.SelectAll[entity.Entity](dbConn, "="+pivot, entity.IdxEntity_GroupType)
	if err != nil {
		return nil, err
	}
	return entities, nil
}
