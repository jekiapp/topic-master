package entity

import (
	"context"
	"errors"
	"fmt"

	"github.com/jekiapp/topic-master/internal/logic/auth"
	usergrouplogic "github.com/jekiapp/topic-master/internal/logic/user_group"
	"github.com/jekiapp/topic-master/internal/model/acl"
	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/jekiapp/topic-master/pkg/db"
	util "github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ClaimEntityRequest struct {
	EntityID  string `json:"entity_id"`
	GroupName string `json:"group_name"`
	Reason    string `json:"reason"`
}

type ClaimEntityResponse struct {
	ApplicationID string `json:"application_id"`
	LinkRedirect  string `json:"link_redirect"`
}

type iClaimEntityRepo interface {
	CreateApplication(app acl.Application) error
	GetGroupByName(name string) (acl.Group, error)
	GetUserGroup(userID, groupID string) (acl.UserGroup, error)
	ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error)
	GetReviewerIDsByGroupID(groupID string) ([]string, error)
	CreateApplicationAssignment(assignment acl.ApplicationAssignment) error
	CreateApplicationHistory(history acl.ApplicationHistory) error
	GetEntityByID(entityID string) (entitymodel.Entity, error)
}

type claimEntityRepo struct {
	db *buntdb.DB
}

func (r *claimEntityRepo) CreateApplication(app acl.Application) error {
	return db.Insert(r.db, &app)
}

func (r *claimEntityRepo) GetGroupByName(name string) (acl.Group, error) {
	return userrepo.GetGroupByName(r.db, name)
}

func (r *claimEntityRepo) GetUserGroup(userID, groupID string) (acl.UserGroup, error) {
	return userrepo.GetUserGroup(r.db, userID, groupID)
}

func (r *claimEntityRepo) ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error) {
	return userrepo.ListUserGroupsByGroupID(r.db, groupID, limit)
}

func (r *claimEntityRepo) GetReviewerIDsByGroupID(groupID string) ([]string, error) {
	return usergrouplogic.GetReviewerIDsByGroupID(r.db, groupID)
}

func (r *claimEntityRepo) CreateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return db.Insert(r.db, &assignment)
}

func (r *claimEntityRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return db.Insert(r.db, &history)
}

func (r *claimEntityRepo) GetEntityByID(entityID string) (entitymodel.Entity, error) {
	entityObj, err := entityrepo.GetEntityByID(r.db, entityID)
	if err != nil {
		return entitymodel.Entity{}, err
	}
	return entityObj, nil
}

type ClaimEntityUsecase struct {
	repo iClaimEntityRepo
}

func NewClaimEntityUsecase(db *buntdb.DB) ClaimEntityUsecase {
	return ClaimEntityUsecase{
		repo: &claimEntityRepo{db: db},
	}
}

func (r ClaimEntityRequest) Validate() error {
	if r.EntityID == "" {
		return errors.New("missing entity_id")
	}
	if r.GroupName == "" {
		return errors.New("missing group_name")
	}
	return nil
}

func (uc ClaimEntityUsecase) Handle(ctx context.Context, req ClaimEntityRequest) (ClaimEntityResponse, error) {
	if err := req.Validate(); err != nil {
		return ClaimEntityResponse{}, err
	}
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ClaimEntityResponse{}, errors.New("user not found in context")
	}
	group, err := uc.repo.GetGroupByName(req.GroupName)
	if err != nil {
		return ClaimEntityResponse{}, errors.New("group not found")
	}
	// Validate user is a member of the group
	_, err = uc.repo.GetUserGroup(user.ID, group.ID)
	if err != nil {
		return ClaimEntityResponse{}, errors.New("user is not a member of the group")
	}
	// get entity by id , then use the entity name as the title
	entityObj, err := uc.repo.GetEntityByID(req.EntityID)
	if err != nil {
		return ClaimEntityResponse{}, errors.New("entity not found")
	}

	input := auth.CreateApplicationInput{
		Title:              fmt.Sprintf("Claim %s:%s for group %s", entityObj.TypeID, entityObj.Name, req.GroupName),
		ApplicationType:    acl.ApplicationType_Claim,
		PermissionIDs:      []string{acl.Permission_Claim_Entity.Name},
		Reason:             req.Reason,
		ReviewerGroupID:    group.ID,
		MetaData:           map[string]string{"group_id": group.ID, "entity_id": req.EntityID},
		HistoryInitAction:  "Create claim ticket",
		HistoryInitComment: fmt.Sprintf("Initial claim %s %s for group %s", entityObj.TypeID, entityObj.Name, req.GroupName),
	}
	out, err := auth.CreateApplication(ctx, input, uc.repo)
	if err != nil {
		return ClaimEntityResponse{}, err
	}
	linkRedirect := fmt.Sprintf("/#ticket-detail?id=%s", out.ApplicationID)
	return ClaimEntityResponse{ApplicationID: out.ApplicationID, LinkRedirect: linkRedirect}, nil
}
