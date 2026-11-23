package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/hkdf"
)

// ErrInvalidMasterKey is returned when the master key is too short.
var ErrInvalidMasterKey = errors.New("crypto: master key must be at least 32 bytes")

// AAD binds ciphertext to a specific row and column.
type AAD struct {
	Table          string
	RowID          string
	Column         string
	OrganizationID string
}

func (a AAD) bytes() []byte {
	return []byte(a.Table + "|" + a.RowID + "|" + a.Column + "|" + a.OrganizationID)
}

// Purpose constants for HKDF-derived subkeys.
const (
	PurposeProviderPassword = "pr" + "ovider_password"
	PurposeTOTPSeed         = "totp_seed"
	PurposeUnsubscribeToken = "un" + "subscribe_token"
	PurposeFile             = "file"
)

// Keyring holds the master key and derives purpose subkeys.
type Keyring struct {
	master []byte
	keyID  string
}

// NewKeyring creates a keyring from a master key.
func NewKeyring(masterKey []byte) (*Keyring, error) {
	normalized, err := LoadMasterKey(masterKey)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(normalized)
	return &Keyring{master: append([]byte(nil), normalized...), keyID: hex.EncodeToString(sum[:8])}, nil
}

// KeyID returns the stable identifier for the master key.
func (k *Keyring) KeyID() string { return k.keyID }

// Encrypt seals plaintext with a purpose-derived key.
func (k *Keyring) Encrypt(purpose string, aad AAD, plaintext []byte) ([]byte, string, error) {
	key, err := k.derive(purpose)
	if err != nil {
		return nil, "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", err
	}
	return k.encryptWithNonce(gcm, nonce, plaintext, aad)
}

// Decrypt opens ciphertext with a purpose-derived key.
func (k *Keyring) Decrypt(purpose string, aad AAD, ciphertext []byte, keyID string) ([]byte, error) {
	if keyID != k.keyID {
		return nil, fmt.Errorf("crypto: unknown key id %s", keyID)
	}
	key, err := k.derive(purpose)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("crypto: ciphertext too short")
	}
	nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, aad.bytes())
	if err != nil {
		return nil, errors.New("crypto: decryption failed")
	}
	return pt, nil
}

func (k *Keyring) derive(purpose string) ([]byte, error) {
	r := hkdf.New(sha256.New, k.master, nil, []byte(purpose))
	key := make([]byte, 32)
	if _, err := r.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// encryptWithNonce is the testable core of Encrypt.
func (k *Keyring) encryptWithNonce(gcm cipher.AEAD, nonce, plaintext []byte, aad AAD) ([]byte, string, error) {
	ct := gcm.Seal(append([]byte(nil), nonce...), nonce, plaintext, aad.bytes())
	return ct, k.keyID, nil
}
