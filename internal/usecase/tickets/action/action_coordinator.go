// I want this usecase to receive json parameter {action: "approve", application_id: "123"}
// then in the logic it will check each permission of the application
// if the permission name contains "signup" then it will be handled by the signup usecase in signup_handler.go
// generate placeholder structure for signup_handler.go
// the signup_handler.go is in the same directory as this file

package action

import (
	"context"
	"strings"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type ActionCoordinator struct {
	repo          iActionCoordinatorRepo
	signupHandler *SignupHandler // placeholder, defined in signup_handler.go
}

type ActionRequest struct {
	Action        string `json:"action"`
	ApplicationID string `json:"application_id"`
}

type ActionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type iActionCoordinatorRepo interface {
	GetApplicationByID(id string) (acl.Application, error)
	GetPermissionByID(id string) (acl.Permission, error)
}

func NewActionCoordinator(db *buntdb.DB) *ActionCoordinator {
	return &ActionCoordinator{repo: &actionCoordinatorRepo{db: db}, signupHandler: NewSignupHandler(db)}
}

func (ac *ActionCoordinator) Handle(ctx context.Context, req ActionRequest) (ActionResponse, error) {
	app, err := ac.repo.GetApplicationByID(req.ApplicationID)
	if err != nil {
		return ActionResponse{}, err
	}
	for _, permID := range app.PermissionIDs {
		if strings.Contains(permID, "signup") {
			if ac.signupHandler != nil {
				return ac.signupHandler.HandleSignup(ctx, req)
			}
		}

		// perm, err := ac.repo.GetPermissionByID(permID)
		// if err != nil {
		// 	return ActionResponse{}, err
		// }

		// else: handle other permissions as needed (placeholder)
	}
	return ActionResponse{Status: "success", Message: "Action completed"}, nil
}

type actionCoordinatorRepo struct {
	db *buntdb.DB
}

func (r *actionCoordinatorRepo) GetApplicationByID(id string) (acl.Application, error) {
	return db.GetByID[acl.Application](r.db, id)
}

func (r *actionCoordinatorRepo) GetPermissionByID(id string) (acl.Permission, error) {
	return db.GetByID[acl.Permission](r.db, id)
}
