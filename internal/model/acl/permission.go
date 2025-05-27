package acl

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (p Permission) GetPrimaryKey() string {
	id := p.ID
	if id == "" {
		id = uuid.NewString()
	}
	return fmt.Sprintf("%s:%s", TablePermission, id)
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

// PermissionApplication represents a user's request to obtain a permission.
type PermissionApplication struct {
	ID           string    // UUID
	UserID       string    // Reference to User.ID (the applicant)
	PermissionID string    // Reference to Permission.ID (the requested permission)
	Reason       string    // Reason for the application
	Status       string    // Overall status (e.g., pending, approved, rejected)
	CreatedAt    time.Time // When the application was created
	UpdatedAt    time.Time // Last update timestamp
}

func (pa PermissionApplication) GetPrefix() string {
	return "permission_application:"
}

func (pa PermissionApplication) GetKey() string {
	return fmt.Sprintf("%s%s:%s", pa.GetPrefix(), pa.UserID, pa.PermissionID)
}

// PermissionApplicationReviewer links a permission application to a reviewer and their review status.
type ApplicationAssignment struct {
	ID            string    // UUID
	ApplicationID string    // Reference to PermissionApplication.ID
	ReviewerID    string    // Reference to User.ID (the reviewer)
	ReviewStatus  string    // Status (e.g., pending, approved, rejected)
	ReviewComment string    // Optional comment from the reviewer
	ReviewedAt    time.Time // When the review was made
	CreatedAt     time.Time // When the mapping was created
	UpdatedAt     time.Time // Last update timestamp
}

func (aa ApplicationAssignment) GetPrefix() string {
	return "app_assign:"
}

func (aa ApplicationAssignment) GetKey() string {
	return fmt.Sprintf("%s%s:%s", aa.GetPrefix(), aa.ApplicationID, aa.ReviewerID)
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
