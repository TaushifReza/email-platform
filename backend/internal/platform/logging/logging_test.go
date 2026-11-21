package logging

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	logger, err := New("info", os.Stderr)
	require.NoError(t, err)
	assert.NotNil(t, logger)

	_, err = New("invalid", os.Stderr)
	require.Error(t, err)
}

func TestRedactSensitiveFields(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New("info", &buf)
	require.NoError(t, err)

	logger.Info("test",
		zap.String("password", "secret123"),
		zap.String("secret_token", "abc123"),
		zap.String("normal_field", "normal_value"),
		zap.String("authorization", "Bearer token123"),
		zap.String("api_key", "key123"))

	logOutput := buf.String()
	assert.Contains(t, logOutput, `"password":"[REDACTED]"`)
	assert.Contains(t, logOutput, `"secret_token":"[REDACTED]"`)
	assert.Contains(t, logOutput, `"authorization":"[REDACTED]"`)
	assert.Contains(t, logOutput, `"api_key":"[REDACTED]"`)
	assert.Contains(t, logOutput, `"normal_field":"normal_value"`)
}

func TestRedactEmailInField(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New("info", &buf)
	require.NoError(t, err)

	logger.Info("test", zap.String("email", "user@example.com"))

	assert.Contains(t, buf.String(), `"email":"u***@example.com"`)
}

func TestRedactionWithVariants(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New("info", &buf)
	require.NoError(t, err)

	logger.Info("test",
		zap.String("secret_key", "secret123"),
		zap.String("passwd", "password123"),
		zap.String("auth_token", "token456"),
		zap.String("bearer_token", "bearer789"))

	logOutput := buf.String()
	assert.Contains(t, logOutput, `"secret_key":"[REDACTED]"`)
	assert.Contains(t, logOutput, `"passwd":"[REDACTED]"`)
	assert.Contains(t, logOutput, `"auth_token":"[REDACTED]"`)
	assert.Contains(t, logOutput, `"bearer_token":"[REDACTED]"`)
}

func TestRedactCore_Write(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New("info", &buf)
	require.NoError(t, err)

	logger.Info("Test message",
		zap.String("password", "secret123"),
		zap.String("email", "test@example.com"),
		zap.String("token", "Bearer abc123"),
		zap.String("normal_field", "normal_value"),
	)

	output := buf.String()
	assert.Contains(t, output, "[REDACTED]")
	assert.Contains(t, output, "t***@example.com")
	assert.NotContains(t, output, "secret123")
	assert.NotContains(t, output, "abc123")
	assert.Contains(t, output, "normal_value")
}

func TestPlantedSecretAndEmailNeverAppear(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New("info", &buf)
	require.NoError(t, err)
	logger = logger.With(zap.String("password", "planted-password-123"))
	logger.Info("contact planted@example.com with token=planted-token-456",
		zap.String("recipient", "planted@example.com"),
		zap.String("note", "secret=planted-secret-789"))

	output := buf.String()
	assert.NotContains(t, output, "planted-password-123")
	assert.NotContains(t, output, "planted-token-456")
	assert.NotContains(t, output, "planted-secret-789")
	assert.NotContains(t, output, "planted@example.com")
}
