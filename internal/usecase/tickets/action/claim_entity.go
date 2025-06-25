package action

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ClaimEntityInput struct {
	EntityID  string `json:"entity_id"`
	GroupName string `json:"group_name"`
}

type ClaimEntityResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ClaimEntityHandler struct {
	repo iClaimEntityRepo
}

func NewClaimEntityHandler(db *buntdb.DB) *ClaimEntityHandler {
	return &ClaimEntityHandler{repo: &claimEntityRepo{db: db}}
}

// Approve claim entity
// Coordinator must pass all assigneeIDs; eligibility is already checked in coordinator
func (h *ClaimEntityHandler) HandleApprove(ctx context.Context, entityID, groupName, appID string, assigneeIDs []string) (ClaimEntityResponse, error) {
	ent, err := h.repo.GetEntityByID(entityID)
	if err != nil {
		return ClaimEntityResponse{Status: "error", Message: "Entity not found"}, err
	}
	ent.GroupOwner = groupName
	ent.Status = entity.EntityStatus_Active
	ent.UpdatedAt = time.Now()
	if err := h.repo.UpdateEntity(ent); err != nil {
		return ClaimEntityResponse{Status: "error", Message: "Failed to update entity"}, err
	}
	// Mark application as completed
	app, err := h.repo.GetApplicationByID(appID)
	if err == nil {
		app.Status = acl.StatusCompleted
		app.UpdatedAt = time.Now()
		h.repo.UpdateApplication(app)
	}
	// Update assignments
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ClaimEntityResponse{Status: "error", Message: "Unauthorized"}, nil
	}
	for _, reviewerID := range assigneeIDs {
		assignment := acl.ApplicationAssignment{
			ApplicationID: appID,
			ReviewerID:    reviewerID,
			UpdatedAt:     time.Now(),
		}
		if reviewerID == user.ID {
			assignment.ReviewStatus = acl.ReviewStatusApproved
			assignment.ReviewedAt = time.Now()
		} else {
			assignment.ReviewStatus = acl.ReviewStatusPassed
		}
		h.repo.UpdateApplicationAssignment(assignment)
	}
	// Add history
	h.repo.CreateApplicationHistory(acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: appID,
		Action:        acl.ActionApprove,
		ActorID:       user.ID,
		Comment:       "Entity claim approved",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	return ClaimEntityResponse{Status: "success", Message: "Entity claim approved"}, nil
}

// Reject claim entity
// Coordinator must pass all assigneeIDs; eligibility is already checked in coordinator
func (h *ClaimEntityHandler) HandleReject(ctx context.Context, entityID, appID string, assigneeIDs []string) (ClaimEntityResponse, error) {
	ent, err := h.repo.GetEntityByID(entityID)
	if err != nil {
		return ClaimEntityResponse{Status: "error", Message: "Entity not found"}, err
	}
	ent.Status = entity.EntityStatus_Deleted
	ent.UpdatedAt = time.Now()
	if err := h.repo.UpdateEntity(ent); err != nil {
		return ClaimEntityResponse{Status: "error", Message: "Failed to update entity"}, err
	}
	// Mark application as completed
	app, err := h.repo.GetApplicationByID(appID)
	if err == nil {
		app.Status = acl.StatusCompleted
		app.UpdatedAt = time.Now()
		h.repo.UpdateApplication(app)
	}
	// Update assignments
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ClaimEntityResponse{Status: "error", Message: "Unauthorized"}, nil
	}
	for _, reviewerID := range assigneeIDs {
		assignment := acl.ApplicationAssignment{
			ApplicationID: appID,
			ReviewerID:    reviewerID,
			UpdatedAt:     time.Now(),
		}
		if reviewerID == user.ID {
			assignment.ReviewStatus = acl.ReviewStatusRejected
			assignment.ReviewedAt = time.Now()
		} else {
			assignment.ReviewStatus = acl.ReviewStatusPassed
		}
		h.repo.UpdateApplicationAssignment(assignment)
	}
	// Add history
	h.repo.CreateApplicationHistory(acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: appID,
		Action:        acl.ActionReject,
		ActorID:       user.ID,
		Comment:       "Entity claim rejected",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	return ClaimEntityResponse{Status: "success", Message: "Entity claim rejected"}, nil
}

// Existing claim handler (for direct claim)
func (h *ClaimEntityHandler) HandleClaimEntity(ctx context.Context, req ClaimEntityInput) (ActionResponse, error) {
	ent, err := h.repo.GetEntityByID(req.EntityID)
	if err != nil {
		return ActionResponse{Status: "error", Message: "Entity not found"}, err
	}

	ent.GroupOwner = req.GroupName
	ent.UpdatedAt = time.Now()
	if err := h.repo.UpdateEntity(ent); err != nil {
		return ActionResponse{Status: "error", Message: "Failed to update entity"}, err
	}

	return ActionResponse{Status: "success", Message: "Entity claimed successfully"}, nil
}

type iClaimEntityRepo interface {
	GetEntityByID(id string) (*entity.Entity, error)
	UpdateEntity(ent *entity.Entity) error
	CreateApplicationHistory(history acl.ApplicationHistory) error
	UpdateApplication(app acl.Application) error
	GetApplicationByID(id string) (acl.Application, error)
	UpdateApplicationAssignment(assignment acl.ApplicationAssignment) error
}

// Concrete implementation of iClaimEntityRepo
// (You may move this to a separate file if needed)
type claimEntityRepo struct {
	db *buntdb.DB
}

func NewClaimEntityRepo(db *buntdb.DB) *claimEntityRepo {
	return &claimEntityRepo{db: db}
}

func (r *claimEntityRepo) GetEntityByID(id string) (*entity.Entity, error) {
	ent, err := db.GetByID[entity.Entity](r.db, id)
	if err != nil {
		return nil, err
	}
	return &ent, nil
}

func (r *claimEntityRepo) UpdateEntity(ent *entity.Entity) error {
	return db.Update(r.db, ent)
}

func (r *claimEntityRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return db.Insert(r.db, &history)
}

func (r *claimEntityRepo) UpdateApplication(app acl.Application) error {
	return db.Update(r.db, &app)
}

func (r *claimEntityRepo) GetApplicationByID(id string) (acl.Application, error) {
	return db.GetByID[acl.Application](r.db, id)
}

func (r *claimEntityRepo) UpdateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return db.Update(r.db, &assignment)
}
