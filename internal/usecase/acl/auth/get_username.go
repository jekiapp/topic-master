package acl

import (
	"context"
	"errors"

	"github.com/jekiapp/topic-master/pkg/util"
)

// GetUsernameResponse is the response for GetUsernameUsecase
type GetUsernameResponse struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

// GetUsernameUsecase provides the username from context
// No dependencies needed
// see util.GetUserInfo()
type GetUsernameUsecase struct{}

func NewGetUsernameUsecase() GetUsernameUsecase {
	return GetUsernameUsecase{}
}

// Handle extracts the username from context
func (uc GetUsernameUsecase) Handle(ctx context.Context, params map[string]string) (GetUsernameResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil || user.Username == "" {
		return GetUsernameResponse{}, errors.New("user is not authenticated")
	}
	return GetUsernameResponse{Name: user.Name, Username: user.Username}, nil
}
