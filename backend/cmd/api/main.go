package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/health"
	"github.com/TaushifReza/email-platform/backend/internal/platform/logging"
	"github.com/jackc/pgx/v5"
	redis "github.com/go-redis/redis/v8"
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

	// Initialize health checker with database
	dbChecker := &health.DBChecker{DSN: cfg.DatabaseDSN}
	
	// Connect to Redis if configured
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr: cfg.RedisURL,
		})
	}

	// Create health server
	h := health.New(dbChecker, redisClient)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.Healthz)
	mux.HandleFunc("/readyz", h.Readyz)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("API service starting on " + cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("API server error", "error", err)
		}
	}()

	// Keep the main goroutine alive
	select {}
}
