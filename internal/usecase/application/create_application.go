package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/nsqper/internal/model/acl"
	repository "github.com/jekiapp/nsqper/internal/repository"
	apprepo "github.com/jekiapp/nsqper/internal/repository/application"
	permissionrepo "github.com/jekiapp/nsqper/internal/repository/permission"
	userrepo "github.com/jekiapp/nsqper/internal/repository/user"
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
	return apprepo.CreateApplication(r.db, app)
}

func (r *applicationRepo) GetApplicationByUserAndPermission(userID, permissionID string) (*acl.PermissionApplication, error) {
	return apprepo.GetApplicationByUserAndPermission(r.db, userID, permissionID)
}

type CreateApplicationUsecase struct {
	repo iApplicationRepo
	db   *buntdb.DB
}

func NewCreateApplicationUsecase(db *buntdb.DB) CreateApplicationUsecase {
	return CreateApplicationUsecase{
		repo: &applicationRepo{db: db},
		db:   db,
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

	// by permission id, get the entity, then check the group owner of this entity
	permission, err := permissionrepo.GetPermissionByID(uc.db, req.PermissionID)
	if err != nil {
		return CreateApplicationResponse{}, errors.New("permission not found")
	}
	entity, err := repository.GetEntityByID(uc.db, permission.EntityID)
	if err != nil {
		return CreateApplicationResponse{}, errors.New("entity not found")
	}
	groupID := entity.GroupOwner
	userIDs, err := userrepo.ListUserIDsByGroupID(uc.db, groupID)
	if err != nil {
		return CreateApplicationResponse{}, errors.New("failed to list group members")
	}
	for _, userID := range userIDs {
		user, err := userrepo.GetUserByID(uc.db, userID)
		if err != nil || user == nil {
			continue
		}
		if user.Type == "admin" {
			assignment := acl.ApplicationAssignment{
				ID:            uuid.NewString(),
				ApplicationID: app.ID,
				ReviewerID:    user.ID,
				ReviewStatus:  "pending",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			_ = apprepo.CreateApplicationAssignment(uc.db, assignment)
		}
	}

	return CreateApplicationResponse{Application: app}, nil
}
