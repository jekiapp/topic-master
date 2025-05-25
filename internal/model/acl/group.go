package acl

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/nsqper/pkg/db"
	"github.com/tidwall/buntdb"
)

// Group represents a user group or role (master)
type Group struct {
	ID          string // Unique identifier
	Name        string // Group name
	Description string // Group description
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	TableGroup    = "group"
	IdxGroup_Name = TableGroup + ":name"
)

func (g Group) GetPrimaryKey() string {
	id := g.ID
	if id == "" {
		id = uuid.NewString()
	}
	return fmt.Sprintf("%s:%s", TableGroup, id)
}

func (g Group) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxGroup_Name,
			Pattern: fmt.Sprintf("%s:*:%s", TableGroup, "name"),
			Type:    buntdb.IndexString,
		},
	}
}

func (g Group) GetIndexValues() map[string]string {
	return map[string]string{
		"name": g.Name,
	}
}

// UserGroup maps users to groups (many-to-many)
type UserGroup struct {
	ID        string // Unique identifier for the mapping
	UserID    string // Reference to User.ID
	GroupID   string // Reference to Group.ID
	Type      string // Type of user group (admin, member, etc.)
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	TableUserGroup       = "user_group"
	IdxUserGroup_ID      = TableUserGroup + ":usergroup_id"
	IdxUserGroup_UserID  = TableUserGroup + ":user_id"
	IdxUserGroup_GroupID = TableUserGroup + ":group_id"
	IdxUserGroup_Type    = TableUserGroup + ":type"
)

func (ug UserGroup) GetPrimaryKey() string {
	id := ug.ID
	if id == "" {
		id = uuid.NewString()
	}
	return fmt.Sprintf("%s:%s", TableUserGroup, id)
}

func (ug UserGroup) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxUserGroup_ID,
			Pattern: fmt.Sprintf("%s:*:%s", TableUserGroup, "usergroup_id"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxUserGroup_UserID,
			Pattern: fmt.Sprintf("%s:*:%s", TableUserGroup, "user_id"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxUserGroup_GroupID,
			Pattern: fmt.Sprintf("%s:*:%s", TableUserGroup, "group_id"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxUserGroup_Type,
			Pattern: fmt.Sprintf("%s:*:%s", TableUserGroup, "type"),
			Type:    buntdb.IndexString,
		},
	}
}

func (ug UserGroup) GetIndexValues() map[string]string {
	return map[string]string{
		"usergroup_id": fmt.Sprintf("%s:%s", ug.GroupID, ug.UserID),
		"user_id":      ug.UserID,
		"group_id":     ug.GroupID,
		"type":         ug.Type,
	}
}
