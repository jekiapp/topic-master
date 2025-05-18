package user

import (
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

func CreateUserGroup(db *buntdb.DB, userGroup acl.UserGroup) error {
	return db.Update(func(tx *buntdb.Tx) error {
		key := "usergroup:" + userGroup.UserID + ":" + userGroup.GroupID
		msgpackValue, err := msgpack.Marshal(userGroup)
		if err != nil {
			return err
		}
		value := string(msgpackValue)
		_, _, err = tx.Set(key, value, nil)
		return err
	})
}

func GetUserGroup(db *buntdb.DB, userID, groupID string) (*acl.UserGroup, error) {
	var userGroup acl.UserGroup
	key := "usergroup:" + userID + ":" + groupID
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
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
		key := "group:" + group.Name
		msgpackValue, err := msgpack.Marshal(group)
		if err != nil {
			return err
		}
		value := string(msgpackValue)
		_, _, err = tx.Set(key, value, nil)
		return err
	})
}

func GetGroupByName(db *buntdb.DB, name string) (*acl.Group, error) {
	var group acl.Group
	key := "group:" + name
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
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
