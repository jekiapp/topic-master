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

	"github.com/jekiapp/nsqper/internal/config"
	"github.com/jekiapp/nsqper/internal/repository"
)

const dbfile = "nsqper.db"

func main() {
	dataDir := flag.String("data_dir", "", "Path to buntdb data directory(required)")
	nsqlookupdHTTPAddr := flag.String("nsqlookupd_http_address", "", "NSQLookupd HTTP address (required if no config)")
	port := flag.String("port", "4181", "Port to listen on")
	flag.Parse()
	if *dataDir == "" {
		fmt.Println("-data_dir is required")
		os.Exit(1)
	}

	dataPath := filepath.Join(*dataDir, dbfile)

	db, err := buntdb.Open(dataPath)
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

	err = config.CheckAndSetupRoot(db)
	if err != nil {
		log.Fatalf("failed to check and setup root: %v", err)
	}

	repository.Init(cfg, db)

	mux := http.NewServeMux()
	handler := initHandler(db, cfg)
	handler.routes(mux)

	// sync all the topics
	_, err = handler.syncTopicsUC.HandleQuery(context.Background(), nil)
	if err != nil {
		log.Fatalf("failed to sync topics: %v", err)
	}

	// Start the server
	fmt.Printf("NSQper is running on port %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
