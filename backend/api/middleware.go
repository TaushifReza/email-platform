package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platformadmin"
	"github.com/gin-gonic/gin"
)

// AuthService is the session dependency used by handlers and middleware.
type AuthService interface {
	SignIn(ctx context.Context, req identity.LoginRequest, info identity.ClientInfo) (identity.SessionResult, error)
	Authenticate(ctx context.Context, token string) (identity.Actor, error)
	AuthenticatePending(ctx context.Context, token string) (identity.PendingSession, error)
	SignOut(ctx context.Context, token string) error
}

// contextKeyActor is the gin context key for the authenticated actor.
const contextKeyActor = "actor"

// contextKeyOperator is the gin context key for the authenticated operator.
const contextKeyOperator = "operator"

// ActorFrom returns the authenticated actor, if any.
func ActorFrom(c *gin.Context) (identity.Actor, bool) {
	v, ok := c.Get(contextKeyActor)
	if !ok {
		return identity.Actor{}, false
	}
	a, ok := v.(identity.Actor)
	return a, ok
}

// OperatorFrom returns the operator identity asserted by the trusted IdP.
func OperatorFrom(c *gin.Context) (platformadmin.Operator, bool) {
	v, ok := c.Get(contextKeyOperator)
	if !ok {
		return platformadmin.Operator{}, false
	}
	op, ok := v.(platformadmin.Operator)
	return op, ok
}

func isOperatorPath(path string) bool {
	return strings.HasPrefix(path, "/operator/")
}

func isPublicPath(path string) bool {
	switch {
	case path == "/healthz" || path == "/readyz":
		return true
	case path == "/auth/login":
		return true
	case path == "/auth/invitations/accept":
		return true
	case path == "/auth/password-reset/request":
		return true
	case path == "/auth/password-reset/confirm":
		return true
	case path == "/docs" || path == "/openapi.yaml":
		return true
	case strings.HasPrefix(path, "/u/"):
		return true
	}
	return false
}

func isPendingMFAPath(path string) bool {
	return path == "/auth/mfa/verify" || path == "/auth/mfa/enroll" ||
		path == "/auth/mfa/enroll/confirm" || path == "/auth/logout" ||
		path == "/auth/step-up"
}

func isCSRFExemptPath(path string) bool {
	return strings.HasPrefix(path, "/u/")
}

func AuthMiddleware(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isOperatorPath(c.Request.URL.Path) {
			s.authenticateOperator(c)
			return
		}
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		cfg := s.cfg
		token, err := c.Cookie(cookieName(cfg))
		if err != nil || token == "" {
			RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
			c.Abort()
			return
		}
		actor, err := s.auth.Authenticate(c.Request.Context(), token)
		if err == nil {
			c.Set(contextKeyActor, actor)
			c.Next()
			return
		}
		if errors.Is(err, identity.ErrMFARequired) && isPendingMFAPath(c.Request.URL.Path) {
			ps, perr := s.auth.AuthenticatePending(c.Request.Context(), token)
			if perr == nil {
				c.Set(contextKeyActor, identity.Actor{
					UserID:         ps.UserID,
					OrganizationID: ps.OrganizationID,
					Role:           ps.Role,
					SessionID:      ps.SessionID,
				})
				c.Next()
				return
			}
		}
		if errors.Is(err, identity.ErrMFARequired) {
			RespondProblem(c, http.StatusUnauthorized, CodeMFARequired, "MFA required", nil)
			c.Abort()
			return
		}
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		c.Abort()
	}
}

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions { //nolint:lll
			c.Next()
			return
		}
		if isCSRFExemptPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		if c.GetHeader(csrfHeaderName) == "" {
			RespondProblem(c, http.StatusForbidden, CodeCSRFRejected, "CSRF rejected", nil)
			c.Abort()
			return
		}
		if origin := c.GetHeader("Origin"); origin != "" {
			if !sameOrigin(origin, c.Request.Host) {
				RespondProblem(c, http.StatusForbidden, CodeCSRFRejected, "CSRF rejected", nil)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func sameOrigin(origin, host string) bool {
	// origin like http://host:port
	o := origin
	if i := strings.Index(o, "://"); i >= 0 {
		o = o[i+3:]
	}
	if i := strings.Index(o, "/"); i >= 0 {
		o = o[:i]
	}
	return strings.EqualFold(o, host)
}
