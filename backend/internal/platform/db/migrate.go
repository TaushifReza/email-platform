package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/TaushifReza/email-platform/backend/db/migrations"
	"github.com/pressly/goose/v3"
)

// Migrate applies all embedded migrations. It is safe to call repeatedly.
func Migrate(ctx context.Context, sqlDB *sql.DB) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	if _, err = provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
