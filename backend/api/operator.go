package api

import (
	"errors"
	"net/http"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/httpx"
	"github.com/TaushifReza/email-platform/backend/internal/platformadmin"
	"github.com/gin-gonic/gin"
)

func (s *Server) GetOperatorSession(c *gin.Context) {
	op, ok := OperatorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	role := identity.PlatformAdminActor(op.ID).Role
	c.JSON(http.StatusOK, OperatorSession{Id: op.ID, Email: op.Email, Role: role})
}

func (s *Server) ListOperatorOrganizations(c *gin.Context, params ListOperatorOrganizationsParams) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	orgs, err := s.platformAdmin.ListOrganizations(c.Request.Context(), actor, operatorPage(params.Limit, params.Offset))
	if err != nil {
		respondOperatorError(c, err)
		return
	}
	data := make([]Organization, 0, len(orgs))
	for _, org := range orgs {
		data = append(data, toAPIOrganization(org))
	}
	c.JSON(http.StatusOK, OrganizationList{
		Data: &data,
		Meta: &struct {
			HasMore *bool `json:"hasMore,omitempty"`
		}{HasMore: boolPtr(len(orgs) >= operatorLimit(params.Limit))},
	})
}

func (s *Server) CreateOperatorOrganization(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	in := platformadmin.CreateOrganizationInput{Name: req.Name}
	if req.DefaultSenderName != nil {
		in.DefaultSenderName = *req.DefaultSenderName
	}
	if req.DefaultReplyTo != nil {
		in.DefaultReplyTo = *req.DefaultReplyTo
	}
	if req.TimeZone != nil {
		in.TimeZone = *req.TimeZone
	}
	org, err := s.platformAdmin.CreateOrganization(c.Request.Context(), actor, in, s.clientInfo(c))
	if err != nil {
		if errors.Is(err, identity.ErrForbidden) {
			RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
			return
		}
		RespondProblem(c, http.StatusBadRequest, CodeValidationFailed, err.Error(), nil)
		return
	}
	c.JSON(http.StatusCreated, toAPIOrganization(org))
}

func (s *Server) SuspendOrganization(c *gin.Context, orgID string) {
	s.changeOperatorOrganizationState(c, orgID, true)
}

func (s *Server) ReinstateOrganization(c *gin.Context, orgID string) {
	s.changeOperatorOrganizationState(c, orgID, false)
}

func (s *Server) changeOperatorOrganizationState(c *gin.Context, orgID string, suspend bool) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var err error
	if suspend {
		err = s.platformAdmin.SuspendOrganization(c.Request.Context(), actor, orgID, s.clientInfo(c))
	} else {
		err = s.platformAdmin.ReinstateOrganization(c.Request.Context(), actor, orgID, s.clientInfo(c))
	}
	switch {
	case err == nil:
		RespondNoContent(c)
	case errors.Is(err, identity.ErrForbidden):
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
	default:
		RespondProblem(c, http.StatusNotFound, CodeNotFound, "Organization not found", nil)
	}
}

func (s *Server) UpdateOrganizationLimits(c *gin.Context, orgID string) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req OrganizationLimitsUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	in := platformadmin.LimitsInput{
		StorageQuotaBytes: req.StorageQuotaBytes,
		CapsEnabled:       req.CapsEnabled,
	}
	if req.DailyEmailCap != nil {
		in.DailyEmailCap = *req.DailyEmailCap
	}
	if req.CampaignRecipientCap != nil {
		in.CampaignRecipientCap = *req.CampaignRecipientCap
	}
	org, err := s.platformAdmin.UpdateLimits(c.Request.Context(), actor, orgID, in, s.clientInfo(c))
	if err != nil {
		respondOperatorError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAPIOrganization(org))
}

func (s *Server) ListOperatorAudit(c *gin.Context, params ListOperatorAuditParams) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	page := operatorPage(params.Limit, params.Offset)
	entries, err := s.platformAdmin.ListOperatorAudit(c.Request.Context(), actor, page)
	if err != nil {
		respondOperatorError(c, err)
		return
	}
	data := make([]AuditEntry, 0, len(entries))
	for _, e := range entries {
		data = append(data, toAPIAuditEntry(e))
	}
	c.JSON(http.StatusOK, AuditEntryList{
		Data: &data,
		Meta: &struct {
			HasMore *bool `json:"hasMore,omitempty"`
		}{HasMore: boolPtr(len(entries) >= operatorLimit(params.Limit))},
	})
}

func (s *Server) clientInfo(c *gin.Context) identity.ClientInfo {
	return identity.ClientInfo{
		IP:        httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs),
		UserAgent: c.Request.UserAgent(),
	}
}

func respondOperatorError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, identity.ErrForbidden):
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
	case errors.Is(err, platformadmin.ErrNotFound):
		RespondProblem(c, http.StatusNotFound, CodeNotFound, "Organization not found", nil)
	default:
		RespondProblem(c, http.StatusBadRequest, CodeValidationFailed, err.Error(), nil)
	}
}

func operatorPage(limit, offset *int) platformadmin.Page {
	page := platformadmin.Page{}
	if limit != nil {
		page.Limit = *limit
	}
	if offset != nil {
		page.Offset = *offset
	}
	return page
}

func operatorLimit(limit *int) int {
	if limit != nil && *limit > 0 {
		return *limit
	}
	return 50
}

func toAPIAuditEntry(e audit.Entry) AuditEntry {
	out := AuditEntry{
		Action: &e.Action, Outcome: &e.Outcome,
		CreatedAt: &e.At,
	}
	if e.ID != "" {
		out.Id = &e.ID
	}
	if e.ActorID != "" {
		out.ActorId = &e.ActorID
	}
	if e.ActorRole != "" {
		out.ActorRole = &e.ActorRole
	}
	if e.ActorType != "" {
		out.ActorType = &e.ActorType
	}
	if e.TargetType != "" {
		out.TargetType = &e.TargetType
	}
	if e.TargetID != "" {
		out.TargetId = &e.TargetID
	}
	if e.SourceIP != "" {
		out.SourceIp = &e.SourceIP
	}
	if e.Client != "" {
		out.Client = &e.Client
	}
	if len(e.Details) > 0 {
		details := e.Details
		out.Details = &details
	}
	return out
}
