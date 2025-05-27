package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/tidwall/buntdb"
)

type ChangePasswordRequest struct {
	UserID      string `json:"user_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Success bool `json:"success"`
}

type iUserPasswordRepo interface {
	GetUserByID(userID string) (acl.User, error)
	UpdateUser(user acl.User) error
}

type changePasswordRepo struct {
	db *buntdb.DB
}

func (r *changePasswordRepo) GetUserByID(userID string) (acl.User, error) {
	return userrepo.GetUserByID(r.db, userID)
}

func (r *changePasswordRepo) UpdateUser(user acl.User) error {
	return userrepo.UpdateUser(r.db, user)
}

type ChangePasswordUsecase struct {
	repo iUserPasswordRepo
}

func NewChangePasswordUsecase(db *buntdb.DB) ChangePasswordUsecase {
	return ChangePasswordUsecase{
		repo: &changePasswordRepo{db: db},
	}
}

func (uc ChangePasswordUsecase) Handle(ctx context.Context, req ChangePasswordRequest) (ChangePasswordResponse, error) {
	if req.UserID == "" || req.OldPassword == "" || req.NewPassword == "" {
		return ChangePasswordResponse{Success: false}, errors.New("missing required fields: user_id, old_password, or new_password")
	}
	user, err := uc.repo.GetUserByID(req.UserID)
	if err != nil {
		return ChangePasswordResponse{Success: false}, errors.New("user not found")
	}
	// Hash the old password and compare
	hash := sha256.Sum256([]byte(req.OldPassword))
	hashedOldPassword := hex.EncodeToString(hash[:])
	if user.Password != hashedOldPassword {
		return ChangePasswordResponse{Success: false}, errors.New("invalid old password")
	}
	// Hash the new password
	newHash := sha256.Sum256([]byte(req.NewPassword))
	hashedNewPassword := hex.EncodeToString(newHash[:])
	user.Password = hashedNewPassword
	user.UpdatedAt = time.Now()
	if err := uc.repo.UpdateUser(user); err != nil {
		return ChangePasswordResponse{Success: false}, err
	}
	return ChangePasswordResponse{Success: true}, nil
}
