package proxy

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/internal/validation"
	"github.com/ed007183/llmgopher/pkg/llm"
)

const dynamicProviderSyncInterval = 5 * time.Second

// StartDynamicProviderSync loads DB-backed OpenAI-compatible providers from the
// hot-swapped state cache and refreshes their registry entries on a timer.
func StartDynamicProviderSync(
	ctx context.Context,
	registry llm.ProviderRegistry,
	stateCache *storage.StateCache,
	db *sql.DB,
	providerCredentialKey []byte,
	logger *slog.Logger,
) {
	if registry == nil || stateCache == nil || db == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}

	syncOnce := func() {
		if err := syncDynamicProviderRegistrations(ctx, registry, stateCache, db, providerCredentialKey, logger); err != nil {
			logger.Warn("dynamic provider sync failed", "error", err)
		}
	}

	syncOnce()

	ticker := time.NewTicker(dynamicProviderSyncInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				syncOnce()
			}
		}
	}()
}

func syncDynamicProviderRegistrations(
	ctx context.Context,
	registry llm.ProviderRegistry,
	stateCache *storage.StateCache,
	db *sql.DB,
	providerCredentialKey []byte,
	logger *slog.Logger,
) error {
	state := stateCache.Load()
	if state == nil {
		return nil
	}

	credentialTokens, err := validation.LoadProviderCredentialTokens(ctx, db, providerCredentialKey)
	if err != nil {
		return err
	}

	RegisterDynamicOpenAICompatProviders(registry, state, credentialTokens, logger)
	return nil
}

// RegisterDynamicOpenAICompatProviders registers DB-configured providers that
// speak OpenAI-compatible APIs and can be handled by OpenAICompatProvider.
func RegisterDynamicOpenAICompatProviders(
	registry llm.ProviderRegistry,
	state *storage.GatewayState,
	credentialTokens map[uuid.UUID]string,
	logger *slog.Logger,
) {
	if registry == nil || state == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}

	for providerID, providerCfg := range state.Providers {
		if providerCfg == nil {
			continue
		}

		authType := strings.ToLower(strings.TrimSpace(providerCfg.AuthType))
		if !supportsOpenAICompatAuthType(authType) {
			continue
		}

		baseURL := ResolveProviderBaseURL(providerCfg.Name, providerCfg.BaseURL)
		if baseURL == "" {
			logger.Warn("skipping dynamic provider registration without base URL",
				"provider_id", providerID.String(),
				"provider_name", providerCfg.Name,
			)
			continue
		}

		apiKey := ""
		if authType != "none" {
			var ok bool
			apiKey, ok = credentialTokens[providerID]
			if !ok || strings.TrimSpace(apiKey) == "" {
				logger.Warn("skipping dynamic provider registration without credentials",
					"provider_id", providerID.String(),
					"provider_name", providerCfg.Name,
					"auth_type", authType,
				)
				continue
			}
		}

		registry.Register(NewOpenAICompatProvider(providerCfg.Name, baseURL, apiKey))
		logger.Debug("registered dynamic OpenAI-compatible provider",
			"provider_id", providerID.String(),
			"provider_name", providerCfg.Name,
			"auth_type", authType,
			"base_url", baseURL,
		)
	}
}

func supportsOpenAICompatAuthType(authType string) bool {
	switch strings.ToLower(strings.TrimSpace(authType)) {
	case "bearer", "openai_compat", "none":
		return true
	default:
		return false
	}
}
