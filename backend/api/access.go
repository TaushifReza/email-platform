package api

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/httpx"
	"github.com/gin-gonic/gin"
)

// AccessService is the invitation/password-reset dependency used by handlers.
type AccessService interface {
	IssueInvitation(ctx context.Context, actor identity.Actor, email, role string) error
	AcceptInvitation(ctx context.Context, token, password string) error
	RequestPasswordReset(ctx context.Context, email string, info identity.ClientInfo) error
	ResetPassword(ctx context.Context, token, password string) error
}

const maxAccessBodyBytes = 4096

type acceptInvitationRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type resetRequest struct {
	Email string `json:"email"`
}

type resetConfirmRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// AcceptInvitation implements POST /auth/invitations/accept.
func (s *Server) AcceptInvitation(c *gin.Context) { //nolint:dupl
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAccessBodyBytes)
	var req acceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	if strings.TrimSpace(req.Token) == "" || req.Password == "" {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request fields", nil)
		return
	}
	if err := s.access.AcceptInvitation(c.Request.Context(), req.Token, req.Password); err != nil {
		if identity.IsPasswordPolicyError(err) {
			RespondProblem(c, http.StatusUnprocessableEntity, CodeValidationFailed, "Password does not meet policy", nil)
			return
		}
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid or expired invitation", nil)
		return
	}
	RespondNoContent(c)
}

// RequestPasswordReset implements POST /auth/password-reset/request.
func (s *Server) RequestPasswordReset(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAccessBodyBytes)
	var req resetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	info := identity.ClientInfo{
		IP:        httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs),
		UserAgent: c.Request.UserAgent(),
	}
	err := s.access.RequestPasswordReset(c.Request.Context(), req.Email, info)
	var rl *identity.RateLimitError
	if errors.As(err, &rl) {
		seconds := int(math.Ceil(rl.RetryAfter.Seconds()))
		if seconds < 1 {
			seconds = 1
		}
		c.Header("Retry-After", strconv.Itoa(seconds))
		RespondProblem(c, http.StatusTooManyRequests, CodeRateLimited, "Too many attempts", nil)
		return
	}
	RespondNoContent(c)
}

// ResetPassword implements POST /auth/password-reset/confirm.
func (s *Server) ResetPassword(c *gin.Context) { //nolint:dupl
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAccessBodyBytes)
	var req resetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	if strings.TrimSpace(req.Token) == "" || req.Password == "" {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request fields", nil)
		return
	}
	if err := s.access.ResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		if identity.IsPasswordPolicyError(err) {
			RespondProblem(c, http.StatusUnprocessableEntity, CodeValidationFailed, "Password does not meet policy", nil)
			return
		}
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid or expired reset link", nil)
		return
	}
	RespondNoContent(c)
}
