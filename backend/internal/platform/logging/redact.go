package logging

import (
	"strings"
)

// Redact takes a string and returns a redacted version
func Redact(in string) string {
	if in == "" {
		return in
	}

	// Key-based redaction for common sensitive keys
	// Password-like fields
	if strings.Contains(strings.ToLower(in), "password") ||
		strings.Contains(strings.ToLower(in), "passwd") ||
		strings.Contains(strings.ToLower(in), "secret") ||
		strings.Contains(strings.ToLower(in), "token") ||
		strings.Contains(strings.ToLower(in), "seed") ||
		strings.Contains(strings.ToLower(in), "authorization") ||
		strings.Contains(strings.ToLower(in), "api_key") ||
		strings.Contains(strings.ToLower(in), "auth") {
		return "[REDACTED]"
	}

	// For simplicity, we redact sensitive fields at the field level in core.Write()
	return in
}
