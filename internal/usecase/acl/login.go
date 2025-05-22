package acl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/jekiapp/nsqper/internal/config"
	"github.com/jekiapp/nsqper/internal/logic/auth"
	"github.com/jekiapp/nsqper/internal/model/acl"
	userrepo "github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/tidwall/buntdb"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  acl.User `json:"user"`
}

type iUserLoginRepo interface {
	GetUserByUsername(username string) (*acl.User, error)
}

type loginRepo struct {
	db *buntdb.DB
}

func (r *loginRepo) GetUserByUsername(username string) (*acl.User, error) {
	return userrepo.GetUserByUsername(r.db, username)
}

type LoginUsecase struct {
	repo   iUserLoginRepo
	config *config.Config
}

func NewLoginUsecase(db *buntdb.DB, cfg *config.Config) LoginUsecase {
	return LoginUsecase{
		repo:   &loginRepo{db: db},
		config: cfg,
	}
}

const ACCESS_TOKEN_COOKIE_NAME = "access_token"

func (uc LoginUsecase) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp, err := uc.doLogin(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	// Set JWT as HttpOnly, Secure cookie
	cookie := &http.Cookie{
		Name:     ACCESS_TOKEN_COOKIE_NAME,
		Value:    resp.Token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
	// Return only user info (no token)
	json.NewEncoder(w).Encode(map[string]interface{}{"user": resp.User})
}

func (uc LoginUsecase) doLogin(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	user, err := uc.repo.GetUserByUsername(req.Username)
	if err != nil {
		return LoginResponse{}, err
	}
	if user == nil {
		return LoginResponse{}, errors.New("user not found")
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
		User:  *user,
	}, nil
}
