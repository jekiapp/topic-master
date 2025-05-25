package user

import (
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/jekiapp/nsqper/pkg/db"
	"github.com/tidwall/buntdb"
)

func GetAllGroups(dbConn *buntdb.DB) ([]acl.Group, error) {
	pivot := acl.Group{} // empty pivot to select all
	return db.SelectAll[acl.Group](dbConn, pivot, acl.IdxGroup_Name)
}
