package usergroup

import (
	"errors"

	"github.com/jekiapp/topic-master/internal/model/acl"
	userRepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/tidwall/buntdb"
)

func GetReviewerIDsByGroupID(db *buntdb.DB, groupID string) ([]string, error) {
	group, err := userRepo.GetGroupByID(db, groupID)
	if err != nil {
		return nil, err
	}

	adminIDs, err := userRepo.GetAdminUserIDsByGroupID(db, group.ID)
	if err != nil {
		return nil, err
	}
	if len(adminIDs) > 0 {
		return adminIDs, nil
	}

	rootGroup, err := userRepo.GetGroupByName(db, acl.GroupRoot)
	if err != nil {
		return nil, errors.New("root group not found")
	}

	rootMembers, err := userRepo.ListUserGroupsByGroupID(db, rootGroup.ID, 0)
	if err != nil {
		return nil, errors.New("failed to list root group members")
	}

	var reviewers []string
	for _, member := range rootMembers {
		reviewers = append(reviewers, member.UserID)
	}
	return reviewers, nil
}
