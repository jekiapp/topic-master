package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/nsqper/internal/model/acl"
	repoapp "github.com/jekiapp/nsqper/internal/repository/application"
	"github.com/tidwall/buntdb"
)

type CreateApplicationRequest struct {
	UserID       string `json:"user_id"`
	PermissionID string `json:"permission_id"`
	Reason       string `json:"reason"`
}

type CreateApplicationResponse struct {
	Application acl.PermissionApplication
}

type iApplicationRepo interface {
	CreateApplication(app acl.PermissionApplication) error
	GetApplicationByUserAndPermission(userID, permissionID string) (*acl.PermissionApplication, error)
}

type applicationRepo struct {
	db *buntdb.DB
}

func (r *applicationRepo) CreateApplication(app acl.PermissionApplication) error {
	return repoapp.CreateApplication(r.db, app)
}

func (r *applicationRepo) GetApplicationByUserAndPermission(userID, permissionID string) (*acl.PermissionApplication, error) {
	return repoapp.GetApplicationByUserAndPermission(r.db, userID, permissionID)
}

type CreateApplicationUsecase struct {
	repo iApplicationRepo
}

func NewCreateApplicationUsecase(db *buntdb.DB) CreateApplicationUsecase {
	return CreateApplicationUsecase{
		repo: &applicationRepo{db: db},
	}
}

func (uc CreateApplicationUsecase) Handle(ctx context.Context, req CreateApplicationRequest) (CreateApplicationResponse, error) {
	if req.UserID == "" {
		return CreateApplicationResponse{}, errors.New("missing required field: user_id")
	}
	if req.PermissionID == "" {
		return CreateApplicationResponse{}, errors.New("missing required field: permission_id")
	}
	if req.Reason == "" {
		return CreateApplicationResponse{}, errors.New("missing required field: reason")
	}
	// Check if application already exists
	existing, err := uc.repo.GetApplicationByUserAndPermission(req.UserID, req.PermissionID)
	if err == nil && existing != nil && existing.Status == "pending" {
		return CreateApplicationResponse{}, errors.New("application already exists for this user and permission and is pending")
	}
	app := acl.PermissionApplication{
		ID:           uuid.NewString(),
		UserID:       req.UserID,
		PermissionID: req.PermissionID,
		Reason:       req.Reason,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.repo.CreateApplication(app); err != nil {
		return CreateApplicationResponse{}, err
	}
	return CreateApplicationResponse{Application: app}, nil
}
