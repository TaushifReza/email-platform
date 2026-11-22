package api

import (
	"net/http"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/httpx"
	"github.com/TaushifReza/email-platform/backend/internal/tenancy"
	"github.com/gin-gonic/gin"
)

func (s *Server) GetOrganization(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	org, err := s.organizations.GetSettings(c.Request.Context(), actor)
	if err != nil {
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	c.JSON(http.StatusOK, toAPIOrganization(org))
}

func (s *Server) UpdateOrganization(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req OrganizationProfileUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	p := tenancy.ProfileUpdate{}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.DefaultSenderName != nil {
		p.DefaultSenderName = *req.DefaultSenderName
	}
	if req.DefaultReplyTo != nil {
		p.DefaultReplyTo = *req.DefaultReplyTo
	}
	if req.TimeZone != nil {
		p.TimeZone = *req.TimeZone
	}
	info := identity.ClientInfo{IP: httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs), UserAgent: c.Request.UserAgent()}
	org, err := s.organizations.UpdateProfile(c.Request.Context(), actor, p, info)
	if err != nil {
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	c.JSON(http.StatusOK, toAPIOrganization(org))
}

func (s *Server) UpdateOrganizationPolicy(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req OrganizationPolicyUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	p := tenancy.PolicyUpdate{}
	if req.RequireApproval != nil {
		p.RequireApproval = *req.RequireApproval
	}
	if req.RetentionDays != nil {
		p.RetentionDays = *req.RetentionDays
	}
	info := identity.ClientInfo{IP: httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs), UserAgent: c.Request.UserAgent()}
	org, err := s.organizations.UpdatePolicy(c.Request.Context(), actor, p, info)
	if err != nil {
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	c.JSON(http.StatusOK, toAPIOrganization(org))
}

func toAPIOrganization(o tenancy.Organization) Organization {
	id := o.ID
	name := o.Name
	dsn := o.DefaultSenderName
	dr := o.DefaultReplyTo
	tz := o.TimeZone
	status := o.Status
	ra := o.RequireApproval
	rd := o.RetentionDays
	sq := o.StorageQuotaBytes
	ce := o.CapsEnabled
	return Organization{
		Id: &id, Name: &name, DefaultSenderName: &dsn, DefaultReplyTo: &dr,
		TimeZone: &tz, Status: &status, RequireApproval: &ra, RetentionDays: &rd,
		StorageQuotaBytes: &sq, CapsEnabled: &ce,
	}
}
