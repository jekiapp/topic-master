package acl

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jekiapp/nsqper/internal/model/acl"
	permissionrepo "github.com/jekiapp/nsqper/internal/repository/permission"
	"github.com/tidwall/buntdb"
)

type CreatePermissionRequest struct {
	Actions  []string `json:"actions"` // format: ["publish","tail","delete"]
	EntityID string   `json:"entity_id"`
	Type     string   `json:"type"`
	Reason   string   `json:"reason"`
}

type CreatePermissionResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type iPermissionRepo interface {
	CreatePermission(permission acl.Permission) error
	GetPermission(name string, entityID string) (*acl.Permission, error)
}

type createPermissionRepo struct {
	db *buntdb.DB
}

func (r *createPermissionRepo) CreatePermission(permission acl.Permission) error {
	return permissionrepo.CreatePermission(r.db, permission)
}

func (r *createPermissionRepo) GetPermission(name string, entityID string) (*acl.Permission, error) {
	return permissionrepo.GetPermission(r.db, name, entityID)
}

type CreatePermissionUsecase struct {
	repo iPermissionRepo
}

func NewCreatePermissionUsecase(db *buntdb.DB) CreatePermissionUsecase {
	return CreatePermissionUsecase{
		repo: &createPermissionRepo{db: db},
	}
}

func (uc CreatePermissionUsecase) Handle(ctx context.Context, req CreatePermissionRequest) (CreatePermissionResponse, error) {
	if len(req.Actions) == 0 {
		return CreatePermissionResponse{}, errors.New("missing required field: actions")
	}
	if req.EntityID == "" {
		return CreatePermissionResponse{}, errors.New("missing required field: entity_id")
	}

	for _, action := range req.Actions {
		// Check if permission already exists
		existingPermission, err := uc.repo.GetPermission(action, req.EntityID)
		if err == nil && existingPermission != nil {
			log.Printf("permission already exists: %s:%s", action, req.EntityID)
			continue
		}
		permission := acl.Permission{
			Name:      action,
			EntityID:  req.EntityID,
			Type:      req.Type,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := uc.repo.CreatePermission(permission); err != nil {
			return CreatePermissionResponse{Status: "error", Error: err.Error()}, err
		}
	}

	return CreatePermissionResponse{Status: "success"}, nil
}
