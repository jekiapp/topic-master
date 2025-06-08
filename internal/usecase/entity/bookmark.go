// this is usecase for toggle bookmark true/false for particular user on particular entity
// it will receive entity id and bookmark:bool

// please create a new model in model/entity/bookmark.go for mapping entity:user bookmark
// learn from other model to see on how to implement model, also the indexing stuff

// create new bookmark repository in repository/entity/bookmark.go
// crete a new func for toggle true/false of user bookmark by deleting/creating the mapping record

package entity

import (
	"context"
	"errors"

	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ToggleBookmarkInput struct {
	EntityID string
	Bookmark bool
}

type ToggleBookmarkResponse struct {
	Message string `json:"message"`
}

type iBookmarkRepo interface {
	ToggleBookmark(entityID, userID string, bookmark bool) error
}

type ToggleBookmarkUsecase struct {
	repo iBookmarkRepo
	db   *buntdb.DB
}

func NewToggleBookmarkUsecase(db *buntdb.DB) ToggleBookmarkUsecase {
	return ToggleBookmarkUsecase{
		repo: &bookmarkRepo{db: db},
		db:   db,
	}
}

func (uc ToggleBookmarkUsecase) Toggle(ctx context.Context, input ToggleBookmarkInput) (ToggleBookmarkResponse, error) {
	userInfo := util.GetUserInfo(ctx)
	if userInfo == nil {
		return ToggleBookmarkResponse{Message: "User info not found in context"}, errors.New("user info not found in context")
	}
	err := uc.repo.ToggleBookmark(input.EntityID, userInfo.ID, input.Bookmark)
	if err != nil {
		return ToggleBookmarkResponse{Message: "Failed to toggle bookmark"}, err
	}
	return ToggleBookmarkResponse{Message: "Bookmark toggled successfully"}, nil
}

type bookmarkRepo struct {
	db *buntdb.DB
}

func (r *bookmarkRepo) ToggleBookmark(entityID, userID string, bookmark bool) error {
	return entityrepo.ToggleBookmark(r.db, entityID, userID, bookmark)
}
