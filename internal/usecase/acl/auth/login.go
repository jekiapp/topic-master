package acl

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jekiapp/topic-master/internal/config"
	"github.com/jekiapp/topic-master/internal/logic/auth"
	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	dbPkg "github.com/jekiapp/topic-master/pkg/db"
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
	ListGroupsForUser(userID string) ([]acl.GroupRole, error)
	InsertResetPassword(rp acl.ResetPassword) error
}

type loginRepo struct {
	db *buntdb.DB
}

func (r *loginRepo) GetUserByUsername(username string) (acl.User, error) {
	return userrepo.GetUserByUsername(r.db, username)
}

func (r *loginRepo) ListGroupsForUser(userID string) ([]acl.GroupRole, error) {
	return userrepo.ListGroupsForUser(r.db, userID)
}

func (r *loginRepo) InsertResetPassword(rp acl.ResetPassword) error {
	return dbPkg.Insert(r.db, &rp)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := uc.repo.GetUserByUsername(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}
	// Hash the provided password and compare
	hash := sha256.Sum256([]byte(req.Password))
	hashedPassword := hex.EncodeToString(hash[:])
	if user.Password != hashedPassword {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid password"})
		return
	}
	if user.Status == acl.StatusUserPending {
		// Generate reset token
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate reset token"})
			return
		}
		token := hex.EncodeToString(tokenBytes)
		expiresAt := time.Now().Add(1 * time.Hour).Unix()
		rp := &acl.ResetPassword{
			Token:     token,
			Username:  user.Username,
			CreatedAt: time.Now().Unix(),
			ExpiresAt: expiresAt,
		}
		if err := uc.repo.InsertResetPassword(*rp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to save reset token"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"redirect": "/reset-password?token=" + token})
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

	// password is ok, now check if user is pending

	if user.Status == acl.StatusUserPending {
		// Generate reset token
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return LoginResponse{}, errors.New("failed to generate reset token")
		}
		token := hex.EncodeToString(tokenBytes)
		expiresAt := time.Now().Add(1 * time.Hour).Unix()
		rp := &acl.ResetPassword{
			Token:     token,
			Username:  user.Username,
			CreatedAt: time.Now().Unix(),
			ExpiresAt: expiresAt,
		}
		if err := dbPkg.Insert(uc.repo.(*loginRepo).db, rp); err != nil {
			return LoginResponse{}, errors.New("failed to save reset token")
		}
		return LoginResponse{}, nil
	}

	// Fetch all groups for the user
	groups, err := uc.repo.ListGroupsForUser(user.ID)
	if err != nil {
		return LoginResponse{}, errors.New("failed to fetch user groups (" + err.Error() + ")")
	}

	// Prepare JWT claims
	claims := &acl.JWTClaims{
		UserID:           user.ID,
		Name:             user.Name,
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
