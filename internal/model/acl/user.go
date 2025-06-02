package acl

import (
	"fmt"
	"time"

	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

// GroupRole represents a user's role in a group
type GroupRole struct {
	GroupID   string `json:"group_id"`   // ID of the group
	GroupName string `json:"group_name"` // Name of the group
	Role      string `json:"role"`       // Role of the user in the group (e.g., admin, user)
}

const (
	StatusUserActive     = "active"
	StatusUserInApproval = "in_approval" // apply to signup, waiting for approval
	StatusUserPending    = "pending"     // created by root, but not yet activate
	StatusUserInactive   = "inactive"
)

// User represents a system user (master)
type User struct {
	ID        string      `json:"id"`         // Unique identifier (e.g., UUID or string key)
	Username  string      `json:"username"`   // Username
	Name      string      `json:"name"`       // Display Name
	Password  string      `json:"password"`   // Password hash
	Status    string      `json:"status"`     // Status (e.g., active, inactive, etc.)
	CreatedAt time.Time   `json:"created_at"` // Creation timestamp
	UpdatedAt time.Time   `json:"updated_at"` // Last update timestamp
	Groups    []GroupRole `json:"groups"`     // List of groups and roles
}

const (
	TableUser        = "user"
	IdxUser_Status   = TableUser + ":status"
	IdxUser_Username = TableUser + ":username"
)

func (u User) GetPrimaryKey() string {
	return fmt.Sprintf("%s:%s", TableUser, u.ID)
}

func (u User) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxUser_Username,
			Pattern: fmt.Sprintf("%s:*:%s", TableUser, "username"),
			Type:    buntdb.IndexString,
		},
	}
}

func (u User) GetIndexValues() map[string]string {
	return map[string]string{
		"username": u.Username,
	}
}

// Authorization maps a user to a permission (for access control checks)
type Authorization struct {
	ID           string    `json:"id"`            // Unique identifier for the mapping
	UserID       string    `json:"user_id"`       // Reference to User.ID
	PermissionID string    `json:"permission_id"` // Reference to Permission.ID
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) SetID(id string) {
	u.ID = id
}
