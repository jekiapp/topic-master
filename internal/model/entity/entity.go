package entity

import (
	"time"

	"github.com/tidwall/buntdb"

	"github.com/jekiapp/topic-master/pkg/db"
)

type Entity struct {
	ID          string
	TypeID      string // e.g EntityType_NSQTopic
	Kind        string // e.g EntityKind_Topic
	GroupOwner  string // Group.Name
	Name        string
	Resource    string
	Status      string
	Description string
	Tags        []string
	Metadata    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	EntityResource_NSQ   = "NSQ"
	EntityResource_Kafka = "Kafka"

	EntityType_NSQTopic   = "nsq_topic"
	EntityType_NSQChannel = "nsq_channel"

	EntityType_KafkaTopic   = "kafka_topic"
	EntityType_KafkaChannel = "kafka_channel"

	EntityKind_Topic   = "topic"
	EntityKind_Channel = "channel"

	EntityStatus_Active  = "active"
	EntityStatus_Deleted = "deleted"
)

// publish, tail, etc.
type EntityDefaultPermission struct {
	ID             string
	EntityID       string
	PermissionName string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

const (
	TableEntity            = "entity"
	IdxEntity_TypeID       = TableEntity + ":typeid"
	IdxEntity_Group        = TableEntity + ":group"
	IdxEntity_Name         = TableEntity + ":name"
	IdxEntity_Status       = TableEntity + ":status"
	IdxEntity_GroupType    = TableEntity + ":group_type"
	IdxEntity_TypeName     = TableEntity + ":type_name"
	IdxEntity_TopicChannel = TableEntity + ":topic_channel"
	IdxEntity_Kind         = TableEntity + ":kind"

	GroupNone = "None"
)

func (e *Entity) GetPrimaryKey(id string) string {
	if e.ID == "" && id != "" {
		e.ID = id
	}
	return TableEntity + ":" + e.ID
}

func (e Entity) GetIndexes() []db.Index {
	return []db.Index{
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
			Name:    IdxEntity_GroupType,
			Pattern: TableEntity + ":*:group_type",
			Type:    buntdb.IndexString,
		},
		{
			Name:    IdxEntity_TypeName,
			Pattern: TableEntity + ":*:type_name",
			Type:    buntdb.IndexString,
		},
		{
			Name:     IdxEntity_TopicChannel,
			Pattern:  TableEntity + ":*:topic_channel",
			Type:     buntdb.IndexString,
			Optional: true,
		},
		{
			Name:    IdxEntity_Kind,
			Pattern: TableEntity + ":*:kind",
			Type:    buntdb.IndexString,
		},
	}
}

func (e Entity) GetIndexValues() map[string]string {
	values := map[string]string{
		"typeid":     e.TypeID,
		"group":      e.GroupOwner,
		"name":       e.Name,
		"status":     e.Status,
		"group_type": e.GroupOwner + ":" + e.TypeID,
		"type_name":  e.TypeID + ":" + e.Name,
		"kind":       e.Kind,
	}

	if e.TypeID == EntityType_NSQChannel && e.Metadata["topic"] != "" {
		values["topic_channel"] = e.Metadata["topic"]
	}

	return values
}

func (e *Entity) SetID(id string) {
	e.ID = id
}
