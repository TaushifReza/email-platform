package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustDerive(t *testing.T, kr *Keyring, purpose string) []byte {
	t.Helper()
	key, err := kr.derive(purpose)
	require.NoError(t, err)
	return key
}

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

func TestKeyringKnownAnswer_SEC04(t *testing.T) {
	kr, err := NewKeyring([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	require.Equal(t, "3eb1bd439947eb76", kr.KeyID())
	aad := AAD{Table: "users", RowID: "u1", Column: "totp_seed", OrganizationID: "o1"}
	block, berr := aes.NewCipher(mustDerive(t, kr, PurposeTOTPSeed))
	require.NoError(t, berr)
	gcm, gerr := cipher.NewGCM(block)
	require.NoError(t, gerr)
	_ = gerr
	nonce := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	ct, keyID, err := kr.encryptWithNonce(gcm, nonce, []byte("secret"), aad)
	require.NoError(t, err)
	require.Equal(t, kr.KeyID(), keyID)
	require.Equal(t, "0102030405060708090a0b0c62fe0d1475ce1367c9f64ef8cae1768683e18f074f25", hex.EncodeToString(ct))
	// round trip via public Decrypt
	pt, err := kr.Decrypt(PurposeTOTPSeed, aad, ct, keyID)
	require.NoError(t, err)
	require.Equal(t, []byte("secret"), pt)
}

func TestKeyringTamperDetection_SEC04(t *testing.T) {
	kr, err := NewKeyring([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	aad := AAD{Table: "users", RowID: "u1", Column: "totp_seed", OrganizationID: "o1"}
	ct, keyID, err := kr.Encrypt(PurposeTOTPSeed, aad, []byte("secret"))
	require.NoError(t, err)
	ct[len(ct)-1] ^= 0xff
	_, err = kr.Decrypt(PurposeTOTPSeed, aad, ct, keyID)
	require.Error(t, err)
}

func TestKeyringWrongKeyID_SEC04(t *testing.T) {
	kr, err := NewKeyring([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	aad := AAD{Table: "users", RowID: "u1", Column: "totp_seed", OrganizationID: "o1"}
	ct, _, err := kr.Encrypt(PurposeTOTPSeed, aad, []byte("secret"))
	require.NoError(t, err)
	_, err = kr.Decrypt(PurposeTOTPSeed, aad, ct, "deadbeef")
	require.Error(t, err)
}

func TestKeyringPurposeSeparation_SEC04(t *testing.T) {
	kr, err := NewKeyring([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	aad := AAD{Table: "users", RowID: "u1", Column: "totp_seed", OrganizationID: "o1"}
	ct, keyID, err := kr.Encrypt(PurposeTOTPSeed, aad, []byte("secret"))
	require.NoError(t, err)
	_, err = kr.Decrypt(PurposeProviderPassword, aad, ct, keyID)
	require.Error(t, err)
}

func TestKeyringFreshNonce_SEC04(t *testing.T) {
	kr, err := NewKeyring([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	aad := AAD{Table: "users", RowID: "u1", Column: "totp_seed", OrganizationID: "o1"}
	ct1, _, err := kr.Encrypt(PurposeTOTPSeed, aad, []byte("secret"))
	require.NoError(t, err)
	ct2, _, err := kr.Encrypt(PurposeTOTPSeed, aad, []byte("secret"))
	require.NoError(t, err)
	require.NotEqual(t, ct1, ct2)
}
