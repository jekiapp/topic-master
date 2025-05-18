package model

import "time"

// User represents a system user (master)
type User struct {
	ID        string    // Unique identifier (e.g., UUID or string key)
	Username  string    // Username
	Name      string    // Display Name
	Password  string    // Password hash
	Email     string    // Email address
	Phone     string    // Phone number
	Type      string    // Type (e.g., admin, user, etc.)
	Status    string    // Status (e.g., active, inactive, etc.)
	CreatedAt time.Time // Creation timestamp
	UpdatedAt time.Time // Last update timestamp
}

// Group represents a user group or role (master)
type Group struct {
	ID        string // Unique identifier
	Name      string // Group name
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserGroup maps users to groups (many-to-many)
type UserGroup struct {
	ID        string // Unique identifier for the mapping
	UserID    string // Reference to User.ID
	GroupID   string // Reference to Group.ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Permission represents an action or resource (master)
type Permission struct {
	ID        string // Unique identifier
	Name      string // Permission name
	EntityID  string // Reference to Entity.ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GroupPermission maps groups to permissions (many-to-many)
type GroupPermission struct {
	ID           string // Unique identifier for the mapping
	GroupID      string // Reference to Group.ID
	PermissionID string // Reference to Permission.ID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Authorization maps a user to a permission (for access control checks)
type Authorization struct {
	ID           string // Unique identifier for the mapping
	UserID       string // Reference to User.ID
	PermissionID string // Reference to Permission.ID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
