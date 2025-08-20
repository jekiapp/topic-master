package topic

import (
	"errors"
	"fmt"
	"log"

	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
	topicmodel "github.com/jekiapp/topic-master/internal/model/topic"
	dbPkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

type ISyncTopics interface {
	GetAllTopics() ([]topicmodel.Topic, error)

	GetAllTopicEntities() ([]entitymodel.Entity, error)

	CreateNsqTopicEntity(topic string) (*entitymodel.Entity, error)
	DeleteNsqTopicEntity(topic string) error

	CreateKafkaTopicEntity(topic string) (*entitymodel.Entity, error)
	DeleteKafkaTopicEntity(topic string) error
}

func SyncTopics(db *buntdb.DB, iSyncTopics ISyncTopics) (topics []string, err error) {
	// Get the list of topics from the source (e.g., nsqlookupd)
	topicsPlatform, err := iSyncTopics.GetAllTopics()
	if err != nil {
		return nil, err
	}

	if len(topicsPlatform) == 0 {
		log.Println("[WARN] No topics found")
	}
	// Build a set for fast lookup of valid topics
	topicSet := make(map[string]topicmodel.Topic, len(topicsPlatform))
	for _, t := range topicsPlatform {
		key := fmt.Sprintf("%s-%s", t.Resource, t.Name)
		topicSet[key] = t
	}

	// Get all topic entities currently in the DB
	dbEntities, err := iSyncTopics.GetAllTopicEntities()
	if err != nil && err != dbPkg.ErrNotFound {
		return nil, err
	}

	var errSet error
	// Build a set for fast lookup of DB topics
	dbTopicSet := make(map[string]struct{}, len(dbEntities))
	for _, entity := range dbEntities {
		key := fmt.Sprintf("%s-%s", entity.Resource, entity.Name)
		dbTopicSet[key] = struct{}{}
		// If a topic exists in DB but not in the source, delete it from DB
		// another option is to mark it as deleted
		if _, ok := topicSet[key]; !ok {
			if entity.Resource == entitymodel.EntityResource_NSQ {
				log.Println("[INFO] Deleting topic from DB: ", entity.Name)
				if delErr := iSyncTopics.DeleteNsqTopicEntity(key); delErr != nil {
					// Collect deletion errors
					errSet = errors.Join(errSet, errors.New("DeleteNsqTopicEntity("+entity.Name+"): "+delErr.Error()))
				}
			}
			if entity.Resource == entitymodel.EntityResource_Kafka {
				log.Println("[INFO] Deleting topic from DB: ", entity.Name)
				if delErr := iSyncTopics.DeleteKafkaTopicEntity(key); delErr != nil {
					// Collect deletion errors
					errSet = errors.Join(errSet, errors.New("DeleteKafkaTopicEntity("+entity.Name+"): "+delErr.Error()))
				}
			}
		}
	}

	// For each topic in the source, if not found in DB, create it in DB
	for _, t := range topicSet {
		key := fmt.Sprintf("%s-%s", t.Resource, t.Name)
		if _, ok := dbTopicSet[key]; !ok {
			log.Println("[INFO] Creating topic in DB: ", t)
			if t.Resource == entitymodel.EntityResource_NSQ {
				if _, createErr := iSyncTopics.CreateNsqTopicEntity(t.Name); createErr != nil {
					// Collect creation errors
					errSet = errors.Join(errSet, errors.New("CreateNsqTopicEntity("+t.Name+"): "+createErr.Error()))
				}
			}
			if t.Resource == entitymodel.EntityResource_Kafka {
				if _, createErr := iSyncTopics.CreateKafkaTopicEntity(t.Name); createErr != nil {
					// Collect creation errors
					errSet = errors.Join(errSet, errors.New("CreateKafkaTopicEntity("+t.Name+"): "+createErr.Error()))
				}
			}
		}
	}

	// Return any collected errors (nil if none)
	return topics, errSet
}
