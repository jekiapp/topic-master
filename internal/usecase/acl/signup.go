package acl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	Username        string `json:"username"`
	Name            string `json:"name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Reason          string `json:"reason"`
	GroupID         string `json:"group_id"`
	GroupType       string `json:"group_type"` // member or admin
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
	if r.Password == "" {
		return errors.New("missing password")
	}
	if r.ConfirmPassword == "" {
		return errors.New("missing confirm_password")
	}
	if r.Password != r.ConfirmPassword {
		return errors.New("password and confirm_password do not match")
	}
	if r.GroupID == "" {
		return errors.New("missing group_id")
	}
	if r.GroupType == "" {
		return errors.New("missing group_type")
	}

	return nil
}

func (uc SignupUsecase) Handle(ctx context.Context, req SignupRequest) (SignupResponse, error) {
	if err := req.Validate(); err != nil {
		return SignupResponse{}, err
	}
	app := acl.Application{
		Title:         fmt.Sprintf("Signup request by %s(%s)", req.Name, req.Username),
		UserID:        acl.ActorSystem,
		PermissionIDs: []string{"signup:" + req.Username},
		Reason:        fmt.Sprintf("Request to become %s to %s", req.GroupType, req.GroupID),
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
	// 2. Create a new user with status "in approval"
	// Hash the password (SHA256)
	hash := sha256.Sum256([]byte(req.Password))
	hashedPassword := hex.EncodeToString(hash[:])
	user := acl.User{
		ID:        uuid.NewString(),
		Username:  req.Username,
		Name:      req.Name,
		Password:  hashedPassword,
		Status:    acl.StatusUserInApproval,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := userrepo.CreateUser(uc.repo.(*signupRepo).db, user); err != nil {
		return SignupResponse{}, err
	}
	// Assign user to the requested group
	userGroup := acl.UserGroup{
		UserID:    user.ID,
		GroupID:   req.GroupID,
		Type:      req.GroupType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := userrepo.CreateUserGroup(uc.repo.(*signupRepo).db, userGroup); err != nil {
		return SignupResponse{}, err
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
