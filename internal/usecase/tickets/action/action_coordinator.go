// I want this usecase to receive json parameter {action: "approve", application_id: "123"}
// then in the logic it will check each permission of the application
// if the permission name contains "signup" then it will be handled by the signup usecase in signup_handler.go
// generate placeholder structure for signup_handler.go
// the signup_handler.go is in the same directory as this file

package action

import (
	"context"
	"errors"
	"strings"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ActionCoordinator struct {
	repo               iActionCoordinatorRepo
	signupHandler      *SignupHandler // placeholder, defined in signup_handler.go
	claimEntityHandler *ClaimEntityHandler
}

type ActionRequest struct {
	Action        string `json:"action"`
	ApplicationID string `json:"application_id"`
}

type ActionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewActionCoordinator(db *buntdb.DB) *ActionCoordinator {
	return &ActionCoordinator{
		repo:               &actionCoordinatorRepo{db: db},
		signupHandler:      NewSignupHandler(db),
		claimEntityHandler: NewClaimEntityHandler(db),
	}
}

func (ac *ActionCoordinator) validateActor(ctx context.Context, appID string) ([]acl.ApplicationAssignment, error) {
	// validate the actor
	user := util.GetUserInfo(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}
	assignments, err := ac.repo.ListAssignmentsByApplicationID(appID)
	if err != nil {
		return nil, err
	}

	isAssignee := false
	for _, assignment := range assignments {
		if assignment.ReviewerID == user.ID {
			isAssignee = true
			break
		}
	}
	if !isAssignee {
		return nil, errors.New("you are not eligible to perform this action")
	}
	return assignments, nil
}

func (ac *ActionCoordinator) Handle(ctx context.Context, req ActionRequest) (ActionResponse, error) {
	app, err := ac.repo.GetApplicationByID(req.ApplicationID)
	if err != nil {
		return ActionResponse{}, err
	}

	assignments, err := ac.validateActor(ctx, req.ApplicationID)
	if err != nil {
		return ActionResponse{}, err
	}

	for _, permID := range app.PermissionIDs {
		if strings.HasPrefix(permID, "signup") {
			return ac.signupHandler.HandleSignup(ctx, req)
		}

		if strings.HasPrefix(permID, "claim") {
			permIDsplit := strings.Split(permID, ":")
			if len(permIDsplit) != 2 {
				return ActionResponse{}, errors.New("invalid permission id")
			}
			entityID := permIDsplit[1]

			groupName := app.MetaData[entityID+":group_name"]

			err := ac.claimEntityHandler.HandleClaimEntity(ctx, ClaimEntityInput{
				Action:      req.Action,
				AppID:       req.ApplicationID,
				EntityID:    entityID,
				GroupName:   groupName,
				Assignments: assignments,
			})
			if err != nil {
				return ActionResponse{}, err
			}
			continue
		}

	}
	return ActionResponse{Status: "success", Message: "Action completed"}, nil
}

type iActionCoordinatorRepo interface {
	GetApplicationByID(id string) (acl.Application, error)
	GetPermissionByID(id string) (acl.Permission, error)
	ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error)
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

func (r *actionCoordinatorRepo) ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error) {
	return db.SelectAll[acl.ApplicationAssignment](r.db, "="+appID, acl.IdxAppAssign_ApplicationID)
}
