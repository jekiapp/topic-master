package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jekiapp/topic-master/internal/config"
	"github.com/tidwall/buntdb"
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

func initConfig(db *buntdb.DB, nsqlookupdHTTPAddr string, kafkaCluster string) (*config.Config, error) {
	cfg, err := config.NewConfig(db)
	if err != nil {
		if nsqlookupdHTTPAddr == "" && kafkaCluster == "" {
			fmt.Println("No config found. Please provide -nsqlookupd_http_address or -kafka_cluster flag.")
			os.Exit(1)
		}

		cfg, err = config.SetupNewConfig(db, nsqlookupdHTTPAddr, kafkaCluster)
		if err != nil {
			log.Fatalf("failed to setup new config: %v", err)
		}
	}

	changed := false
	if cfg.NSQLookupdHTTPAddr != nsqlookupdHTTPAddr {
		fmt.Printf("nsqlookupd_http_address is changed from \"%s\" to \"%s\"\n", cfg.NSQLookupdHTTPAddr, nsqlookupdHTTPAddr)
		changed = true
	}
	if cfg.KafkaCluster != kafkaCluster {
		fmt.Printf("kafka_cluster is changed from \"%s\" to \"%s\"\n", cfg.KafkaCluster, kafkaCluster)
		changed = true
	}

	if changed {
		fmt.Printf("The topic will be synced from the new address. Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Aborted by user.")
			os.Exit(1)
		}

		fmt.Printf("Saving new config...\n")
		cfg.KafkaCluster = kafkaCluster
		cfg.NSQLookupdHTTPAddr = nsqlookupdHTTPAddr
		err = cfg.SaveConfig(db)
		if err != nil {
			log.Fatalf("failed to save config: %v", err)
		}
	}

	return cfg, nil
}
