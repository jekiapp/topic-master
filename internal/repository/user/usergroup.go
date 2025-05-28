package user

import (
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
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

func GetUserGroup(dbConn *buntdb.DB, userID, groupID string) (acl.UserGroup, error) {
	pivot := groupID + ":" + userID
	return db.SelectOne[acl.UserGroup](dbConn, pivot, acl.IdxUserGroup_ID)
}

func CreateGroup(dbConn *buntdb.DB, group acl.Group) error {
	return db.Insert(dbConn, group)
}

func GetGroupByID(dbConn *buntdb.DB, id string) (acl.Group, error) {
	return db.GetByID[acl.Group](dbConn, id)
}

func GetGroupByName(dbConn *buntdb.DB, name string) (acl.Group, error) {
	return db.SelectOne[acl.Group](dbConn, name, acl.IdxGroup_Name)
}

func ListUserGroupsByGroupID(dbConn *buntdb.DB, groupID string, limit int) ([]acl.UserGroup, error) {
	all, err := db.SelectPaginated[acl.UserGroup](dbConn, groupID, acl.IdxUserGroup_GroupID, &db.Pagination{Limit: limit})
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(all) > limit {
		return all[:limit], nil
	}
	return all, nil
}

func GetAdminUserIDsByGroupID(dbConn *buntdb.DB, groupID string) ([]string, error) {
	userGroups, err := ListUserGroupsByGroupID(dbConn, groupID, 0)
	if err != nil {
		return nil, err
	}
	var adminIDs []string
	for _, userGroup := range userGroups {
		if userGroup.Role == acl.RoleGroupAdmin {
			adminIDs = append(adminIDs, userGroup.UserID)
		}
	}
	return adminIDs, nil
}

func ListUserGroupsByUserID(dbConn *buntdb.DB, userID string) ([]acl.UserGroup, error) {
	return db.SelectAll[acl.UserGroup](dbConn, userID, acl.IdxUserGroup_UserID)
}
