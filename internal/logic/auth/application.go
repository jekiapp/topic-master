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

// iApproveApplicationRepo abstracts the repo methods needed for approval logic
// (should be implemented by the repo used in usecase)
type iApproveApplicationRepo interface {
	GetApplicationByID(id string) (acl.Application, error)
	UpdateApplication(app acl.Application) error
	UpdateApplicationAssignment(assignment acl.ApplicationAssignment) error
	CreateApplicationHistory(history acl.ApplicationHistory) error
}

// ApproveApplication marks the application as completed, updates assignments, and adds history
func ApproveApplication(
	ctx context.Context,
	repo iApproveApplicationRepo,
	appID string,
	assignments []acl.ApplicationAssignment,
	comment string,
) error {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return errors.New("unauthorized")
	}
	app, err := repo.GetApplicationByID(appID)
	if err != nil {
		return err
	}
	app.Status = acl.StatusCompleted
	app.UpdatedAt = time.Now()
	if err := repo.UpdateApplication(app); err != nil {
		return err
	}
	for _, assignment := range assignments {
		if assignment.ReviewerID == user.ID {
			assignment.ReviewStatus = acl.ReviewStatusApproved
			assignment.ReviewedAt = time.Now()
		} else {
			assignment.ReviewStatus = acl.ReviewStatusPassed
		}
		assignment.UpdatedAt = time.Now()
		repo.UpdateApplicationAssignment(assignment)
	}
	history := acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: appID,
		Action:        acl.ActionApprove,
		ActorID:       user.ID,
		Comment:       comment,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := repo.CreateApplicationHistory(history); err != nil {
		// log but do not fail
		log.Println("failed to create application history", err)
	}
	return nil
}

// RejectApplication marks the application as completed, updates assignments as rejected, and adds history
func RejectApplication(
	ctx context.Context,
	repo iApproveApplicationRepo,
	appID string,
	assignments []acl.ApplicationAssignment,
	comment string,
) error {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return errors.New("unauthorized")
	}
	app, err := repo.GetApplicationByID(appID)
	if err != nil {
		return err
	}
	app.Status = acl.StatusCompleted
	app.UpdatedAt = time.Now()
	if err := repo.UpdateApplication(app); err != nil {
		return err
	}
	for _, assignment := range assignments {
		if assignment.ReviewerID == user.ID {
			assignment.ReviewStatus = acl.ReviewStatusRejected
			assignment.ReviewedAt = time.Now()
		} else {
			assignment.ReviewStatus = acl.ReviewStatusPassed
		}
		assignment.UpdatedAt = time.Now()
		repo.UpdateApplicationAssignment(assignment)
	}
	history := acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: appID,
		Action:        acl.ActionReject,
		ActorID:       user.ID,
		Comment:       comment,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := repo.CreateApplicationHistory(history); err != nil {
		// log but do not fail
		log.Println("failed to create application history", err)
	}
	return nil
}

/*

 */
