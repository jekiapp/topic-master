package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"

	"github.com/jekiapp/topic-master/internal/model"
	"github.com/jekiapp/topic-master/internal/model/acl"
)

func InitJWTMiddleware(secret string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return JWTMiddleware(next, secret)
	}
}

func JWTMiddleware(next http.HandlerFunc, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		isAjax := r.Header.Get("X-Requested-With") == "XMLHttpRequest"

		var tokenString string
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Try to get access_token from cookies
			cookie, err := r.Cookie("access_token")
			if err == nil && cookie.Value != "" {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Missing or invalid Authorization header and access_token cookie"}`))
			} else {
				http.Error(w, "Missing or invalid Authorization header and access_token cookie", http.StatusUnauthorized)
			}
			return
		}

		// Decode base64 secret key before using for JWT verification
		decodedSecret, err := base64.StdEncoding.DecodeString(secret)
		if err != nil {
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Failed to decode secret key"}`))
			} else {
				http.Error(w, "Failed to decode secret key", http.StatusUnauthorized)
			}
			return
		}
		token := &acl.JWTClaims{}
		parsedToken, err := jwt.ParseWithClaims(tokenString, token, func(token *jwt.Token) (interface{}, error) {
			return decodedSecret, nil
		})
		if err != nil || !parsedToken.Valid {
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(fmt.Sprintf(`{"error": "Invalid token %s"}`, err.Error())))
			} else {
				http.Error(w, fmt.Sprintf("Invalid token %s", err.Error()), http.StatusUnauthorized)
			}
			return
		}
		claims, ok := parsedToken.Claims.(*acl.JWTClaims)
		if !ok {
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Invalid token claims"}`))
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			}
			return
		}
		ctx := context.WithValue(r.Context(), model.UserInfoKey, claims)
		next(w, r.WithContext(ctx))
	}
}
