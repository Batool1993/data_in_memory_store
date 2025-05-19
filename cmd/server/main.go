package main

import (
	"data_storage/config"
	"data_storage/server/adapters"
	"data_storage/server/storage"
	"data_storage/server/store_service"
	"fmt"
	"log"
	"net/http"
)

func main() {

	// load configs
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	fmt.Sprintln("config", cfg.CleanUpInterval)

	// 2) Wire up repository, service, and handlers
	repo := storage.NewDataRepo(cfg.CleanUpInterval)

	//Ensure invalidator stops on exit
	defer repo.ShutDownInvalidation()

	svc := store_service.NewStoreService(repo, cfg.DefaultTTL)

	// 3) Build the router+middleware
	handler := adapters.NewHandler(svc, cfg.APIToken)

	// 4) Start HTTP server
	log.Printf("listening on :8080")
	if err = http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
