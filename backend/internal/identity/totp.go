package identity

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" // #nosec G505 -- TOTP RFC 6238 uses SHA-1
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	totpDigits = 6
	totpPeriod = 30 * time.Second
	totpSkew   = 1
)

// TOTPSecret generates a 20-byte secret encoded as base32 without padding.
func TOTPSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// TOTPCode computes the code for the given secret and time.
func TOTPCode(secret string, t time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", err
	}
	counter := uint64(t.UTC().Unix() / int64(totpPeriod.Seconds())) // #nosec G115
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	h := mac.Sum(nil)
	off := h[len(h)-1] & 0x0f
	code := (uint32(h[off])&0x7f)<<24 |
		(uint32(h[off+1])&0xff)<<16 |
		(uint32(h[off+2])&0xff)<<8 |
		(uint32(h[off+3]) & 0xff)
	code %= 1000000
	return fmt.Sprintf("%06d", code), nil
}

// TOTPStep returns the time step for t.
func TOTPStep(t time.Time) int64 {
	return t.UTC().Unix() / int64(totpPeriod.Seconds())
}

// ValidateTOTP checks code against secret within ±1 step and returns the accepted step.
func ValidateTOTP(secret, code string, now time.Time) (int64, bool) {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return 0, false
	}
	for i := -totpSkew; i <= totpSkew; i++ {
		t := now.Add(time.Duration(i) * totpPeriod)
		step := TOTPStep(t)
		got, err := TOTPCode(secret, t)
		if err != nil {
			continue
		}
		if hmac.Equal([]byte(got), []byte(code)) {
			return step, true
		}
	}
	return 0, false
}

// TOTPOtpauthURI builds an otpauth URI.
func TOTPOtpauthURI(issuer, account, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		issuer, account, secret, issuer, totpDigits, int(totpPeriod.Seconds()))
}
