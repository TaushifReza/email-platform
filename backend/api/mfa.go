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
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	if s.mfa == nil {
		RespondProblem(c, http.StatusServiceUnavailable, CodeUnavailable, "Service unavailable", nil)
		return
	}
	res, err := s.mfa.BeginEnrollment(c.Request.Context(), actor)
	if err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request", nil)
		return
	}
	c.JSON(http.StatusOK, mfaEnrollResponse{Secret: res.Secret, OtpauthURI: res.OtpauthURI})
}

// ConfirmMFAEnrollment implements POST /auth/mfa/enroll/confirm.
func (s *Server) ConfirmMFAEnrollment(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	if s.mfa == nil {
		RespondProblem(c, http.StatusServiceUnavailable, CodeUnavailable, "Service unavailable", nil)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMFABodyBytes)
	var req mfaEnrollConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request", nil)
		return
	}
	codes, err := s.mfa.ConfirmEnrollment(c.Request.Context(), actor, req.Code)
	if err != nil {
		RespondProblem(c, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials", nil)
		return
	}
	c.JSON(http.StatusOK, mfaRecoveryResponse{RecoveryCodes: codes})
}

// VerifyMFA implements POST /auth/mfa/verify.
func (s *Server) VerifyMFA(c *gin.Context) {
	token, _ := c.Cookie(cookieName(s.cfg)) //nolint:errcheck
	ps, err := s.auth.AuthenticatePending(c.Request.Context(), token)
	if err != nil {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	if s.mfa == nil {
		RespondProblem(c, http.StatusServiceUnavailable, CodeUnavailable, "Service unavailable", nil)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMFABodyBytes)
	var req mfaVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request", nil)
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
			RespondProblem(c, http.StatusLocked, CodeLocked, "Account locked", nil)
			return
		}
		var rl *identity.RateLimitError
		if errors.As(err, &rl) {
			seconds := int(rl.RetryAfter.Seconds())
			if seconds < 1 {
				seconds = 1
			}
			c.Header("Retry-After", itoa(seconds))
			RespondProblem(c, http.StatusTooManyRequests, CodeRateLimited, "Too many attempts", nil)
			return
		}
		RespondProblem(c, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials", nil)
		return
	}
	RespondNoContent(c)
}

// StepUp implements POST /auth/step-up.
func (s *Server) StepUp(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	if s.mfa == nil {
		RespondProblem(c, http.StatusServiceUnavailable, CodeUnavailable, "Service unavailable", nil)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMFABodyBytes)
	var req stepUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request", nil)
		return
	}
	if err := s.mfa.StepUp(c.Request.Context(), actor, req.Password, req.Code); err != nil {
		RespondProblem(c, http.StatusUnauthorized, CodeInvalidCredentials, "Invalid credentials", nil)
		return
	}
	RespondNoContent(c)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
