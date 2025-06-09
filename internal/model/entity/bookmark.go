package entity

import (
	"time"

	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type Bookmark struct {
	EntityID  string
	UserID    string
	CreatedAt time.Time
}

const (
	TableBookmark        = "bookmark"
	IdxBookmark_EntityID = TableBookmark + ":entityid"
	IdxBookmark_UserID   = TableBookmark + ":userid"
	IdxBookmark_EntUser  = TableBookmark + ":entuser"
)

func (b *Bookmark) GetPrimaryKey(id string) string {
	if id != "" {
		return id
	}
	return TableBookmark + ":" + b.EntityID + ":" + b.UserID
}

func (b Bookmark) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxBookmark_EntityID,
			Pattern: TableBookmark + ":*:entityid",
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxBookmark_UserID,
			Pattern: TableBookmark + ":*:userid",
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxBookmark_EntUser,
			Pattern: TableBookmark + ":*:entuser",
			Type:    buntdb.IndexString,
		},
	}
}

func (b Bookmark) GetIndexValues() map[string]string {
	return map[string]string{
		"entityid": b.EntityID,
		"userid":   b.UserID,
		"entuser":  b.EntityID + ":" + b.UserID,
	}
}
