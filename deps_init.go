package main

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type deps struct {
	kafka_adm_client *kafka.AdminClient
}

func initDeps(kafkaCluster string) (*deps, error) {
	deps := &deps{}
	if kafkaCluster != "" {
		admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": kafkaCluster})
		if err != nil {
			return nil, fmt.Errorf("failed to create kafka admin client: %w", err)
		}
		deps.kafka_adm_client = admin
	}

	return deps, nil
}

func (d *deps) close() {
	if d.kafka_adm_client != nil {
		d.kafka_adm_client.Close()
	}
}
