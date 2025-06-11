package entity

import (
	"errors"
	"time"

	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func ToggleBookmark(dbConn *buntdb.DB, entityID, entityType, userID string, bookmark bool) error {
	b := &entity.Bookmark{
		EntityID:   entityID,
		EntityType: entityType,
		UserID:     userID,
	}
	if bookmark {
		// Create mapping
		b.CreatedAt = time.Now()
		return db.Insert(dbConn, b)
	} else {
		// Delete mapping
		return db.DeleteByID[entity.Bookmark](dbConn, b.GetPrimaryKey(""))
	}
}

func IsBookmarked(dbConn *buntdb.DB, entityID, userID string) (bool, error) {
	b := &entity.Bookmark{
		EntityID: entityID,
		UserID:   userID,
	}
	key := b.GetPrimaryKey("")
	_, err := db.GetByID[entity.Bookmark](dbConn, key)
	if err != nil {
		if errors.Is(err, buntdb.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ListBookmarkedTopicIDsByUser(dbConn *buntdb.DB, userID, entityType string) ([]string, error) {
	bookmarks, err := db.SelectAll[entity.Bookmark](dbConn, "="+userID+":"+entityType, entity.IdxBookmark_UserID_EntityType)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(bookmarks))
	for _, b := range bookmarks {
		ids = append(ids, b.EntityID)
	}
	return ids, nil
}
