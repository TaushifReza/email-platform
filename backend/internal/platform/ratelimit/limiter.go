package ratelimit

import (
	"context"
	"time"
)

// Key describes one rate limit dimension.
type Key struct {
	Name   string
	Limit  int
	Window time.Duration
}

// Decision reports whether an attempt may proceed.
type Decision struct {
	Allowed    bool
	RetryAfter time.Duration
}

// Counter increments a named key for a window.
type Counter interface {
	Incr(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error)
}

// Limiter applies rate limits.
type Limiter struct {
	counter Counter
}

// New creates a limiter.
func New(counter Counter) *Limiter {
	return &Limiter{counter: counter}
}

// Allow checks every key and increments each one.
func (l *Limiter) Allow(ctx context.Context, keys []Key) (Decision, error) {
	if l.counter == nil {
		return Decision{}, ErrCounterUnavailable
	}
	allowed := true
	var retry time.Duration
	for _, key := range keys {
		count, ttl, err := l.counter.Incr(ctx, key.Name, key.Window)
		if err != nil {
			return Decision{}, err
		}
		if count > int64(key.Limit) {
			allowed = false
			if ttl > retry {
				retry = ttl
			}
		}
	}
	return Decision{Allowed: allowed, RetryAfter: retry}, nil
}

// ErrCounterUnavailable is returned when no counter is configured.
