package user

import (
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

func CreateUser(db *buntdb.DB, user acl.User) error {
	return db.Update(func(tx *buntdb.Tx) error {
		key := "user:" + user.ID
		msgpackValue, err := msgpack.Marshal(user)
		if err != nil {
			return err
		}
		value := string(msgpackValue)
		_, _, err = tx.Set(key, value, nil)
		return err
	})
}

func GetUserByID(db *buntdb.DB, id string) (*acl.User, error) {
	var user acl.User
	key := "user:" + id
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		return msgpack.Unmarshal([]byte(val), &user)
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(db *buntdb.DB, user acl.User) error {
	return db.Update(func(tx *buntdb.Tx) error {
		key := "user:" + user.ID
		msgpackValue, err := msgpack.Marshal(user)
		if err != nil {
			return err
		}
		value := string(msgpackValue)
		_, _, err = tx.Set(key, value, nil)
		return err
	})
}

func GetUserByUsername(db *buntdb.DB, username string) (*acl.User, error) {
	var user acl.User
	found := false
	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys("user:*", func(key, value string) bool {
			if err := msgpack.Unmarshal([]byte(value), &user); err == nil && user.Username == username {
				found = true
				return false // found
			}
			return true // keep searching
		})
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &user, nil
}

// ListGroupsForUser fetches all groups for a user and returns []acl.GroupRole
func ListGroupsForUser(db *buntdb.DB, userID, userType string) ([]acl.GroupRole, error) {
	var groups []acl.GroupRole
	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys("usergroup:"+userID+":*", func(key, value string) bool {
			var ug acl.UserGroup
			if err := msgpack.Unmarshal([]byte(value), &ug); err != nil {
				return true
			}
			// Get group name
			group, err := GetGroupByName(db, ug.GroupID)
			groupName := ug.GroupID
			if err == nil && group != nil {
				groupName = group.Name
			}
			groups = append(groups, acl.GroupRole{GroupName: groupName, Role: userType})
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	// If user has no groups, still add their type as a role with empty group
	if len(groups) == 0 {
		groups = append(groups, acl.GroupRole{GroupName: "", Role: userType})
	}
	return groups, nil
}
