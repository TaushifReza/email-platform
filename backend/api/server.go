package api

import (
	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/health"
	"go.uber.org/zap"
)

// Server implements the generated OpenAPI server interface.
type Server struct {
	health *health.Service
	auth   AuthService
	logger *zap.Logger
	cfg    config.Config
}

// NewServer creates an API server.
func NewServer(healthService *health.Service, auth AuthService, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{health: healthService, auth: auth, logger: logger, cfg: config.Config{CookieSecure: true}}
}

// SetConfig attaches runtime configuration.
func (s *Server) SetConfig(cfg config.Config) { s.cfg = cfg }
