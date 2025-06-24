package action

import (
	"context"
	"time"

	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type ClaimEntityInput struct {
	GroupName    string `json:"group_name"`
	EntityName   string `json:"entity_name"`
	EntityTypeID string `json:"entity_type_id"`
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

func (h *ClaimEntityHandler) HandleClaimEntity(ctx context.Context, req ClaimEntityInput) (ClaimEntityResponse, error) {
	ent, err := h.repo.FindEntityByNameAndTypeID(req.EntityName, req.EntityTypeID)
	if err != nil {
		return ClaimEntityResponse{Status: "error", Message: "Entity not found"}, err
	}

	ent.GroupOwner = req.GroupName
	ent.UpdatedAt = time.Now()
	if err := h.repo.UpdateEntity(ent); err != nil {
		return ClaimEntityResponse{Status: "error", Message: "Failed to update entity"}, err
	}

	return ClaimEntityResponse{Status: "success", Message: "Entity claimed successfully"}, nil
}

type iClaimEntityRepo interface {
	FindEntityByNameAndTypeID(name, typeID string) (*entity.Entity, error)
	UpdateEntity(ent *entity.Entity) error
}

// Concrete implementation of iClaimEntityRepo
// (You may move this to a separate file if needed)
type claimEntityRepo struct {
	db *buntdb.DB
}

func NewClaimEntityRepo(db *buntdb.DB) *claimEntityRepo {
	return &claimEntityRepo{db: db}
}

func (r *claimEntityRepo) FindEntityByNameAndTypeID(name, typeID string) (*entity.Entity, error) {
	pivot := typeID + ":" + name
	ent, err := db.SelectOne[entity.Entity](r.db, pivot, entity.IdxEntity_TypeName)
	if err != nil {
		return nil, err
	}
	return &ent, nil
}

func (r *claimEntityRepo) UpdateEntity(ent *entity.Entity) error {
	return db.Update(r.db, ent)
}
