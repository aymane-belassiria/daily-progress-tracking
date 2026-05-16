package main

import (
	"log"

	"daily-progress-tracking/backend/internal/app"
)

func main() {
	cfg, err := app.LoadConfig(".env")
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	store, err := app.NewStore(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer store.Close()

	server := app.NewServer(cfg, store)
	log.Printf("Backend listening on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
