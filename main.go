package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tidwall/buntdb"

	"github.com/jekiapp/topic-master/internal/config"
	"github.com/jekiapp/topic-master/internal/repository"
)

const dataFilename = "topic-master.db"

func main() {
	dataPath := flag.String("data_path", "", "Path to topic-master data directory(required)")
	nsqlookupdHTTPAddr := flag.String("nsqlookupd_http_address", "", "NSQLookupd HTTP address")
	kafkaCluster := flag.String("kafka_bootstrap_server", "", "Kafka bootstrap server")
	skipSync := flag.Bool("skip_sync", false, "Skip sync topics")
	port := flag.String("port", "4181", "Port to listen on")
	flag.Parse()
	if *dataPath == "" {
		fmt.Println("-data_path is required")
		os.Exit(1)
	}

	db, err := buntdb.Open(filepath.Join(*dataPath, dataFilename))
	if err != nil {
		log.Fatalf("failed to open data directory: %v", err)
	}
	defer db.Close()

	cfg, err := initConfig(db, *nsqlookupdHTTPAddr, *kafkaCluster)
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	// make sure indexes are created before checking and setting up root
	repository.Init(cfg, db)

	err = config.CheckAndSetupRoot(db)
	if err != nil {
		log.Fatalf("failed to check and setup root: %v", err)
	}

	deps, err := initDeps(cfg.KafkaCluster)
	if err != nil {
		log.Fatalf("failed to init deps: %v", err)
	}
	defer deps.close()

	mux := http.NewServeMux()
	handler := initHandler(db, deps, cfg)
	handler.routes(mux)

	// sync all the topics
	if !*skipSync {
		_, err = handler.syncTopicsUC.HandleQuery(context.Background(), nil)
		if err != nil {
			log.Fatalf("failed to sync topics: %v", err)
		}
	}

	// Start the server
	fmt.Printf("topic-master is running on port %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
