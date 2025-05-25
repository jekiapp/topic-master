package acl

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/nsqper/pkg/db"
	"github.com/tidwall/buntdb"
)

// GroupRole represents a user's role in a group
type GroupRole struct {
	GroupName string // Name of the group
	Role      string // Role of the user in the group (e.g., admin, user)
}

// User represents a system user (master)
type User struct {
	ID        string      // Unique identifier (e.g., UUID or string key)
	Username  string      // Username
	Name      string      // Display Name
	Password  string      // Password hash
	Email     string      // Email address
	Phone     string      // Phone number
	Type      string      // Type (e.g., admin, user, etc.)
	Status    string      // Status (e.g., active, inactive, etc.)
	CreatedAt time.Time   // Creation timestamp
	UpdatedAt time.Time   // Last update timestamp
	Groups    []GroupRole // List of groups and roles
}

const (
	TableUser        = "user"
	IdxUser_Type     = TableUser + ":type"
	IdxUser_Status   = TableUser + ":status"
	IdxUser_Username = TableUser + ":username"
	IdxUser_Email    = TableUser + ":email"
)

func (u User) GetPrimaryKey() string {
	id := u.ID
	if id == "" {
		id = uuid.NewString()
	}
	return fmt.Sprintf("%s:%s", TableUser, id)
}

func (u User) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxUser_Type,
			Pattern: fmt.Sprintf("%s:*:%s", TableUser, "type"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxUser_Username,
			Pattern: fmt.Sprintf("%s:*:%s", TableUser, "username"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxUser_Email,
			Pattern: fmt.Sprintf("%s:*:%s", TableUser, "email"),
			Type:    buntdb.IndexString,
		},
	}
}

func (u User) GetIndexValues() map[string]string {
	return map[string]string{
		"type":     u.Type,
		"username": u.Username,
		"email":    u.Email,
	}
}

// Authorization maps a user to a permission (for access control checks)
type Authorization struct {
	ID           string // Unique identifier for the mapping
	UserID       string // Reference to User.ID
	PermissionID string // Reference to Permission.ID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
