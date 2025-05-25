package user

import (
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/jekiapp/nsqper/pkg/db"
	"github.com/tidwall/buntdb"
)

func InitIndexGroup(db *buntdb.DB) error {
	indexes := acl.Group{}.GetIndexes()
	for _, index := range indexes {
		err := db.CreateIndex(index.Name, index.Pattern, index.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitIndexUserGroup(db *buntdb.DB) error {
	indexes := acl.UserGroup{}.GetIndexes()
	for _, index := range indexes {
		err := db.CreateIndex(index.Name, index.Pattern, index.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateUserGroup(dbConn *buntdb.DB, userGroup acl.UserGroup) error {
	return db.Insert(dbConn, userGroup)
}

func GetUserGroup(dbConn *buntdb.DB, userID, groupID string) (*acl.UserGroup, error) {
	tmp := acl.UserGroup{UserID: userID, GroupID: groupID}
	userGroup, err := db.SelectOne[acl.UserGroup](dbConn, tmp, acl.IdxUserGroup_ID)
	if err != nil {
		return nil, err
	}
	return &userGroup, nil
}

func CreateGroup(dbConn *buntdb.DB, group acl.Group) error {
	return db.Insert(dbConn, group)
}

func GetGroupByID(dbConn *buntdb.DB, id string) (*acl.Group, error) {
	tmp := acl.Group{ID: id}
	group, err := db.GetByID[acl.Group](dbConn, tmp)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func GetGroupByName(dbConn *buntdb.DB, name string) (acl.Group, error) {
	tmp := acl.Group{Name: name}
	group, err := db.SelectOne[acl.Group](dbConn, tmp, acl.IdxGroup_Name)
	if err != nil {
		return acl.Group{}, err
	}
	return group, nil
}

func ListUserGroupsByGroupID(dbConn *buntdb.DB, groupID string) ([]acl.UserGroup, error) {
	pivot := acl.UserGroup{GroupID: groupID}
	return db.SelectAll[acl.UserGroup](dbConn, pivot, acl.IdxUserGroup_GroupID)
}

func GetAdminUserIDsByGroupID(dbConn *buntdb.DB, groupID string) ([]string, error) {
	userGroups, err := ListUserGroupsByGroupID(dbConn, groupID)
	if err != nil {
		return nil, err
	}
	var adminIDs []string
	for _, userGroup := range userGroups {
		if userGroup.Type == "admin" {
			adminIDs = append(adminIDs, userGroup.UserID)
		}
	}
	return adminIDs, nil
}
