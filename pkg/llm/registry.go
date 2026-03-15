package llm

import (
	"fmt"
	"strings"
	"sync"
)

// DefaultRegistry is a thread-safe ProviderRegistry that maps model prefixes to providers.
type DefaultRegistry struct {
	mu       sync.RWMutex
	exact    map[string]Provider // exact model name -> provider
	prefixes []prefixEntry       // prefix -> provider, checked in registration order
	byName   map[string]Provider // canonical provider name -> provider
}

type prefixEntry struct {
	prefix   string
	provider Provider
}

func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		exact:  make(map[string]Provider),
		byName: make(map[string]Provider),
	}
}

func (r *DefaultRegistry) Register(provider Provider, models ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byName[strings.ToLower(provider.Name())] = provider
	for _, m := range models {
		if strings.HasSuffix(m, "*") {
			r.prefixes = append(r.prefixes, prefixEntry{
				prefix:   strings.TrimSuffix(m, "*"),
				provider: provider,
			})
		} else {
			r.exact[m] = provider
		}
	}
}

func (r *DefaultRegistry) Resolve(model string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if p, ok := r.exact[model]; ok {
		return p, nil
	}
	for _, pe := range r.prefixes {
		if strings.HasPrefix(model, pe.prefix) {
			return pe.provider, nil
		}
	}
	return nil, fmt.Errorf("no provider registered for model %q", model)
}

func (r *DefaultRegistry) ResolveProvider(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if p, ok := r.byName[strings.ToLower(name)]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("no provider registered with name %q", name)
}
