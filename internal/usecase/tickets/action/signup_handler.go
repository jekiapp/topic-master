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
		return h.handleApprove(ctx, req)
	case acl.ActionReject:
		return h.handleReject(ctx, req)
	default:
		return ActionResponse{Status: "error", Message: fmt.Sprintf("Invalid action: %s", req.Action)}, nil
	}
}

func (h *SignupHandler) handleApprove(ctx context.Context, req ActionRequest) (ActionResponse, error) {
	// 1. Get application
	app, err := h.repo.GetApplicationByID(req.ApplicationID)
	if err != nil {
		return ActionResponse{Status: "error", Message: "Application not found"}, err
	}

	// 2. Mark application as completed
	app.Status = acl.StatusCompleted
	app.UpdatedAt = time.Now()
	if err := h.repo.UpdateApplication(app); err != nil {
		return ActionResponse{Status: "error", Message: "Failed to update application"}, err
	}

	// 3. Get all assignments for this application
	assignments, err := h.repo.ListAssignmentsByApplicationID(app.ID)
	if err != nil {
		return ActionResponse{Status: "error", Message: "Failed to get assignments"}, err
	}

	// 4. Get current user from context
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ActionResponse{Status: "error", Message: "Unauthorized"}, nil
	}

	for i, assignment := range assignments {
		if assignment.ReviewerID == user.ID {
			assignments[i].ReviewStatus = acl.ReviewStatusApproved
			assignments[i].ReviewedAt = time.Now()
			assignments[i].UpdatedAt = time.Now()
		} else {
			assignments[i].ReviewStatus = acl.ReviewStatusPassed
			assignments[i].UpdatedAt = time.Now()
		}
		if err := h.repo.UpdateApplicationAssignment(assignments[i]); err != nil {
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
	if err == nil {
		applicant.Status = acl.UserStatusActive
		applicant.CreatedAt = time.Now()
		applicant.UpdatedAt = time.Now()
		if err = h.repo.CreateUser(applicant.User); err != nil {
			return ActionResponse{Status: "error", Message: "Failed to create user"}, err
		}
	}

	return ActionResponse{Status: "success", Message: "Signup completed"}, nil
}

func (h *SignupHandler) handleReject(ctx context.Context, req ActionRequest) (ActionResponse, error) {
	// 1. Get application by id
	app, err := h.repo.GetApplicationByID(req.ApplicationID)
	if err != nil {
		return ActionResponse{Status: "error", Message: "Application not found"}, err
	}

	// 2. Mark application as completed
	app.Status = acl.StatusCompleted
	app.UpdatedAt = time.Now()
	if err := h.repo.UpdateApplication(app); err != nil {
		return ActionResponse{Status: "error", Message: "Failed to update application"}, err
	}

	// 3. Get all assignments for this application
	assignments, err := h.repo.ListAssignmentsByApplicationID(app.ID)
	if err != nil {
		return ActionResponse{Status: "error", Message: "Failed to get assignments"}, err
	}

	// 4. Get current user from context
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ActionResponse{Status: "error", Message: "Unauthorized"}, nil
	}

	for i, assignment := range assignments {
		if assignment.ReviewerID == user.ID {
			assignments[i].ReviewStatus = acl.ReviewStatusRejected
			assignments[i].ReviewedAt = time.Now()
			assignments[i].UpdatedAt = time.Now()
		} else {
			assignments[i].ReviewStatus = acl.ReviewStatusPassed
			assignments[i].UpdatedAt = time.Now()
		}
		if err := h.repo.UpdateApplicationAssignment(assignments[i]); err != nil {
			log.Println("Failed to update assignment", err)
		}
	}

	// 5. Add application history
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

	// 6. Delete user by id
	// update user pending status to inactive
	applicant, err := h.repo.GetUserPendingByID(app.UserID)
	if err == nil {
		applicant.Status = acl.UserStatusInactive
		applicant.UpdatedAt = time.Now()
		if err = h.repo.UpdateUserPending(applicant); err != nil {
			return ActionResponse{Status: "error", Message: "Failed to update user"}, err
		}
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
	return db.Update(r.db, app)
}

func (r *signupRepo) ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error) {
	return db.SelectAll[acl.ApplicationAssignment](r.db, "="+appID, acl.IdxAppAssign_ApplicationID)
}

func (r *signupRepo) UpdateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return db.Update(r.db, assignment)
}

func (r *signupRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return db.Insert(r.db, history)
}

func (r *signupRepo) GetUserPendingByID(userID string) (acl.UserPending, error) {
	return db.GetByID[acl.UserPending](r.db, userID)
}

func (r *signupRepo) CreateUser(user acl.User) error {
	return db.Insert(r.db, user)
}

func (r *signupRepo) UpdateUserPending(user acl.UserPending) error {
	return db.Update(r.db, &user)
}
