package logging

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	bearerTokenPattern      = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/\-=]+`)
	authorizationPattern    = regexp.MustCompile(`(?i)\bauthorization\s*[:=]\s*.*`)
	secretAssignmentPattern = regexp.MustCompile(
		`(?i)\b(password|passwd|secret|token|api[_-]?key|seed)\s*[:=]\s*[^\s,;]+`,
	)
	emailPattern = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
)

// Redact masks secrets and recipient email addresses in a free-form string.
func Redact(in string) string {
	if in == "" {
		return in
	}

	out := bearerTokenPattern.ReplaceAllString(in, "[REDACTED]")
	out = authorizationPattern.ReplaceAllString(out, "[REDACTED]")
	out = secretAssignmentPattern.ReplaceAllString(out, "[REDACTED]")
	out = emailPattern.ReplaceAllStringFunc(out, maskEmail)

	return out
}

// maskEmail keeps the first character of the local part and the full domain.
func maskEmail(in string) string {
	email := strings.TrimSpace(in)
	if email == "" {
		return ""
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return "[REDACTED]"
	}

	first, _ := utf8.DecodeRuneInString(email[:at])
	return string(first) + "***@" + email[at+1:]
}
