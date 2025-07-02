package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

//go:generate mockgen -source=signup.go -destination=mock/mock_signup_repo.go -package=user_mock

type SignupRequest struct {
	Username        string `json:"username"`
	Name            string `json:"name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Reason          string `json:"reason"`
	GroupID         string `json:"group_id"`
	GroupName       string `json:"group_name"`
	GroupRole       string `json:"group_role"`
}

type SignupResponse struct {
	ApplicationID string `json:"application_id"`
}

type SignupUsecase struct {
	repo ISignupRepo
}

func NewSignupUsecase(db *buntdb.DB) SignupUsecase {
	return SignupUsecase{
		repo: &signupRepo{db: db},
	}
}

func (uc SignupUsecase) Validate(r SignupRequest) error {
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
	if len(r.Password) < acl.MinPasswordLength {
		return errors.New("password must be at least " + strconv.Itoa(acl.MinPasswordLength) + " characters long")
	}
	if r.GroupID == "" {
		return errors.New("missing group_id")
	}
	if r.GroupRole == "" {
		return errors.New("missing group_role")
	}

	if r.GroupName == acl.GroupRoot && r.GroupRole == acl.RoleGroupAdmin {
		return errors.New("cannot apply to become admin of root group")
	}

	user, err := uc.repo.GetUserByUsername(r.Username)
	if err != nil && err != buntdb.ErrNotFound {
		return errors.New("failed to get user by username: " + err.Error())
	}
	if user.ID != "" {
		return errors.New("username already exists")
	}

	return nil
}

func (uc SignupUsecase) Handle(ctx context.Context, req SignupRequest) (SignupResponse, error) {
	if err := uc.Validate(req); err != nil {
		return SignupResponse{}, err
	}

	userID := uuid.NewString()
	app := &acl.Application{
		ID:            uuid.NewString(),
		Title:         fmt.Sprintf("Signup request by %s (%s)", req.Name, req.Username),
		UserID:        userID,
		Type:          acl.ApplicationType_Signup,
		PermissionIDs: []string{acl.Permission_Signup_User.Name},
		Reason:        fmt.Sprintf("Request to become %s of group %s", req.GroupRole, req.GroupName),
		Status:        acl.StatusWaitingForApproval,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := uc.repo.CreateApplication(*app); err != nil {
		return SignupResponse{}, err
	}

	rootGroup, err := uc.repo.GetGroupByName(acl.GroupRoot)
	if err != nil {
		return SignupResponse{}, errors.New("root group not found")
	}

	rootMembers, err := uc.repo.ListUserGroupsByGroupID(rootGroup.ID, 0)
	if err != nil {
		return SignupResponse{}, errors.New("failed to list root group members")
	}

	adminUserIDs := []string{}
	if req.GroupName != acl.GroupRoot {
		adminUserIDs, err = uc.repo.GetAdminUserIDsByGroupID(req.GroupID)
		if err != nil && err != buntdb.ErrNotFound {
			return SignupResponse{}, errors.New("failed to get admin user ids: " + err.Error())
		}
	}

	for _, member := range rootMembers {
		adminUserIDs = append(adminUserIDs, member.UserID)
	}

	if len(adminUserIDs) == 0 {
		return SignupResponse{}, errors.New("no active reviewers found: no group admin and no root group members")
	}

	hasActiveReviewer := false
	for _, userID := range adminUserIDs {
		user, err := uc.repo.GetUserByID(userID)
		if err != nil {
			log.Printf("failed to get user by id %s: %v", userID, err)
			continue
		}
		if user.Status != acl.StatusUserActive {
			continue
		}

		hasActiveReviewer = true
		assignment := &acl.ApplicationAssignment{
			ID:            uuid.NewString(),
			ApplicationID: app.ID,
			ReviewerID:    userID,
			ReviewStatus:  acl.ActionWaitingForApproval,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := uc.repo.CreateApplicationAssignment(*assignment); err != nil {
			return SignupResponse{}, fmt.Errorf("failed to create application assignment: %w", err)
		}
	}

	if !hasActiveReviewer {
		return SignupResponse{}, errors.New("no active reviewers found")
	}

	hash := sha256.Sum256([]byte(req.Password))
	hashedPassword := hex.EncodeToString(hash[:])
	user := acl.UserPending{
		User: acl.User{
			ID:       userID,
			Username: req.Username,
			Name:     req.Name,
			Password: hashedPassword,
			Status:   acl.StatusUserInApproval,
			Groups: []acl.GroupRole{
				{
					GroupID:   req.GroupID,
					GroupName: req.GroupName,
					Role:      req.GroupRole,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	if err := uc.repo.CreateUserPending(user); err != nil {
		return SignupResponse{}, fmt.Errorf("failed to create user pending: %w", err)
	}
	userGroup := acl.UserGroup{
		ID:        uuid.NewString(),
		UserID:    userID,
		GroupID:   req.GroupID,
		Role:      req.GroupRole,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.CreateUserGroup(userGroup); err != nil {
		return SignupResponse{}, fmt.Errorf("failed to create user group: %w", err)
	}
	history := &acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		Action:        "Create ticket",
		ActorID:       userID,
		Comment:       fmt.Sprintf("Initial signup by %s", user.Name),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_ = uc.repo.CreateApplicationHistory(*history)
	return SignupResponse{ApplicationID: app.ID}, nil
}

// --- ISignupRepo interface and implementation ---
type ISignupRepo interface {
	CreateApplication(app acl.Application) error
	GetGroupByName(name string) (acl.Group, error)
	ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error)
	GetAdminUserIDsByGroupID(groupID string) ([]string, error)
	CreateApplicationAssignment(assignment acl.ApplicationAssignment) error
	CreateApplicationHistory(history acl.ApplicationHistory) error
	GetUserByID(userID string) (acl.User, error)
	CreateUserPending(user acl.UserPending) error
	CreateUserGroup(userGroup acl.UserGroup) error
	GetUserByUsername(username string) (acl.User, error)
}

type signupRepo struct {
	db *buntdb.DB
}

func (r *signupRepo) CreateApplication(app acl.Application) error {
	return db.Insert(r.db, &app)
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
	return db.Insert(r.db, &assignment)
}

func (r *signupRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return db.Insert(r.db, &history)
}

func (r *signupRepo) GetUserByID(userID string) (acl.User, error) {
	return db.GetByID[acl.User](r.db, userID)
}

func (r *signupRepo) CreateUserPending(user acl.UserPending) error {
	return db.Insert(r.db, &user)
}

func (r *signupRepo) CreateUserGroup(userGroup acl.UserGroup) error {
	return db.Insert(r.db, &userGroup)
}

func (r *signupRepo) GetUserByUsername(username string) (acl.User, error) {
	return db.SelectOne[acl.User](r.db, username, acl.IdxUser_Username)
}
