package api

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxLoginBodyBytes = 4096

// LoginService is the login dependency used by handlers.
type LoginService interface {
	Login(ctx context.Context, req identity.LoginRequest) (identity.LoginResult, error)
}

// Login implements POST /auth/login.
func (s *Server) Login(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLoginBodyBytes)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: InvalidRequest})
		return
	}
	if strings.TrimSpace(req.Email) == "" || len(req.Email) > 254 || req.Password == "" || len(req.Password) > 1024 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: InvalidRequest})
		return
	}

	result, err := s.auth.Login(c.Request.Context(), identity.LoginRequest{Email: req.Email, Password: req.Password})
	if err == nil {
		c.JSON(http.StatusOK, LoginResponse{MfaRequired: result.MFARequired})
		return
	}

	var rl *identity.RateLimitError
	if errors.As(err, &rl) {
		seconds := int(math.Ceil(rl.RetryAfter.Seconds()))
		if seconds < 1 {
			seconds = 1
		}
		c.Header("Retry-After", strconv.Itoa(seconds))
		c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: RateLimited})
		return
	}
	if errors.Is(err, identity.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: InvalidCredentials})
		return
	}
	if errors.Is(err, identity.ErrRateLimiterUnavailable) {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: Unavailable})
		return
	}
	s.logger.Error("login failed", zap.Error(err))
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: InternalError})
}
