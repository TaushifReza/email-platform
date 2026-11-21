package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultListenAddr    = ":8080"
	defaultLogLevel      = "info"
	defaultRedisURL      = "redis://localhost:6379"
	defaultStorageURL    = "http://localhost:9000"
	defaultStorageName   = "email-platform"
	defaultSMTPPort      = 587
	defaultSenderEmail   = "noreply@example.com"
	defaultDevMode       = false
	defaultOpenAPISpec   = "api/openapi.yaml"
	secretFileSuffix     = "_FILE"
	trimSecretWhitespace = true
)

var (
	ErrDatabaseURLRequired      = errors.New("DATABASE_URL is required")
	ErrStorageAccessKeyRequired = errors.New("STORAGE_ACCESS_KEY is required")
	ErrStorageSecretKeyRequired = errors.New("STORAGE_SECRET_KEY is required")
	ErrSMTPHostRequired         = errors.New("SMTP_HOST is required")
	ErrSMTPPortInvalid          = errors.New("SMTP_PORT must be a positive integer")
)

// Config holds the application configuration.
type Config struct {
	ListenAddr           string
	LogLevel             string
	DatabaseURL          string
	MigrationDatabaseURL string
	RedisURL             string
	StorageEndpoint      string
	StorageBucket        string
	StorageAccessKey     string
	StorageSecretKey     string
	MasterKey            []byte
	SMTPHost             string
	SMTPPort             int
	SMTPUsername         string
	SMTPPassword         string
	EmailFrom            string
	DevMode              bool
	OpenAPISpecPath      string
}

// Load reads configuration from environment variables and *_FILE secrets.
// It validates only the database connection because every process needs it.
// Storage and SMTP settings are validated by the services that use them.
func Load() (Config, error) {
	cfg := Config{
		ListenAddr:           getEnv("LISTEN_ADDR", defaultListenAddr),
		LogLevel:             getEnv("LOG_LEVEL", defaultLogLevel),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		MigrationDatabaseURL: getEnv("MIGRATION_DATABASE_URL", ""),
		RedisURL:             getEnv("REDIS_URL", defaultRedisURL),
		StorageEndpoint:      getEnv("STORAGE_ENDPOINT", defaultStorageURL),
		StorageBucket:        getEnv("STORAGE_BUCKET", defaultStorageName),
		StorageAccessKey:     getEnv("STORAGE_ACCESS_KEY", ""),
		StorageSecretKey:     getEnv("STORAGE_SECRET_KEY", ""),
		MasterKey:            []byte(getEnv("MASTER_KEY", "")),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             getEnvInt("SMTP_PORT", defaultSMTPPort),
		SMTPUsername:         getEnv("SMTP_USERNAME", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		EmailFrom:            getEnv("EMAIL_FROM", defaultSenderEmail),
		DevMode:              getEnvBool("DEV_MODE", defaultDevMode),
		OpenAPISpecPath:      getEnv("OPENAPI_SPEC_PATH", defaultOpenAPISpec),
	}

	secrets := []struct {
		fileKey string
		target  *string
	}{
		{"DATABASE_URL_FILE", &cfg.DatabaseURL},
		{"MIGRATION_DATABASE_URL_FILE", &cfg.MigrationDatabaseURL},
		{"STORAGE_ACCESS_KEY_FILE", &cfg.StorageAccessKey},
		{"STORAGE_SECRET_KEY_FILE", &cfg.StorageSecretKey},
		{"SMTP_PASSWORD_FILE", &cfg.SMTPPassword},
		{"MASTER_KEY_FILE", nil},
	}
	for _, secret := range secrets {
		value, err := readSecret(secret.fileKey)
		if err != nil {
			return cfg, err
		}
		if secret.target != nil {
			if value != "" {
				*secret.target = value
			}
			continue
		}
		if value != "" {
			cfg.MasterKey = []byte(value)
		}
	}

	if cfg.MigrationDatabaseURL == "" {
		cfg.MigrationDatabaseURL = cfg.DatabaseURL
	}

	if err := cfg.ValidateDatabase(); err != nil {
		return cfg, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the full configuration.
func (c Config) Validate() error {
	if err := c.ValidateDatabase(); err != nil {
		return err
	}
	if err := c.ValidateStorage(); err != nil {
		return err
	}
	return c.ValidateSMTP()
}

// ValidateDatabase validates the database connection settings.
func (c Config) ValidateDatabase() error {
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return ErrDatabaseURLRequired
	}
	return nil
}

// ValidateStorage validates object storage settings.
func (c Config) ValidateStorage() error {
	if strings.TrimSpace(c.StorageAccessKey) == "" {
		return ErrStorageAccessKeyRequired
	}
	if strings.TrimSpace(c.StorageSecretKey) == "" {
		return ErrStorageSecretKeyRequired
	}
	return nil
}

// ValidateSMTP validates outgoing SMTP settings. SMTP username is optional
// because local Mailpit and some providers accept unauthenticated SMTP.
func (c Config) ValidateSMTP() error {
	if strings.TrimSpace(c.SMTPHost) == "" {
		return ErrSMTPHostRequired
	}
	if c.SMTPPort <= 0 {
		return ErrSMTPPortInvalid
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return defaultValue
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func getEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func readSecret(fileKey string) (string, error) {
	path := getEnv(fileKey, "")
	if path == "" {
		return "", nil
	}
	// #nosec G304 -- Secret file paths come from trusted operator configuration.
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fileKey, err)
	}
	if trimSecretWhitespace {
		return strings.TrimSpace(string(data)), nil
	}
	return string(data), nil
}

// SecretFileSuffix is the suffix used for file-backed secrets.
func SecretFileSuffix() string {
	return secretFileSuffix
}
