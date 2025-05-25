package user

import (
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/jekiapp/nsqper/pkg/db"
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
	tmp := acl.User{ID: id}
	user, err := db.GetByID[acl.User](dbConn, tmp)
	if err != nil {
		return acl.User{}, err
	}
	return user, nil
}

func UpdateUser(dbConn *buntdb.DB, user acl.User) error {
	return db.Update(dbConn, user)
}

func UpsertUser(dbConn *buntdb.DB, user acl.User) error {
	return db.Upsert(dbConn, user)
}

func GetUserByUsername(dbConn *buntdb.DB, username string) (acl.User, error) {
	tmp := acl.User{Username: username}
	user, err := db.SelectOne[acl.User](dbConn, tmp, acl.IdxUser_Username)
	if err != nil {
		return acl.User{}, err
	}
	return user, nil
}

// ListGroupsForUser fetches all groups for a user and returns []acl.GroupRole
func ListGroupsForUser(dbConn *buntdb.DB, userID, userType string) ([]acl.GroupRole, error) {
	// Fetch all user groups for the user
	ugPivot := acl.UserGroup{UserID: userID}
	userGroups, err := db.SelectAll[acl.UserGroup](dbConn, ugPivot, acl.IdxUserGroup_UserID)
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
		groups = append(groups, acl.GroupRole{GroupName: groupName, Role: userType})
	}
	if len(groups) == 0 {
		groups = append(groups, acl.GroupRole{GroupName: "", Role: userType})
	}
	return groups, nil
}

func GetAllUsers(dbConn *buntdb.DB) ([]acl.User, error) {
	pivot := acl.User{} // empty pivot to select all
	return db.SelectAll[acl.User](dbConn, pivot, acl.IdxUser_Username)
}
