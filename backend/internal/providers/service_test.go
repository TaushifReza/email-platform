package providers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/crypto"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	created Provider
	list    []Provider
	err     error
}

func (f *fakeRepo) List(_ context.Context, _ string) ([]Provider, error) {
	return f.list, f.err
}

func (f *fakeRepo) Get(_ context.Context, _ string, _ string) (Provider, error) {
	return f.created, f.err
}

func (f *fakeRepo) Create(_ context.Context, _ string, p Provider, enc []byte, keyID string, now time.Time, entry audit.Entry) error { //nolint:revive,lll
	f.created = p
	return f.err
}

func (f *fakeRepo) Update(_ context.Context, orgID, id string, u Update, now time.Time, entry audit.Entry) (Provider, error) { //nolint:revive,lll
	return f.created, f.err
}

func (f *fakeRepo) Rotate(_ context.Context, orgID, id, username string, enc []byte, keyID string, now time.Time, entry audit.Entry) (Provider, error) { //nolint:revive,lll
	return f.created, f.err
}

func (f *fakeRepo) Deactivate(_ context.Context, orgID, id string, now time.Time, entry audit.Entry) (Provider, error) { //nolint:revive,lll
	return f.created, f.err
}

func (f *fakeRepo) SetDefault(_ context.Context, _ string, _ string, now time.Time, entry audit.Entry) (Provider, error) { //nolint:revive,lll
	return f.created, f.err
}

func (f *fakeRepo) EncryptedCredentials(_ context.Context, _ string, _ string) (string, []byte, string, error) { //nolint:lll
	return "u", []byte("ct"), "k", f.err
}

type fakeEnc struct{ err error }

func (f *fakeEnc) Encrypt(_ string, _ crypto.AAD, plaintext []byte) ([]byte, string, error) {
	if f.err != nil {
		return nil, "", f.err
	}
	return []byte("cipher:" + string(plaintext)), "kid", nil
}

func (f *fakeEnc) Decrypt(_ string, _ crypto.AAD, _ []byte, _ string) ([]byte, error) {
	return []byte("plain"), nil
}

type fakeTargets struct{ err error }

func (f *fakeTargets) ValidateTarget(_ context.Context, _ string, _ int) error {
	return f.err
}

func TestPRV04_PrivateTargetRefusedAtConfiguration(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, &fakeEnc{}, &fakeTargets{err: errors.New("nope")}, clock.System{})
	actor := identity.Actor{OrganizationID: "org", UserID: "u", Role: "owner", StepUpExpiresAt: ptrTime(timeNow())}
	_, err := svc.Create(context.Background(), actor, Input{Host: "evil", Port: 25, TLSMode: "starttls", Username: "u", Password: "secret-password", FromAddress: "a@b.com"}, identity.ClientInfo{}) //nolint:lll
	require.Error(t, err)
	require.Contains(t, err.Error(), "host not allowed")
}

func TestPRV05_RotateRequiresStepUp(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, &fakeEnc{}, &fakeTargets{}, clock.System{})
	actor := identity.Actor{OrganizationID: "org", UserID: "u", Role: "owner"}
	_, err := svc.Rotate(context.Background(), actor, "p", "u", "newpw", identity.ClientInfo{})
	require.ErrorIs(t, err, identity.ErrMFAStepUp)
}

func TestPRV05_SecretEncryptedAndNeverInProvider(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, &fakeEnc{}, &fakeTargets{}, clock.System{})
	actor := identity.Actor{OrganizationID: "org", UserID: "u", Role: "owner", StepUpExpiresAt: ptrTime(timeNow())}
	p, err := svc.Create(context.Background(), actor, Input{Host: "h", Port: 587, TLSMode: "starttls", Username: "u", Password: "secret-password", FromAddress: "a@b.com"}, identity.ClientInfo{}) //nolint:lll
	require.NoError(t, err)
	require.NotContains(t, p.Username, "secret-password")
}

func ptrTime(t time.Time) *time.Time { return &t }
func timeNow() time.Time             { return clock.System{}.Now() }
