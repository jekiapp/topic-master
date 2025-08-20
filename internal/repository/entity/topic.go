package entity

import (
	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func GetNsqTopicEntity(dbConn *buntdb.DB, topic string) (*entity.Entity, error) {
	pivot := entity.EntityType_NSQTopic + ":" + topic
	entityObj, err := db.SelectOne[entity.Entity](dbConn, pivot, entity.IdxEntity_TypeName)
	if err != nil {
		return nil, err
	}
	return &entityObj, nil
}

func GetAllNsqTopicEntities(dbConn *buntdb.DB) ([]entity.Entity, error) {
	entities, err := db.SelectAll[entity.Entity](dbConn, ">="+entity.EntityType_NSQTopic, entity.IdxEntity_TypeName)
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func GetAllTopicEntities(dbConn *buntdb.DB) ([]entity.Entity, error) {
	entities, err := db.SelectAll[entity.Entity](dbConn, "="+entity.EntityKind_Topic, entity.IdxEntity_Kind)
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

func GetTopicEntitiesByIDs(dbConn *buntdb.DB, ids []string) ([]entity.Entity, error) {
	entities := make([]entity.Entity, 0, len(ids))
	for _, id := range ids {
		ent, err := GetEntityByID(dbConn, id)
		if err == nil && ent.Kind == entity.EntityKind_Topic {
			entities = append(entities, ent)
		}
	}
	return entities, nil
}
