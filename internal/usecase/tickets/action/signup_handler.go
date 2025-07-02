package action

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type iSignupRepo interface {
	GetApplicationByID(id string) (acl.Application, error)
	UpdateApplication(app acl.Application) error
	ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error)
	UpdateApplicationAssignment(assignment acl.ApplicationAssignment) error
	CreateApplicationHistory(history acl.ApplicationHistory) error
	GetUserPendingByID(userID string) (acl.UserPending, error)
	CreateUser(user acl.User) error
	UpdateUserPending(user acl.UserPending) error
	CreateUserGroup(userGroup acl.UserGroup) error
}

type SignupHandler struct {
	repo iSignupRepo
}

func NewSignupHandler(db *buntdb.DB) *SignupHandler {
	return &SignupHandler{repo: &signupRepo{db: db}}
}

func (h *SignupHandler) HandleSignup(ctx context.Context, req ActionRequest) (ActionResponse, error) {
	switch req.Action {
	case acl.ActionApprove:
		return h.handleApprove(ctx, req, nil)
	case acl.ActionReject:
		return h.handleReject(ctx, req, nil)
	default:
		return ActionResponse{Status: "error", Message: fmt.Sprintf("Invalid action: %s", req.Action)}, nil
	}
}

// Coordinator must pass all assigneeIDs; eligibility and assignment fetching is already checked in coordinator
func (h *SignupHandler) handleApprove(ctx context.Context, req ActionRequest, assigneeIDs []string) (ActionResponse, error) {
	// validate eligibility
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ActionResponse{Status: "error", Message: "Unauthorized"}, nil
	}
	// 1. Get application
	app, err := h.repo.GetApplicationByID(req.ApplicationID)
	if err != nil {
		return ActionResponse{Status: "error", Message: "Application not found"}, err
	}

	// 3. Update assignments using provided assigneeIDs
	for _, reviewerID := range assigneeIDs {
		assignment := acl.ApplicationAssignment{
			ApplicationID: app.ID,
			ReviewerID:    reviewerID,
			UpdatedAt:     time.Now(),
		}
		if reviewerID == user.ID {
			assignment.ReviewStatus = acl.ReviewStatusApproved
			assignment.ReviewedAt = time.Now()
		} else {
			assignment.ReviewStatus = acl.ReviewStatusPassed
		}
		if err := h.repo.UpdateApplicationAssignment(assignment); err != nil {
			log.Println("Failed to update assignment", err)
		}
	}

	// 5. Add application history
	history := acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		Action:        acl.ActionApprove,
		ActorID:       user.ID,
		Comment:       "Signup approved",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := h.repo.CreateApplicationHistory(history); err != nil {
		log.Println("Failed to create application history", err)
	}

	// 6. Create user from pending user
	applicant, err := h.repo.GetUserPendingByID(app.UserID)
	if err != nil {
		return ActionResponse{Status: "error", Message: "Failed to get user pending"}, err
	}

	applicant.Status = acl.UserStatusActive
	applicant.CreatedAt = time.Now()
	applicant.UpdatedAt = time.Now()
	if err = h.repo.CreateUser(applicant.User); err != nil {
		return ActionResponse{Status: "error", Message: "Failed to create user"}, err
	}

	groupID := app.MetaData["group_id"]
	groupRole := app.MetaData["group_role"]

	userGroup := acl.UserGroup{
		ID:        uuid.NewString(),
		UserID:    applicant.User.ID,
		GroupID:   groupID,
		Role:      groupRole,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.repo.CreateUserGroup(userGroup); err != nil {
		return ActionResponse{Status: "error", Message: "Failed to create user group"}, err
	}

	// Mark application as completed
	app.Status = acl.StatusCompleted
	app.UpdatedAt = time.Now()
	if err := h.repo.UpdateApplication(app); err != nil {
		return ActionResponse{Status: "error", Message: "Failed to update application"}, err
	}

	return ActionResponse{Status: "success", Message: "Signup completed"}, nil
}

// Coordinator must pass all assigneeIDs; eligibility and assignment fetching is already checked in coordinator
func (h *SignupHandler) handleReject(ctx context.Context, req ActionRequest, assigneeIDs []string) (ActionResponse, error) {
	// validate eligibility
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ActionResponse{Status: "error", Message: "Unauthorized"}, nil
	}
	// 1. Get application by id
	app, err := h.repo.GetApplicationByID(req.ApplicationID)
	if err != nil {
		return ActionResponse{Status: "error", Message: "Application not found"}, err
	}

	// Delete user pending by id
	applicant, err := h.repo.GetUserPendingByID(app.UserID)
	if err == nil {
		applicant.Status = acl.UserStatusInactive
		applicant.UpdatedAt = time.Now()
		if err = h.repo.UpdateUserPending(applicant); err != nil {
			return ActionResponse{Status: "error", Message: "Failed to update user"}, err
		}
	}

	// Update assignments using provided assigneeIDs
	for _, reviewerID := range assigneeIDs {
		assignment := acl.ApplicationAssignment{
			ApplicationID: app.ID,
			ReviewerID:    reviewerID,
			UpdatedAt:     time.Now(),
		}
		if reviewerID == user.ID {
			assignment.ReviewStatus = acl.ReviewStatusRejected
			assignment.ReviewedAt = time.Now()
		} else {
			assignment.ReviewStatus = acl.ReviewStatusPassed
		}
		if err := h.repo.UpdateApplicationAssignment(assignment); err != nil {
			log.Println("Failed to update assignment", err)
		}
	}

	// Add application history
	history := acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		Action:        acl.ActionReject,
		ActorID:       user.ID,
		Comment:       "Signup rejected",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := h.repo.CreateApplicationHistory(history); err != nil {
		log.Println("Failed to create application history", err)
	}

	// Mark application as completed
	app.Status = acl.StatusCompleted
	app.UpdatedAt = time.Now()
	if err := h.repo.UpdateApplication(app); err != nil {
		return ActionResponse{Status: "error", Message: "Failed to update application"}, err
	}

	return ActionResponse{Status: "success", Message: "Signup rejected and user deleted"}, nil
}

type signupRepo struct {
	db *buntdb.DB
}

func (r *signupRepo) GetApplicationByID(id string) (acl.Application, error) {
	return db.GetByID[acl.Application](r.db, id)
}

func (r *signupRepo) UpdateApplication(app acl.Application) error {
	return db.Update(r.db, &app)
}

func (r *signupRepo) ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error) {
	return db.SelectAll[acl.ApplicationAssignment](r.db, "="+appID, acl.IdxAppAssign_ApplicationID)
}

func (r *signupRepo) UpdateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return db.Update(r.db, &assignment)
}

func (r *signupRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return db.Insert(r.db, &history)
}

func (r *signupRepo) GetUserPendingByID(userID string) (acl.UserPending, error) {
	return db.GetByID[acl.UserPending](r.db, userID)
}

func (r *signupRepo) CreateUser(user acl.User) error {
	return db.Insert(r.db, &user)
}

func (r *signupRepo) UpdateUserPending(user acl.UserPending) error {
	return db.Update(r.db, &user)
}

func (r *signupRepo) CreateUserGroup(userGroup acl.UserGroup) error {
	return db.Insert(r.db, &userGroup)
}
