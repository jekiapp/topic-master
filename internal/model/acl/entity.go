package acl

import (
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/buntdb"

	"github.com/jekiapp/nsqper/internal/model"
)

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

const (
	TableEntity         = "entity"
	IdxEntity_TypeID    = TableEntity + ":typeid"
	IdxEntity_Group     = TableEntity + ":group"
	IdxEntity_Name      = TableEntity + ":name"
	IdxEntity_Status    = TableEntity + ":status"
	IdxEntity_GroupName = TableEntity + ":group_name"
)

func (e Entity) GetPrimaryKey() string {
	id := e.ID
	if id == "" {
		id = uuid.NewString()
	}
	return TableEntity + ":" + id
}

func (e Entity) GetIndexes() []model.Index {
	return []model.Index{
		{
			Name:    IdxEntity_TypeID,
			Pattern: TableEntity + ":*:typeid",
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxEntity_Group,
			Pattern: TableEntity + ":*:group",
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxEntity_Name,
			Pattern: TableEntity + ":*:name",
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxEntity_Status,
			Pattern: TableEntity + ":*:status",
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxEntity_GroupName,
			Pattern: TableEntity + ":*:group_name",
			Type:    buntdb.IndexString,
		},
	}
}

func (e Entity) GetIndexValues() map[string]string {
	return map[string]string{
		"typeid":     e.TypeID,
		"group":      e.GroupOwner,
		"name":       e.Name,
		"status":     e.Status,
		"group_name": e.GroupOwner + ":" + e.Name,
	}
}
