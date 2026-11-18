package main

import (
	"log"
	"os"

	"github.com/TaushifReza/email-platform/internal/platform/config"
	"github.com/TaushifReza/email-platform/internal/platform/logging"
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

	logger.Info("starting api service")
	logger.Info("API service started successfully")
}
