package entity

import (
	"time"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/jekiapp/nsqper/pkg/db"
	"github.com/tidwall/buntdb"
)

func CreateNsqTopicEntity(dbConn *buntdb.DB, topic string) (*acl.Entity, error) {
	entity := &acl.Entity{
		TypeID:     acl.EntityType_NSQTopic,
		Name:       topic,
		Resource:   "NSQ",
		Status:     "active",
		GroupOwner: "root",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := db.Insert(dbConn, *entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func GetNsqTopicEntity(dbConn *buntdb.DB, topic string) (*acl.Entity, error) {
	tmp := acl.Entity{TypeID: acl.EntityType_NSQTopic, Name: topic}
	entity, err := db.SelectOne[acl.Entity](dbConn, tmp, acl.IdxEntity_TypeName)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func GetAllNsqTopicEntities(dbConn *buntdb.DB) ([]acl.Entity, error) {
	tmp := acl.Entity{TypeID: acl.EntityType_NSQTopic}
	entities, err := db.SelectAll[acl.Entity](dbConn, tmp, acl.IdxEntity_TypeID)
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func DeleteNsqTopicEntity(dbConn *buntdb.DB, topic string) error {
	tmp := acl.Entity{TypeID: acl.EntityType_NSQTopic, Name: topic}
	return db.DeleteByIndex(dbConn, tmp, acl.IdxEntity_TypeName)
}

// ListNsqTopicEntitiesByGroup returns all nsq topic entities owned by the given group. If group is "root", returns all topics.
func ListNsqTopicEntitiesByGroup(dbConn *buntdb.DB, group string) ([]acl.Entity, error) {
	tmp := acl.Entity{GroupOwner: group, TypeID: acl.EntityType_NSQTopic}
	entities, err := db.SelectAll[acl.Entity](dbConn, tmp, acl.IdxEntity_GroupType)
	if err != nil {
		return nil, err
	}
	return entities, nil
}
