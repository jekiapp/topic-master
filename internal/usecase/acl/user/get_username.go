package user

import (
	"context"
	"errors"

	acl "github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type GetUsernameResponse struct {
	Name         string        `json:"name"`
	Username     string        `json:"username"`
	Root         bool          `json:"root"`
	Groups       []string      `json:"groups"`
	GroupDetails []GroupDetail `json:"group_details"`
}

// iUserGroupRepo abstracts user group lookup
//
//go:generate mockgen -source=get_username.go -destination=mock/mock_get_username_repo.go -package=user_mock
type iUserGroupRepo interface {
	ListUserGroupsByUserID(userID string) ([]acl.UserGroup, error)
	GetGroupByID(groupID string) (acl.Group, error)
}

type GetUsernameUsecase struct {
	repo iUserGroupRepo
}

func NewGetUsernameUsecase(db *buntdb.DB) GetUsernameUsecase {
	return GetUsernameUsecase{repo: &userGroupRepoImpl{db: db}}
}

// Handle extracts the username from context and fetches group info from repo
func (uc GetUsernameUsecase) Handle(ctx context.Context, params map[string]string) (GetUsernameResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil || user.Username == "" {
		return GetUsernameResponse{}, errors.New("user is not authenticated")
	}

	userGroups, err := uc.repo.ListUserGroupsByUserID(user.ID)
	if err != nil {
		return GetUsernameResponse{}, err
	}

	isRoot := false
	groups := []string{}
	groupDetails := []GroupDetail{}
	for _, ug := range userGroups {
		group, err := uc.repo.GetGroupByID(ug.GroupID)
		if err != nil {
			return GetUsernameResponse{}, err
		}
		groupName := group.Name
		groups = append(groups, groupName)
		groupDetails = append(groupDetails, GroupDetail{
			GroupID:   ug.GroupID,
			GroupName: groupName,
			Role:      ug.Role,
		})
		if groupName == acl.GroupRoot {
			isRoot = true
		}
	}

	return GetUsernameResponse{
		Name:         user.Name,
		Username:     user.Username,
		Root:         isRoot,
		Groups:       groups,
		GroupDetails: groupDetails,
	}, nil
}

type userGroupRepoImpl struct {
	db *buntdb.DB
}

func (r *userGroupRepoImpl) ListUserGroupsByUserID(userID string) ([]acl.UserGroup, error) {
	return user.ListUserGroupsByUserID(r.db, userID)
}

func (r *userGroupRepoImpl) GetGroupByID(groupID string) (acl.Group, error) {
	return user.GetGroupByID(r.db, groupID)
}
