package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/httpx"
	"github.com/gin-gonic/gin"
)

const maxMFABodyBytes = 1024

type mfaVerifyRequest struct {
	Code string `binding:"required" json:"code"`
}

type mfaEnrollConfirmRequest struct {
	Code string `binding:"required" json:"code"`
}

type stepUpRequest struct {
	Password string `binding:"required" json:"password"`
	Code     string `binding:"required" json:"code"`
}

type mfaEnrollResponse struct {
	Secret     string `json:"secret"`
	OtpauthURI string `json:"otpauth_uri"`
}

type mfaRecoveryResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// EnrollMFA implements POST /auth/mfa/enroll.
func (s *Server) EnrollMFA(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: Unauthorized})
		return
	}
	if s.mfa == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: Unavailable})
		return
	}
	res, err := s.mfa.BeginEnrollment(c.Request.Context(), actor)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: InvalidRequest})
		return
	}
	c.JSON(http.StatusOK, mfaEnrollResponse{Secret: res.Secret, OtpauthURI: res.OtpauthURI})
}

// ConfirmMFAEnrollment implements POST /auth/mfa/enroll/confirm.
func (s *Server) ConfirmMFAEnrollment(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: Unauthorized})
		return
	}
	if s.mfa == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: Unavailable})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMFABodyBytes)
	var req mfaEnrollConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: InvalidRequest})
		return
	}
	codes, err := s.mfa.ConfirmEnrollment(c.Request.Context(), actor, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: InvalidCredentials})
		return
	}
	c.JSON(http.StatusOK, mfaRecoveryResponse{RecoveryCodes: codes})
}

// VerifyMFA implements POST /auth/mfa/verify.
func (s *Server) VerifyMFA(c *gin.Context) {
	token, _ := c.Cookie(cookieName(s.cfg)) //nolint:errcheck
	ps, err := s.auth.AuthenticatePending(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: Unauthorized})
		return
	}
	if s.mfa == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: Unavailable})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMFABodyBytes)
	var req mfaVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: InvalidRequest})
		return
	}
	info := identity.ClientInfo{
		IP:        httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs),
		UserAgent: c.Request.UserAgent(),
	}
	pending := identity.PendingSession{
		SessionID:      ps.SessionID,
		OrganizationID: ps.OrganizationID,
		UserID:         ps.UserID,
		Role:           ps.Role,
	}
	if err := s.mfa.VerifyMFA(c.Request.Context(), pending, req.Code, info); err != nil {
		if errors.Is(err, identity.ErrMFALocked) {
			c.JSON(http.StatusLocked, ErrorResponse{Error: "locked"})
			return
		}
		var rl *identity.RateLimitError
		if errors.As(err, &rl) {
			seconds := int(rl.RetryAfter.Seconds())
			if seconds < 1 {
				seconds = 1
			}
			c.Header("Retry-After", itoa(seconds))
			c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: RateLimited})
			return
		}
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: InvalidCredentials})
		return
	}
	c.Status(http.StatusNoContent)
}

// StepUp implements POST /auth/step-up.
func (s *Server) StepUp(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: Unauthorized})
		return
	}
	if s.mfa == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: Unavailable})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMFABodyBytes)
	var req stepUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: InvalidRequest})
		return
	}
	if err := s.mfa.StepUp(c.Request.Context(), actor, req.Password, req.Code); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: InvalidCredentials})
		return
	}
	c.Status(http.StatusNoContent)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
