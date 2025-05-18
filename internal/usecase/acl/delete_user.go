package acl

import (
	"context"

	"github.com/tidwall/buntdb"
)

type DeleteUserRequest struct {
	UserID string
}

type DeleteUserResponse struct {
	Success bool
}

type iUserDeleteRepo interface {
	DeleteUser(userID string) error
}

type userDeleteRepo struct {
	db *buntdb.DB
}

func (r *userDeleteRepo) DeleteUser(userID string) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		// Find the user key by userID (UUID)
		var userKey string
		tx.AscendKeys("user:*", func(key, value string) bool {
			// The value is a CSV, the first field is username, but we need to match by userID
			// For this example, let's assume the key is not the UUID, so we need to scan values
			// In production, store by UUID as key for efficiency
			if value[:len(userID)] == userID {
				userKey = key
				return false
			}
			return true
		})
		if userKey == "" {
			return buntdb.ErrNotFound
		}
		_, err := tx.Delete(userKey)
		return err
	})
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
