package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	// Test creating a logger with valid level
	logger, err := New("info", os.Stderr)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	
	// Test creating a logger with invalid level
	_, err = New("invalid-level", os.Stderr)
	assert.Error(t, err)
}

func TestRedactCore_Write(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	
	// Create a zap logger that uses our redaction core
	logger, err := New("info", &buf)
	assert.NoError(t, err)
	
	// Test logging with sensitive information
	logger.Info("Test message", 
		zap.String("password", "secret123"),
		zap.String("email", "test@example.com"),
		zap.String("token", "Bearer abc123"),
		zap.String("normal_field", "normal_value"),
	)
	
	// Get the output
	output := buf.String()
	
	// Verify that sensitive information was redacted
	assert.Contains(t, output, "[REDACTED]")
	assert.Contains(t, output, "u***@example.com")
	assert.NotContains(t, output, "secret123")
	assert.NotContains(t, output, "abc123")
	assert.Contains(t, output, "normal_value")
}

func TestRedact(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "password in field",
			in:   "password=test123",
			want: "[REDACTED]",
		},
		{
			name: "secret in field",
			in:   "secret=secret123",
			want: "[REDACTED]",
		},
		{
			name: "email address redacted",
			in:   "user@example.com",
			want: "u***@example.com",
		},
		{
			name: "bearer token redacted",
			in:   "Bearer abc123def456",
			want: "[REDACTED]",
		},
		{
			name: "normal text unchanged",
			in:   "This is a normal message",
			want: "This is a normal message",
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
			name: "simple email",
			in:   "user@example.com",
			want: "u***@example.com",
		},
		{
			name: "complex email with subdomain",
			in:   "user.name@test.co.uk",
			want: "u***@co.uk",
		},
		{
			name: "empty string",
			in:   "",
			want: "[REDACTED]",
		},
		{
			name: "invalid email format",
			in:   "notanemail",
			want: "[REDACTED]",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, maskEmail(tt.in))
		})
	}
}