package acl

import (
	"context"

	"github.com/jekiapp/topic-master/internal/logic/auth"
	authlogic "github.com/jekiapp/topic-master/internal/logic/auth"
	"github.com/jekiapp/topic-master/internal/model/acl"
	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
	entityrepo "github.com/jekiapp/topic-master/internal/repository/entity"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	util "github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type CheckActionAuthRequest struct {
	EntityID string `json:"entity_id"`
	Action   string `json:"action"`
}

type CheckActionAuthResponse struct {
	Allowed bool   `json:"allowed"`
	Error   string `json:"error,omitempty"`
}

type iCheckActionAuthRepo interface {
	authlogic.ICheckUserActionPermission
}

type checkActionAuthRepo struct {
	db *buntdb.DB
}

func (r *checkActionAuthRepo) GetEntityByID(id string) (*entitymodel.Entity, error) {
	entityObj, err := entityrepo.GetEntityByID(r.db, id)
	if err != nil {
		return nil, err
	}
	return &entityObj, nil
}

func (r *checkActionAuthRepo) GetGroupsByUserID(userID string) ([]acl.GroupRole, error) {
	return userrepo.ListGroupsForUser(r.db, userID)
}

func (r *checkActionAuthRepo) GetPermissionByActionEntity(userID, entityID, action string) (acl.PermissionMap, error) {
	return entityrepo.GetPermissionMapByActionEntityUser(r.db, userID, entityID, action)
}

type CheckActionAuthUsecase struct {
	repo iCheckActionAuthRepo
}

func NewCheckActionAuthUsecase(db *buntdb.DB) CheckActionAuthUsecase {
	return CheckActionAuthUsecase{
		repo: &checkActionAuthRepo{db: db},
	}
}

func (uc CheckActionAuthUsecase) Handle(ctx context.Context, req CheckActionAuthRequest) (CheckActionAuthResponse, error) {
	user := util.GetUserInfo(ctx)
	err := auth.CheckUserActionPermission(user, req.EntityID, req.Action, uc.repo)
	if err != nil {
		return CheckActionAuthResponse{Allowed: false, Error: err.Error()}, nil
	}
	return CheckActionAuthResponse{Allowed: true}, nil
}
