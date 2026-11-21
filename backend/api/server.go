package api

import (
	"github.com/TaushifReza/email-platform/backend/internal/platform/health"
	"go.uber.org/zap"
)

// Server implements the generated OpenAPI server interface.
type Server struct {
	health *health.Service
	auth   LoginService
	logger *zap.Logger
}

// NewServer creates an API server.
func NewServer(healthService *health.Service, auth LoginService, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{health: healthService, auth: auth, logger: logger}
}
