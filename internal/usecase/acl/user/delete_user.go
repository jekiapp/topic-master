package user

import (
	"context"

	"github.com/jekiapp/topic-master/internal/model/acl"
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
}

type userDeleteRepo struct {
	db *buntdb.DB
}

func (r *userDeleteRepo) DeleteUser(userID string) error {
	return db.DeleteByID(r.db, acl.User{ID: userID})
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
	return DeleteUserResponse{Success: true}, nil
}
