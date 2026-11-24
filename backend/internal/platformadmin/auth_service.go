package platformadmin

import (
	"context"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
)

// OperatorLoginService authenticates operators and issues sessions.
type OperatorLoginService struct {
	repo  Repository
	ident *IdentityService
	clock clock.Clock
}

func NewOperatorLoginService(repo Repository, ident *IdentityService, clk clock.Clock) *OperatorLoginService {
	if clk == nil {
		clk = clock.System{}
	}
	return &OperatorLoginService{repo: repo, ident: ident, clock: clk}
}

func (s *OperatorLoginService) Login(ctx context.Context, email, password string,
	info identity.ClientInfo,
) (OperatorSessionResult, error) {
	op, err := s.ident.AuthenticateWithPassword(ctx, email, password)
	if err != nil {
		return OperatorSessionResult{}, err
	}
	return IssueOperatorSession(ctx, s.repo, s.clock, op, info.IP, info.UserAgent, 12*time.Hour, 24*time.Hour)
}
