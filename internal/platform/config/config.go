package config

import (
	"os"
	"strconv"
)

const defaultSMTPPort = 587

var (
	errDatabaseURLRequired      = &configError{"DATABASE_URL is required"}
	errStorageAccessKeyRequired = &configError{"STORAGE_ACCESS_KEY is required"}
	errStorageSecretKeyRequired = &configError{"STORAGE_SECRET_KEY is required"}
	errSMTPHostRequired         = &configError{"SMTP_HOST is required"}
	errSMTPUsernameRequired     = &configError{"SMTP_USERNAME is required"}
)

type configError struct {
	msg string
}

func (e *configError) Error() string {
	return e.msg
}

// Config holds the application configuration.
type Config struct {
	ListenAddr       string
	DatabaseURL      string
	StorageAccount   string
	StorageAccessKey string
	StorageSecretKey string
	SMTPHost         string
	SMTPUsername     string
	SMTPPassword     string
	SMTPPort         int
}

// Load loads configuration from environment variables and files.
func Load() (*Config, error) {
	cfg := Config{
		ListenAddr:       getEnv("LISTEN_ADDR", ":8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		StorageAccount:   getEnv("STORAGE_ACCOUNT", ""),
		StorageAccessKey: getEnv("STORAGE_ACCESS_KEY", ""),
		StorageSecretKey: getEnv("STORAGE_SECRET_KEY", ""),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPUsername:     getEnv("SMTP_USERNAME", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		SMTPPort:         getEnvInt("SMTP_PORT", defaultSMTPPort),
	}

	return &cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return errDatabaseURLRequired
	}

	if c.StorageAccessKey == "" {
		return errStorageAccessKeyRequired
	}

	if c.StorageSecretKey == "" {
		return errStorageSecretKeyRequired
	}

	if c.SMTPHost == "" {
		return errSMTPHostRequired
	}

	if c.SMTPUsername == "" {
		return errSMTPUsernameRequired
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}

	return defaultValue
}
