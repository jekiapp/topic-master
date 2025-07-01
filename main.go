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
	nsqlookupdHTTPAddr := flag.String("nsqlookupd_http_address", "", "NSQLookupd HTTP address (required)")
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

	if err == nil && cfg.NSQLookupdHTTPAddr != *nsqlookupdHTTPAddr {
		fmt.Printf("nsqlookupd_http_address is changed from %s to %s\n", cfg.NSQLookupdHTTPAddr, *nsqlookupdHTTPAddr)
		log.Fatalf("either provide new data path or remove %s to reset the data", *dataPath)
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
