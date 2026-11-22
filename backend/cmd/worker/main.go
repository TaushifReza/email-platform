package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/notifications"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/logging"
	"github.com/TaushifReza/email-platform/backend/internal/platform/outbox"
	"go.uber.org/zap"
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
	defer func() { _ = logger.Sync() }() //nolint:errcheck

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewWithSystem(ctx, cfg.DatabaseURL, cfg.SystemDatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	sender := notifications.NewSMTPSender(cfg)
	processor := notifications.NewProcessor(pool, outbox.NewStore(pool), sender, clock.System{}, logger)
	go processor.Run(ctx)

	logger.Info("worker started")
	<-ctx.Done()
	// allow in-flight send to finish briefly
	time.Sleep(200 * time.Millisecond)
	logger.Info("worker shutting down")
}
