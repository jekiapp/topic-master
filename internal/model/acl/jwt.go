package acl

import "github.com/golang-jwt/jwt/v4"

type JWTClaims struct {
	UserID   string      `json:"user_id"`
	Username string      `json:"username"`
	Groups   []GroupRole `json:"groups"`
	jwt.RegisteredClaims
}
