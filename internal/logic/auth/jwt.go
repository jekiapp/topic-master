package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jekiapp/nsqper/internal/model/acl"
)

// ValidateJWT parses and validates a JWT token string.
func ValidateJWT(tokenString string, jwtSecret []byte) (*acl.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &acl.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	// Validate the token and its claims
	if claims, ok := token.Claims.(*acl.JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// GenerateJWT creates a signed JWT token string for the given claims and secret key.
func GenerateJWT(claims *acl.JWTClaims, jwtSecret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// DefaultRegisteredClaims is a helper to set standard JWT claims (exp, iat, sub, etc.)
func DefaultRegisteredClaims(userID string) jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
	}
}
