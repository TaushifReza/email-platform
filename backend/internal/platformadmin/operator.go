package platformadmin

import (
	"context"
	"errors"
	"strings"
)

// ErrOperatorNotFound covers unknown and suspended operators alike. It must
// not reveal which emails are registered.
var ErrOperatorNotFound = errors.New("platformadmin: operator not found")

// IdentityService authenticates the identity asserted by Cloudflare Access.
// The header is trusted only after the transport layer has verified the peer;
// this service validates the allowlist and returns the operator's identity.
type IdentityService struct {
	repo Repository
}

func NewIdentityService(repo Repository) *IdentityService {
	return &IdentityService{repo: repo}
}

// Authenticate resolves a normalized IdP email to a registered operator.
func (s *IdentityService) Authenticate(ctx context.Context, email string) (Operator, error) {
	normalized := NormalizeEmail(email)
	if !validEmail(normalized) {
		return Operator{}, ErrOperatorNotFound
	}
	op, err := s.repo.LookupOperatorByEmail(ctx, normalized)
	if err != nil {
		return Operator{}, err
	}
	if op.ID == "" {
		return Operator{}, ErrOperatorNotFound
	}
	return op, nil
}

// NormalizeEmail matches the database's generated normalized column.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	if email == "" || len(email) > 254 {
		return false
	}
	at := strings.IndexByte(email, '@')
	if at < 1 || at == len(email)-1 || strings.Contains(email[at+1:], "@") {
		return false
	}
	return !strings.ContainsAny(email, " \t\r\n")
}
