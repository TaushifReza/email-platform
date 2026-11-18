package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Save original environment
	origDatabaseURL := os.Getenv("DATABASE_URL")
	origS3AccessKey := os.Getenv("S3_ACCESS_KEY")
	origS3SecretKey := os.Getenv("S3_SECRET_KEY")
	origSMTPHost := os.Getenv("SMTP_HOST")
	origSMTPUsername := os.Getenv("SMTP_USERNAME")

	// Set up test environment
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
	os.Setenv("S3_ACCESS_KEY", "test-access-key")
	os.Setenv("S3_SECRET_KEY", "test-secret-key")
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("SMTP_USERNAME", "test@example.com")

	defer func() {
		// Restore original environment
		os.Setenv("DATABASE_URL", origDatabaseURL)
		os.Setenv("S3_ACCESS_KEY", origS3AccessKey)
		os.Setenv("S3_SECRET_KEY", origS3SecretKey)
		os.Setenv("SMTP_HOST", origSMTPHost)
		os.Setenv("SMTP_USERNAME", origSMTPUsername)
	}()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", cfg.DatabaseDSN)
	assert.Equal(t, "test-access-key", cfg.S3AccessKey)
	assert.Equal(t, "test-secret-key", cfg.S3SecretKey)
	assert.Equal(t, "smtp.example.com", cfg.SMTPHost)
	assert.Equal(t, "test@example.com", cfg.SMTPUsername)
}

func TestLoad_WithFileSecrets(t *testing.T) {
	// Save original environment
	origDatabaseURLFile := os.Getenv("DATABASE_URL_FILE")
	origS3AccessKeyFile := os.Getenv("S3_ACCESS_KEY_FILE")
	origS3SecretKeyFile := os.Getenv("S3_SECRET_KEY_FILE")

	// Create temporary files with secrets
	tmpDir := t.TempDir()
	databaseURLFile := filepath.Join(tmpDir, "database_url")
	s3AccessKeyFile := filepath.Join(tmpDir, "s3_access_key")
	s3SecretKeyFile := filepath.Join(tmpDir, "s3_secret_key")

	os.WriteFile(databaseURLFile, []byte("postgres://file:test@localhost:5432/testdb"), 0644)
	os.WriteFile(s3AccessKeyFile, []byte("file-access-key"), 0644)
	os.WriteFile(s3SecretKeyFile, []byte("file-secret-key"), 0644)

	// Set up test environment
	os.Setenv("DATABASE_URL_FILE", databaseURLFile)
	os.Setenv("S3_ACCESS_KEY_FILE", s3AccessKeyFile)
	os.Setenv("S3_SECRET_KEY_FILE", s3SecretKeyFile)

	defer func() {
		// Restore original environment
		os.Setenv("DATABASE_URL_FILE", origDatabaseURLFile)
		os.Setenv("S3_ACCESS_KEY_FILE", origS3AccessKeyFile)
		os.Setenv("S3_SECRET_KEY_FILE", origS3SecretKeyFile)
	}()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "postgres://file:test@localhost:5432/testdb", cfg.DatabaseDSN)
	assert.Equal(t, "file-access-key", cfg.S3AccessKey)
	assert.Equal(t, "file-secret-key", cfg.S3SecretKey)
}

func TestLoad_InvalidConfig(t *testing.T) {
	// Save original environment
	origDatabaseURL := os.Getenv("DATABASE_URL")
	origS3AccessKey := os.Getenv("S3_ACCESS_KEY")
	origS3SecretKey := os.Getenv("S3_SECRET_KEY")
	origSMTPHost := os.Getenv("SMTP_HOST")
	origSMTPUsername := os.Getenv("SMTP_USERNAME")

	defer func() {
		// Restore original environment
		os.Setenv("DATABASE_URL", origDatabaseURL)
		os.Setenv("S3_ACCESS_KEY", origS3AccessKey)
		os.Setenv("S3_SECRET_KEY", origS3SecretKey)
		os.Setenv("SMTP_HOST", origSMTPHost)
		os.Setenv("SMTP_USERNAME", origSMTPUsername)
	}()

	// Test missing database URL
	os.Unsetenv("DATABASE_URL")
	os.Setenv("S3_ACCESS_KEY", "test-access-key")
	os.Setenv("S3_SECRET_KEY", "test-secret-key")
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("SMTP_USERNAME", "test@example.com")

	_, err := Load()
	assert.Error(t, err)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want error
	}{
		{
			name: "valid config",
			cfg: Config{
				DatabaseDSN:  "postgres://test:test@localhost:5432/testdb",
				S3AccessKey:  "test-access-key",
				S3SecretKey:  "test-secret-key",
				SMTPHost:     "smtp.example.com",
				SMTPUsername: "test@example.com",
			},
			want: nil,
		},
		{
			name: "missing database DSN",
			cfg: Config{
				S3AccessKey:  "test-access-key",
				S3SecretKey:  "test-secret-key",
				SMTPHost:     "smtp.example.com",
				SMTPUsername: "test@example.com",
			},
			want: fmt.Errorf("DATABASE_URL is required"),
		},
		{
			name: "missing S3 access key",
			cfg: Config{
				DatabaseDSN:  "postgres://test:test@localhost:5432/testdb",
				S3SecretKey:  "test-secret-key",
				SMTPHost:     "smtp.example.com",
				SMTPUsername: "test@example.com",
			},
			want: fmt.Errorf("S3_ACCESS_KEY is required"),
		},
		{
			name: "missing S3 secret key",
			cfg: Config{
				DatabaseDSN:  "postgres://test:test@localhost:5432/testdb",
				S3AccessKey:  "test-access-key",
				SMTPHost:     "smtp.example.com",
				SMTPUsername: "test@example.com",
			},
			want: fmt.Errorf("S3_SECRET_KEY is required"),
		},
		{
			name: "missing SMTP host",
			cfg: Config{
				DatabaseDSN:  "postgres://test:test@localhost:5432/testdb",
				S3AccessKey:  "test-access-key",
				S3SecretKey:  "test-secret-key",
				SMTPUsername: "test@example.com",
			},
			want: fmt.Errorf("SMTP_HOST is required"),
		},
		{
			name: "missing SMTP username",
			cfg: Config{
				DatabaseDSN:  "postgres://test:test@localhost:5432/testdb",
				S3AccessKey:  "test-access-key",
				S3SecretKey:  "test-secret-key",
				SMTPHost:     "smtp.example.com",
			},
			want: fmt.Errorf("SMTP_USERNAME is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.want != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.want.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}