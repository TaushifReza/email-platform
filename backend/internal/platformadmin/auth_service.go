package platformadmin

import (
	"context"
	"crypto/sha256"
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

func (s *OperatorLoginService) ValidateSession(ctx context.Context, token string) (Operator, error) {
	if token == "" {
		return Operator{}, ErrOperatorNotFound
	}
	h := sha256.Sum256([]byte(token))
	rec, err := s.repo.LookupOperatorSession(ctx, h[:])
	if err != nil {
		return Operator{}, ErrOperatorNotFound
	}
	now := s.clock.Now()
	if rec.RevokedAt != nil {
		return Operator{}, ErrOperatorNotFound
	}
	if !rec.IdleExpiresAt.After(now) || !rec.AbsoluteExpiresAt.After(now) {
		return Operator{}, ErrOperatorNotFound
	}
	return Operator{ID: rec.OperatorID, Email: rec.OperatorEmail}, nil
}
