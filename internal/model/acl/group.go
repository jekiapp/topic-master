package acl

import (
	"fmt"
	"time"

	"github.com/jekiapp/topic-master/pkg/db"
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
	GroupRoot  = "root"
	TableGroup = "group"

	IdxGroup_Name = TableGroup + ":name"

	RoleGroupAdmin  = "admin"
	RoleGroupMember = "member"
)

func (g Group) GetPrimaryKey() string {
	return fmt.Sprintf("%s:%s", TableGroup, g.ID)
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
	Role      string // Role of user group (admin, member, etc.)
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	TableUserGroup       = "user_group"
	IdxUserGroup_ID      = TableUserGroup + ":usergroup_id"
	IdxUserGroup_UserID  = TableUserGroup + ":user_id"
	IdxUserGroup_GroupID = TableUserGroup + ":group_id"
	IdxUserGroup_Role    = TableUserGroup + ":role"
)

func (ug UserGroup) GetPrimaryKey() string {
	return fmt.Sprintf("%s:%s", TableUserGroup, ug.ID)
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
			Name:    IdxUserGroup_Role,
			Pattern: fmt.Sprintf("%s:*:%s", TableUserGroup, "role"),
			Type:    buntdb.IndexString,
		},
	}
}

func (ug UserGroup) GetIndexValues() map[string]string {
	return map[string]string{
		"usergroup_id": fmt.Sprintf("%s:%s", ug.GroupID, ug.UserID),
		"user_id":      ug.UserID,
		"group_id":     ug.GroupID,
		"role":         ug.Role,
	}
}

func (g *Group) SetID(id string) {
	g.ID = id
}

func (ug *UserGroup) SetID(id string) {
	ug.ID = id
}
