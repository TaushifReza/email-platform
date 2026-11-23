package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeySetRotation_SEC04(t *testing.T) {
	current, err := NewKeyring([]byte("current-master-key-0123456789abcdef"))
	require.NoError(t, err)
	var previous *Keyring
	previous, err = NewKeyring([]byte("previous-master-key-abcdef0123456789"))
	require.NoError(t, err)
	require.NoError(t, err)
	set, err := NewKeySet(current, previous)
	require.NoError(t, err)

	aad := AAD{Table: "users", RowID: "u1", Column: "totp_seed", OrganizationID: "o1"}
	ct, keyID, err := previous.Encrypt(PurposeTOTPSeed, aad, []byte("secret"))
	require.NoError(t, err)
	require.Equal(t, previous.KeyID(), keyID)

	pt, err := set.Decrypt(PurposeTOTPSeed, aad, ct, keyID)
	require.NoError(t, err)
	require.Equal(t, []byte("secret"), pt)

	newCT, newKeyID, err := set.Encrypt(PurposeTOTPSeed, aad, []byte("secret2"))
	require.NoError(t, err)
	require.Equal(t, current.KeyID(), newKeyID)
	pt2, err := set.Decrypt(PurposeTOTPSeed, aad, newCT, newKeyID)
	require.NoError(t, err)
	require.Equal(t, []byte("secret2"), pt2)

	_, err = set.Decrypt(PurposeTOTPSeed, aad, newCT, "deadbeef")
	require.Error(t, err)
}

func TestKeySetRejectsDuplicateKeyIDs_SEC04(t *testing.T) {
	k, err := NewKeyring([]byte("same-master-key-0123456789abcdef"))
	require.NoError(t, err)
	_, err = NewKeySet(k, k)
	require.Error(t, err)
}
