package application

import (
	"encoding/json"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
)

// ListApplicationsByUserID returns all PermissionApplications for a given userID.
func ListApplicationsByUserID(db *buntdb.DB, userID string) ([]acl.PermissionApplication, error) {
	var apps []acl.PermissionApplication
	prefix := permissionApplicationPrefix + userID + ":"
	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(prefix+"*", func(key, value string) bool {
			var app acl.PermissionApplication
			if err := json.Unmarshal([]byte(value), &app); err == nil {
				apps = append(apps, app)
			}
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	return apps, nil
}
