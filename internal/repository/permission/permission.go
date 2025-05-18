package permission

import (
	"encoding/json"
	"fmt"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
)

const permissionPrefix = "permission:"

func CreatePermission(db *buntdb.DB, permission acl.Permission) error {
	key := permissionPrefix + permission.Name + ":" + permission.EntityID
	value, err := json.Marshal(permission)
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		_, replaced, err := tx.Set(key, string(value), nil)
		if err != nil {
			return err
		}
		if replaced {
			return fmt.Errorf("permission already exists")
		}
		return nil
	})
}

func GetPermission(db *buntdb.DB, name string, entityID string) (*acl.Permission, error) {
	key := permissionPrefix + name + ":" + entityID
	var permission acl.Permission
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &permission)
	})
	if err != nil {
		return nil, err
	}
	return &permission, nil
}
