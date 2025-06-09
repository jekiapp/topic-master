package acl

import (
	"fmt"
	"time"

	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

// Permission represents an action or resource (master)
type Permission struct {
	ID          string // UUID
	Name        string // Permission action name (publish, tail, delete, etc)
	EntityID    string // Reference to Entity.ID
	Type        string // Type of the permission (e.g. "group", "user")
	Description string // Description of the permission ("publishing topic", "tailing topic", etc)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	TablePermission          = "permission"
	IdxPermission_Type       = TablePermission + ":type"
	IdxPermission_NameEntity = TablePermission + ":name_entity"
)

func (p Permission) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxPermission_Type,
			Pattern: fmt.Sprintf("%s:*:%s", TablePermission, "type"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxPermission_NameEntity,
			Pattern: fmt.Sprintf("%s:*:%s", TablePermission, "name_entity"),
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
		"type":        p.Type,
		"name_entity": fmt.Sprintf("%s:%s", p.Name, p.EntityID),
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
