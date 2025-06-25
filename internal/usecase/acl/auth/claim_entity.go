package acl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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
}

type iClaimEntityRepo interface {
	CreateApplication(app acl.Application) error
	GetGroupByName(name string) (acl.Group, error)
	GetUserGroup(userID, groupID string) (acl.UserGroup, error)
	ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error)
	GetAdminUserIDsByGroupID(groupID string) ([]string, error)
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

func (r *claimEntityRepo) GetAdminUserIDsByGroupID(groupID string) ([]string, error) {
	return userrepo.GetAdminUserIDsByGroupID(r.db, groupID)
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

	// Create application
	app := &acl.Application{
		ID:            uuid.NewString(),
		Title:         fmt.Sprintf("Claim %s:%s for group %s", entityObj.TypeID, entityObj.Name, req.GroupName),
		UserID:        user.ID,
		PermissionIDs: []string{"claim:" + req.EntityID},
		Reason:        req.Reason,
		Status:        acl.StatusWaitingForApproval,
		MetaData:      map[string]string{req.EntityID + ":group_name": req.GroupName},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := uc.repo.CreateApplication(*app); err != nil {
		return ClaimEntityResponse{}, err
	}
	// Approvers: admins of group + root group members
	rootGroup, err := uc.repo.GetGroupByName(acl.GroupRoot)
	if err != nil {
		return ClaimEntityResponse{}, errors.New("root group not found")
	}
	rootMembers, err := uc.repo.ListUserGroupsByGroupID(rootGroup.ID, 0)
	if err != nil {
		return ClaimEntityResponse{}, errors.New("failed to list root group members")
	}
	adminUserIDs, err := uc.repo.GetAdminUserIDsByGroupID(group.ID)
	if err != nil {
		return ClaimEntityResponse{}, errors.New("failed to get admin user ids")
	}
	for _, member := range rootMembers {
		adminUserIDs = append(adminUserIDs, member.UserID)
	}
	for _, userID := range adminUserIDs {
		assignment := &acl.ApplicationAssignment{
			ID:            uuid.NewString(),
			ApplicationID: app.ID,
			ReviewerID:    userID,
			ReviewStatus:  acl.ActionWaitingForApproval,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := uc.repo.CreateApplicationAssignment(*assignment); err != nil {
			return ClaimEntityResponse{}, err
		}
	}
	// Insert application history
	history := &acl.ApplicationHistory{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		Action:        "Create claim entity ticket",
		ActorID:       user.ID,
		Comment:       fmt.Sprintf("Initial claim %s %s by %s", entityObj.TypeID, entityObj.Name, user.Name),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_ = uc.repo.CreateApplicationHistory(*history)
	return ClaimEntityResponse{ApplicationID: app.ID}, nil
}
