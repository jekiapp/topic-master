// this usecase will return simple []Group{id,name} of group
// learn from get_group_list.go this will serve GET query
// attach to handler.go

package group

import (
	"context"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/tidwall/buntdb"
)

type GetGroupListSimpleRequest struct{}

type GetGroupListSimpleResponse struct {
	Groups []GroupSimpleItem `json:"groups"`
}

type GroupSimpleItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type iGroupListSimpleRepo interface {
	GetAllGroups() ([]acl.Group, error)
}

type GetGroupListSimpleUsecase struct {
	groupRepo iGroupListSimpleRepo
}

func NewGetGroupListSimpleUsecase(db *buntdb.DB) GetGroupListSimpleUsecase {
	return GetGroupListSimpleUsecase{
		groupRepo: &groupRepoImpl{db: db},
	}
}

func (uc GetGroupListSimpleUsecase) Handle(ctx context.Context, req GetGroupListSimpleRequest) (GetGroupListSimpleResponse, error) {
	groups, err := uc.groupRepo.GetAllGroups()
	if err != nil {
		return GetGroupListSimpleResponse{}, err
	}
	var result []GroupSimpleItem
	for _, g := range groups {
		result = append(result, GroupSimpleItem{
			ID:   g.ID,
			Name: g.Name,
		})
	}
	return GetGroupListSimpleResponse{Groups: result}, nil
}
