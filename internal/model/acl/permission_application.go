package model

import "time"

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
type PermissionApplicationReviewer struct {
	ID            string    // Unique identifier for the mapping
	ApplicationID string    // Reference to PermissionApplication.ID
	ReviewerID    string    // Reference to User.ID (the reviewer)
	ReviewStatus  string    // Status (e.g., pending, approved, rejected)
	ReviewComment string    // Optional comment from the reviewer
	ReviewedAt    time.Time // When the review was made
	CreatedAt     time.Time // When the mapping was created
	UpdatedAt     time.Time // Last update timestamp
}

// PermissionApplicationHistory tracks the history of actions taken on a permission application.
type PermissionApplicationHistory struct {
	ID            string    // Unique identifier for the history record
	ApplicationID string    // Reference to PermissionApplication.ID
	Action        string    // Action taken (e.g., submitted, reviewed, approved, rejected)
	ActorID       string    // Reference to User.ID (who performed the action)
	Comment       string    // Optional comment or reason for the action
	CreatedAt     time.Time // When the action was taken
	UpdatedAt     time.Time // Last update timestamp
}
