package application

import (
	"encoding/json"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/tidwall/buntdb"
)

// ListApplicationsByUserID returns all Applications for a given userID.
func ListApplicationsByUserID(db *buntdb.DB, userID string) ([]acl.Application, error) {
	var apps []acl.Application
	prefix := permissionApplicationPrefix + userID + ":"
	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(prefix+"*", func(key, value string) bool {
			var app acl.Application
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

// ListAssignmentsByReviewerID returns all ApplicationAssignments for a given reviewerID.
func ListAssignmentsByReviewerID(db *buntdb.DB, reviewerID string) ([]acl.ApplicationAssignment, error) {
	var assignments []acl.ApplicationAssignment
	prefix := "app_assign:"
	err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(prefix+"*", func(key, value string) bool {
			var assignment acl.ApplicationAssignment
			if err := json.Unmarshal([]byte(value), &assignment); err == nil {
				if assignment.ReviewerID == reviewerID {
					assignments = append(assignments, assignment)
				}
			}
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	return assignments, nil
}
