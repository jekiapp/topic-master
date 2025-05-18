package repository

import (
	"encoding/json"
	"fmt"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
)

const entityPrefix = "entity:"

func GetEntityByID(db *buntdb.DB, id string) (*acl.Entity, error) {
	var entity acl.Entity
	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(entityPrefix+"*", func(key, value string) bool {
			if err := json.Unmarshal([]byte(value), &entity); err == nil && entity.ID == id {
				return false // found
			}
			return true // keep searching
		})
	})
	if err != nil {
		return nil, err
	}
	if entity.ID != id {
		return nil, fmt.Errorf("entity not found")
	}
	return &entity, nil
}
