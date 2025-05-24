package entity

import (
	"fmt"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

func InitIndexEntity(db *buntdb.DB) error {
	indexes := acl.Entity{}.GetIndexes()
	for _, index := range indexes {
		err := db.CreateIndex(index.Name, index.Pattern, index.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

// format id = nsqtopic:topic_name
func GetEntityByID(db *buntdb.DB, id string) (*acl.Entity, error) {
	var entity = acl.Entity{
		ID: id,
	}
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(entity.GetPrimaryKey())
		if err != nil {
			return err
		}
		return msgpack.Unmarshal([]byte(val), &entity)
	})
	if err != nil {
		return nil, err
	}
	if entity.ID != id {
		return nil, fmt.Errorf("entity not found")
	}
	return &entity, nil
}
