package api

import (
	"context"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/health"
	"github.com/TaushifReza/email-platform/backend/internal/platformadmin"
	"github.com/TaushifReza/email-platform/backend/internal/tenancy"
	"go.uber.org/zap"
)

// MFAService is the MFA dependency used by handlers.
type MFAService interface {
	BeginEnrollment(ctx context.Context, actor identity.Actor) (identity.MFAEnrollment, error)
	ConfirmEnrollment(ctx context.Context, actor identity.Actor, code string) ([]string, error)
	VerifyMFA(ctx context.Context, ps identity.PendingSession, code string, info identity.ClientInfo) error
	StepUp(ctx context.Context, actor identity.Actor, password, code string) error
}

// OrganizationService is the organization dependency.
type OrganizationService interface {
	GetSettings(ctx context.Context, actor identity.Actor) (tenancy.Organization, error)
	UpdateProfile(ctx context.Context, actor identity.Actor, p tenancy.ProfileUpdate, info identity.ClientInfo) (tenancy.Organization, error) //nolint:lll
	UpdatePolicy(ctx context.Context, actor identity.Actor, p tenancy.PolicyUpdate, info identity.ClientInfo) (tenancy.Organization, error)   //nolint:lll
}

// UserService is the user management dependency.
type UserService interface {
	List(ctx context.Context, actor identity.Actor) ([]identity.User, error)
	Invite(ctx context.Context, actor identity.Actor, email, role string, info identity.ClientInfo) error
	Update(ctx context.Context, actor identity.Actor, userID string, change identity.UserChange, info identity.ClientInfo) error //nolint:lll
}

// OperatorIdentityService resolves the trusted IdP identity to an operator.
type OperatorIdentityService interface {
	Authenticate(ctx context.Context, email string) (platformadmin.Operator, error)
	AuthenticateWithPassword(ctx context.Context, email, password string) (platformadmin.Operator, error)
}

// OperatorAuthService issues operator sessions for frontend login.
type OperatorAuthService interface {
	Login(ctx context.Context, email, password string,
		info identity.ClientInfo) (platformadmin.OperatorSessionResult, error)
	ValidateSession(ctx context.Context, token string) (platformadmin.Operator, error)
}

// PlatformAdminService implements ADM-01 and operator audit access.
type PlatformAdminService interface {
	ListOrganizations(ctx context.Context, actor identity.Actor, page platformadmin.Page) ([]tenancy.Organization, error) //nolint:lll
	CreateOrganization(ctx context.Context, actor identity.Actor, in platformadmin.CreateOrganizationInput,
		info identity.ClientInfo) (tenancy.Organization, error)
	SuspendOrganization(ctx context.Context, actor identity.Actor, orgID string, info identity.ClientInfo) error
	ReinstateOrganization(ctx context.Context, actor identity.Actor, orgID string, info identity.ClientInfo) error
	UpdateLimits(ctx context.Context, actor identity.Actor, orgID string, in platformadmin.LimitsInput,
		info identity.ClientInfo) (tenancy.Organization, error)
	ListOperatorAudit(ctx context.Context, actor identity.Actor, page platformadmin.Page) ([]audit.Entry, error)
}

// Server implements the generated OpenAPI server interface.
type Server struct {
	health           *health.Service
	auth             AuthService
	mfa              MFAService
	access           AccessService
	organizations    OrganizationService
	users            UserService
	operatorIdentity OperatorIdentityService
	operatorAuth     OperatorAuthService
	platformAdmin    PlatformAdminService
	providerSvc      ProviderService
	logger           *zap.Logger
	cfg              config.Config
}

// SetAccessService attaches the access service.
func (s *Server) SetAccessService(a AccessService) { s.access = a }

// SetOrganizationService attaches the organization service.
func (s *Server) SetOrganizationService(o OrganizationService) { s.organizations = o }

// SetUserService attaches the user service.
func (s *Server) SetUserService(u UserService) { s.users = u }

// SetMFAService attaches the MFA service.
func (s *Server) SetMFAService(m MFAService) { s.mfa = m }

// SetOperatorIdentityService attaches the operator identity provider.
func (s *Server) SetOperatorIdentityService(o OperatorIdentityService) { s.operatorIdentity = o }

// SetOperatorAuthService attaches the operator login service.
func (s *Server) SetOperatorAuthService(a OperatorAuthService) { s.operatorAuth = a }

// SetPlatformAdminService attaches operator organization management.
func (s *Server) SetPlatformAdminService(p PlatformAdminService) { s.platformAdmin = p }

// SetProviderService attaches the provider service.
func (s *Server) SetProviderService(p ProviderService) { s.providerSvc = p }

// NewServer creates an API server.
func NewServer(healthService *health.Service, auth AuthService, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{health: healthService, auth: auth, logger: logger, cfg: config.Config{CookieSecure: true}}
}

// SetConfig attaches runtime configuration.
func (s *Server) SetConfig(cfg config.Config) { s.cfg = cfg }
