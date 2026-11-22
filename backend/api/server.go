package api

import (
	"context"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/health"
	"go.uber.org/zap"
)

// MFAService is the MFA dependency used by handlers.
type MFAService interface {
	BeginEnrollment(ctx context.Context, actor identity.Actor) (identity.MFAEnrollment, error)
	ConfirmEnrollment(ctx context.Context, actor identity.Actor, code string) ([]string, error)
	VerifyMFA(ctx context.Context, ps identity.PendingSession, code string, info identity.ClientInfo) error
	StepUp(ctx context.Context, actor identity.Actor, password, code string) error
}

// Server implements the generated OpenAPI server interface.
type Server struct {
	health *health.Service
	auth   AuthService
	mfa    MFAService
	access AccessService
	logger *zap.Logger
	cfg    config.Config
}

// SetAccessService attaches the access service.
func (s *Server) SetAccessService(a AccessService) { s.access = a }

// SetMFAService attaches the MFA service.
func (s *Server) SetMFAService(m MFAService) { s.mfa = m }

// NewServer creates an API server.
func NewServer(healthService *health.Service, auth AuthService, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{health: healthService, auth: auth, logger: logger, cfg: config.Config{CookieSecure: true}}
}

// SetConfig attaches runtime configuration.
func (s *Server) SetConfig(cfg config.Config) { s.cfg = cfg }
