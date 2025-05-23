package handler

import (
	"net/http"
	"strings"

	"context"

	"github.com/golang-jwt/jwt/v4"

	"github.com/jekiapp/nsqper/internal/model"
	"github.com/jekiapp/nsqper/internal/model/acl"
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
		if !strings.HasPrefix(authHeader, "Bearer ") {
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Missing or invalid Authorization header"}`))
			} else {
				http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			}
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token := &acl.JWTClaims{}
		parsedToken, err := jwt.ParseWithClaims(tokenString, token, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !parsedToken.Valid {
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Invalid token"}`))
			} else {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
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
