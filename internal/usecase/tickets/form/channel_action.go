package form

import (
	"context"
	"errors"

	usergrouplogic "github.com/jekiapp/topic-master/internal/logic/user_group"
	"github.com/jekiapp/topic-master/internal/model/acl"
	entityRepo "github.com/jekiapp/topic-master/internal/repository/entity"
	userRepo "github.com/jekiapp/topic-master/internal/repository/user"
	util "github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ChannelActionUsecase struct {
	db *buntdb.DB
}

func NewFormChannelActionUsecase(db *buntdb.DB) ChannelActionUsecase {
	return ChannelActionUsecase{
		db: db,
	}
}

func (uc ChannelActionUsecase) getChannelForm(ctx context.Context, entityID string) (NewApplicationResponse, error) {
	userInfo := util.GetUserInfo(ctx)
	if userInfo == nil {
		return NewApplicationResponse{}, errors.New("user not found")
	}

	// get entity from db
	channelEntity, err := entityRepo.GetEntityByID(uc.db, entityID)
	if err != nil {
		return NewApplicationResponse{}, err
	}

	group, err := userRepo.GetGroupByName(uc.db, channelEntity.GroupOwner)
	if err != nil {
		return NewApplicationResponse{}, err
	}

	// get admins for assignee list (reviewers)
	adminIDs, err := usergrouplogic.GetReviewerIDsByGroupID(uc.db, group.ID)
	if err != nil {
		return NewApplicationResponse{}, err
	}
	var reviewers []reviewerResponse
	for _, adminID := range adminIDs {
		user, err := userRepo.GetUserByID(uc.db, adminID)
		if err == nil {
			reviewers = append(reviewers, reviewerResponse{Username: user.Username, Name: user.Name})
		}
	}

	// hardcoded permissions
	permissions := acl.ChannelActionPermissions

	fields := []fieldResponse{
		{Label: "Channel Name", Type: "label", DefaultValue: channelEntity.Name, Editable: false},
		{Label: "Channel Owner", Type: "label", DefaultValue: channelEntity.GroupOwner, Editable: false},
		{Label: "Reason", Type: "textarea", DefaultValue: "", Editable: true},
	}

	return NewApplicationResponse{
		Title:       "Channel Action Form",
		Applicant:   applicantResponse{Username: userInfo.Username, Name: userInfo.Name},
		Type:        acl.ApplicationType_ChannelForm,
		Reviewers:   reviewers,
		Fields:      fields,
		Permissions: permissions,
	}, nil
}
