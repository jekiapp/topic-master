package user

import (
	"context"

	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type DeleteUserRequest struct {
	UserID string `json:"user_id"`
}

type DeleteUserResponse struct {
	Success bool `json:"success"`
}

type iUserDeleteRepo interface {
	DeleteUser(userID string) error
	ListUserGroupsByUserID(userID string) ([]acl.UserGroup, error)
	DeleteUserGroup(userGroupID string) error
}

type userDeleteRepo struct {
	db *buntdb.DB
}

func (r *userDeleteRepo) DeleteUser(userID string) error {
	return db.DeleteByID[acl.User](r.db, userID)
}

func (r *userDeleteRepo) ListUserGroupsByUserID(userID string) ([]acl.UserGroup, error) {
	return userrepo.ListUserGroupsByUserID(r.db, userID)
}

func (r *userDeleteRepo) DeleteUserGroup(userGroupID string) error {
	return db.DeleteByID[acl.UserGroup](r.db, userGroupID)
}

type DeleteUserUsecase struct {
	repo iUserDeleteRepo
}

func NewDeleteUserUsecase(db *buntdb.DB) DeleteUserUsecase {
	return DeleteUserUsecase{
		repo: &userDeleteRepo{db: db},
	}
}

func (uc DeleteUserUsecase) Handle(ctx context.Context, req DeleteUserRequest) (DeleteUserResponse, error) {
	err := uc.repo.DeleteUser(req.UserID)
	if err != nil {
		return DeleteUserResponse{Success: false}, err
	}

	// delete all user groups
	userGroups, err := uc.repo.ListUserGroupsByUserID(req.UserID)
	if err != nil {
		return DeleteUserResponse{Success: false}, err
	}
	for _, userGroup := range userGroups {
		err = uc.repo.DeleteUserGroup(userGroup.ID)
		if err != nil {
			return DeleteUserResponse{Success: false}, err
		}
	}

	return DeleteUserResponse{Success: true}, nil
}
