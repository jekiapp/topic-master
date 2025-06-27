package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/acl"
	util "github.com/jekiapp/topic-master/pkg/util"
)

type CreateApplicationInput struct {
	Title              string
	ApplicationType    string
	PermissionIDs      []string
	Reason             string
	ReviewerGroupID    string
	MetaData           map[string]string
	HistoryInitAction  string
	HistoryInitComment string
}

type CreateApplicationOutput struct {
	ApplicationID string
}

type iCreateApplicationRepo interface {
	CreateApplication(app acl.Application) error
	GetReviewerIDsByGroupID(groupID string) ([]string, error)
	CreateApplicationAssignment(assignment acl.ApplicationAssignment) error
	CreateApplicationHistory(history acl.ApplicationHistory) error
}

func CreateApplication(ctx context.Context, req CreateApplicationInput, repo iCreateApplicationRepo) (CreateApplicationOutput, error) {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return CreateApplicationOutput{}, errors.New("user not found in context")
	}
	// Create application
	app := acl.Application{
		ID:            uuid.NewString(),
		Title:         req.Title,
		UserID:        user.ID,
		Type:          req.ApplicationType,
		PermissionIDs: req.PermissionIDs,
		Reason:        req.Reason,
		Status:        acl.StatusWaitingForApproval,
		MetaData:      req.MetaData,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := repo.CreateApplication(app); err != nil {
		return CreateApplicationOutput{}, err
	}

	adminUserIDs, err := repo.GetReviewerIDsByGroupID(req.ReviewerGroupID)
	if err != nil {
		return CreateApplicationOutput{}, errors.New("failed to get admin user ids")
	}

	for _, userID := range adminUserIDs {
		assignment := acl.ApplicationAssignment{
			ID:            uuid.NewString(),
			ApplicationID: app.ID,
			ReviewerID:    userID,
			ReviewStatus:  acl.ActionWaitingForApproval,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := repo.CreateApplicationAssignment(assignment); err != nil {
			return CreateApplicationOutput{}, err
		}
	}
	// Insert application history
	history := acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		Action:        req.HistoryInitAction,
		ActorID:       user.ID,
		Comment:       req.HistoryInitComment,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = repo.CreateApplicationHistory(history)
	if err != nil {
		log.Println("failed to create application history", err)
	}

	return CreateApplicationOutput{ApplicationID: app.ID}, nil
}
