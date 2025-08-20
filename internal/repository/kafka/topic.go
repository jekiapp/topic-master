package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

// GetAllTopics lists all topics from Kafka using the admin client
func GetAllTopics(kCli *kafka.AdminClient) ([]string, error) {
	topicsMetadata, err := kCli.GetMetadata(nil, true, 5000)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	topics := make([]string, 0, len(topicsMetadata.Topics))
	for topic := range topicsMetadata.Topics {
		topics = append(topics, topic)
	}
	return topics, nil
}

func CreateKafkaTopicEntity(dbConn *buntdb.DB, topic string) (*entity.Entity, error) {
	entityObj := &entity.Entity{
		ID:         uuid.NewString(),
		Kind:       entity.EntityKind_Topic,
		TypeID:     entity.EntityType_KafkaTopic,
		Name:       topic,
		Resource:   entity.EntityResource_Kafka,
		Status:     entity.EntityStatus_Active,
		GroupOwner: entity.GroupNone,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := db.Insert(dbConn, entityObj); err != nil {
		return nil, err
	}
	return entityObj, nil
}

func DeleteKafkaTopicEntity(dbConn *buntdb.DB, topic string) error {
	return db.DeleteByIndex(dbConn, &entity.Entity{TypeID: entity.EntityType_KafkaTopic, Name: topic}, entity.IdxEntity_TypeName)
}
