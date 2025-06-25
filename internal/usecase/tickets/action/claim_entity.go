package action

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ClaimEntityInput struct {
	Action      string   `json:"action"`
	AppID       string   `json:"app_id"`
	EntityID    string   `json:"entity_id"`
	GroupName   string   `json:"group_name"`
	AssigneeIDs []string `json:"assignee_ids"`
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
func (h *ClaimEntityHandler) HandleApprove(ctx context.Context, appID, entityID, groupName string, assigneeIDs []string) error {
	ent, err := h.repo.GetEntityByID(entityID)
	if err != nil {
		return err
	}
	ent.GroupOwner = groupName
	ent.Status = entity.EntityStatus_Active
	ent.UpdatedAt = time.Now()
	if err := h.repo.UpdateEntity(ent); err != nil {
		return err
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
		return errors.New("unauthorized")
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
		Comment:       ent.TypeID + " claim approved",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	return nil
}

// Reject claim entity
// Coordinator must pass all assigneeIDs; eligibility is already checked in coordinator
func (h *ClaimEntityHandler) HandleReject(ctx context.Context, appID, entityID string, assigneeIDs []string) error {
	ent, err := h.repo.GetEntityByID(entityID)
	if err != nil {
		return err
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
		return errors.New("unauthorized")
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
		Comment:       ent.TypeID + " claim rejected",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	return nil
}

// Existing claim handler (for direct claim)
func (h *ClaimEntityHandler) HandleClaimEntity(ctx context.Context, req ClaimEntityInput) error {
	if req.Action == acl.ActionApprove {
		return h.HandleApprove(ctx, req.AppID, req.EntityID, req.GroupName, req.AssigneeIDs)
	} else if req.Action == acl.ActionReject {
		return h.HandleReject(ctx, req.AppID, req.EntityID, req.AssigneeIDs)
	}
	return nil
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
