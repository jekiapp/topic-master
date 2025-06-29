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

type TopicActionUsecase struct {
	db *buntdb.DB
}

func NewFormTopicActionUsecase(db *buntdb.DB) TopicActionUsecase {
	return TopicActionUsecase{
		db: db,
	}
}

func (uc TopicActionUsecase) getTopicForm(ctx context.Context, entityID string) (NewApplicationResponse, error) {
	userInfo := util.GetUserInfo(ctx)
	if userInfo == nil {
		return NewApplicationResponse{}, errors.New("user not found")
	}

	// get entity from db
	topicEntity, err := entityRepo.GetEntityByID(uc.db, entityID)
	if err != nil {
		return NewApplicationResponse{}, err
	}

	group, err := userRepo.GetGroupByName(uc.db, topicEntity.GroupOwner)
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
	permissions := []acl.Permission{
		acl.Permission_Topic_Publish,
		acl.Permission_Topic_Tail,
		acl.Permission_Topic_Drain,
		acl.Permission_Topic_Pause,
		acl.Permission_Topic_Delete,
	}

	// hardcoded fields
	fields := []fieldResponse{
		{Label: "Topic Name", Type: "label", DefaultValue: topicEntity.Name, Editable: false},
		{Label: "Topic Description", Type: "label-multiline", DefaultValue: topicEntity.Description, Editable: false},
		{Label: "Topic Owner", Type: "label", DefaultValue: topicEntity.GroupOwner, Editable: false},
		{Label: "Reason", Type: "textarea", DefaultValue: "", Editable: true},
	}

	return NewApplicationResponse{
		Title:       "Topic Action Form",
		Applicant:   applicantResponse{Username: userInfo.Username, Name: userInfo.Name},
		Type:        acl.ApplicationType_TopicForm,
		Reviewers:   reviewers,
		Fields:      fields,
		Permissions: permissions,
	}, nil
}
