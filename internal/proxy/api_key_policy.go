package proxy

import (
	"context"
	"errors"
	"slices"

	"github.com/ed007183/llmgopher/internal/middleware"
	"github.com/ed007183/llmgopher/internal/storage"
)

var errModelNotAllowed = errors.New("model not allowed for this API key")

func enforceAllowedModel(ctx context.Context, stateCache *storage.StateCache, model string) error {
	if stateCache == nil {
		return nil
	}

	apiKeyID := middleware.GetAPIKeyID(ctx)
	if apiKeyID == "" {
		return nil
	}

	state := stateCache.Load()
	if state == nil {
		return nil
	}

	key, ok := state.APIKeysByID[apiKeyID]
	if !ok {
		return nil
	}

	if len(key.AllowedModels) > 0 && !slices.Contains(key.AllowedModels, model) {
		return errModelNotAllowed
	}

	return nil
}
