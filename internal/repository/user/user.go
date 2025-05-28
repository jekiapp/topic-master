package user

import (
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func InitIndexUser(db *buntdb.DB) error {
	indexes := acl.User{}.GetIndexes()
	for _, index := range indexes {
		err := db.CreateIndex(index.Name, index.Pattern, index.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateUser(dbConn *buntdb.DB, user acl.User) error {
	return db.Insert(dbConn, user)
}

func GetUserByID(dbConn *buntdb.DB, id string) (acl.User, error) {
	return db.GetByID[acl.User](dbConn, id)
}

func UpdateUser(dbConn *buntdb.DB, user acl.User) error {
	return db.Update(dbConn, user)
}

func UpsertUser(dbConn *buntdb.DB, user acl.User) error {
	return db.Upsert(dbConn, user)
}

func GetUserByUsername(dbConn *buntdb.DB, username string) (acl.User, error) {
	return db.SelectOne[acl.User](dbConn, username, acl.IdxUser_Username)
}

// ListGroupsForUser fetches all groups for a user and returns []acl.GroupRole
func ListGroupsForUser(dbConn *buntdb.DB, userID string) ([]acl.GroupRole, error) {
	// Fetch all user groups for the user
	userGroups, err := db.SelectAll[acl.UserGroup](dbConn, userID, acl.IdxUserGroup_UserID)
	if err != nil {
		return nil, err
	}
	var groups []acl.GroupRole
	for _, ug := range userGroups {
		group, err := GetGroupByID(dbConn, ug.GroupID)
		groupName := ug.GroupID
		if err == nil {
			groupName = group.Name
		}
		groups = append(groups, acl.GroupRole{GroupID: ug.GroupID, GroupName: groupName, Role: ug.Role})
	}
	return groups, nil
}

func GetAllUsers(dbConn *buntdb.DB) ([]acl.User, error) {
	return db.SelectAll[acl.User](dbConn, "", acl.IdxUser_Username)
}
