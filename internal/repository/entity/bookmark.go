package entity

import (
	"errors"
	"time"

	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func ToggleBookmark(dbConn *buntdb.DB, entityID, userID string, bookmark bool) error {
	b := &entity.Bookmark{
		EntityID: entityID,
		UserID:   userID,
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
