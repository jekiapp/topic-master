package acl

import (
	"fmt"
	"time"

	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

const (
	PermissionAction_Publish = "publish"
	PermissionAction_Tail    = "tail"
	PermissionAction_Delete  = "delete"
	PermissionAction_Empty   = "empty"
	PermissionAction_Pause   = "pause"
)

// PermissionMap represents an action or resource (master)
type PermissionMap struct {
	ID        string // UUID
	Action    string // Permission action name (publish, tail, delete, etc)
	UserID    string // Reference to User.ID
	EntityID  string // Reference to Entity.ID
	CreatedAt time.Time
}

const (
	TablePermissionMap                = "permission_map"
	IdxPermissionMap_Entity           = TablePermissionMap + ":entity"
	IdxPermissionMap_ActionEntityUser = TablePermissionMap + ":action_entity_user"
)

func (p PermissionMap) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxPermissionMap_ActionEntityUser,
			Pattern: fmt.Sprintf("%s:*:%s", TablePermissionMap, "action_entity_user"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxPermissionMap_Entity,
			Pattern: fmt.Sprintf("%s:*:%s", TablePermissionMap, "entity"),
			Type:    buntdb.IndexString,
		},
	}
}

func (p *PermissionMap) GetPrimaryKey(id string) string {
	if p.ID == "" && id != "" {
		p.ID = id
	}
	return fmt.Sprintf("%s:%s", TablePermissionMap, p.ID)
}

func (p PermissionMap) GetIndexValues() map[string]string {
	return map[string]string{
		"entity":             p.EntityID,
		"action_entity_user": fmt.Sprintf("%s:%s:%s", p.Action, p.EntityID, p.UserID),
	}
}

// GroupPermission maps groups to permissions (many-to-many)
type GroupPermission struct {
	ID           string // UUID
	GroupID      string // Reference to Group.ID
	PermissionID string // Reference to Permission.ID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (gp GroupPermission) GetPrefix() string {
	return "group_permission:"
}

func (gp GroupPermission) GetKey() string {
	return fmt.Sprintf("%s%s:%s", gp.GetPrefix(), gp.GroupID, gp.PermissionID)
}

func (p *PermissionMap) SetID(id string) {
	p.ID = id
}
