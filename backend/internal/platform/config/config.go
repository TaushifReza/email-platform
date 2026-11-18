package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	ListenAddr     string
	LogLevel       string
	DatabaseDSN    string
	RedisURL       string
	StorageEndpoint     string
	StorageBucket       string
	StorageAccessKey    string
	StorageSecretKey    string
	MasterKey      []byte
	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	EmailFrom      string
}

// Load loads configuration from environment variables and files
func Load() (Config, error) {
	cfg := Config{
		ListenAddr:  getEnv("LISTEN_ADDR", ":8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DatabaseDSN: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		StorageEndpoint:  getEnv("STORAGE_ENDPOINT", "https://s3.amazonaws.com"),
		StorageBucket:    getEnv("STORAGE_BUCKET", "email-platform"),
		MasterKey:      []byte(getEnv("MASTER_KEY", "")),
		SMTPHost:       getEnv("SMTP_HOST", ""),
		SMTPPort:       getEnvInt("SMTP_PORT", 587),
		SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		EmailFrom:      getEnv("EMAIL_FROM", "noreply@example.com"),
	}

	// Handle file-based secrets
	cfg.DatabaseDSN = getFileOrEnv("DATABASE_URL_FILE", cfg.DatabaseDSN)
	cfg.StorageAccessKey = getFileOrEnv("STORAGE_ACCESS_KEY_FILE", cfg.StorageAccessKey)
	cfg.StorageSecretKey = getFileOrEnv("STORAGE_SECRET_KEY_FILE", cfg.StorageSecretKey)
	cfg.SMTPPassword = getFileOrEnv("SMTP_PASSWORD_FILE", cfg.SMTPPassword)
	cfg.MasterKey = []byte(getFileOrEnv("MASTER_KEY_FILE", string(cfg.MasterKey)))

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c Config) Validate() error {
	if c.DatabaseDSN == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.StorageAccessKey == "" {
		return fmt.Errorf("STORAGE_ACCESS_KEY is required")
	}
	if c.StorageSecretKey == "" {
		return fmt.Errorf("STORAGE_SECRET_KEY is required")
	}
	if c.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}
	if c.SMTPUsername == "" {
		return fmt.Errorf("SMTP_USERNAME is required")
	}
	return nil
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt returns the integer value of an environment variable or a default value
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// getFileOrEnv returns the content of a file or environment variable if file is not found
func getFileOrEnv(fileKey, envValue string) string {
	// If we have a file path, try to read from it
	if filePath := getEnv(fileKey, ""); filePath != "" {
		if data, err := os.ReadFile(filePath); err == nil {
			return strings.TrimSpace(string(data))
		}
		// If file reading fails, fall back to the environment variable value
	}
	return envValue
}