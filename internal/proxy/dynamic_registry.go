package proxy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
		switch {
		case supportsOpenAICompatAuthType(authType):
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

		case authType == "aws_bedrock":
			region := resolveBedrockRegion(providerCfg.BaseURL)
			if region == "" {
				logger.Warn("skipping bedrock provider registration without region",
					"provider_id", providerID.String(),
					"provider_name", providerCfg.Name,
				)
				continue
			}

			creds, err := parseBedrockCredentialPayload(credentialTokens[providerID])
			if err != nil {
				logger.Warn("skipping bedrock provider registration with invalid credentials",
					"provider_id", providerID.String(),
					"provider_name", providerCfg.Name,
					"error", err,
				)
				continue
			}

			provider, err := NewBedrockProvider(region, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)
			if err != nil {
				logger.Warn("skipping bedrock provider registration",
					"provider_id", providerID.String(),
					"provider_name", providerCfg.Name,
					"region", region,
					"error", err,
				)
				continue
			}

			registry.Register(provider)
			logger.Debug("registered dynamic Bedrock provider",
				"provider_id", providerID.String(),
				"provider_name", providerCfg.Name,
				"auth_type", authType,
				"region", region,
			)
		}
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

type bedrockCredentialPayload struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token,omitempty"`
}

func parseBedrockCredentialPayload(raw string) (bedrockCredentialPayload, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return bedrockCredentialPayload{}, nil
	}

	var payload bedrockCredentialPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return bedrockCredentialPayload{}, fmt.Errorf("credential_token must be JSON")
	}

	payload.AccessKeyID = strings.TrimSpace(payload.AccessKeyID)
	payload.SecretAccessKey = strings.TrimSpace(payload.SecretAccessKey)
	payload.SessionToken = strings.TrimSpace(payload.SessionToken)

	if (payload.AccessKeyID == "") != (payload.SecretAccessKey == "") {
		return bedrockCredentialPayload{}, fmt.Errorf("access_key_id and secret_access_key must both be provided")
	}

	return payload, nil
}
