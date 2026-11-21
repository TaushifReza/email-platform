package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readinessTimeout = 5 * time.Second

// GetHealth implements GET /healthz.
func (s *Server) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}

// GetReady implements GET /readyz.
func (s *Server) GetReady(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
	defer cancel()

	status := s.health.Ready(ctx)
	body := ReadyResponse{
		Status: Ok,
		Checks: struct {
			Db    string `json:"db"` //nolint:stylecheck // Matches the generated OpenAPI field name.
			Redis string `json:"redis"`
		}{
			Db:    status.Database,
			Redis: status.Redis,
		},
	}

	if !status.Ready {
		body.Status = Degraded
		c.JSON(http.StatusServiceUnavailable, body)
		return
	}

	c.JSON(http.StatusOK, body)
}
