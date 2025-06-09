package user

import (
	"fmt"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func GetAllGroups(dbConn *buntdb.DB) ([]acl.Group, error) {
	return db.SelectAll[acl.Group](dbConn, "*", acl.IdxGroup_Name)
}

// DeleteGroupByID deletes a group by its ID
func DeleteGroupByID(dbConn *buntdb.DB, id string) error {
	if id == "" {
		return fmt.Errorf("missing group id")
	}
	return db.DeleteByID[acl.Group](dbConn, id)
}

// UpdateGroup updates a group by its ID
func UpdateGroup(dbConn *buntdb.DB, group acl.Group) error {
	return db.Update(dbConn, &group)
}
