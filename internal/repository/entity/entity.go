package entity

import (
	"fmt"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
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
func GetEntityByID(dbConn *buntdb.DB, id string) (*acl.Entity, error) {
	entity, err := db.GetByID[acl.Entity](dbConn, id)
	if err != nil {
		return nil, err
	}
	if entity.ID != id {
		return nil, fmt.Errorf("entity not found")
	}
	return &entity, nil
}
