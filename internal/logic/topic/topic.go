package topic

import (
	"errors"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
)

type ISyncTopics interface {
	GetAllTopics() ([]string, error)
	GetAllNsqTopicEntities() ([]*acl.Entity, error)
	CreateNsqTopicEntity(topic string) (*acl.Entity, error)
	DeleteNsqTopicEntity(topic string) error
}

func SyncTopics(db *buntdb.DB, iSyncTopics ISyncTopics) error {
	// Get the list of topics from the source (e.g., config, external system)
	topics, err := iSyncTopics.GetAllTopics()
	if err != nil {
		return err
	}
	// Build a set for fast lookup of valid topics
	topicSet := make(map[string]struct{}, len(topics))
	for _, t := range topics {
		topicSet[t] = struct{}{}
	}

	// Get all topic entities currently in the DB
	dbEntities, err := iSyncTopics.GetAllNsqTopicEntities()
	if err != nil {
		return err
	}
	// Build a set for fast lookup of DB topics
	dbTopicSet := make(map[string]struct{}, len(dbEntities))
	for _, entity := range dbEntities {
		dbTopicSet[entity.Name] = struct{}{}
		// If a topic exists in DB but not in the source, delete it from DB
		if _, ok := topicSet[entity.Name]; !ok {
			if delErr := iSyncTopics.DeleteNsqTopicEntity(entity.Name); delErr != nil {
				// Collect deletion errors
				err = errors.Join(err, errors.New("DeleteNsqTopicEntity("+entity.Name+"): "+delErr.Error()))
			}
		}
	}

	// For each topic in the source, if not found in DB, create it in DB
	for t := range topicSet {
		if _, ok := dbTopicSet[t]; !ok {
			if _, createErr := iSyncTopics.CreateNsqTopicEntity(t); createErr != nil {
				// Collect creation errors
				err = errors.Join(err, errors.New("CreateNsqTopicEntity("+t+"): "+createErr.Error()))
			}
		}
	}

	// Return any collected errors (nil if none)
	return err
}
