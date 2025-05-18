package acl

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/nsqper/internal/model/acl"
	grouprepo "github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/tidwall/buntdb"
)

type CreateUserGroupRequest struct {
	Name string `json:"name"`
}

type CreateUserGroupResponse struct {
	Group acl.Group
}

type iGroupRepo interface {
	CreateGroup(group acl.Group) error
	GetGroupByName(name string) (*acl.Group, error)
}

type createGroupRepo struct {
	db *buntdb.DB
}

func (r *createGroupRepo) CreateGroup(group acl.Group) error {
	return grouprepo.CreateGroup(r.db, group)
}

func (r *createGroupRepo) GetGroupByName(name string) (*acl.Group, error) {
	return grouprepo.GetGroupByName(r.db, name)
}

type CreateUserGroupUsecase struct {
	repo iGroupRepo
}

func NewCreateUserGroupUsecase(db *buntdb.DB) CreateUserGroupUsecase {
	return CreateUserGroupUsecase{
		repo: &createGroupRepo{db: db},
	}
}

func (uc CreateUserGroupUsecase) Handle(ctx context.Context, req CreateUserGroupRequest) (CreateUserGroupResponse, error) {
	if req.Name == "" {
		return CreateUserGroupResponse{}, errors.New("missing required field: name")
	}
	// Check if group already exists
	existingGroup, err := uc.repo.GetGroupByName(req.Name)
	if err == nil && existingGroup != nil {
		return CreateUserGroupResponse{}, errors.New("group already exists")
	}
	group := acl.Group{
		ID:        uuid.NewString(),
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.CreateGroup(group); err != nil {
		return CreateUserGroupResponse{}, err
	}
	return CreateUserGroupResponse{Group: group}, nil
}
