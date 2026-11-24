package api

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/httpx"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxLoginBodyBytes = 4096

// Login implements POST /auth/login.
func (s *Server) Login(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLoginBodyBytes)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	if strings.TrimSpace(req.Email) == "" || len(req.Email) > 254 || req.Password == "" || len(req.Password) > 1024 {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request fields", nil)
		return
	}

	info := identity.ClientInfo{
		IP:        httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs),
		UserAgent: c.Request.UserAgent(),
	}
	result, err := s.auth.SignIn(c.Request.Context(), identity.LoginRequest{Email: req.Email, Password: req.Password}, info) //nolint:lll
	if err == nil {
		setSessionCookie(c.Writer, s.cfg, result.Token, int(time.Until(result.AbsoluteExpiresAt).Seconds()))
		c.JSON(http.StatusOK, LoginResponse{
			MfaRequired:           result.MFARequired,
			MfaEnrollmentRequired: result.MFAEnrollmentRequired,
		})
		return
	}

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
	if errors.Is(err, identity.ErrInvalidCredentials) {
		RespondProblem(c, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials", nil)
		return
	}
	if errors.Is(err, identity.ErrRateLimiterUnavailable) {
		RespondProblem(c, http.StatusServiceUnavailable, CodeUnavailable, "Authentication unavailable", nil)
		return
	}
	s.logger.Error("login failed", zap.Error(err))
	RespondProblem(c, http.StatusInternalServerError, CodeInternalError, "Internal server error", nil)
}

// OperatorLogin implements POST /auth/operator/login.
func (s *Server) OperatorLogin(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLoginBodyBytes)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	if strings.TrimSpace(req.Email) == "" || req.Password == "" {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request fields", nil)
		return
	}
	if s.operatorAuth == nil {
		RespondProblem(c, http.StatusServiceUnavailable, CodeUnavailable, "Operator authentication unavailable", nil)
		return
	}
	info := identity.ClientInfo{
		IP:        httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs),
		UserAgent: c.Request.UserAgent(),
	}
	result, err := s.operatorAuth.Login(c.Request.Context(), req.Email, req.Password, info)
	if err != nil {
		RespondProblem(c, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials", nil)
		return
	}
	setSessionCookie(c.Writer, s.cfg, result.Token, int(time.Until(result.AbsoluteExpiresAt).Seconds()))
	c.JSON(http.StatusOK, gin.H{"operator_id": result.OperatorID})
}

// Logout implements POST /auth/logout.
func (s *Server) Logout(c *gin.Context) {
	token, _ := c.Cookie(cookieName(s.cfg))        //nolint:errcheck
	_ = s.auth.SignOut(c.Request.Context(), token) //nolint:errcheck
	clearSessionCookie(c.Writer, s.cfg)
	RespondNoContent(c)
}

// GetSession implements GET /auth/session.
func (s *Server) GetSession(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	c.JSON(http.StatusOK, SessionResponse{
		UserId:         actor.UserID,
		OrganizationId: actor.OrganizationID,
		Role:           actor.Role,
		MfaVerified:    true,
	})
}
