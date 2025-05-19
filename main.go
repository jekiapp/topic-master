package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tidwall/buntdb"

	"github.com/jekiapp/nsqper/internal/config"
	"github.com/jekiapp/nsqper/internal/repository"
)

func main() {
	dataPath := flag.String("data_path", "", "Path to buntdb data file (required)")
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
		fmt.Println("No config found, please enter configuration:")
		cfg = config.PromptConfig()
		if err := config.SaveConfig(db, cfg); err != nil {
			log.Fatalf("failed to save config: %v", err)
		}
		fmt.Println("Config saved. Reloading...")
		cfg, err = config.NewConfig(db)
		if err != nil {
			log.Fatalf("failed to load config after saving: %v", err)
		}
	}

	err = config.CheckAndSetupRoot(db)
	if err != nil {
		log.Fatalf("failed to check and setup root: %v", err)
	}

	fmt.Printf("Loaded config: %+v\n", cfg)

	repository.Init(cfg)

	mux := http.NewServeMux()
	handler := initHandler(db, cfg)
	handler.routes(mux)

	// sync all the topics
	handler.syncTopicsUC.HandleQuery(context.Background(), nil)

	// Start the server
	fmt.Printf("NSQper is running on port %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
