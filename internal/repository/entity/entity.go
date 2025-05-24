package entity

import (
	"fmt"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

const entityPrefix = "entity:"

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
