package main

import (
	"log"
	"net/http"

	"github.com/jekiapp/hi-mod-arch/config"
	"github.com/jekiapp/hi-mod-arch/internal/logic"
	"github.com/jekiapp/hi-mod-arch/internal/repository"
	"github.com/jekiapp/hi-mod-arch/internal/usecase/example"
	"github.com/jekiapp/hi-mod-arch/pkg/db"
)

func InitApplication() Handler {
	cfg := config.InitConfig()

	err := logic.Init(cfg)
	if err != nil {
		log.Fatalf("error init logic %s", err.Error())
	}

	err = repository.Init(cfg)
	if err != nil {
		log.Fatalf("error init repository %s", err.Error())
	}

	// init database
	dbCli, err := db.InitDatabase(db.DbConfig{Host: cfg.Database.Host})
	if err != nil {
		log.Fatal(err)
	}

	productCli := &http.Client{}
	userCli := &http.Client{}
	promoCli := &http.Client{}

	newHandler := Handler{
		ExampleHandler: example.NewExampleUsecase(dbCli, promoCli, productCli, userCli),
	}

	return newHandler
}
