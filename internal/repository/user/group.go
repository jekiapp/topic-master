package user

import (
	"fmt"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/jekiapp/nsqper/pkg/db"
	"github.com/tidwall/buntdb"
)

func GetAllGroups(dbConn *buntdb.DB) ([]acl.Group, error) {
	pivot := acl.Group{} // empty pivot to select all
	return db.SelectAll[acl.Group](dbConn, pivot, acl.IdxGroup_Name)
}

// DeleteGroupByID deletes a group by its ID
func DeleteGroupByID(dbConn *buntdb.DB, id string) error {
	if id == "" {
		return fmt.Errorf("missing group id")
	}
	group := acl.Group{ID: id}
	return db.DeleteByID(dbConn, group)
}

// UpdateGroup updates a group by its ID
func UpdateGroup(dbConn *buntdb.DB, group acl.Group) error {
	return db.Update(dbConn, group)
}
