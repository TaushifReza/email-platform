package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAuthService struct {
	result identity.SessionResult
	err    error
	actor  identity.Actor
}

func (f *fakeAuthService) SignIn(_ context.Context, _ identity.LoginRequest, _ identity.ClientInfo) (identity.SessionResult, error) { //nolint:lll
	return f.result, f.err
}

func (f *fakeAuthService) Authenticate(_ context.Context, _ string) (identity.Actor, error) {
	if f.err != nil {
		return identity.Actor{}, f.err
	}
	return f.actor, nil
}

func (f *fakeAuthService) AuthenticatePending(_ context.Context, _ string) (identity.PendingSession, error) {
	if f.err != nil {
		return identity.PendingSession{}, f.err
	}
	return identity.PendingSession{}, f.err
}

func (f *fakeAuthService) SignOut(_ context.Context, _ string) error {
	return f.err
}

func testAuthRouter(auth AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHandlers(router, NewServer(nil, auth, nil))
	return router
}

func TestLoginSuccess(t *testing.T) {
	fake := &fakeAuthService{result: identity.SessionResult{MFARequired: true, Token: "tok", AbsoluteExpiresAt: time.Now().Add(time.Hour)}} //nolint:lll
	router := testAuthRouter(fake)
	rec := httptest.NewRecorder()
	const body = `{"email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Csrf-Token", "1")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var resp LoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.MfaRequired)
}

func TestLoginEnrollmentRequired(t *testing.T) {
	fake := &fakeAuthService{result: identity.SessionResult{MFAEnrollmentRequired: true, Token: "tok", AbsoluteExpiresAt: time.Now().Add(time.Hour)}} //nolint:lll
	router := testAuthRouter(fake)
	rec := httptest.NewRecorder()
	const body = `{"email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Csrf-Token", "1")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var resp LoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.MfaEnrollmentRequired)
}

func TestLoginInvalidCredentials(t *testing.T) {
	fake := &fakeAuthService{err: identity.ErrInvalidCredentials}
	router := testAuthRouter(fake)
	rec := httptest.NewRecorder()
	const body = `{"email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Csrf-Token", "1")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLoginRateLimited(t *testing.T) {
	fake := &fakeAuthService{err: &identity.RateLimitError{RetryAfter: 30 * time.Second}}
	router := testAuthRouter(fake)
	rec := httptest.NewRecorder()
	const body = `{"email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Csrf-Token", "1")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestLoginUnavailable(t *testing.T) {
	fake := &fakeAuthService{err: identity.ErrRateLimiterUnavailable}
	router := testAuthRouter(fake)
	rec := httptest.NewRecorder()
	const body = `{"email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Csrf-Token", "1")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestLoginBadInput(t *testing.T) {
	fake := &fakeAuthService{}
	router := testAuthRouter(fake)
	rec := httptest.NewRecorder()
	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Csrf-Token", "1")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
