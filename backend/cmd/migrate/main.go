package main

import (
	"context"
	"log"
	"os"

	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/pressly/goose/v3"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Run migrations
	goose.SetBasePath("backend/db/migrations")
	if err = goose.Up(ctx, cfg.DatabaseDSN, "backend/db/migrations"); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("Migrations completed successfully")
}
