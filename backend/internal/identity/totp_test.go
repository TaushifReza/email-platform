package identity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTOTPRFC6238Vector(t *testing.T) {
	// RFC 6238 SHA-1 test vector for 59 seconds.
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ" //nolint:gosec
	code, err := TOTPCode(secret, time.Unix(59, 0).UTC())
	require.NoError(t, err)
	require.Equal(t, "287082", code)
}

func TestTOTPReplayStep(t *testing.T) {
	secret, err := TOTPSecret()
	require.NoError(t, err)
	now := time.Now().UTC()
	step, ok := ValidateTOTP(secret, mustCode(t, secret, now), now)
	require.True(t, ok)
	require.Positive(t, step)
}

func mustCode(t *testing.T, secret string, now time.Time) string {
	t.Helper()
	c, err := TOTPCode(secret, now)
	require.NoError(t, err)
	return c
}
