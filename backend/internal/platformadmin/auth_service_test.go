package platformadmin

import (
	"context"
	"testing"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/stretchr/testify/require"
)

func TestOperatorLoginServiceLoginSuccess(t *testing.T) {
	hasher, err := identity.NewHasher(identity.DefaultArgon2Params)
	require.NoError(t, err)
	hash, err := hasher.Hash("secret")
	require.NoError(t, err)

	repo := &fakeRepository{operator: Operator{ID: "op-1", Email: "op@example.com"}, operatorPasswordHash: hash}
	ident := NewIdentityServiceWithHasher(repo, hasher)
	svc := NewOperatorLoginService(repo, ident, clock.System{})
	res, err := svc.Login(context.Background(), "op@example.com", "secret", identity.ClientInfo{})
	require.NoError(t, err)
	require.Equal(t, "op-1", res.OperatorID)
	require.NotEmpty(t, res.Token)
}
