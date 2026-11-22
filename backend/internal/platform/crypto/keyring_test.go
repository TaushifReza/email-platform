package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyringRoundTripAndAADBound(t *testing.T) {
	kr, err := NewKeyring(make([]byte, 32))
	require.NoError(t, err)
	aad := AAD{Table: "users", RowID: "u1", Column: "totp_seed", OrganizationID: "o1"}
	ct, keyID, err := kr.Encrypt("totp_seed", aad, []byte("secret"))
	require.NoError(t, err)
	pt, err := kr.Decrypt("totp_seed", aad, ct, keyID)
	require.NoError(t, err)
	require.Equal(t, []byte("secret"), pt)
	// wrong AAD fails
	_, err = kr.Decrypt("totp_seed", AAD{Table: "users", RowID: "u2", Column: "totp_seed", OrganizationID: "o1"}, ct, keyID) //nolint:lll
	require.Error(t, err)
}

func TestKeyringRejectsShortMasterKey(t *testing.T) {
	_, err := NewKeyring([]byte("short"))
	require.Error(t, err)
}
