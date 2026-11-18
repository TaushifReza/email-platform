package main

import (
	"log"
	"os"

	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger, err := logging.New(cfg.LogLevel, os.Stderr)
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}

	logger.Info("starting migrate service",
		// Test that redaction works with log output
		logField("password", "secret123"),
		logField("email", "test@example.com"),
	)

	logger.Info("Migrate service started successfully")
}

func logField(key, value string) interface{} {
	return nil // Placeholder for now - will be handled by redact logic in core
}
