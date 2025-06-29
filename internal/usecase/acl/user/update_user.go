//go:generate mockgen -source=update_user.go -destination=mock/mock_update_user_repo.go -package=user_mock
package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type UpdateUserRequest struct {
	Username      string `json:"username"`
	Name          string `json:"name"`
	ResetPassword bool   `json:"reset_password"`
	Groups        []struct {
		GroupID string `json:"group_id"`
		Role    string `json:"role"`
		Name    string `json:"name"`
	} `json:"groups"`
}

type UpdateUserResponse struct {
	Username          string `json:"username"`
	GeneratedPassword string `json:"generated_password"`
}

type iUpdateUserRepo interface {
	GetUserByID(userID string) (acl.User, error)
	GetUserByUsername(username string) (acl.User, error)
	UpdateUser(user acl.User) error
	ListUserGroupsByUserID(userID string) ([]acl.UserGroup, error)
	DeleteUserGroupsByUserID(userID string) error
	CreateUserGroup(userGroup acl.UserGroup) error
}

type UpdateUserUsecase struct {
	repo iUpdateUserRepo
}

func NewUpdateUserUsecase(db *buntdb.DB) UpdateUserUsecase {
	return UpdateUserUsecase{
		repo: &updateUserRepo{db: db},
	}
}

func validateUpdateUserRequest(req UpdateUserRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	if len(req.Groups) == 0 {
		return errors.New("at least one group is required")
	}
	rootCount := 0
	for i, g := range req.Groups {
		if g.GroupID == "" {
			return errors.New("group_id is required for group index " + strconv.Itoa(i))
		}
		if g.Role == "" {
			return errors.New("role is required for group index " + strconv.Itoa(i))
		}
		if g.Name == acl.GroupRoot {
			rootCount++
		}
	}
	if rootCount > 0 && len(req.Groups) > 1 {
		return errors.New("if a group with type 'root' is present, it must be the only group")
	}
	return nil
}

func (uc UpdateUserUsecase) Handle(ctx context.Context, req UpdateUserRequest) (UpdateUserResponse, error) {
	// Basic input validation
	if err := validateUpdateUserRequest(req); err != nil {
		return UpdateUserResponse{}, err
	}

	existingUser, err := uc.repo.GetUserByUsername(req.Username)
	if err != nil {
		return UpdateUserResponse{}, errors.New("user not found : " + req.Username)
	}

	// Update user fields
	updatedUser := existingUser
	updatedUser.Name = req.Name
	updatedUser.UpdatedAt = time.Now()

	var newPassword string
	// Update password if provided
	if req.ResetPassword {
		password, err := generateRandomPassword(12)
		if err != nil {
			return UpdateUserResponse{}, err
		}
		hash := sha256.Sum256([]byte(password))
		hashedPassword := hex.EncodeToString(hash[:])
		newPassword = password
		updatedUser.Password = hashedPassword
		updatedUser.Status = acl.StatusUserPending
	}

	// Update user in database
	if err := uc.repo.UpdateUser(updatedUser); err != nil {
		return UpdateUserResponse{}, err
	}

	// Update group mappings
	// First, delete all existing user group mappings
	if err := uc.repo.DeleteUserGroupsByUserID(existingUser.ID); err != nil {
		return UpdateUserResponse{}, err
	}

	// Create new user group mappings
	for _, group := range req.Groups {
		userGroup := acl.UserGroup{
			ID:        uuid.NewString(),
			UserID:    existingUser.ID,
			GroupID:   group.GroupID,
			Role:      group.Role,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := uc.repo.CreateUserGroup(userGroup); err != nil {
			return UpdateUserResponse{}, err
		}
	}

	return UpdateUserResponse{Username: req.Username, GeneratedPassword: newPassword}, nil
}

type updateUserRepo struct {
	db *buntdb.DB
}

func (r *updateUserRepo) GetUserByID(userID string) (acl.User, error) {
	return userrepo.GetUserByID(r.db, userID)
}

func (r *updateUserRepo) GetUserByUsername(username string) (acl.User, error) {
	return userrepo.GetUserByUsername(r.db, username)
}

func (r *updateUserRepo) UpdateUser(user acl.User) error {
	return userrepo.UpdateUser(r.db, user)
}

func (r *updateUserRepo) ListUserGroupsByUserID(userID string) ([]acl.UserGroup, error) {
	return userrepo.ListUserGroupsByUserID(r.db, userID)
}

func (r *updateUserRepo) DeleteUserGroupsByUserID(userID string) error {
	pivot := acl.UserGroup{UserID: userID}
	return db.DeleteByIndex(r.db, &pivot, acl.IdxUserGroup_UserID)
}

func (r *updateUserRepo) CreateUserGroup(userGroup acl.UserGroup) error {
	return userrepo.CreateUserGroup(r.db, userGroup)
}
