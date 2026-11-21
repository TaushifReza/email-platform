package db

import (
	"context"
	"database/sql"
	"time"

	// Register the pgx driver for database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB owns the database handle. Query methods are intentionally not exposed:
// tenant SQL must run inside WithOrg so row-level security is always scoped.
type DB struct {
	sql *sql.DB
}

// New creates a new database handle from DSN with appropriate settings.
func New(ctx context.Context, dsn string) (*DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(30 * time.Second)

	if err = sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return &DB{sql: sqlDB}, nil
}

// Ping verifies the database is reachable.
func (d *DB) Ping(ctx context.Context) error {
	return d.sql.PingContext(ctx)
}

// Close releases the database handle.
func (d *DB) Close() error {
	return d.sql.Close()
}
