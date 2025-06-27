package acl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type ResetPasswordRequest struct {
	Token           string `json:"token"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ResetPasswordResponse struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Redirect string `json:"redirect,omitempty"`
}

type IResetPasswordRepo interface {
	GetResetPasswordByToken(token string) (acl.ResetPassword, error)
	GetResetPasswordByUsername(username string) (acl.ResetPassword, error)
	DeleteResetPasswordByToken(token string) error
}

type IUserRepo interface {
	GetUserByUsername(username string) (acl.User, error)
	UpdateUser(user acl.User) error
}

type resetPasswordRepo struct {
	db *buntdb.DB
}

func (r *resetPasswordRepo) GetResetPasswordByToken(token string) (acl.ResetPassword, error) {
	rp, err := db.GetByID[acl.ResetPassword](r.db, token)
	if err != nil {
		return acl.ResetPassword{}, err
	}
	return rp, nil
}

func (r *resetPasswordRepo) GetResetPasswordByUsername(username string) (acl.ResetPassword, error) {
	rp, err := db.SelectOne[acl.ResetPassword](r.db, username, acl.IdxResetPassword_Username)
	if err != nil {
		return acl.ResetPassword{}, err
	}
	return rp, nil
}

func (r *resetPasswordRepo) DeleteResetPasswordByToken(token string) error {
	return db.DeleteByID[acl.ResetPassword](r.db, token)
}

type userRepo struct {
	db *buntdb.DB
}

func (r *userRepo) GetUserByUsername(username string) (acl.User, error) {
	return userrepo.GetUserByUsername(r.db, username)
}
func (r *userRepo) UpdateUser(user acl.User) error {
	return userrepo.UpdateUser(r.db, user)
}

type ResetPasswordUsecase struct {
	rpRepo   IResetPasswordRepo
	userRepo IUserRepo
}

func NewResetPasswordUsecase(db *buntdb.DB) ResetPasswordUsecase {
	return ResetPasswordUsecase{
		rpRepo:   &resetPasswordRepo{db: db},
		userRepo: &userRepo{db: db},
	}
}

type ResetPasswordGetResponse struct {
	Error    string `json:"error,omitempty"`
	Username string `json:"username,omitempty"`
}

func (uc ResetPasswordUsecase) HandleGet(ctx context.Context, req map[string]string) (ResetPasswordGetResponse, error) {
	token := req["token"]
	if token == "" {
		return ResetPasswordGetResponse{Error: "Token is required"}, nil
	}
	rp, err := uc.rpRepo.GetResetPasswordByToken(token)
	if err != nil {
		return ResetPasswordGetResponse{Error: "Invalid or expired token"}, nil
	}
	return ResetPasswordGetResponse{Username: rp.Username}, nil
}

func (uc ResetPasswordUsecase) HandlePost(ctx context.Context, req ResetPasswordRequest) (ResetPasswordResponse, error) {
	if req.NewPassword == "" || req.ConfirmPassword == "" {
		return ResetPasswordResponse{Success: false, Error: "Password fields required"}, nil
	}
	if req.NewPassword != req.ConfirmPassword {
		return ResetPasswordResponse{Success: false, Error: "Passwords do not match"}, nil
	}
	if len(req.NewPassword) < 8 {
		return ResetPasswordResponse{Success: false, Error: "Password too short (min 8)"}, nil
	}

	rp, err := uc.rpRepo.GetResetPasswordByToken(req.Token)
	if err != nil {
		return ResetPasswordResponse{Success: false, Error: "Invalid or expired token"}, nil
	}
	if rp.ExpiresAt > 0 && time.Now().Unix() > rp.ExpiresAt {
		return ResetPasswordResponse{Success: false, Error: "Token expired"}, nil
	}

	user, err := uc.userRepo.GetUserByUsername(rp.Username)
	if err != nil {
		return ResetPasswordResponse{Success: false, Error: "User not found"}, nil
	}

	hash := sha256.Sum256([]byte(req.NewPassword))
	hashedPassword := hex.EncodeToString(hash[:])
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()
	user.Status = acl.StatusUserActive
	if err := uc.userRepo.UpdateUser(user); err != nil {
		return ResetPasswordResponse{Success: false, Error: "Failed to update password"}, nil
	}
	_ = uc.rpRepo.DeleteResetPasswordByToken(req.Token)
	return ResetPasswordResponse{Success: true, Redirect: "/login"}, nil
}
