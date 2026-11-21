//go:build redis

package ratelimit

import (
	"context"
	"os"
	"testing"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestSEC_12_RedisWindow(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Fatal("REDIS_URL is required")
	}
	opts, err := redis.ParseURL(url)
	require.NoError(t, err)
	client := redis.NewClient(opts)
	defer client.Close()
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis unavailable: %v", err)
	}
	key := "test:ratelimit:" + t.Name()
	defer client.Del(ctx, key)
	counter := NewRedisCounter(client)
	count, ttl, err := counter.Incr(ctx, key, 100*time.Millisecond)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	require.Greater(t, ttl, time.Duration(0))
	count, _, err = counter.Incr(ctx, key, 100*time.Millisecond)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
	time.Sleep(150 * time.Millisecond)
	count, _, err = counter.Incr(ctx, key, 100*time.Millisecond)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}
