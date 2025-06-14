package nsq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/entity"
	dbPkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

// GetAllChannels fetches all channels for a topic from lookupd
func GetAllChannels(topic string) ([]string, error) {
	if lookupdAddr == "" {
		return nil, fmt.Errorf("lookupd address not initialized")
	}
	url := fmt.Sprintf("%s/channels?topic=%s", lookupdAddr, topic)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lookupd returned status %d", resp.StatusCode)
	}
	var result struct {
		Channels []string `json:"channels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Channels, nil
}

// GetAllNsqTopicChannels gets all channel entities from the database for a topic
func GetAllNsqTopicChannels(db *buntdb.DB, topic string) ([]entity.Entity, error) {
	entities, err := dbPkg.SelectAll[entity.Entity](db, "="+topic, entity.IdxEntity_TopicChannel)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return nil, dbPkg.ErrNotFound
	}
	return entities, nil
}

// CreateNsqChannelEntity creates a channel entity in the database
func CreateNsqChannelEntity(db *buntdb.DB, topic, channel string) (*entity.Entity, error) {
	entity := &entity.Entity{
		ID:          uuid.NewString(),
		Name:        channel,
		TypeID:      entity.EntityType_NSQChannel,
		Resource:    entity.EntityResource_NSQ,
		Status:      entity.EntityStatus_Active,
		Description: "NSQ channel",
		Metadata:    map[string]string{"topic": topic},
		GroupOwner:  entity.GroupNone,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := dbPkg.Insert(db, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

// DeleteNsqChannelEntity deletes a channel entity from the database
func DeleteNsqChannelEntity(db *buntdb.DB, topic, channel string) error {
	tmp := &entity.Entity{TypeID: entity.EntityType_NSQChannel, Name: channel}
	return dbPkg.DeleteByIndex(db, tmp, entity.IdxEntity_TypeName)
}
