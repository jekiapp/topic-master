// this is for updating a group by id
// only allow update for the description only
// check delete_group.go for the reference
// only root group can do this

package group

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jekiapp/nsqper/internal/model/acl"
	grouprepo "github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/jekiapp/nsqper/pkg/util"
	"github.com/tidwall/buntdb"
)

type UpdateGroupByIDRequest struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type UpdateGroupByIDResponse struct {
	Group acl.Group
}

type iUpdateGroupRepo interface {
	GetGroupByID(id string) (acl.Group, error)
	UpdateGroup(group acl.Group) error
}

type updateGroupRepo struct {
	db *buntdb.DB
}

func (r *updateGroupRepo) GetGroupByID(id string) (acl.Group, error) {
	return grouprepo.GetGroupByID(r.db, id)
}

func (r *updateGroupRepo) UpdateGroup(group acl.Group) error {
	return grouprepo.UpdateGroup(r.db, group)
}

type UpdateGroupByIDUsecase struct {
	repo iUpdateGroupRepo
}

func NewUpdateGroupByIDUsecase(db *buntdb.DB) UpdateGroupByIDUsecase {
	return UpdateGroupByIDUsecase{
		repo: &updateGroupRepo{db: db},
	}
}

func (uc UpdateGroupByIDUsecase) Handle(ctx context.Context, req UpdateGroupByIDRequest) (UpdateGroupByIDResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return UpdateGroupByIDResponse{}, errors.New("unauthorized: user info not found")
	}
	isRoot := false
	for _, g := range user.Groups {
		if g.GroupName == acl.GroupRoot {
			isRoot = true
			break
		}
	}
	if !isRoot {
		return UpdateGroupByIDResponse{}, errors.New("forbidden: only root group can update groups")
	}
	if req.ID == "" {
		return UpdateGroupByIDResponse{}, errors.New("missing required field: id")
	}
	group, err := uc.repo.GetGroupByID(req.ID)
	if err != nil {
		return UpdateGroupByIDResponse{}, fmt.Errorf("failed to get group by id: %w", err)
	}
	group.Description = req.Description
	group.UpdatedAt = time.Now()
	if err := uc.repo.UpdateGroup(group); err != nil {
		return UpdateGroupByIDResponse{}, err
	}
	return UpdateGroupByIDResponse{Group: group}, nil
}
