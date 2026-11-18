package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedact(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "password field value",
			in:   "password=test123",
			want: "[REDACTED]",
		},
		{
			name: "secret field value",
			in:   "secret=test123",
			want: "[REDACTED]",
		},
		{
			name: "email address",
			in:   "user@example.com",
			want: "u***@example.com",
		},
		{
			name: "email in text",
			in:   "Contact user@example.com for help",
			want: "Contact u***@example.com for help",
		},
		{
			name: "bearer token",
			in:   "Bearer abc123def456",
			want: "[REDACTED]",
		},
		{
			name: "bearer token in text",
			in:   "Authorization: Bearer abc123def456",
			want: "[REDACTED]",
		},
		{
			name: "normal text",
			in:   "This is a normal text string",
			want: "This is a normal text string",
		},
		{
			name: "multiple emails",
			in:   "Contact john@company.com or jane@example.com",
			want: "Contact j***@company.com or j***@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Redact(tt.in))
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "valid email",
			in:   "user@example.com",
			want: "u***@example.com",
		},
		{
			name: "email with multiple dots",
			in:   "user.name@example.co.uk",
			want: "u***@co.uk",
		},
		{
			name: "empty string",
			in:   "",
			want: "[REDACTED]",
		},
		{
			name: "invalid email format (no @)",
			in:   "user.example.com",
			want: "[REDACTED]",
		},
		{
			name: "invalid email format (no domain)",
			in:   "user@",
			want: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, maskEmail(tt.in))
		})
	}
}