package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TaushifReza/email-platform/backend/internal/platform/health"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pingFunc func(context.Context) error

func (f pingFunc) Ping(ctx context.Context) error {
	return f(ctx)
}

func TestHealthEndpoint(t *testing.T) {
	router := testRouter(pingFunc(func(context.Context) error { return nil }), nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var body HealthResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, "ok", body.Status)
}

func TestReadyEndpointReportsDatabaseFailure(t *testing.T) {
	router := testRouter(pingFunc(func(context.Context) error { return errors.New("db down") }), nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	var body ReadyResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, Degraded, body.Status)
	assert.Equal(t, "error", body.Checks.Db)
}

func TestReadyEndpointDoesNotFailWhenRedisIsDown(t *testing.T) {
	router := testRouter(
		pingFunc(func(context.Context) error { return nil }),
		pingFunc(func(context.Context) error { return errors.New("redis down") }),
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var body ReadyResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, Ok, body.Status)
	assert.Equal(t, "ok", body.Checks.Db)
	assert.Equal(t, "error", body.Checks.Redis)
}

func testRouter(db, redis health.Pinger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHandlers(router, NewHealthServer(health.New(db, redis)))
	return router
}
