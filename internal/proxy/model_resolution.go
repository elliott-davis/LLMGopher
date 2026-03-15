package proxy

import (
	"strings"

	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/pkg/llm"
)

func resolveConfiguredModel(state *storage.GatewayState, requestedModel string) (*llm.Model, bool) {
	if state == nil {
		return nil, false
	}

	requestedModel = strings.TrimSpace(requestedModel)
	if requestedModel == "" {
		return nil, false
	}

	if modelCfg, ok := state.Models[requestedModel]; ok {
		return modelCfg, true
	}

	for _, candidate := range state.Models {
		if strings.EqualFold(candidate.Name, requestedModel) {
			return candidate, true
		}
	}

	providerName, modelName, ok := splitRequestedModel(requestedModel)
	if !ok {
		return nil, false
	}

	for _, candidate := range state.Models {
		if !modelNameMatches(candidate.Name, modelName) {
			continue
		}

		providerID, err := uuid.Parse(candidate.ProviderID)
		if err != nil {
			continue
		}
		providerCfg, ok := state.Providers[providerID]
		if !ok {
			continue
		}

		if providerNameMatches(providerCfg, providerName) {
			return candidate, true
		}
	}

	return nil, false
}

func splitRequestedModel(requestedModel string) (providerName string, modelName string, ok bool) {
	requestedModel = strings.TrimSpace(requestedModel)
	idx := strings.Index(requestedModel, "/")
	if idx <= 0 || idx >= len(requestedModel)-1 {
		return "", "", false
	}

	providerName = strings.TrimSpace(requestedModel[:idx])
	modelName = strings.TrimSpace(requestedModel[idx+1:])
	if providerName == "" || modelName == "" {
		return "", "", false
	}
	return providerName, modelName, true
}

func modelNameMatches(candidateName, requested string) bool {
	candidateName = strings.TrimSpace(candidateName)
	requested = strings.TrimSpace(requested)
	if candidateName == "" || requested == "" {
		return false
	}

	if strings.EqualFold(candidateName, requested) {
		return true
	}

	if slashIdx := strings.LastIndex(candidateName, "/"); slashIdx >= 0 && slashIdx < len(candidateName)-1 {
		return strings.EqualFold(candidateName[slashIdx+1:], requested)
	}

	return false
}

func preferredProviderRegistryName(providerCfg *llm.ProviderConfig, requestedModel string) string {
	if providerCfg == nil {
		if requestedProvider, _, ok := splitRequestedModel(requestedModel); ok {
			return requestedProvider
		}
		return ""
	}

	if requestedProvider, _, ok := splitRequestedModel(requestedModel); ok {
		return requestedProvider
	}

	if inferred := inferProviderRegistryName(providerCfg); inferred != "" {
		return inferred
	}

	return strings.TrimSpace(providerCfg.Name)
}

func providerNameMatches(providerCfg *llm.ProviderConfig, requestedProvider string) bool {
	requestedProvider = strings.TrimSpace(requestedProvider)
	if providerCfg == nil || requestedProvider == "" {
		return false
	}

	if strings.EqualFold(strings.TrimSpace(providerCfg.Name), requestedProvider) {
		return true
	}

	inferred := inferProviderRegistryName(providerCfg)
	return inferred != "" && strings.EqualFold(inferred, requestedProvider)
}

func inferProviderRegistryName(providerCfg *llm.ProviderConfig) string {
	if providerCfg == nil {
		return ""
	}

	name := strings.ToLower(strings.TrimSpace(providerCfg.Name))
	baseURL := strings.ToLower(strings.TrimSpace(providerCfg.BaseURL))
	authType := strings.ToLower(strings.TrimSpace(providerCfg.AuthType))

	switch {
	case strings.Contains(name, "vertex"),
		strings.Contains(baseURL, "aiplatform.googleapis.com"),
		authType == "vertex_service_account":
		return "vertex"
	case strings.Contains(name, "openai"),
		strings.Contains(baseURL, "api.openai.com"):
		return "openai"
	case strings.Contains(name, "anthropic"),
		strings.Contains(baseURL, "api.anthropic.com"):
		return "anthropic"
	default:
		return ""
	}
}
