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
	"github.com/jekiapp/topic-master/pkg/util"

	aclusecase "github.com/jekiapp/topic-master/internal/usecase/acl/auth"
)

func InitSessionMiddleware(secret string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			var tokenString string
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				cookie, err := r.Cookie("access_token")
				if err == nil && cookie.Value != "" {
					tokenString = cookie.Value
				}
			}

			if tokenString != "" {
				decodedSecret, err := base64.StdEncoding.DecodeString(secret)
				if err == nil {
					token := &acl.JWTClaims{}
					parsedToken, err := jwt.ParseWithClaims(tokenString, token, func(token *jwt.Token) (interface{}, error) {
						return decodedSecret, nil
					})
					if err == nil && parsedToken.Valid {
						claims, ok := parsedToken.Claims.(*acl.JWTClaims)
						if ok {
							ctx := context.WithValue(r.Context(), model.UserInfoKey, claims)
							r = r.WithContext(ctx)
						}
					}
				}
			}
			next(w, r)
		}
	}
}

func InitJWTMiddleware(secret string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return JWTMiddleware(next, secret)
	}
}

func InitJWTMiddlewareWithRoot(secret string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		rootNext := func(w http.ResponseWriter, r *http.Request) {

			claims := util.GetUserInfo(r.Context())
			// check if the user is root
			isRoot := false
			for _, group := range claims.Groups {
				if group.GroupName == acl.GroupRoot {
					isRoot = true
					break
				}
			}

			if !isRoot {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next(w, r)
		}
		return JWTMiddleware(rootNext, secret)
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

func InitActionAuthMiddleware(secret string, usecase aclusecase.CheckActionAuthUsecase) func(next http.HandlerFunc, action string) http.HandlerFunc {
	return func(next http.HandlerFunc, action string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user := util.GetUserInfo(r.Context())
			if user == nil {
				http.Error(w, "Unauthorized: user not found", http.StatusUnauthorized)
				return
			}

			entityID := r.URL.Query().Get("entity_id")
			if entityID == "" {
				http.Error(w, "Unauthorized: missing entity_id", http.StatusUnauthorized)
				return
			}

			resp, err := usecase.Handle(r.Context(), aclusecase.CheckActionAuthRequest{
				EntityID: entityID,
				Action:   action,
			})
			if err != nil || !resp.Allowed {
				errMsg := "Unauthorized"
				if resp.Error != "" {
					errMsg += ": " + resp.Error
				} else if err != nil {
					errMsg += ": " + err.Error()
				}
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
}
