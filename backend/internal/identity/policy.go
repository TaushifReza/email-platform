package identity

import (
	_ "embed"
	"errors"
	"strings"
	"unicode/utf8"
)

// ErrPasswordTooShort is returned for short passwords.
var ErrPasswordTooShort = errors.New("identity: password too short")

// ErrPasswordTooLong is returned for long passwords.
var ErrPasswordTooLong = errors.New("identity: password too long")

// ErrPasswordWhitespace is returned for whitespace-only passwords.
var ErrPasswordWhitespace = errors.New("identity: password must contain non-whitespace")

// ErrPasswordCommon is returned for common passwords.
var ErrPasswordCommon = errors.New("identity: password is too common")

//go:embed common_passwords.txt
var commonPasswordsRaw string

//nolint:gochecknoglobals
var commonPasswords = buildCommonPasswordSet(commonPasswordsRaw)

func buildCommonPasswordSet(raw string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, line := range strings.Split(raw, "\n") {
		value := strings.ToLower(strings.TrimSpace(line))
		if value == "" || strings.HasPrefix(value, "#") {
			continue
		}
		set[value] = struct{}{}
	}
	return set
}

// ValidatePassword checks password policy.
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < 12 {
		return ErrPasswordTooShort
	}
	if len(password) > 1024 {
		return ErrPasswordTooLong
	}
	if strings.TrimSpace(password) == "" {
		return ErrPasswordWhitespace
	}
	if _, ok := commonPasswords[strings.ToLower(strings.TrimSpace(password))]; ok {
		return ErrPasswordCommon
	}
	return nil
}
