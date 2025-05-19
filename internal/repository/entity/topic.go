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
