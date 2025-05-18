package model

import "time"

type Entity struct {
	ID          string
	TypeID      string
	GroupOwner  string
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
