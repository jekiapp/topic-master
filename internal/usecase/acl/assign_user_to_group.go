package acl

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/acl"
	usergrouprepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/tidwall/buntdb"
)

type AssignUserToGroupRequest struct {
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

type AssignUserToGroupResponse struct {
	UserGroup acl.UserGroup
}

type iUserGroupRepo interface {
	CreateUserGroup(userGroup acl.UserGroup) error
	GetUserGroup(userID, groupID string) (acl.UserGroup, error)
}

type userGroupRepo struct {
	db *buntdb.DB
}

func (r *userGroupRepo) CreateUserGroup(userGroup acl.UserGroup) error {
	return usergrouprepo.CreateUserGroup(r.db, userGroup)
}

func (r *userGroupRepo) GetUserGroup(userID, groupID string) (acl.UserGroup, error) {
	return usergrouprepo.GetUserGroup(r.db, userID, groupID)
}

type AssignUserToGroupUsecase struct {
	repo iUserGroupRepo
}

func NewAssignUserToGroupUsecase(db *buntdb.DB) AssignUserToGroupUsecase {
	return AssignUserToGroupUsecase{
		repo: &userGroupRepo{db: db},
	}
}

func (uc AssignUserToGroupUsecase) Handle(ctx context.Context, req AssignUserToGroupRequest) (AssignUserToGroupResponse, error) {
	if req.UserID == "" || req.GroupID == "" {
		return AssignUserToGroupResponse{}, errors.New("missing required fields: user_id or group_id")
	}
	// Check if mapping already exists
	_, err := uc.repo.GetUserGroup(req.UserID, req.GroupID)
	if err == nil {
		return AssignUserToGroupResponse{}, errors.New("user is already assigned to this group")
	}
	userGroup := acl.UserGroup{
		ID:        uuid.NewString(),
		UserID:    req.UserID,
		GroupID:   req.GroupID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.CreateUserGroup(userGroup); err != nil {
		return AssignUserToGroupResponse{}, err
	}
	return AssignUserToGroupResponse{UserGroup: userGroup}, nil
}
