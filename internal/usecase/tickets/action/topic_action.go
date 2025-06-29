package action

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/logic/auth"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type TopicActionInput struct {
	Action      string
	Application acl.Application
	Assignments []acl.ApplicationAssignment
}

type TopicActionHandler struct {
	repo iTopicActionRepo
}

func NewTopicActionHandler(db *buntdb.DB) *TopicActionHandler {
	return &TopicActionHandler{repo: &topicActionRepo{db: db}}
}

func (h *TopicActionHandler) HandleTopicAction(ctx context.Context, req TopicActionInput) (ActionResponse, error) {
	if req.Action == acl.ActionApprove {
		err := h.HandleApprove(ctx, req)
		if err != nil {
			return ActionResponse{}, err
		}
		return ActionResponse{
			Status:  "success",
			Message: "Permissions inserted for topic",
		}, nil
	} else if req.Action == acl.ActionReject {
		err := h.HandleReject(ctx, req)
		if err != nil {
			return ActionResponse{}, err
		}
		return ActionResponse{
			Status:  "rejected",
			Message: "Action rejected",
		}, nil
	}
	return ActionResponse{}, errors.New("invalid action")
}

func (h *TopicActionHandler) HandleApprove(ctx context.Context, input TopicActionInput) error {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return errors.New("user unathorized")
	}
	app := input.Application
	userID := app.UserID
	for _, permID := range app.PermissionIDs {
		perm := acl.PermissionMap{
			ID:        uuid.NewString(),
			UserID:    userID,
			EntityID:  app.MetaData["entity_id"],
			Action:    permID,
			CreatedAt: time.Now(),
		}
		if err := h.repo.InsertPermission(perm); err != nil {
			return err
		}
	}
	auth.ApproveApplication(ctx, h.repo, input.Application.ID, input.Assignments, "topic action permissions approved")
	return nil
}

func (h *TopicActionHandler) HandleReject(ctx context.Context, input TopicActionInput) error {
	auth.RejectApplication(ctx, h.repo, input.Application.ID, input.Assignments, "topic action permissions rejected")
	return nil
}

type iTopicActionRepo interface {
	auth.IApplicationAction
	InsertPermission(perm acl.PermissionMap) error
}

type topicActionRepo struct {
	db *buntdb.DB
}

func (r *topicActionRepo) InsertPermission(perm acl.PermissionMap) error {
	return db.Insert(r.db, &perm)
}

func (r *topicActionRepo) GetApplicationByID(id string) (acl.Application, error) {
	return db.GetByID[acl.Application](r.db, id)
}

func (r *topicActionRepo) UpdateApplication(app acl.Application) error {
	return db.Update(r.db, &app)
}

func (r *topicActionRepo) UpdateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return db.Update(r.db, &assignment)
}

func (r *topicActionRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return db.Insert(r.db, &history)
}
