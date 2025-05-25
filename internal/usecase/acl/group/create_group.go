package group

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/nsqper/internal/model/acl"
	grouprepo "github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/tidwall/buntdb"
)

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateGroupResponse struct {
	Group acl.Group
}

type iCreateGroupRepo interface {
	CreateGroup(group acl.Group) error
	GetGroupByName(name string) (acl.Group, error)
}

type createGroupRepo struct {
	db *buntdb.DB
}

func (r *createGroupRepo) CreateGroup(group acl.Group) error {
	return grouprepo.CreateGroup(r.db, group)
}

func (r *createGroupRepo) GetGroupByName(name string) (acl.Group, error) {
	return grouprepo.GetGroupByName(r.db, name)
}

type CreateGroupUsecase struct {
	repo iCreateGroupRepo
}

func NewCreateGroupUsecase(db *buntdb.DB) CreateGroupUsecase {
	return CreateGroupUsecase{
		repo: &createGroupRepo{db: db},
	}
}

func (uc CreateGroupUsecase) Handle(ctx context.Context, req CreateGroupRequest) (CreateGroupResponse, error) {
	if req.Name == "" {
		return CreateGroupResponse{}, errors.New("missing required field: name")
	}
	// Check if group already exists
	_, err := uc.repo.GetGroupByName(req.Name)
	if err == nil {
		return CreateGroupResponse{}, errors.New("group already exists")
	}
	group := acl.Group{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := uc.repo.CreateGroup(group); err != nil {
		return CreateGroupResponse{}, err
	}
	return CreateGroupResponse{Group: group}, nil
}
