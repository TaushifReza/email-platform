package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New creates a new pgx pool from DSN with appropriate settings.
func New(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Configure the pool.
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Verify connection works.
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
