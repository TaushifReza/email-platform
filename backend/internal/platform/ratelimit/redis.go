package ratelimit

import (
	"context"
	"time"

	redis "github.com/go-redis/redis/v8"
)

// RedisCounter increments keys in Redis with a fixed window.
type RedisCounter struct {
	client *redis.Client
}

// NewRedisCounter creates a Redis counter.
func NewRedisCounter(client *redis.Client) *RedisCounter {
	return &RedisCounter{client: client}
}

//nolint:gochecknoglobals
var incrScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
return {current, ttl}
`)

// Incr increments the key and returns the new count and remaining TTL.
func (c *RedisCounter) Incr(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	if c.client == nil {
		return 0, 0, ErrCounterUnavailable
	}
	ms := window.Milliseconds()
	if ms <= 0 {
		ms = 1
	}
	res, err := incrScript.Run(ctx, c.client, []string{key}, ms).Int64Slice()
	if err != nil {
		return 0, 0, err
	}
	if len(res) != 2 {
		return 0, 0, ErrCounterUnavailable
	}
	return res[0], time.Duration(res[1]) * time.Millisecond, nil
}
