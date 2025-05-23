// extract user info from the request context

package util

import (
	"context"

	"github.com/jekiapp/nsqper/internal/model"
	"github.com/jekiapp/nsqper/internal/model/acl"
)

func GetUserInfo(ctx context.Context) *acl.User {
	claims := ctx.Value(model.UserInfoKey).(*acl.JWTClaims)
	if claims == nil {
		return nil
	}
	user := &acl.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Groups:   claims.Groups,
	}
	return user
}
