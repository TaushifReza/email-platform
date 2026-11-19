package db

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping database test")
	}

	ctx := context.Background()
	pool, err := New(ctx, dsn)
	assert.NoError(t, err)
	assert.NotNil(t, pool)

	// Test that we can ping
	err = pool.Ping(ctx)
	assert.NoError(t, err)

	// Clean up
	pool.Close()
}