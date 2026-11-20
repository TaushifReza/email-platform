package logging

import (
	"io"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger with secret and email redaction.
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

// redactCore wraps a zapcore.Core and redacts sensitive values.
type redactCore struct {
	core zapcore.Core
}

func (c *redactCore) Enabled(level zapcore.Level) bool {
	return c.core.Enabled(level)
}

func (c *redactCore) With(fields []zapcore.Field) zapcore.Core {
	return &redactCore{core: c.core.With(redactFields(fields))}
}

func (c *redactCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return checked
	}
	return checked.AddCore(entry, c)
}

func (c *redactCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	entry.Message = Redact(entry.Message)
	return c.core.Write(entry, redactFields(fields))
}

func (c *redactCore) Sync() error {
	return c.core.Sync()
}

func redactFields(fields []zapcore.Field) []zapcore.Field {
	redacted := make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		redacted = append(redacted, redactField(field))
	}
	return redacted
}

func redactField(field zapcore.Field) zapcore.Field {
	if isSensitiveKey(field.Key) {
		return zap.String(field.Key, "[REDACTED]")
	}

	//nolint:exhaustive // Only string-like fields can contain secrets.
	switch field.Type {
	case zapcore.StringType:
		return zap.String(field.Key, Redact(field.String))
	case zapcore.ByteStringType:
		return zap.ByteString(field.Key, []byte(Redact(field.String)))
	default:
		return field
	}
}

func isSensitiveKey(key string) bool {
	normalized := strings.NewReplacer("_", "", "-", "", ".", "", " ", "").Replace(strings.ToLower(key))
	sensitiveParts := []string{
		"password", "passwd", "secret", "token", "apikey", "authorization", "seed", "bearer",
	}
	for _, part := range sensitiveParts {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return normalized == "auth"
}
