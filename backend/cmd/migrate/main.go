package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	sqlDB, err := sql.Open("pgx", cfg.MigrationDatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer sqlDB.Close()

	if err = db.Migrate(ctx, sqlDB); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
