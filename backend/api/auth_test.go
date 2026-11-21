package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeLoginService struct {
	result identity.LoginResult
	err    error
	got    identity.LoginRequest
}

func (f *fakeLoginService) Login(_ context.Context, req identity.LoginRequest) (identity.LoginResult, error) {
	f.got = req
	return f.result, f.err
}

func testAuthRouter(auth LoginService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHandlers(router, NewServer(nil, auth, zap.NewNop()))
	return router
}

func TestAUTH_01_LoginEndpoint(t *testing.T) {
	fake := &fakeLoginService{result: identity.LoginResult{MFARequired: true}}
	router := testAuthRouter(fake)
	body := bytes.NewBufferString(`{"email":"a@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var resp LoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.MfaRequired)
	assert.Equal(t, "a@example.com", fake.got.Email)
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

func TestAUTH_04_LoginEndpointUniform(t *testing.T) {
	fake := &fakeLoginService{err: identity.ErrInvalidCredentials}
	router := testAuthRouter(fake)
	payloads := []string{
		`{"email":"a@example.com","password":"wrong"}`,
		`{"email":"missing@example.com","password":"wrong"}`,
	}
	for _, payload := range payloads {
		body := bytes.NewBufferString(payload)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.JSONEq(t, `{"error":"invalid_credentials"}`, rec.Body.String())
	}
}

func TestSEC_12_LoginEndpointRetryAfter(t *testing.T) {
	fake := &fakeLoginService{err: &identity.RateLimitError{RetryAfter: 30 * time.Second}}
	router := testAuthRouter(fake)
	body := bytes.NewBufferString(`{"email":"a@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "30", rec.Header().Get("Retry-After"))
	assert.JSONEq(t, `{"error":"rate_limited"}`, rec.Body.String())
}

func TestSEC_12_LoginEndpointUnavailable(t *testing.T) {
	fake := &fakeLoginService{err: identity.ErrRateLimiterUnavailable}
	router := testAuthRouter(fake)
	body := bytes.NewBufferString(`{"email":"a@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(t, `{"error":"unavailable"}`, rec.Body.String())
}

func TestAUTH_01_LoginEndpointNoStore(t *testing.T) {
	fake := &fakeLoginService{result: identity.LoginResult{}}
	router := testAuthRouter(fake)
	body := bytes.NewBufferString(`{"email":"a@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

func TestAUTH_04_LoginEndpointInvalidRequest(t *testing.T) {
	fake := &fakeLoginService{}
	router := testAuthRouter(fake)
	body := bytes.NewBufferString(`{"email":"","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":"invalid_request"}`, rec.Body.String())
}
