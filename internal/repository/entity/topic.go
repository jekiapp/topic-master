package entity

import (
	"strings"
	"time"

	"github.com/jekiapp/nsqper/internal/model"
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

func CreateNsqTopicEntity(db *buntdb.DB, topic string) (*acl.Entity, error) {
	entity := &acl.Entity{
		TypeID:    "nsq_topic",
		Name:      topic,
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
		_, _, err := tx.Set(entity.GetPrimaryKey(), string(data), nil)
		//set index
		for key, value := range entity.GetIndexValues() {
			if value == "" {
				continue
			}
			tx.Set(key, value, nil)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func GetNsqTopicEntity(db *buntdb.DB, topic string) (*acl.Entity, error) {
	var entity acl.Entity
	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(acl.IdxEntity_Name, topic, func(key, value string) bool {
			foundKey := strings.TrimSuffix(key, ":name")
			val, err := tx.Get(foundKey)
			if err != nil {
				return false
			}
			err = msgpack.Unmarshal([]byte(val), &entity)
			if err != nil {
				return false
			}
			return false
		})
	})
	if err != nil {
		return nil, model.ErrNotFound
	}
	return &entity, nil
}

func GetAllNsqTopicEntities(db *buntdb.DB) ([]*acl.Entity, error) {
	var entities []*acl.Entity
	var firstErr error

	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(acl.IdxEntity_TypeID, "nsq_topic", func(key, value string) bool {
			var entity acl.Entity
			foundKey := strings.TrimSuffix(key, ":typeid")
			val, err := tx.Get(foundKey)
			if err != nil {
				return false
			}
			err = msgpack.Unmarshal([]byte(val), &entity)
			if err != nil {
				return false
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

	groupName := group + ":nsq_topic"

	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(acl.IdxEntity_GroupName, groupName, func(key, value string) bool {
			var entity acl.Entity
			foundKey := strings.TrimSuffix(key, ":group_name")
			val, err := tx.Get(foundKey)
			if err != nil {
				return false
			}
			if err := msgpack.Unmarshal([]byte(val), &entity); err != nil {
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
