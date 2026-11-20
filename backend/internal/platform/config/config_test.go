package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb?sslmode=disable")
	t.Setenv("STORAGE_ACCESS_KEY", "test-access-key")
	t.Setenv("STORAGE_SECRET_KEY", "test-secret-key")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_USERNAME", "test@example.com")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb?sslmode=disable", cfg.DatabaseURL)
	assert.Equal(t, "test-access-key", cfg.StorageAccessKey)
	assert.Equal(t, "test-secret-key", cfg.StorageSecretKey)
	assert.Equal(t, "smtp.example.com", cfg.SMTPHost)
	assert.Equal(t, "test@example.com", cfg.SMTPUsername)
}

func TestLoadFileSecretsTakePrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	databaseURLFile := writeSecret(t, tmpDir, "database_url", "postgres://file:test@localhost:5432/testdb")
	accessKeyFile := writeSecret(t, tmpDir, "access_key", "file-access-key")
	secretKeyFile := writeSecret(t, tmpDir, "secret_key", "file-secret-key")
	passwordFile := writeSecret(t, tmpDir, "smtp_password", "file-smtp-password")
	masterKeyFile := writeSecret(t, tmpDir, "master_key", "file-master-key")

	t.Setenv("DATABASE_URL", "postgres://env:test@localhost:5432/testdb")
	t.Setenv("DATABASE_URL_FILE", databaseURLFile)
	t.Setenv("STORAGE_ACCESS_KEY", "env-access-key")
	t.Setenv("STORAGE_ACCESS_KEY_FILE", accessKeyFile)
	t.Setenv("STORAGE_SECRET_KEY", "env-secret-key")
	t.Setenv("STORAGE_SECRET_KEY_FILE", secretKeyFile)
	t.Setenv("SMTP_PASSWORD", "env-smtp-password")
	t.Setenv("SMTP_PASSWORD_FILE", passwordFile)
	t.Setenv("MASTER_KEY", "env-master-key")
	t.Setenv("MASTER_KEY_FILE", masterKeyFile)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "postgres://file:test@localhost:5432/testdb", cfg.DatabaseURL)
	assert.Equal(t, "file-access-key", cfg.StorageAccessKey)
	assert.Equal(t, "file-secret-key", cfg.StorageSecretKey)
	assert.Equal(t, "file-smtp-password", cfg.SMTPPassword)
	assert.Equal(t, "file-master-key", string(cfg.MasterKey))
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_FILE", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL is required")
}

func TestLoadFailsWhenSecretFileIsMissing(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
	t.Setenv("STORAGE_ACCESS_KEY_FILE", filepath.Join(t.TempDir(), "missing"))

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "STORAGE_ACCESS_KEY_FILE")
}

func TestStorageAndSMTPAreValidatedOnlyWhenRequested(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
	t.Setenv("STORAGE_ACCESS_KEY", "")
	t.Setenv("STORAGE_SECRET_KEY", "")
	t.Setenv("STORAGE_ACCESS_KEY_FILE", "")
	t.Setenv("STORAGE_SECRET_KEY_FILE", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_USERNAME", "")

	cfg, err := Load()
	require.NoError(t, err)

	require.Error(t, cfg.ValidateStorage())
	assert.Contains(t, cfg.ValidateStorage().Error(), "STORAGE_ACCESS_KEY is required")

	require.Error(t, cfg.ValidateSMTP())
	assert.Contains(t, cfg.ValidateSMTP().Error(), "SMTP_HOST is required")
}

func TestValidateDatabase(t *testing.T) {
	cfg := Config{}
	require.Error(t, cfg.ValidateDatabase())
	assert.Contains(t, cfg.ValidateDatabase().Error(), "DATABASE_URL is required")

	cfg.DatabaseURL = "postgres://localhost/db"
	require.NoError(t, cfg.ValidateDatabase())
}

func TestSMTPValidationDoesNotRequireUsernameForMailpit(t *testing.T) {
	cfg := Config{SMTPHost: "mailpit", SMTPPort: 1025}
	require.NoError(t, cfg.ValidateSMTP())
}

func writeSecret(t *testing.T, dir, name, value string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(value+"\n"), 0o600))

	return path
}
