package api

import (
	"context"
	"net/http"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/health"
	"github.com/gin-gonic/gin"
)

const readinessTimeout = 5 * time.Second

// HealthServer maps the generated OpenAPI health contract to the health service.
type HealthServer struct {
	health *health.Service
}

// NewHealthServer creates the health endpoint implementation.
func NewHealthServer(service *health.Service) *HealthServer {
	return &HealthServer{health: service}
}

// GetHealth implements GET /healthz.
func (s *HealthServer) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}

// GetReady implements GET /readyz.
func (s *HealthServer) GetReady(c *gin.Context) {
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
