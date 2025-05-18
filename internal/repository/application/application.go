package application

import (
	"encoding/json"
	"fmt"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
)

const permissionApplicationPrefix = "permission_application:"

func CreateApplication(db *buntdb.DB, app acl.PermissionApplication) error {
	key := permissionApplicationPrefix + app.UserID + ":" + app.PermissionID
	value, err := json.Marshal(app)
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		_, replaced, err := tx.Set(key, string(value), nil)
		if err != nil {
			return err
		}
		if replaced {
			return fmt.Errorf("application already exists")
		}
		return nil
	})
}

func GetApplicationByUserAndPermission(db *buntdb.DB, userID, permissionID string) (*acl.PermissionApplication, error) {
	key := permissionApplicationPrefix + userID + ":" + permissionID
	var app acl.PermissionApplication
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &app)
	})
	if err != nil {
		return nil, err
	}
	return &app, nil
}
