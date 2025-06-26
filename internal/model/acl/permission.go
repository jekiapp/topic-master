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

// Permission represents an action or resource (master)
type Permission struct {
	ID        string // UUID
	Action    string // Permission action name (publish, tail, delete, etc)
	UserID    string // Reference to User.ID
	EntityID  string // Reference to Entity.ID
	CreatedAt time.Time
}

const (
	TablePermission                = "permission"
	IdxPermission_Entity           = TablePermission + ":entity"
	IdxPermission_ActionEntityUser = TablePermission + ":action_entity_user"
)

func (p Permission) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxPermission_ActionEntityUser,
			Pattern: fmt.Sprintf("%s:*:%s", TablePermission, "action_entity_user"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxPermission_Entity,
			Pattern: fmt.Sprintf("%s:*:%s", TablePermission, "entity"),
			Type:    buntdb.IndexString,
		},
	}
}

func (p *Permission) GetPrimaryKey(id string) string {
	if p.ID == "" && id != "" {
		p.ID = id
	}
	return fmt.Sprintf("%s:%s", TablePermission, p.ID)
}

func (p Permission) GetIndexValues() map[string]string {
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

func (p *Permission) SetID(id string) {
	p.ID = id
}
