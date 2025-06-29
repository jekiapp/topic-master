package action

import (
	"context"
	"errors"
	"time"

	"github.com/jekiapp/topic-master/internal/logic/auth"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type ClaimEntityInput struct {
	Application acl.Application
	Action      string
	Assignments []acl.ApplicationAssignment
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
func (req ClaimEntityInput) Validate() error {
	if req.Action == "" {
		return errors.New("missing action")
	}
	if req.Application.ID == "" {
		return errors.New("missing app_id")
	}
	if len(req.Assignments) == 0 {
		return errors.New("missing assignments")
	}
	return nil
}

// Existing claim handler (for direct claim)
func (h *ClaimEntityHandler) HandleClaimEntity(ctx context.Context, req ClaimEntityInput) (ActionResponse, error) {
	if err := req.Validate(); err != nil {
		return ActionResponse{}, err
	}

	switch req.Action {
	case acl.ActionApprove:
		err := h.HandleApprove(ctx, req)
		if err != nil {
			return ActionResponse{}, err
		}
		return ActionResponse{
			Status:  "success",
			Message: "Claim entity approved",
		}, nil
	case acl.ActionReject:
		err := h.HandleReject(ctx, req)
		if err != nil {
			return ActionResponse{}, err
		}
		return ActionResponse{
			Status:  "success",
			Message: "Claim entity rejected",
		}, nil
	}
	return ActionResponse{}, errors.New("invalid action")
}

// Approve claim entity
// Coordinator must pass all assigneeIDs; eligibility is already checked in coordinator
func (h *ClaimEntityHandler) HandleApprove(ctx context.Context, input ClaimEntityInput) error {

	entityID := input.Application.MetaData["entity_id"]
	ent, err := h.repo.GetEntityByID(entityID)
	if err != nil {
		return err
	}

	groupID := input.Application.MetaData["group_id"]
	group, err := h.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}
	ent.GroupOwner = group.Name
	ent.Status = entity.EntityStatus_Active
	ent.UpdatedAt = time.Now()
	if err := h.repo.UpdateEntity(ent); err != nil {
		return err
	}
	// Use reusable approval logic
	comment := ent.TypeID + " claim approved"
	err = auth.ApproveApplication(ctx, h.repo, input.Application.ID, input.Assignments, comment)
	if err != nil {
		return err
	}
	return nil
}

// Reject claim entity
// Coordinator must pass all assigneeIDs; eligibility is already checked in coordinator
func (h *ClaimEntityHandler) HandleReject(ctx context.Context, input ClaimEntityInput) error {
	entityID := input.Application.MetaData["entity_id"]
	ent, err := h.repo.GetEntityByID(entityID)
	if err != nil {
		return err
	}
	// Use reusable rejection logic
	comment := ent.TypeID + " claim rejected"
	err = auth.RejectApplication(ctx, h.repo, input.Application.ID, input.Assignments, comment)
	if err != nil {
		return err
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
	GetGroupByID(id string) (acl.Group, error)
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

func (r *claimEntityRepo) GetGroupByID(id string) (acl.Group, error) {
	return db.GetByID[acl.Group](r.db, id)
}
