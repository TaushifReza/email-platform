package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platformadmin"
	"github.com/TaushifReza/email-platform/backend/internal/tenancy"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOperatorIdentity struct {
	operator platformadmin.Operator
	err      error
	gotEmail string
}

func (f *fakeOperatorIdentity) Authenticate(_ context.Context, email string) (platformadmin.Operator, error) {
	f.gotEmail = email
	return f.operator, f.err
}

type fakePlatformAdmin struct {
	err      error
	created  bool
	listed   bool
	audited  bool
	limitSet bool
}

func (f *fakePlatformAdmin) ListOrganizations(context.Context, identity.Actor,
	platformadmin.Page,
) ([]tenancy.Organization, error) {
	f.listed = true
	return nil, f.err
}

func (f *fakePlatformAdmin) CreateOrganization(context.Context, identity.Actor,
	platformadmin.CreateOrganizationInput, identity.ClientInfo,
) (tenancy.Organization, error) {
	f.created = true
	return tenancy.Organization{ID: "org-1", Name: "Acme"}, f.err
}

func (f *fakePlatformAdmin) SuspendOrganization(context.Context, identity.Actor, string, identity.ClientInfo) error {
	return f.err
}

func (f *fakePlatformAdmin) ReinstateOrganization(context.Context, identity.Actor, string, identity.ClientInfo) error {
	return f.err
}

func (f *fakePlatformAdmin) UpdateLimits(context.Context, identity.Actor, string, platformadmin.LimitsInput,
	identity.ClientInfo,
) (tenancy.Organization, error) {
	f.limitSet = true
	return tenancy.Organization{ID: "org-1"}, f.err
}

func (f *fakePlatformAdmin) ListOperatorAudit(context.Context, identity.Actor,
	platformadmin.Page,
) ([]audit.Entry, error) {
	f.audited = true
	return []audit.Entry{{ID: "a1", Action: audit.ActionOrgCreated}}, f.err
}

func operatorTestRouter(identitySvc OperatorIdentityService, admin PlatformAdminService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	server := NewServer(nil, nil, nil)
	server.SetConfig(config.Config{
		TrustedProxyCIDRs:      "172.16.0.0/12",
		OperatorIdentityHeader: "Cf-Access-Authenticated-User-Email",
	})
	server.SetOperatorIdentityService(identitySvc)
	server.SetPlatformAdminService(admin)
	router := gin.New()
	router.Use(RequestIDMiddleware(), CSRFMiddleware(), AuthMiddleware(server))
	RegisterHandlers(router, server)
	return router
}

func operatorRequest(router *gin.Engine, path, peer, email string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = peer
	if email != "" {
		req.Header.Set("Cf-Access-Authenticated-User-Email", email)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestOperatorAuthTrustBoundary_SEC02(t *testing.T) {
	const trustedPeer = "172.18.0.5:1234"

	t.Run("untrusted peer is rejected even with a header", func(t *testing.T) {
		svc := &fakeOperatorIdentity{operator: platformadmin.Operator{ID: "op-1", Email: "a@b.com"}}
		rec := operatorRequest(operatorTestRouter(svc, nil),
			"/operator/session", "203.0.113.9:1234", "a@b.com")
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.NotContains(t, rec.Body.String(), "a@b.com")
	})

	t.Run("missing identity is rejected", func(t *testing.T) {
		rec := operatorRequest(operatorTestRouter(&fakeOperatorIdentity{}, nil),
			"/operator/session", trustedPeer, "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("unknown operator is rejected without enumeration", func(t *testing.T) {
		svc := &fakeOperatorIdentity{err: platformadmin.ErrOperatorNotFound}
		rec := operatorRequest(operatorTestRouter(svc, nil),
			"/operator/session", trustedPeer, "nobody@example.com")
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Equal(t, "nobody@example.com", svc.gotEmail)
		assert.NotContains(t, rec.Body.String(), "nobody@example.com")
		assert.NotContains(t, rec.Body.String(), "not_found")
	})

	t.Run("registered operator sees its identity", func(t *testing.T) {
		svc := &fakeOperatorIdentity{operator: platformadmin.Operator{ID: "op-1", Email: "a@b.com"}}
		rec := operatorRequest(operatorTestRouter(svc, nil),
			"/operator/session", trustedPeer, "a@b.com")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"role":"platform_admin"`)
		assert.Contains(t, rec.Body.String(), `"email":"a@b.com"`)
	})
}

func TestOperatorHandlersRequireCSRFAndMapErrors(t *testing.T) {
	const trustedPeer = "172.18.0.5:1234"
	svc := &fakeOperatorIdentity{operator: platformadmin.Operator{ID: "op-1", Email: "a@b.com"}}

	t.Run("mutation without header is rejected", func(t *testing.T) {
		router := operatorTestRouter(svc, &fakePlatformAdmin{})
		req := httptest.NewRequest(http.MethodPost, "/operator/organizations", nil)
		req.RemoteAddr = trustedPeer
		req.Header.Set("Cf-Access-Authenticated-User-Email", "a@b.com")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "csrf_rejected")
	})

	t.Run("forbidden service error maps to 403", func(t *testing.T) {
		admin := &fakePlatformAdmin{err: identity.ErrForbidden}
		rec := operatorRequest(operatorTestRouter(svc, admin),
			"/operator/organizations", trustedPeer, "a@b.com")
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.True(t, admin.listed)
	})

	t.Run("dashboard and audit read", func(t *testing.T) {
		admin := &fakePlatformAdmin{}
		rec := operatorRequest(operatorTestRouter(svc, admin),
			"/operator/audit", trustedPeer, "a@b.com")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, admin.audited)
		assert.Contains(t, rec.Body.String(), "org.created")
	})

	t.Run("unexpected lookup error fails closed", func(t *testing.T) {
		failing := &fakeOperatorIdentity{err: errors.New("db down")}
		rec := operatorRequest(operatorTestRouter(failing, nil),
			"/operator/session", trustedPeer, "a@b.com")
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})
}
