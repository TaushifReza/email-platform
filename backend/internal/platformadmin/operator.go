package platformadmin

import (
	"context"
	"errors"
	"strings"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
)

// ErrOperatorNotFound covers unknown and suspended operators alike. It must
// not reveal which emails are registered.
var ErrOperatorNotFound = errors.New("platformadmin: operator not found")

// ErrInvalidCredentials is returned for any failed operator login without
// revealing whether the email is registered.
var ErrInvalidCredentials = errors.New("platformadmin: invalid credentials")

// IdentityService authenticates operators. It supports both the legacy
// Cloudflare-Access header assertion and frontend password login.
type IdentityService struct {
	repo   Repository
	hasher *identity.Hasher
}

func NewIdentityService(repo Repository) *IdentityService {
	h, err := identity.NewHasher(identity.DefaultArgon2Params)
	if err != nil {
		return &IdentityService{repo: repo}
	}
	return &IdentityService{repo: repo, hasher: h}
}

func NewIdentityServiceWithHasher(repo Repository, hasher *identity.Hasher) *IdentityService {
	return &IdentityService{repo: repo, hasher: hasher}
}

// AuthenticateWithPassword verifies an operator's password for frontend login.
func (s *IdentityService) AuthenticateWithPassword(ctx context.Context, email, password string) (Operator, error) {
	normalized := NormalizeEmail(email)
	if !validEmail(normalized) || password == "" {
		return Operator{}, ErrInvalidCredentials
	}
	creds, err := s.repo.LookupOperatorCredentials(ctx, normalized)
	if err != nil {
		return Operator{}, ErrInvalidCredentials
	}
	if creds.PasswordHash == "" {
		return Operator{}, ErrInvalidCredentials
	}
	ok, err := s.hasher.Verify(creds.PasswordHash, password)
	if err != nil || !ok {
		return Operator{}, ErrInvalidCredentials
	}
	return creds.Operator, nil
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
