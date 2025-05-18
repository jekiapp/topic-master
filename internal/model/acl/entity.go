package acl

import "time"

type Entity struct {
	ID          string
	TypeID      string
	GroupOwner  string // Group.ID
	Name        string
	Resource    string
	Status      string
	Description string
	Tags        []string
	Metadata    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// e.g nsq topic
type EntityType struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// publish, tail, etc.
type EntityDefaultPermission struct {
	ID             string
	EntityID       string
	PermissionName string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
