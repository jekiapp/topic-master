package acl

import "time"

// Permission represents an action or resource (master)
type Permission struct {
	// Unique identifier is the combination of Name:EntityID
	Name      string // Permission name
	EntityID  string // Reference to Entity.ID
	Type      string // Type of the permission (e.g. "group", "user")
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

// PermissionApplication represents a user's request to obtain a permission.
type PermissionApplication struct {
	ID           string    // Unique identifier for the application
	UserID       string    // Reference to User.ID (the applicant)
	PermissionID string    // Reference to Permission.ID (the requested permission)
	Reason       string    // Reason for the application
	Status       string    // Overall status (e.g., pending, approved, rejected)
	CreatedAt    time.Time // When the application was created
	UpdatedAt    time.Time // Last update timestamp
}

// PermissionApplicationReviewer links a permission application to a reviewer and their review status.
type ApplicationAssignment struct {
	ID            string    // Unique identifier for the mapping
	ApplicationID string    // Reference to PermissionApplication.ID
	ReviewerID    string    // Reference to User.ID (the reviewer)
	ReviewStatus  string    // Status (e.g., pending, approved, rejected)
	ReviewComment string    // Optional comment from the reviewer
	ReviewedAt    time.Time // When the review was made
	CreatedAt     time.Time // When the mapping was created
	UpdatedAt     time.Time // Last update timestamp
}

// ApplicationHistory tracks the history of actions taken on a permission application.
type ApplicationHistory struct {
	ID            string    // Unique identifier for the history record
	ApplicationID string    // Reference to PermissionApplication.ID
	Action        string    // Action taken (e.g., submitted, reviewed, approved, rejected)
	ActorID       string    // Reference to User.ID (who performed the action)
	Comment       string    // Optional comment or reason for the action
	CreatedAt     time.Time // When the action was taken
	UpdatedAt     time.Time // Last update timestamp
}
