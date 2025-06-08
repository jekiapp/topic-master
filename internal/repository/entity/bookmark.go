package entity

import (
	"errors"
	"time"

	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func ToggleBookmark(dbConn *buntdb.DB, entityID, userID string, bookmark bool) error {
	primaryKey := entity.TableBookmark + ":" + entityID + ":" + userID
	if bookmark {
		// Create mapping
		b := entity.Bookmark{
			EntityID:  entityID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}
		return db.Insert(dbConn, b)
	} else {
		// Delete mapping
		return db.DeleteByID[entity.Bookmark](dbConn, primaryKey)
	}
}

func IsBookmarked(dbConn *buntdb.DB, entityID, userID string) (bool, error) {
	primaryKey := entity.TableBookmark + ":" + entityID + ":" + userID
	_, err := db.GetByID[entity.Bookmark](dbConn, primaryKey)
	if err != nil {
		if errors.Is(err, buntdb.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
