package storage

import (
	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// GatewayState is a read-optimized in-memory snapshot of dynamic routing
// and auth configuration. A full snapshot is atomically swapped in by
// StateCache so request paths can do lock-free lookups.
type GatewayState struct {
	APIKeys     map[string]*llm.APIKey
	APIKeysByID map[string]*llm.APIKey
	Models      map[string]*llm.Model
	Providers   map[uuid.UUID]*llm.ProviderConfig
}
