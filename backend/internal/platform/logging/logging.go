package logging

import (
	"io"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger with redaction
func New(level string, writer io.Writer) (*zap.Logger, error) {
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		logLevel,
	)

	return zap.New(&redactCore{core: core}), nil
}

// redactCore wraps a zapcore.Core and redacts sensitive values in field content
type redactCore struct {
	core zapcore.Core
}

func (c *redactCore) Enabled(level zapcore.Level) bool {
	return c.core.Enabled(level)
}

func (c *redactCore) With(fields []zapcore.Field) zapcore.Core {
	// Create a new core with the same underlying core and updated fields
	return &redactCore{core: c.core.With(fields)}
}

func (c *redactCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return c.core.Check(entry, checked)
}

func (c *redactCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Redact sensitive values in fields
	var redactedFields []zapcore.Field

	for _, field := range fields {
		if strings.Contains(strings.ToLower(field.Key), "password") ||
			strings.Contains(strings.ToLower(field.Key), "passwd") ||
			strings.Contains(strings.ToLower(field.Key), "secret") ||
			strings.Contains(strings.ToLower(field.Key), "token") ||
			strings.Contains(strings.ToLower(field.Key), "seed") ||
			strings.Contains(strings.ToLower(field.Key), "authorization") ||
			strings.Contains(strings.ToLower(field.Key), "api_key") {
			// Redact the value of sensitive fields
			redactedFields = append(redactedFields, zap.String(field.Key, "[REDACTED]"))
		} else {
			redactedFields = append(redactedFields, field)
		}
	}

	return c.core.Write(entry, redactedFields)
}

func (c *redactCore) Sync() error {
	return c.core.Sync()
}