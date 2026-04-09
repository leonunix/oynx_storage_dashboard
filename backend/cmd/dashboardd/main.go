package main

import (
	"log"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/app"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/config"
)

func main() {
	cfg := config.Load()
	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("onyx dashboard backend listening on %s", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
