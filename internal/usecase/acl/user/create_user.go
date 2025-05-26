package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/nsqper/internal/model/acl"
	userrepo "github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/tidwall/buntdb"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	GroupID  string `json:"group_id"`
	Type     string `json:"type"`
}

type CreateUserResponse struct {
	Username          string `json:"username"`
	GeneratedPassword string `json:"generated_password"`
}

type iUserRepo interface {
	CreateUser(user acl.User) error
	GetUserByUsername(username string) (acl.User, error)
	CreateUserGroup(userGroup acl.UserGroup) error
}

type CreateUserUsecase struct {
	repo iUserRepo
}

func NewCreateUserUsecase(db *buntdb.DB) CreateUserUsecase {
	return CreateUserUsecase{
		repo: &createUserRepo{db: db},
	}
}

func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

func validateCreateUserRequest(req CreateUserRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if req.GroupID == "" {
		return errors.New("group_id is required")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Type == "" {
		return errors.New("type is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

func (uc CreateUserUsecase) Handle(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {
	// Basic input validation
	if err := validateCreateUserRequest(req); err != nil {
		return CreateUserResponse{}, err
	}
	// Check if user already exists
	_, err := uc.repo.GetUserByUsername(req.Username)
	if err == nil {
		return CreateUserResponse{}, errors.New("user already exists")
	}
	// Generate UUID for user ID
	userID := uuid.NewString()
	password := req.Password
	if password == "" {
		var err error
		password, err = generateRandomPassword(12)
		if err != nil {
			return CreateUserResponse{}, err
		}
	}
	// Hash the password (simple SHA256 for demonstration)
	hash := sha256.Sum256([]byte(password))
	hashedPassword := hex.EncodeToString(hash[:])
	user := acl.User{
		ID:        userID,
		Username:  req.Username,
		Name:      req.Name,
		Password:  hashedPassword,
		Email:     req.Email,
		Phone:     req.Phone,
		Type:      req.Type,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.CreateUser(user); err != nil {
		return CreateUserResponse{}, err
	}

	// create user group mapping
	userGroup := acl.UserGroup{
		UserID:    user.ID,
		GroupID:   req.GroupID,
		Type:      req.Type,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.CreateUserGroup(userGroup); err != nil {
		return CreateUserResponse{}, err
	}

	return CreateUserResponse{Username: user.Username, GeneratedPassword: password}, nil
}

type createUserRepo struct {
	db *buntdb.DB
}

func (r *createUserRepo) CreateUser(user acl.User) error {
	return userrepo.CreateUser(r.db, user)
}

func (r *createUserRepo) GetUserByUsername(username string) (acl.User, error) {
	return userrepo.GetUserByUsername(r.db, username)
}

func (r *createUserRepo) CreateUserGroup(userGroup acl.UserGroup) error {
	return userrepo.CreateUserGroup(r.db, userGroup)
}
