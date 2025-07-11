//go:generate mockgen -source=get_user_list.go -destination=mock/mock_get_user_list_repo.go -package=user_mock
// learn from get_group_list.go
// this is for listing users
// the column response is username, name, email,  groups, status

package user

import (
	"context"
	"log"
	"strings"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/tidwall/buntdb"
)

type GetUserListRequest struct{}

type GetUserListResponse struct {
	Users []UserListItem `json:"users"`
}

type UserListItem struct {
	ID           string        `json:"id"`
	Username     string        `json:"username"`
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	Groups       string        `json:"groups"`
	Status       string        `json:"status"`
	GroupDetails []GroupDetail `json:"group_details"`
}

type GroupDetail struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Role      string `json:"role"`
}

type iUserDataRepo interface {
	GetAllUsers() ([]acl.User, error)
	ListUserGroupsByUserID(userID string) ([]acl.UserGroup, error)
	GetGroupByID(groupID string) (acl.Group, error)
}

type GetUserListUsecase struct {
	dataRepo iUserDataRepo
}

func NewGetUserListUsecase(db *buntdb.DB) GetUserListUsecase {
	return GetUserListUsecase{
		dataRepo: &userDataRepoImpl{db: db},
	}
}

func (uc GetUserListUsecase) Handle(ctx context.Context, req GetUserListRequest) (GetUserListResponse, error) {
	users, err := uc.dataRepo.GetAllUsers()
	if err != nil {
		return GetUserListResponse{}, err
	}
	var result []UserListItem
	for _, u := range users {
		userGroups, err := uc.dataRepo.ListUserGroupsByUserID(u.ID)
		if err != nil {
			log.Printf("error listing user groups: %s", err)
		}
		var groupNames []string
		var groupDetails []GroupDetail
		if err == nil {
			for _, ug := range userGroups {
				group, err := uc.dataRepo.GetGroupByID(ug.GroupID)
				if err == nil {
					groupNames = append(groupNames, group.Name)
					groupDetails = append(groupDetails, GroupDetail{
						GroupID:   ug.GroupID,
						GroupName: group.Name,
						Role:      ug.Role,
					})
				}
			}
		}
		result = append(result, UserListItem{
			ID:           u.ID,
			Username:     u.Username,
			Name:         u.Name,
			Groups:       strings.Join(groupNames, ","),
			Status:       u.Status,
			GroupDetails: groupDetails,
		})
	}
	return GetUserListResponse{Users: result}, nil
}

// --- unified repo implementation ---
type userDataRepoImpl struct {
	db *buntdb.DB
}

func (r *userDataRepoImpl) GetAllUsers() ([]acl.User, error) {
	return user.GetAllUsers(r.db)
}

func (r *userDataRepoImpl) ListUserGroupsByUserID(userID string) ([]acl.UserGroup, error) {
	return user.ListUserGroupsByUserID(r.db, userID)
}

func (r *userDataRepoImpl) GetGroupByID(groupID string) (acl.Group, error) {
	return user.GetGroupByID(r.db, groupID)
}
