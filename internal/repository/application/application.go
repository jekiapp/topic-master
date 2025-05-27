package application

import (
	"encoding/json"

	"github.com/jekiapp/topic-master/internal/model/acl"
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

const permissionApplicationPrefix = "permission_application:"
const applicationAssignmentPrefix = "application_assignment:"

func CreateApplication(db *buntdb.DB, app acl.Application) error {
	return dbpkg.Insert(db, app)
}

func GetApplicationByUserAndPermission(db *buntdb.DB, userID, permissionID string) (acl.Application, error) {
	key := permissionApplicationPrefix + userID + ":" + permissionID
	var app acl.Application
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &app)
	})
	if err != nil {
		return acl.Application{}, err
	}
	return app, nil
}

func CreateApplicationAssignment(db *buntdb.DB, assignment acl.ApplicationAssignment) error {
	return dbpkg.Insert(db, assignment)
}
