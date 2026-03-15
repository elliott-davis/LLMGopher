package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"

	"github.com/ed007183/llmgopher/internal"
	"github.com/ed007183/llmgopher/internal/api"
	"github.com/ed007183/llmgopher/internal/middleware"
	"github.com/ed007183/llmgopher/internal/proxy"
	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/internal/validation"
	"github.com/ed007183/llmgopher/pkg/config"
	"github.com/ed007183/llmgopher/pkg/llm"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gateway",
		Short: "LLMGopher AI API Gateway",
		Long:  "A high-performance AI API Gateway that proxies requests to multiple LLM providers.",
		RunE:  run,
	}

	config.BindFlags(rootCmd.Flags())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load(cmd.Flags())
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	providerCredentialKey, err := parseProviderCredentialKey(cfg.Security.ProviderCredentialsKey)
	if err != nil {
		return fmt.Errorf("parse provider credential key: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := storage.NewPostgresDB(ctx, cfg.Postgres, logger)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	if err := storage.Migrate(ctx, db); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	stateCache := storage.NewStateCache(logger)
	stateCache.StartPoller(ctx, db, 5*time.Second)

	auditLogger := storage.NewPgAuditLogger(db, logger, 4096)
	defer auditLogger.Close()

	budgetTracker := storage.NewPgBudgetTracker(db, logger)

	pricingStore := storage.NewPgPricingStore(db, logger)
	if err := pricingStore.Load(ctx); err != nil {
		return fmt.Errorf("load pricing cache: %w", err)
	}
	pricingStore.StartRefresh(cfg.Gateway.PricingCacheInterval)
	defer pricingStore.Close()

	if cfg.Gateway.PricingSyncEnabled {
		var syncOpts []storage.PricingSyncerOption
		if cfg.Gateway.PricingSyncURL != "" {
			syncOpts = append(syncOpts, storage.WithUpstreamURL(cfg.Gateway.PricingSyncURL))
		}
		syncer := storage.NewPricingSyncer(pricingStore, logger, syncOpts...)
		syncer.Start(cfg.Gateway.PricingSyncInterval)
		defer syncer.Close()
	}

	var rateLimiter llm.RateLimiter
	if cfg.Redis.Enabled {
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Warn("Redis unavailable, falling back to in-memory rate limiter", "error", err)
			rateLimiter = middleware.NewInMemoryRateLimiter(cfg.Gateway.RateLimitRPS, cfg.Gateway.RateLimitBurst)
		} else {
			logger.Info("connected to Redis", "addr", cfg.Redis.Addr)
			rateLimiter = middleware.NewRedisRateLimiter(rdb, cfg.Gateway.RateLimitRPS, cfg.Gateway.RateLimitBurst, logger)
			defer rdb.Close()
		}
	} else {
		logger.Info("Redis disabled, using in-memory rate limiter")
		rateLimiter = middleware.NewInMemoryRateLimiter(cfg.Gateway.RateLimitRPS, cfg.Gateway.RateLimitBurst)
	}

	var guardrail llm.Guardrail
	if cfg.Gateway.GuardrailEndpoint != "" {
		guardrail = middleware.NewNemoGuardrail(cfg.Gateway.GuardrailEndpoint, cfg.Gateway.GuardrailTimeout, logger)
		logger.Info("guardrail enabled", "endpoint", cfg.Gateway.GuardrailEndpoint)
	} else {
		guardrail = internal.NoopGuardrail{}
		logger.Info("guardrail disabled (no endpoint configured)")
	}

	registry := llm.NewRegistry()

	if cfg.Providers.OpenAI.APIKey != "" {
		openaiProvider := proxy.NewOpenAIProvider(
			cfg.Providers.OpenAI.APIKey,
			cfg.Providers.OpenAI.BaseURL,
		)
		registry.Register(openaiProvider, "gpt-4*", "gpt-3.5*", "o1*", "o3*", "chatgpt*")
		logger.Info("registered OpenAI provider")
	}

	if cfg.Providers.Anthropic.APIKey != "" {
		anthropicProvider := proxy.NewAnthropicProvider(
			cfg.Providers.Anthropic.APIKey,
			cfg.Providers.Anthropic.BaseURL,
		)
		registry.Register(anthropicProvider, "claude*")
		logger.Info("registered Anthropic provider")
	}

	if cfg.Providers.Vertex.ProjectID != "" {
		gcpTokenSource, err := proxy.NewGoogleCloudTokenSource(ctx)
		if err != nil {
			logger.Warn(
				"Vertex AI not registered: failed to initialize Google ADC credentials",
				"error", err,
				"project_id", cfg.Providers.Vertex.ProjectID,
				"region", cfg.Providers.Vertex.Region,
			)
		} else {
			vertexProvider := proxy.NewVertexProvider(
				cfg.Providers.Vertex.ProjectID,
				cfg.Providers.Vertex.Region,
				gcpTokenSource,
			)
			registry.Register(vertexProvider, "vertex/*")
			logger.Info("registered Vertex AI provider", "project_id", cfg.Providers.Vertex.ProjectID, "region", cfg.Providers.Vertex.Region)
		}
	}

	deps := &api.Dependencies{
		Registry:               registry,
		RateLimiter:            rateLimiter,
		Guardrail:              guardrail,
		AuditLogger:            auditLogger,
		BudgetTracker:          budgetTracker,
		Pricing:                pricingStore,
		StateCache:             stateCache,
		DB:                     db,
		CredentialValidator:    validation.NewCredentialValidator(&http.Client{Timeout: 12 * time.Second}),
		ProviderCredentialsKey: providerCredentialKey,
		APIKeys:                cfg.Auth.APIKeys,
		Logger:                 logger,
	}

	handler := api.NewRouter(deps)

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting gateway", "addr", cfg.Server.Addr)
		errCh <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("received shutdown signal", "signal", sig)
		cancel()
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	logger.Info("gateway stopped gracefully")
	return nil
}

func parseProviderCredentialKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, nil
	}

	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("decoded key must be 32 bytes, got %d", len(key))
	}
	return key, nil
}
