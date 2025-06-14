// learn from topic.go, this file is the logic to sync all the channels from nsqds
// and create the channel entities in the database

package topic

import (
	"errors"
	"log"

	"github.com/jekiapp/topic-master/internal/model/entity"
	dbPkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type ISyncChannels interface {
	GetAllChannels(topic string) ([]string, error)
	GetAllNsqChannelEntities(topic string) ([]entity.Entity, error)
	CreateNsqChannelEntity(topic, channel string) (*entity.Entity, error)
	DeleteNsqChannelEntity(topic, channel string) error
}

func SyncChannels(db *buntdb.DB, topic string, iSyncChannels ISyncChannels) error {
	// Get the list of channels from the source for the given topic
	channels, err := iSyncChannels.GetAllChannels(topic)
	if err != nil {
		return err
	}

	if len(channels) == 0 {
		log.Printf("[WARN] No channels found for topic: %s", topic)
	}

	// Build a set for fast lookup of valid channels
	channelSet := make(map[string]struct{}, len(channels))
	for _, c := range channels {
		channelSet[c] = struct{}{}
	}

	// Get all channel entities currently in the DB for the topic
	dbEntities, err := iSyncChannels.GetAllNsqChannelEntities(topic)
	if err != nil && err != dbPkg.ErrNotFound {
		return err
	}

	var errSet error
	// Build a set for fast lookup of DB channels
	dbChannelSet := make(map[string]struct{}, len(dbEntities))
	for _, entity := range dbEntities {
		dbChannelSet[entity.Name] = struct{}{}
		// If a channel exists in DB but not in the source, delete it from DB
		if _, ok := channelSet[entity.Name]; !ok {
			if delErr := iSyncChannels.DeleteNsqChannelEntity(topic, entity.Name); delErr != nil {
				// Collect deletion errors
				errSet = errors.Join(errSet, errors.New("DeleteNsqChannelEntity("+topic+","+entity.Name+"): "+delErr.Error()))
			}
		}
	}

	// For each channel in the source, if not found in DB, create it in DB
	for c := range channelSet {
		if _, ok := dbChannelSet[c]; !ok {
			if _, createErr := iSyncChannels.CreateNsqChannelEntity(topic, c); createErr != nil {
				// Collect creation errors
				errSet = errors.Join(errSet, errors.New("CreateNsqChannelEntity("+topic+","+c+"): "+createErr.Error()))
			}
		}
	}

	// Return any collected errors (nil if none)
	return errSet
}
