package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TaushifReza/email-platform/backend/api"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/crypto"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/health"
	"github.com/TaushifReza/email-platform/backend/internal/platform/logging"
	"github.com/TaushifReza/email-platform/backend/internal/platform/ratelimit"
	"github.com/gin-gonic/gin"
	redis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const shutdownTimeout = 10 * time.Second

type redisPinger struct {
	client *redis.Client
}

func (p redisPinger) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

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

	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		options, parseErr := redis.ParseURL(cfg.RedisURL)
		if parseErr != nil {
			logger.Warn("Redis URL is invalid; continuing without Redis", zap.Error(parseErr))
		} else {
			redisClient = redis.NewClient(options)
			defer redisClient.Close()
		}
	}

	var redisHealth health.Pinger
	if redisClient != nil {
		redisHealth = redisPinger{client: redisClient}
	}

	identityRepo := identity.NewRepository(pool)
	sessionRepo := identity.NewSessionRepository(pool)
	mfaRepo := identity.NewMFARepository(pool)
	hasher, err := identity.NewHasher(identity.DefaultArgon2Params)
	if err != nil {
		logger.Fatal("Failed to create password hasher", zap.Error(err))
	}
	keyring, err := crypto.NewKeyring(cfg.MasterKey)
	if err != nil {
		logger.Fatal("Failed to create crypto keyring", zap.Error(err))
	}
	var loginLimiter identity.RateLimiter
	if redisClient != nil {
		loginLimiter = ratelimit.New(ratelimit.NewRedisCounter(redisClient))
	}
	identityService := identity.NewService(identityRepo, sessionRepo, hasher,
		loginLimiter, clock.System{}, identity.DefaultLockoutPolicy, identity.DefaultSessionPolicy)
	mfaService := identity.NewMFAService(mfaRepo, sessionRepo, hasher,
		loginLimiter, clock.System{}, keyring)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	server := api.NewServer(health.New(pool, redisHealth),
		identityService,
		logger)
	server.SetMFAService(mfaService)
	server.SetConfig(cfg)
	router.Use(api.RequestIDMiddleware())
	router.Use(api.CSRFMiddleware())
	router.Use(api.AuthMiddleware(server))
	api.RegisterHandlersWithOptions(router, server, api.GinServerOptions{
		ErrorHandler: func(c *gin.Context, err error, statusCode int) {
			api.RespondProblem(c, statusCode, api.CodeInvalidRequest, err.Error(), nil)
		},
	})
	if cfg.DevMode {
		api.RegisterDevRoutes(router, cfg.OpenAPISpecPath)
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("API service starting", zap.String("addr", cfg.ListenAddr))
		if serveErr := srv.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErr <- serveErr
		}
	}()

	select {
	case serveErr := <-serverErr:
		logger.Error("API server error", zap.Error(serveErr))
		return
	case <-ctx.Done():
		logger.Info("API service shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("API shutdown failed", zap.Error(err))
	}
}
