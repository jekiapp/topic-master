package acl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/jekiapp/nsqper/internal/config"
	"github.com/jekiapp/nsqper/internal/logic/auth"
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  acl.User `json:"user"`
}

type LoginUsecase struct {
	db     *buntdb.DB
	config *config.Config
}

func NewLoginUsecase(db *buntdb.DB, cfg *config.Config) LoginUsecase {
	return LoginUsecase{db: db, config: cfg}
}

func (uc LoginUsecase) Handle(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	var user acl.User
	key := "user:" + req.Username
	err := uc.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return errors.New("user not found")
		}
		if err := msgpack.Unmarshal([]byte(val), &user); err != nil {
			return errors.New("failed to decode user data")
		}
		return nil
	})
	if err != nil {
		return LoginResponse{}, err
	}

	// Hash the provided password and compare
	hash := sha256.Sum256([]byte(req.Password))
	hashedPassword := hex.EncodeToString(hash[:])
	if user.Password != hashedPassword {
		return LoginResponse{}, errors.New("invalid password")
	}

	// Prepare JWT claims
	claims := &acl.JWTClaims{
		UserID:           user.ID,
		Username:         user.Username,
		Roles:            []string{user.Type},                   // treat user.Type as a single role for now
		RegisteredClaims: auth.DefaultRegisteredClaims(user.ID), // helper to set standard claims
	}

	token, err := auth.GenerateJWT(claims, uc.config.SecretKey)
	if err != nil {
		return LoginResponse{}, errors.New("failed to generate token")
	}

	return LoginResponse{
		Token: token,
		User:  user,
	}, nil
}
