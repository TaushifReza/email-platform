package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	logger, err := logging.New(cfg.LogLevel, os.Stderr)
	if err != nil {
		log.Fatal("Failed to create logger: ", err)
	}
	defer func() {
		_ = logger.Sync() //nolint:errcheck // Best-effort flush on shutdown.
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("worker started")
	<-ctx.Done()
	logger.Info("worker shutting down")
}
