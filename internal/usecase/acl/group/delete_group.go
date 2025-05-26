package group

import (
	"context"
	"errors"

	"github.com/jekiapp/nsqper/internal/model/acl"
	grouprepo "github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/jekiapp/nsqper/pkg/util"
	"github.com/tidwall/buntdb"
)

type DeleteGroupRequest struct {
	ID string `json:"id"`
}

type DeleteGroupResponse struct {
	Success bool `json:"success"`
}

type iDeleteGroupRepo interface {
	DeleteGroupByID(id string) error
	GetGroupByID(id string) (acl.Group, error)
}

type deleteGroupRepo struct {
	db *buntdb.DB
}

func (r *deleteGroupRepo) DeleteGroupByID(id string) error {
	return grouprepo.DeleteGroupByID(r.db, id)
}

func (r *deleteGroupRepo) GetGroupByID(id string) (acl.Group, error) {
	return grouprepo.GetGroupByID(r.db, id)
}

type DeleteGroupUsecase struct {
	repo iDeleteGroupRepo
}

func NewDeleteGroupUsecase(db *buntdb.DB) DeleteGroupUsecase {
	return DeleteGroupUsecase{
		repo: &deleteGroupRepo{db: db},
	}
}

func (uc DeleteGroupUsecase) Handle(ctx context.Context, req DeleteGroupRequest) (DeleteGroupResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return DeleteGroupResponse{Success: false}, errors.New("unauthorized: user info not found")
	}
	isRoot := false
	for _, g := range user.Groups {
		if g.GroupName == acl.GroupRoot {
			isRoot = true
			break
		}
	}
	if !isRoot {
		return DeleteGroupResponse{Success: false}, errors.New("forbidden: only root group can delete groups")
	}
	if req.ID == "" {
		return DeleteGroupResponse{Success: false}, errors.New("missing required field: id")
	}

	// get group by id
	// if group is root, return error
	group, err := uc.repo.GetGroupByID(req.ID)
	if err != nil {
		return DeleteGroupResponse{Success: false}, err
	}
	if group.Name == acl.GroupRoot {
		return DeleteGroupResponse{Success: false}, errors.New("forbidden: root group cannot be deleted")
	}

	err = uc.repo.DeleteGroupByID(req.ID)
	if err != nil {
		return DeleteGroupResponse{Success: false}, err
	}
	return DeleteGroupResponse{Success: true}, nil
}
