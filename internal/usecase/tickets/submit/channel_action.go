package submit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	auth "github.com/jekiapp/topic-master/internal/logic/auth"
	usergrouplogic "github.com/jekiapp/topic-master/internal/logic/user_group"
	"github.com/jekiapp/topic-master/internal/model/acl"
	apprepo "github.com/jekiapp/topic-master/internal/repository/application"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

// Usecase struct

type ChannelActionSubmitUsecase struct {
	repo *channelActionRepo
}

func NewChannelActionSubmitUsecase(db *buntdb.DB) ChannelActionSubmitUsecase {
	return ChannelActionSubmitUsecase{
		repo: &channelActionRepo{db: db},
	}
}

func (uc ChannelActionSubmitUsecase) validateRequestedPermissions(ctx context.Context, userID, entityID string, requestedPerms []string) error {
	// For each requested permission, check if the user already has it for the entity
	var alreadyOwned []string
	for _, perm := range requestedPerms {
		pm, err := entityrepo.GetPermissionMapByActionEntityUser(uc.repo.db, userID, entityID, perm)
		if err == nil && pm.UserID == userID {
			alreadyOwned = append(alreadyOwned, perm)
		}
	}
	if len(alreadyOwned) > 0 {
		return errors.New("user already has permissions: " + strings.Join(alreadyOwned, ", "))
	}
	return nil
}

func (uc ChannelActionSubmitUsecase) Handle(ctx context.Context, req SubmitApplicationRequest) (SubmitApplicationResponse, error) {
	// Extract userID from context (JWTClaims)
	user := util.GetUserInfo(ctx)
	if user == nil {
		return SubmitApplicationResponse{}, errors.New("user unathorized")
	}

	// Load entity to get group owner
	entity, err := entityrepo.GetEntityByID(uc.repo.db, req.EntityID)
	if err != nil {
		return SubmitApplicationResponse{}, err
	}

	// Validate requested permissions
	if err := uc.validateRequestedPermissions(ctx, user.ID, req.EntityID, req.Permission); err != nil {
		return SubmitApplicationResponse{}, err
	}

	// get group by name
	group, err := uc.repo.GetGroupByName(entity.GroupOwner)
	if err != nil {
		return SubmitApplicationResponse{}, errors.New("reviewer group not found: " + entity.GroupOwner)
	}
	// Use group ID as reviewer group
	reviewerGroupID := group.ID
	input := auth.CreateApplicationInput{
		Title:              fmt.Sprintf("Application to action for channel %s", entity.Name),
		ApplicationType:    req.ApplicationType,
		PermissionIDs:      req.Permission,
		Reason:             req.Reason,
		ReviewerGroupID:    reviewerGroupID,
		MetaData:           map[string]string{"entity_id": req.EntityID},
		HistoryInitAction:  "Apply channel action permission",
		HistoryInitComment: req.Reason,
	}
	out, err := auth.CreateApplication(ctx, input, uc.repo)
	if err != nil {
		return SubmitApplicationResponse{}, err
	}
	appURL := fmt.Sprintf("#ticket-detail?id=%s", out.ApplicationID)
	return SubmitApplicationResponse{
		AppID:  out.ApplicationID,
		AppURL: appURL,
	}, nil
}

type channelActionRepo struct {
	db *buntdb.DB
}

func (r *channelActionRepo) CreateApplication(app acl.Application) error {
	return apprepo.CreateApplication(r.db, app)
}

func (r *channelActionRepo) GetReviewerIDsByGroupID(groupID string) ([]string, error) {
	return usergrouplogic.GetReviewerIDsByGroupID(r.db, groupID)
}

func (r *channelActionRepo) CreateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return apprepo.CreateApplicationAssignment(r.db, assignment)
}

func (r *channelActionRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return apprepo.CreateApplicationHistory(r.db, history)
}

func (r *channelActionRepo) GetGroupByName(name string) (acl.Group, error) {
	return userrepo.GetGroupByName(r.db, name)
}
