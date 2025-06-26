package auth

import (
	"errors"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/model/entity"
)

// EntityByIDFetcher abstracts fetching an entity by ID.
type ICheckUserActionPermission interface {
	GetEntityByID(id string) (*entity.Entity, error)
	GetGroupsByUserID(userID string) ([]acl.GroupRole, error)
	GetPermissionByActionEntity(userID, entityID, action string) (acl.Permission, error)
}

// CheckUserEntityActionPermission checks if a user can perform an action on an entity.
func CheckUserActionPermission(user acl.User, entityID string, action string, deps ICheckUserActionPermission) error {
	// 1. Fetch entity by ID
	ent, err := deps.GetEntityByID(entityID)
	if err != nil {
		return err
	}

	// 2. If entity GroupOwner is empty or None, allow
	if ent.GroupOwner == "" || ent.GroupOwner == entity.GroupNone {
		return nil
	}

	// 3. If owner is not nil, check group membership
	userGroups := user.Groups
	if len(userGroups) == 0 {
		userGroups, err = deps.GetGroupsByUserID(user.ID)
		if err != nil {
			return err
		}
	}

	// 4. If entity is owned by a group and user is a member, allow
	for _, g := range userGroups {
		if ent.GroupOwner == g.GroupName {
			return nil
		}
	}

	// 5. Query permission by action/entity/user
	perm, err := deps.GetPermissionByActionEntity(user.ID, entityID, action)
	if err == nil && perm.UserID == user.ID {
		return nil
	}

	// 6. Deny if all checks fail
	return errors.New("permission denied")
}
