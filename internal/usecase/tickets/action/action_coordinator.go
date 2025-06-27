package action

import (
	"context"
	"errors"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ActionCoordinator struct {
	repo               iActionCoordinatorRepo
	signupHandler      *SignupHandler // placeholder, defined in signup_handler.go
	claimEntityHandler *ClaimEntityHandler
	topicActionHandler *TopicActionHandler
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
		topicActionHandler: NewTopicActionHandler(db),
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

	switch app.Type {
	case acl.ApplicationType_Signup:
		return ac.signupHandler.HandleSignup(ctx, req)
	case acl.ApplicationType_Claim:
		return ac.claimEntityHandler.HandleClaimEntity(ctx, ClaimEntityInput{
			Action:      req.Action,
			AppID:       req.ApplicationID,
			Assignments: assignments,
		})
	case acl.ApplicationType_TopicForm:
		ac.topicActionHandler.HandleTopicAction(ctx, TopicActionInput{
			Action:      req.Action,
			AppID:       req.ApplicationID,
			Assignments: assignments,
		})
		if err != nil {
			return ActionResponse{}, err
		}
	}

	return ActionResponse{}, errors.New("application type not supported")
}

type iActionCoordinatorRepo interface {
	GetApplicationByID(id string) (acl.Application, error)
	GetPermissionByID(id string) (acl.PermissionMap, error)
	ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error)
}

type actionCoordinatorRepo struct {
	db *buntdb.DB
}

func (r *actionCoordinatorRepo) GetApplicationByID(id string) (acl.Application, error) {
	return db.GetByID[acl.Application](r.db, id)
}

func (r *actionCoordinatorRepo) GetPermissionByID(id string) (acl.PermissionMap, error) {
	return db.GetByID[acl.PermissionMap](r.db, id)
}

func (r *actionCoordinatorRepo) ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error) {
	return db.SelectAll[acl.ApplicationAssignment](r.db, "="+appID, acl.IdxAppAssign_ApplicationID)
}
