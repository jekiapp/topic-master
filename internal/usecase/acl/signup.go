// this is a signup usecase
// it will receive a request from signup/new/script.js
// it will create a new application, using the model/acl/application.go Application struct
// select all the member of root group, then each one of them will create a new applicationAssignment
// then select all the admin member of the requested group, then each one of them will create a new applicationAssignment
// then create a new record of ApplicationHistory as "waiting for approval"

package acl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type SignupRequest struct {
	Username  string `json:"username"`
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	GroupID   string `json:"group_id"`
	GroupType string `json:"group_type"` // member or admin
}

type SignupResponse struct {
	Application acl.Application
}

type iSignupRepo interface {
	CreateApplication(app acl.Application) error
	GetGroupByName(name string) (acl.Group, error)
	ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error)
	GetAdminUserIDsByGroupID(groupID string) ([]string, error)
	CreateApplicationAssignment(assignment acl.ApplicationAssignment) error
	CreateApplicationHistory(history acl.ApplicationHistory) error
}

type signupRepo struct {
	db *buntdb.DB
}

func (r *signupRepo) CreateApplication(app acl.Application) error {
	return db.Insert(r.db, app)
}

func (r *signupRepo) GetGroupByName(name string) (acl.Group, error) {
	return userrepo.GetGroupByName(r.db, name)
}

func (r *signupRepo) ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error) {
	return userrepo.ListUserGroupsByGroupID(r.db, groupID, limit)
}

func (r *signupRepo) GetAdminUserIDsByGroupID(groupID string) ([]string, error) {
	return userrepo.GetAdminUserIDsByGroupID(r.db, groupID)
}

func (r *signupRepo) CreateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return db.Insert(r.db, assignment)
}

func (r *signupRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return db.Insert(r.db, history)
}

type SignupUsecase struct {
	repo iSignupRepo
}

func NewSignupUsecase(db *buntdb.DB) SignupUsecase {
	return SignupUsecase{
		repo: &signupRepo{db: db},
	}
}

func (r SignupRequest) Validate() error {
	if r.Username == "" {
		return errors.New("missing username")
	}
	if r.Name == "" {
		return errors.New("missing name")
	}
	if r.GroupID == "" {
		return errors.New("missing group_id")
	}
	if r.GroupType == "" {
		return errors.New("missing group_type")
	}
	if r.GroupType != "member" && r.GroupType != "admin" {
		return errors.New("invalid group_type: must be 'member' or 'admin'")
	}
	return nil
}

func (uc SignupUsecase) Handle(ctx context.Context, req SignupRequest) (SignupResponse, error) {
	if err := req.Validate(); err != nil {
		return SignupResponse{}, err
	}
	app := acl.Application{
		ID:            uuid.NewString(),
		Title:         fmt.Sprintf("Signup request by %s(%s)", req.Name, req.Username),
		UserID:        acl.ActorSystem,
		PermissionIDs: []string{"signup:" + req.Username},
		Reason:        req.Reason,
		Status:        acl.StatusWaitingForApproval,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := uc.repo.CreateApplication(app); err != nil {
		return SignupResponse{}, err
	}
	// 1. Select all members of root group
	rootGroup, err := uc.repo.GetGroupByName(acl.GroupRoot)
	if err != nil {
		return SignupResponse{}, errors.New("root group not found")
	}
	rootMembers, err := uc.repo.ListUserGroupsByGroupID(rootGroup.ID, 0)
	if err != nil {
		return SignupResponse{}, errors.New("failed to list root group members")
	}
	for _, member := range rootMembers {
		assignment := acl.ApplicationAssignment{
			ID:            uuid.NewString(),
			ApplicationID: app.ID,
			ReviewerID:    member.UserID,
			ReviewStatus:  acl.StatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := uc.repo.CreateApplicationAssignment(assignment); err != nil {
			return SignupResponse{}, err
		}
	}
	// 2. Select all admin members of the requested group
	adminIDs, err := uc.repo.GetAdminUserIDsByGroupID(req.GroupID)
	if err != nil {
		return SignupResponse{}, errors.New("failed to list admin group members")
	}
	for _, adminID := range adminIDs {
		assignment := acl.ApplicationAssignment{
			ID:            uuid.NewString(),
			ApplicationID: app.ID,
			ReviewerID:    adminID,
			ReviewStatus:  acl.StatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_ = uc.repo.CreateApplicationAssignment(assignment)
	}
	// 3. Create ApplicationHistory as "waiting for approval"
	history := acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		Action:        acl.ActionWaitingForApproval,
		ActorID:       acl.ActorSystem,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_ = uc.repo.CreateApplicationHistory(history)
	return SignupResponse{Application: app}, nil
}
