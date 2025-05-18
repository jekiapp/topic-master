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
	GetAdminUserIDsByGroupID(groupID string) ([]string, error)
	GetGroupByName(name string) (*acl.Group, error)
	ListUserIDsByGroupID(groupID string) ([]string, error)
	CreateApplicationAssignment(assignment acl.ApplicationAssignment) error
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

func (r *applicationRepo) GetAdminUserIDsByGroupID(groupID string) ([]string, error) {
	return userrepo.GetAdminUserIDsByGroupID(r.db, groupID)
}

func (r *applicationRepo) GetGroupByName(name string) (*acl.Group, error) {
	return userrepo.GetGroupByName(r.db, name)
}

func (r *applicationRepo) ListUserIDsByGroupID(groupID string) ([]string, error) {
	return userrepo.ListUserIDsByGroupID(r.db, groupID)
}

func (r *applicationRepo) CreateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return apprepo.CreateApplicationAssignment(r.db, assignment)
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
	// 1. Validate input
	if req.UserID == "" {
		return CreateApplicationResponse{}, errors.New("missing required field: user_id")
	}
	if req.PermissionID == "" {
		return CreateApplicationResponse{}, errors.New("missing required field: permission_id")
	}
	if req.Reason == "" {
		return CreateApplicationResponse{}, errors.New("missing required field: reason")
	}

	// 2. Check for duplicate pending application
	existing, err := uc.repo.GetApplicationByUserAndPermission(req.UserID, req.PermissionID)
	if err == nil && existing != nil && existing.Status == "pending" {
		return CreateApplicationResponse{}, errors.New("application already exists for this user and permission and is pending")
	}

	// 3. Create the application
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

	// 4. Lookup permission, entity, and group owner for assignment
	permission, err := permissionrepo.GetPermissionByID(uc.db, req.PermissionID)
	if err != nil {
		return CreateApplicationResponse{}, errors.New("permission not found")
	}
	entity, err := repository.GetEntityByID(uc.db, permission.EntityID)
	if err != nil {
		return CreateApplicationResponse{}, errors.New("entity not found")
	}
	groupID := entity.GroupOwner

	// 5. Get admin user IDs in the group
	adminIDs, err := uc.repo.GetAdminUserIDsByGroupID(groupID)
	if err != nil {
		return CreateApplicationResponse{}, errors.New("failed to list admin group members")
	}

	// 6. Get all user IDs in the root group
	rootGroup, err := uc.repo.GetGroupByName("root")
	if err != nil {
		return CreateApplicationResponse{}, errors.New("root group not found")
	}
	rootUserIDs, err := uc.repo.ListUserIDsByGroupID(rootGroup.ID)
	if err != nil {
		return CreateApplicationResponse{}, errors.New("failed to list root group members")
	}
	adminIDs = append(adminIDs, rootUserIDs...)

	// 7. Create assignment for each admin/root user
	for _, adminID := range adminIDs {
		assignment := acl.ApplicationAssignment{
			ID:            uuid.NewString(),
			ApplicationID: app.ID,
			ReviewerID:    adminID,
			ReviewStatus:  "pending",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_ = uc.repo.CreateApplicationAssignment(assignment)
	}

	return CreateApplicationResponse{Application: app}, nil
}
