package acl

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/jekiapp/nsqper/pkg/db"
	"github.com/tidwall/buntdb"
)

type JWTClaims struct {
	UserID   string      `json:"user_id"`
	Username string      `json:"username"`
	Groups   []GroupRole `json:"groups"`
	jwt.RegisteredClaims
}

// ResetPassword represents a password reset token and its association to a username
// and is used for password reset flows.
type ResetPassword struct {
	Token     string `json:"token"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
}

const (
	TableResetPassword        = "reset_password"
	IdxResetPassword_Username = TableResetPassword + ":username"
)

func (r ResetPassword) GetPrimaryKey() string {
	return TableResetPassword + ":" + r.Token
}

func (r ResetPassword) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxResetPassword_Username,
			Pattern: TableResetPassword + ":*:username",
			Type:    buntdb.IndexString,
		},
	}
}

func (r ResetPassword) GetIndexValues() map[string]string {
	return map[string]string{
		"username": r.Username,
	}
}
