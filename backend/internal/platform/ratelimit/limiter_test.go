package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCounter struct {
	count int64
	ttl   time.Duration
	err   error
}

func (f *fakeCounter) Incr(_ context.Context, _ string, _ time.Duration) (int64, time.Duration, error) {
	if f.err != nil {
		return 0, 0, f.err
	}
	f.count++
	return f.count, f.ttl, nil
}

func TestSEC_12_AllowsUnderLimit(t *testing.T) {
	c := &fakeCounter{ttl: time.Minute}
	l := New(c)
	d, err := l.Allow(context.Background(), []Key{{Name: "k", Limit: 2, Window: time.Minute}})
	require.NoError(t, err)
	assert.True(t, d.Allowed)
}

func TestSEC_12_BlocksOverLimit(t *testing.T) {
	c := &fakeCounter{count: 2, ttl: 30 * time.Second}
	l := New(c)
	d, err := l.Allow(context.Background(), []Key{{Name: "k", Limit: 2, Window: time.Minute}})
	require.NoError(t, err)
	assert.False(t, d.Allowed)
	assert.Equal(t, 30*time.Second, d.RetryAfter)
}

func TestSEC_12_CounterError(t *testing.T) {
	c := &fakeCounter{err: errors.New("redis down")}
	l := New(c)
	_, err := l.Allow(context.Background(), []Key{{Name: "k", Limit: 2, Window: time.Minute}})
	assert.Error(t, err)
}

func TestSEC_12_UnavailableCounter(t *testing.T) {
	l := New(nil)
	_, err := l.Allow(context.Background(), []Key{{Name: "k", Limit: 2, Window: time.Minute}})
	assert.ErrorIs(t, err, ErrCounterUnavailable)
}
