// Package main starts the API server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ppabimanyu/sso/internal/config"
	"github.com/ppabimanyu/sso/internal/domain"
	"github.com/ppabimanyu/sso/internal/httphandler"
	"github.com/ppabimanyu/sso/internal/infra"
	"github.com/ppabimanyu/sso/internal/middleware"
	"github.com/ppabimanyu/sso/internal/repository"
)

// @title SSO Service API
// @version 0.1.0
// @description Identity provider API for every project.
// @BasePath /
// @schemes http https
func main() {
	if err := run(); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := infra.NewLogger(cfg.SlogLevel()).With("service", cfg.OTelServiceName, "env", cfg.AppEnv)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	telemetry, err := infra.SetupTelemetry(ctx, cfg)
	if err != nil {
		return fmt.Errorf("set up telemetry: %w", err)
	}

	pool, err := infra.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	redisClient, err := infra.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer func() { _ = redisClient.Close() }()

	readiness := map[string]domain.HealthChecker{
		"postgres": repository.NewPgHealth(pool),
		"redis":    repository.NewRedisHealth(redisClient),
	}

	handler := httphandler.NewRouter(httphandler.RouterOptions{
		Logger:         logger,
		Readiness:      readiness,
		AllowedOrigins: cfg.CORSAllowedOrigins,
		TrustedProxies: cfg.TrustedProxies,
		RateLimit:      middleware.RateLimit(redisClient, cfg.RateLimitRequests, cfg.RateLimitWindow, logger),
		MaxBodyBytes:   cfg.HTTPMaxBodyBytes,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("listen: %w", err)
		}
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		logger.Info("shutdown requested")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	if err := telemetry.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown telemetry: %w", err)
	}

	logger.Info("shutdown complete")
	return nil
}
