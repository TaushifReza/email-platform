package identity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAUTH_01_HashRoundTrip(t *testing.T) {
	h, err := NewHasher(Params{Memory: 1024, Time: 1, Threads: 1, SaltLength: 8, KeyLength: 16})
	require.NoError(t, err)
	hash, err := h.Hash("correct horse battery staple")
	require.NoError(t, err)
	ok, err := h.Verify(hash, "correct horse battery staple")
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = h.Verify(hash, "wrong")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestAUTH_01_HashParams(t *testing.T) {
	h, err := NewHasher(Params{Memory: 2048, Time: 2, Threads: 2, SaltLength: 8, KeyLength: 16})
	require.NoError(t, err)
	hash, err := h.Hash("password")
	require.NoError(t, err)
	assert.Contains(t, hash, "m=2048,t=2,p=2")
	ok, err := h.Verify(hash, "password")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestAUTH_01_VerifyDummy(t *testing.T) {
	h, err := NewHasher(Params{Memory: 1024, Time: 1, Threads: 1, SaltLength: 8, KeyLength: 16})
	require.NoError(t, err)
	assert.NoError(t, h.VerifyDummy("anything"))
}

func TestAUTH_04_PasswordPolicyMinLength(t *testing.T) {
	require.ErrorIs(t, ValidatePassword(strings.Repeat("a", 11)), ErrPasswordTooShort)
	assert.NoError(t, ValidatePassword(strings.Repeat("a", 12)))
}

func TestAUTH_04_PasswordPolicyCommon(t *testing.T) {
	require.ErrorIs(t, ValidatePassword("passwordpassword"), ErrPasswordCommon)
}

func TestAUTH_04_PasswordPolicyMaxLength(t *testing.T) {
	require.ErrorIs(t, ValidatePassword(strings.Repeat("a", 1025)), ErrPasswordTooLong)
}
