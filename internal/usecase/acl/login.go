package acl

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/jekiapp/nsqper/internal/config"
	"github.com/jekiapp/nsqper/internal/logic/auth"
	"github.com/jekiapp/nsqper/internal/model/acl"
	usergrouprepo "github.com/jekiapp/nsqper/internal/repository/user"
	userrepo "github.com/jekiapp/nsqper/internal/repository/user"
	dbPkg "github.com/jekiapp/nsqper/pkg/db"
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
	GetUserByUsername(username string) (acl.User, error)
	ListGroupsForUser(userID, userType string) ([]acl.GroupRole, error)
}

type loginRepo struct {
	db *buntdb.DB
}

func (r *loginRepo) GetUserByUsername(username string) (acl.User, error) {
	return userrepo.GetUserByUsername(r.db, username)
}

func (r *loginRepo) ListGroupsForUser(userID, userType string) ([]acl.GroupRole, error) {
	return usergrouprepo.ListGroupsForUser(r.db, userID, userType)
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
		if err == dbPkg.ErrNotFound {
			return LoginResponse{}, errors.New("user not found")
		}
		return LoginResponse{}, err
	}

	// Hash the provided password and compare
	hash := sha256.Sum256([]byte(req.Password))
	hashedPassword := hex.EncodeToString(hash[:])
	if user.Password != hashedPassword {
		return LoginResponse{}, errors.New("invalid password")
	}

	// Fetch all groups for the user
	groups, err := uc.repo.ListGroupsForUser(user.ID, user.Type)
	if err != nil {
		return LoginResponse{}, errors.New("failed to fetch user groups (" + err.Error() + ")")
	}

	// Prepare JWT claims
	claims := &acl.JWTClaims{
		UserID:           user.ID,
		Username:         user.Username,
		Groups:           groups,
		RegisteredClaims: auth.DefaultRegisteredClaims(user.ID),
	}

	// Decode base64 secret key before using for JWT
	secret, err := base64.StdEncoding.DecodeString(string(uc.config.SecretKey))
	if err != nil {
		return LoginResponse{}, errors.New("failed to decode secret key")
	}
	token, err := auth.GenerateJWT(claims, secret)
	if err != nil {
		return LoginResponse{}, errors.New("failed to generate token")
	}

	return LoginResponse{
		Token: token,
		User:  user,
	}, nil
}
