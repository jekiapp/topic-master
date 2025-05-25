// this usecase will be used to get the list of groups
// follow the same pattern as create_user.go
// the field will be: name, description, members (comma separated list of top 3 usernames)
// create get all groups func in repository/user/group.go

package group

import (
	"context"
	"log"
	"strings"

	"github.com/jekiapp/nsqper/internal/model/acl"
	grouprepo "github.com/jekiapp/nsqper/internal/repository/user"
	userrepo "github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/tidwall/buntdb"
)

type GetGroupListRequest struct{}

type GetGroupListResponse struct {
	Groups []GroupListItem `json:"groups"`
}

type GroupListItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Members     string `json:"members"`
}

type iGroupListRepo interface {
	GetAllGroups() ([]acl.Group, error)
}

type iUserGroupRepo interface {
	ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error)
	GetUserByID(userID string) (acl.User, error)
}

type GetGroupListUsecase struct {
	groupRepo     iGroupListRepo
	userGroupRepo iUserGroupRepo
}

func NewGetGroupListUsecase(db *buntdb.DB) GetGroupListUsecase {
	return GetGroupListUsecase{
		groupRepo:     &groupRepoImpl{db: db},
		userGroupRepo: &userGroupRepoImpl{db: db},
	}
}

func (uc GetGroupListUsecase) Handle(ctx context.Context, req GetGroupListRequest) (GetGroupListResponse, error) {
	groups, err := uc.groupRepo.GetAllGroups()
	if err != nil {
		return GetGroupListResponse{}, err
	}
	var result []GroupListItem
	for _, g := range groups {
		userGroups, err := uc.userGroupRepo.ListUserGroupsByGroupID(g.ID, 3)
		if err != nil {
			log.Printf("error listing user groups: %s", err)
		}
		var usernames []string
		if err == nil {
			for _, ug := range userGroups {
				user, err := uc.userGroupRepo.GetUserByID(ug.UserID)
				if err == nil {
					usernames = append(usernames, user.Username)
				}
			}
		}
		result = append(result, GroupListItem{
			Name:        g.Name,
			Description: g.Description,
			Members:     strings.Join(usernames, ","),
		})
	}
	return GetGroupListResponse{Groups: result}, nil
}

// --- repo implementations ---
type groupRepoImpl struct {
	db *buntdb.DB
}

func (r *groupRepoImpl) GetAllGroups() ([]acl.Group, error) {
	return grouprepo.GetAllGroups(r.db)
}

type userGroupRepoImpl struct {
	db *buntdb.DB
}

func (r *userGroupRepoImpl) ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error) {
	return userrepo.ListUserGroupsByGroupID(r.db, groupID, limit)
}

func (r *userGroupRepoImpl) GetUserByID(userID string) (acl.User, error) {
	return userrepo.GetUserByID(r.db, userID)
}
