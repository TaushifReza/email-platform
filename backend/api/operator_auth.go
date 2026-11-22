package api

import (
	"errors"
	"net/http"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/httpx"
	"github.com/TaushifReza/email-platform/backend/internal/platformadmin"
	"github.com/gin-gonic/gin"
)

// authenticateOperator is the trust boundary for the operator console. The
// identity header is only considered when the immediate peer is a configured
// trusted proxy; Cloudflare Access (or the dev shim) is the authenticator and
// the operator table is the allowlist.
func (s *Server) authenticateOperator(c *gin.Context) {
	if s.operatorIdentity == nil {
		RespondProblem(c, http.StatusServiceUnavailable, CodeUnavailable, "Operator authentication unavailable", nil)
		c.Abort()
		return
	}
	if !httpx.IsTrustedPeer(c.Request, s.cfg.TrustedProxyCIDRs) {
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Untrusted proxy", nil)
		c.Abort()
		return
	}
	email := c.GetHeader(s.cfg.OperatorHeader())
	if email == "" {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "No operator identity", nil)
		c.Abort()
		return
	}
	op, err := s.operatorIdentity.Authenticate(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, platformadmin.ErrOperatorNotFound) {
			RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		} else {
			RespondProblem(c, http.StatusServiceUnavailable, CodeUnavailable, "Operator lookup unavailable", nil)
		}
		c.Abort()
		return
	}
	c.Set(contextKeyOperator, op)
	c.Set(contextKeyActor, identity.PlatformAdminActor(op.ID))
	c.Next()
}
