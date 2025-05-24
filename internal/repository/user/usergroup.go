package user

import (
	"fmt"
	"log"
	"strings"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

func CreateUserGroup(db *buntdb.DB, userGroup acl.UserGroup) error {
	return db.Update(func(tx *buntdb.Tx) error {
		key := userGroup.GetPrimaryKey()
		msgpackValue, err := msgpack.Marshal(userGroup)
		if err != nil {
			return err
		}
		// set index
		for name, value := range userGroup.GetIndexValues() {
			_, _, err = tx.Set(key+":"+name, value, nil)
			if err != nil {
				log.Printf("error setting index: %s", err)
			}
		}
		value := string(msgpackValue)
		_, _, err = tx.Set(key, value, nil)
		return err
	})
}

func GetUserGroup(db *buntdb.DB, userID, groupID string) (*acl.UserGroup, error) {
	var userGroup acl.UserGroup
	id := fmt.Sprintf("%s:%s", groupID, userID)
	err := db.View(func(tx *buntdb.Tx) error {
		var foundKey string
		err := tx.AscendEqual(acl.IdxUserGroup_ID, id, func(key, value string) bool {
			if value == id {
				foundKey = strings.TrimSuffix(key, ":usergroup_id")
				return false
			}
			return true
		})
		if err != nil {
			return err
		}
		val, err := tx.Get(foundKey)
		if err != nil {
			return err
		}
		return msgpack.Unmarshal([]byte(val), &userGroup)
	})
	if err != nil {
		return nil, err
	}
	return &userGroup, nil
}

func CreateGroup(db *buntdb.DB, group acl.Group) error {
	return db.Update(func(tx *buntdb.Tx) error {
		key := group.GetPrimaryKey()
		msgpackValue, err := msgpack.Marshal(group)
		if err != nil {
			return err
		}
		value := string(msgpackValue)
		_, _, err = tx.Set(key, value, nil)
		return err
	})
}

// get group by id
func GetGroupByID(db *buntdb.DB, id string) (*acl.Group, error) {
	var group = acl.Group{ID: id}
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(group.GetPrimaryKey())
		if err != nil {
			return err
		}
		return msgpack.Unmarshal([]byte(val), &group)
	})
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// get group by name
func GetGroupByName(db *buntdb.DB, name string) (*acl.Group, error) {
	var group acl.Group
	err := db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendEqual(acl.IdxGroup_Name, name, func(key, value string) bool {
			if value != name {
				return true
			}

			foundKey := strings.TrimSuffix(key, ":name")
			val, err := tx.Get(foundKey)
			if err != nil {
				log.Printf("error getting group: %s", err)
				return false
			}
			err = msgpack.Unmarshal([]byte(val), &group)
			if err != nil {
				log.Printf("error unmarshalling group: %s", err)
			}
			return false
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func ListUserGroupsByGroupID(db *buntdb.DB, groupID string) ([]acl.UserGroup, error) {
	var userGroups []acl.UserGroup
	err := db.View(func(tx *buntdb.Tx) error {
		err := tx.AscendEqual(acl.IdxUserGroup_GroupID, groupID, func(key, value string) bool {
			foundKey := strings.TrimSuffix(key, ":group_id")
			val, err := tx.Get(foundKey)
			if err != nil {
				log.Printf("error getting user group: %s", err)
				return false
			}
			var userGroup acl.UserGroup
			err = msgpack.Unmarshal([]byte(val), &userGroup)
			if err != nil {
				log.Printf("error unmarshalling user group: %s", err)
				return false
			}
			userGroups = append(userGroups, userGroup)
			return true
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return userGroups, nil
}

func GetAdminUserIDsByGroupID(db *buntdb.DB, groupID string) ([]string, error) {
	userGroups, err := ListUserGroupsByGroupID(db, groupID)
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
