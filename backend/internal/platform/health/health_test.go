package health

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error {
	return f.err
}

func TestReadyRequiresDatabase(t *testing.T) {
	service := New(fakePinger{err: errors.New("down")}, fakePinger{})

	status := service.Ready(context.Background())

	assert.False(t, status.Ready)
	assert.Equal(t, "error", status.Database)
	assert.Equal(t, "ok", status.Redis)
}

func TestReadyDoesNotFailWhenRedisIsDown(t *testing.T) {
	service := New(fakePinger{}, fakePinger{err: errors.New("down")})

	status := service.Ready(context.Background())

	assert.True(t, status.Ready)
	assert.Equal(t, "ok", status.Database)
	assert.Equal(t, "error", status.Redis)
}

func TestReadyWithoutRedis(t *testing.T) {
	service := New(fakePinger{}, nil)

	status := service.Ready(context.Background())

	assert.True(t, status.Ready)
	assert.Equal(t, "ok", status.Database)
	assert.Equal(t, "disabled", status.Redis)
}

func TestLiveIsAlwaysHealthy(t *testing.T) {
	assert.True(t, New(fakePinger{}, nil).Live())
}
