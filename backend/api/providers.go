package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/httpx"
	"github.com/TaushifReza/email-platform/backend/internal/providers"
	"github.com/gin-gonic/gin"
)

type ProviderService interface {
	List(ctx context.Context, actor identity.Actor) ([]providers.Provider, error)
	Create(ctx context.Context, actor identity.Actor, in providers.Input, info identity.ClientInfo) (providers.Provider, error)            //nolint:lll
	Update(ctx context.Context, actor identity.Actor, id string, u providers.Update, info identity.ClientInfo) (providers.Provider, error) //nolint:lll
	Rotate(ctx context.Context, actor identity.Actor, id, username, password string, info identity.ClientInfo) (providers.Provider, error) //nolint:lll
	Deactivate(ctx context.Context, actor identity.Actor, id string, info identity.ClientInfo) (providers.Provider, error)
}

func toAPIProvider(p providers.Provider) Provider {
	id := p.ID
	host := p.Host
	port := p.Port
	tls := p.TLSMode
	user := p.Username
	from := p.FromAddress
	def := p.IsDefault
	status := p.Status
	daily := p.DailyLimit
	minute := p.MinuteLimit
	hour := p.HourLimit
	cv := p.ConfigVersion
	created := p.CreatedAt
	updated := p.UpdatedAt
	out := Provider{
		Id: &id, Host: &host, Port: &port, TlsMode: &tls, Username: &user,
		FromAddress: &from, IsDefault: &def, Status: &status, DailyLimit: &daily,
		MinuteLimit: &minute, HourLimit: &hour, ConfigVersion: &cv, CreatedAt: &created, UpdatedAt: &updated,
	}
	if p.VerifiedAt != nil {
		v := *p.VerifiedAt
		out.VerifiedAt = &v
	}
	return out
}

func (s *Server) ListProviders(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	list, err := s.providerSvc.List(c.Request.Context(), actor)
	if err != nil {
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	data := make([]Provider, 0, len(list))
	for _, p := range list {
		data = append(data, toAPIProvider(p))
	}
	c.JSON(http.StatusOK, ProviderList{Data: &data})
}

func (s *Server) CreateProvider(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	in := providers.Input{
		Host: req.Host, Port: req.Port,
		TLSMode: req.TlsMode, Username: req.Username,
		Password: req.Password, FromAddress: req.FromAddress,
	}
	if req.IsDefault != nil {
		in.IsDefault = *req.IsDefault
	}
	if req.DailyLimit != nil {
		in.DailyLimit = *req.DailyLimit
	}
	if req.MinuteLimit != nil {
		in.MinuteLimit = *req.MinuteLimit
	}
	if req.HourLimit != nil {
		in.HourLimit = *req.HourLimit
	}
	info := identity.ClientInfo{IP: httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs), UserAgent: c.Request.UserAgent()}
	p, err := s.providerSvc.Create(c.Request.Context(), actor, in, info)
	if err != nil {
		if err.Error() == "providers: host not allowed" {
			RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "host not allowed", nil)
			return
		}
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	c.JSON(http.StatusCreated, toAPIProvider(p))
}

func (s *Server) UpdateProvider(c *gin.Context, providerID string) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	u := providers.Update{}
	if req.Host != nil {
		v := *req.Host
		u.Host = &v
	}
	if req.Port != nil {
		v := *req.Port
		u.Port = &v
	}
	if req.TlsMode != nil {
		v := *req.TlsMode
		u.TLSMode = &v
	}
	if req.FromAddress != nil {
		v := *req.FromAddress
		u.FromAddress = &v
	}
	if req.IsDefault != nil {
		v := *req.IsDefault
		u.IsDefault = &v
	}
	if req.DailyLimit != nil {
		v := *req.DailyLimit
		u.DailyLimit = &v
	}
	if req.MinuteLimit != nil {
		v := *req.MinuteLimit
		u.MinuteLimit = &v
	}
	if req.HourLimit != nil {
		v := *req.HourLimit
		u.HourLimit = &v
	}
	info := identity.ClientInfo{IP: httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs), UserAgent: c.Request.UserAgent()}
	p, err := s.providerSvc.Update(c.Request.Context(), actor, providerID, u, info)
	if err != nil {
		if err.Error() == "providers: host not allowed" {
			RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "host not allowed", nil)
			return
		}
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	c.JSON(http.StatusOK, toAPIProvider(p))
}

func (s *Server) RotateProviderSecret(c *gin.Context, providerID string) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req RotateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	username := ""
	if req.Username != nil {
		username = *req.Username
	}
	info := identity.ClientInfo{IP: httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs), UserAgent: c.Request.UserAgent()}
	p, err := s.providerSvc.Rotate(c.Request.Context(), actor, providerID, username, req.Password, info)
	if err != nil {
		if errors.Is(err, identity.ErrMFAStepUp) {
			RespondProblem(c, http.StatusForbidden, CodeForbidden, "step-up required", nil)
			return
		}
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	c.JSON(http.StatusOK, toAPIProvider(p))
}

func (s *Server) DeactivateProvider(c *gin.Context, providerID string) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	info := identity.ClientInfo{IP: httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs), UserAgent: c.Request.UserAgent()}
	_, err := s.providerSvc.Deactivate(c.Request.Context(), actor, providerID, info)
	if err != nil {
		if errors.Is(err, providers.ErrDefault) {
			RespondProblem(c, http.StatusConflict, CodeConflict, "provider is default", nil)
			return
		}
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	RespondNoContent(c)
}
