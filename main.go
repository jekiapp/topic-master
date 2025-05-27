package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tidwall/buntdb"

	"github.com/jekiapp/topic-master/internal/config"
	"github.com/jekiapp/topic-master/internal/repository"
)

func main() {
	dataPath := flag.String("data_path", "", "Path to buntdb data directory(required)")
	nsqlookupdHTTPAddr := flag.String("nsqlookupd_http_address", "", "NSQLookupd HTTP address (required if no config)")
	port := flag.String("port", "4181", "Port to listen on")
	flag.Parse()
	if *dataPath == "" {
		fmt.Println("-data_path is required")
		os.Exit(1)
	}

	db, err := buntdb.Open(*dataPath)
	if err != nil {
		log.Fatalf("failed to open buntdb: %v", err)
	}
	defer db.Close()
	cfg, err := config.NewConfig(db)
	if err != nil {
		if *nsqlookupdHTTPAddr == "" {
			fmt.Println("No config found. Please provide -nsqlookupd_http_address flag.")
			os.Exit(1)
		}
		cfg, err = config.SetupNewConfig(db, *nsqlookupdHTTPAddr)
		if err != nil {
			log.Fatalf("failed to setup new config: %v", err)
		}
	}

	// make sure indexes are created before checking and setting up root
	repository.Init(cfg, db)

	err = config.CheckAndSetupRoot(db)
	if err != nil {
		log.Fatalf("failed to check and setup root: %v", err)
	}

	mux := http.NewServeMux()
	handler := initHandler(db, cfg)
	handler.routes(mux)

	// sync all the topics
	_, err = handler.syncTopicsUC.HandleQuery(context.Background(), nil)
	if err != nil {
		log.Fatalf("failed to sync topics: %v", err)
	}

	// Start the server
	fmt.Printf("topic-master is running on port %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
