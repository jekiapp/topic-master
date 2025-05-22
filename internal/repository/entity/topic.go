package entity

import (
	"time"

	"github.com/jekiapp/nsqper/internal/model"
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

func CreateNsqTopicEntity(db *buntdb.DB, topic string) (*acl.Entity, error) {
	id := entityPrefix + "nsq_topic:" + topic
	entity := &acl.Entity{
		ID:        id,
		TypeID:    "nsq_topic",
		Name:      "Topic: " + topic,
		Resource:  "NSQ",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	data, err := msgpack.Marshal(entity)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(id, string(data), nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func GetNsqTopicEntity(db *buntdb.DB, topic string) (*acl.Entity, error) {
	id := entityPrefix + "nsq_topic:" + topic
	var entity acl.Entity
	err := db.View(func(tx *buntdb.Tx) error {
		resp, err := tx.Get(id)
		if err != nil {
			return err
		}
		if resp == "" {
			return model.ErrNotFound
		}
		err = msgpack.Unmarshal([]byte(resp), &entity)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, model.ErrNotFound
	}
	return &entity, nil
}

func GetAllNsqTopicEntities(db *buntdb.DB) ([]*acl.Entity, error) {
	var entities []*acl.Entity
	var firstErr error
	prefix := entityPrefix + "nsq_topic:"

	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(prefix+"*", func(key, value string) bool {
			var entity acl.Entity
			if err := msgpack.Unmarshal([]byte(value), &entity); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return false // stop the iteration
			}
			entities = append(entities, &entity)
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return entities, nil
}

func DeleteNsqTopicEntity(db *buntdb.DB, topic string) error {
	id := entityPrefix + "nsq_topic:" + topic
	return db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(id)
		return err
	})
}

// ListNsqTopicEntitiesByGroup returns all nsq topic entities owned by the given group. If group is "root", returns all topics.
func ListNsqTopicEntitiesByGroup(db *buntdb.DB, group string) ([]*acl.Entity, error) {
	var entities []*acl.Entity
	var firstErr error
	prefix := entityPrefix + "nsq_topic:"

	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(prefix+"*", func(key, value string) bool {
			var entity acl.Entity
			if err := msgpack.Unmarshal([]byte(value), &entity); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return false // stop the iteration
			}
			if entity.GroupOwner == group {
				entities = append(entities, &entity)
			}
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return entities, nil
}
