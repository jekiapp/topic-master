package util

import (
	"context"

	"github.com/jekiapp/topic-master/internal/model"
	"github.com/jekiapp/topic-master/internal/model/acl"
)

// MockContextWithUser returns a context containing the given user as JWTClaims for testing.
func MockContextWithUser(ctx context.Context, user *acl.User) context.Context {
	claims := &acl.JWTClaims{
		UserID:   user.ID,
		Name:     user.Name,
		Username: user.Username,
		Groups:   user.Groups,
	}
	return context.WithValue(ctx, model.UserInfoKey, claims)
}
