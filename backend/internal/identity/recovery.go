package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"strings"
)

// GenerateRecoveryCodes returns n plaintext codes.
func GenerateRecoveryCodes(n int) ([]string, error) {
	out := make([]string, 0, n)
	for range n {
		b := make([]byte, 10)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		out = append(out, strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)))
	}
	return out, nil
}

// HashRecoveryCode returns the SHA-256 hex hash of a normalized code.
func HashRecoveryCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(code, " ", "")))
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// NormalizeRecoveryCode normalizes user input.
func NormalizeRecoveryCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(code, " ", "")))
}
