package entity

import (
	"errors"
	"fmt"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func InitIndexEntity(db *buntdb.DB) error {
	indexes := entity.Entity{}.GetIndexes()
	bookmarkIndexes := entity.Bookmark{}.GetIndexes()

	indexes = append(indexes, bookmarkIndexes...)
	for _, index := range indexes {
		err := db.CreateIndex(index.Name, index.Pattern, index.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

// format id = nsqtopic:topic_name
func GetEntityByID(dbConn *buntdb.DB, id string) (entity.Entity, error) {
	entityObj, err := db.GetByID[entity.Entity](dbConn, id)
	if err != nil {
		return entity.Entity{}, err
	}
	if entityObj.ID != id {
		return entity.Entity{}, fmt.Errorf("entity not found")
	}
	return entityObj, nil
}

// GetPermissionMapByActionEntityUser fetches a PermissionMap by action, entityID, and userID
func GetPermissionMapByActionEntityUser(dbConn *buntdb.DB, userID, entityID, action string) (acl.PermissionMap, error) {
	pivot := action + ":" + entityID + ":" + userID
	perm, err := db.SelectOne[acl.PermissionMap](dbConn, "="+pivot, acl.IdxPermissionMap_ActionEntityUser)
	if err != nil {
		return acl.PermissionMap{}, errors.New("permission not found")
	}
	return perm, nil
}
