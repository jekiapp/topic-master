package acl

import (
	"context"
	"errors"

	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type AuthorizationInput struct {
	EntityID string `json:"entity_id"`
	Action   string `json:"action"`
}

type AuthorizationResponse struct {
	Authorized bool   `json:"authorized"`
	Message    string `json:"message"`
}

type iAuthorizationRepo interface {
	GetEntityByID(db *buntdb.DB, entityID string) (entitymodel.Entity, error)
}

type AuthorizationUsecase struct {
	repo iAuthorizationRepo
	db   *buntdb.DB
}

func NewAuthorizationUsecase(db *buntdb.DB, repo iAuthorizationRepo) AuthorizationUsecase {
	return AuthorizationUsecase{
		repo: repo,
		db:   db,
	}
}

func (uc AuthorizationUsecase) Check(ctx context.Context, input AuthorizationInput) (AuthorizationResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return AuthorizationResponse{Authorized: false, Message: "User not found in context"}, errors.New("user not found in context")
	}
	entity, err := uc.repo.GetEntityByID(uc.db, input.EntityID)
	if err != nil {
		return AuthorizationResponse{Authorized: false, Message: "Entity not found"}, err
	}
	groupOwner := entity.GroupOwner
	if groupOwner == "" {
		return AuthorizationResponse{Authorized: false, Message: "Entity has no group owner"}, nil
	}
	for _, group := range user.Groups {
		if group.GroupID == groupOwner {
			return AuthorizationResponse{Authorized: true, Message: "User is authorized"}, nil
		}
	}
	return AuthorizationResponse{Authorized: false, Message: "User is not a member of the group owner"}, nil
}
