package acl

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/nsqper/internal/model/acl"
	permissionrepo "github.com/jekiapp/nsqper/internal/repository/permission"
	"github.com/tidwall/buntdb"
)

type CreatePermissionRequest struct {
	Name     string `json:"name"`
	EntityID string `json:"entity_id"`
}

type CreatePermissionResponse struct {
	Permission acl.Permission
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
	if req.Name == "" {
		return CreatePermissionResponse{}, errors.New("missing required field: name")
	}
	if req.EntityID == "" {
		return CreatePermissionResponse{}, errors.New("missing required field: entity_id")
	}
	// Check if permission already exists
	existingPermission, err := uc.repo.GetPermission(req.Name, req.EntityID)
	if err == nil && existingPermission != nil {
		return CreatePermissionResponse{}, errors.New("permission already exists")
	}
	permission := acl.Permission{
		ID:        uuid.NewString(),
		Name:      req.Name,
		EntityID:  req.EntityID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.CreatePermission(permission); err != nil {
		return CreatePermissionResponse{}, err
	}
	return CreatePermissionResponse{Permission: permission}, nil
}
