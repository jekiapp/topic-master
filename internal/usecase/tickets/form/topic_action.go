package form

import (
	"context"

	usergrouplogic "github.com/jekiapp/topic-master/internal/logic/user_group"
	entityRepo "github.com/jekiapp/topic-master/internal/repository/entity"
	userRepo "github.com/jekiapp/topic-master/internal/repository/user"
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
	// get entity from db
	topicEntity, err := entityRepo.GetEntityByID(uc.db, entityID)
	if err != nil {
		return NewApplicationResponse{}, err
	}

	// get admins for assignee list (reviewers)
	adminIDs, err := usergrouplogic.GetReviewerIDsByGroupID(uc.db, topicEntity.GroupOwner)
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
	permissions := []permissionResponse{
		{Name: "publish", Description: "Publish messages"},
		{Name: "tail", Description: "Tail messages"},
		{Name: "drain", Description: "Drain topic"},
		{Name: "pause", Description: "Pause topic"},
		{Name: "delete", Description: "Delete topic"},
	}

	// hardcoded fields
	fields := []fieldResponse{
		{Label: "Topic Name", Type: "text", Required: true, DefaultValue: topicEntity.Name, Editable: false},
		{Label: "Topic Description", Type: "textarea", Required: false, DefaultValue: topicEntity.Description, Editable: false},
		{Label: "Topic Owner", Type: "text", Required: true, DefaultValue: topicEntity.GroupOwner, Editable: false},
		{Label: "Reason", Type: "textarea", Required: true, DefaultValue: "", Editable: true},
	}

	return NewApplicationResponse{
		Title:       "Topic Action Form",
		Type:        TopicFormType,
		Reviewers:   reviewers,
		Fields:      fields,
		Permissions: permissions,
	}, nil
}
