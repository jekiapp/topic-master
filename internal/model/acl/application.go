package acl

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

const (
	TableApplication      = "application"
	IdxApplication_UserID = TableApplication + ":user_id"
	IdxApplication_Status = TableApplication + ":status"

	// Status constants
	StatusWaitingForApproval = "waiting for approval"
	StatusPending            = "pending"

	// Action constants
	ActionWaitingForApproval = "waiting for approval"

	// Actor constants
	ActorSystem = "system"
)

// Application represents a user's request to obtain a permission.
// or a new signup request
type Application struct {
	ID            string    // UUID
	Title         string    // Title of the application
	UserID        string    // Reference to User.ID (the applicant)
	PermissionIDs []string  // Reference to Permission.ID (the requested permission)
	Reason        string    // Reason for the application
	Status        string    // Overall status (e.g., pending, approved, rejected)
	CreatedAt     time.Time // When the application was created
	UpdatedAt     time.Time // Last update timestamp
}

func (a Application) GetPrimaryKey() string {
	id := a.ID
	if id == "" {
		id = uuid.NewString()
	}
	return fmt.Sprintf("%s:%s", TableApplication, id)
}

func (a Application) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxApplication_UserID,
			Pattern: fmt.Sprintf("%s:*:%s", TableApplication, "user_id"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxApplication_Status,
			Pattern: fmt.Sprintf("%s:*:%s", TableApplication, "status"),
			Type:    buntdb.IndexString,
		},
	}
}

func (a Application) GetIndexValues() map[string]string {
	return map[string]string{
		"user_id": a.UserID,
		"status":  a.Status,
	}
}

func (pa Application) GetPrefix() string {
	return "application:"
}

func (pa Application) GetKey() string {
	return fmt.Sprintf("%s%s:%s", pa.GetPrefix(), pa.UserID, pa.PermissionIDs)
}

// PermissionApplicationReviewer links a permission application to a reviewer and their review status.
const (
	TableApplicationAssignment = "app_assign"
	IdxAppAssign_ApplicationID = TableApplicationAssignment + ":application_id"
	IdxAppAssign_ReviewerID    = TableApplicationAssignment + ":reviewer_id"
	IdxAppAssign_ReviewStatus  = TableApplicationAssignment + ":review_status"
)

type ApplicationAssignment struct {
	ID            string    // UUID
	ApplicationID string    // Reference to Application.ID
	ReviewerID    string    // Reference to User.ID (the reviewer)
	ReviewStatus  string    // Status (e.g., pending, approved, rejected)
	ReviewComment string    // Optional comment from the reviewer
	ReviewedAt    time.Time // When the review was made
	CreatedAt     time.Time // When the mapping was created
	UpdatedAt     time.Time // Last update timestamp
}

func (aa ApplicationAssignment) GetPrimaryKey() string {
	id := aa.ID
	if id == "" {
		id = uuid.NewString()
	}
	return fmt.Sprintf("%s:%s", TableApplicationAssignment, id)
}

func (aa ApplicationAssignment) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxAppAssign_ApplicationID,
			Pattern: fmt.Sprintf("%s:*:%s", TableApplicationAssignment, "application_id"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxAppAssign_ReviewerID,
			Pattern: fmt.Sprintf("%s:*:%s", TableApplicationAssignment, "reviewer_id"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxAppAssign_ReviewStatus,
			Pattern: fmt.Sprintf("%s:*:%s", TableApplicationAssignment, "review_status"),
			Type:    buntdb.IndexString,
		},
	}
}

func (aa ApplicationAssignment) GetIndexValues() map[string]string {
	return map[string]string{
		"application_id": aa.ApplicationID,
		"reviewer_id":    aa.ReviewerID,
		"review_status":  aa.ReviewStatus,
	}
}

// ApplicationHistory tracks the history of actions taken on a permission application.
const (
	TableApplicationHistory     = "app_history"
	IdxAppHistory_ApplicationID = TableApplicationHistory + ":application_id"
	IdxAppHistory_ActorID       = TableApplicationHistory + ":actor_id"
	IdxAppHistory_Action        = TableApplicationHistory + ":action"
)

type ApplicationHistory struct {
	ID            string    // Unique identifier for the history record
	ApplicationID string    // Reference to Application.ID
	Action        string    // Action taken (e.g., submitted, reviewed, approved, rejected)
	ActorID       string    // Reference to User.ID (who performed the action)
	Comment       string    // Optional comment or reason for the action
	CreatedAt     time.Time // When the action was taken
	UpdatedAt     time.Time // Last update timestamp
}

func (ah ApplicationHistory) GetPrimaryKey() string {
	id := ah.ID
	if id == "" {
		id = uuid.NewString()
	}
	return fmt.Sprintf("%s:%s", TableApplicationHistory, id)
}

func (ah ApplicationHistory) GetIndexes() []db.Index {
	return []db.Index{
		{
			Name:    IdxAppHistory_ApplicationID,
			Pattern: fmt.Sprintf("%s:*:%s", TableApplicationHistory, "application_id"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxAppHistory_ActorID,
			Pattern: fmt.Sprintf("%s:*:%s", TableApplicationHistory, "actor_id"),
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxAppHistory_Action,
			Pattern: fmt.Sprintf("%s:*:%s", TableApplicationHistory, "action"),
			Type:    buntdb.IndexString,
		},
	}
}

func (ah ApplicationHistory) GetIndexValues() map[string]string {
	return map[string]string{
		"application_id": ah.ApplicationID,
		"actor_id":       ah.ActorID,
		"action":         ah.Action,
	}
}
