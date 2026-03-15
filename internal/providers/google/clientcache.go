package google

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	gemini "github.com/google/generative-ai-go/genai"
	vertex "cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"
)

// GeminiClientCache is a thread-safe cache of Gemini API clients keyed by API key.
type GeminiClientCache struct {
	mu      sync.RWMutex
	clients map[string]*gemini.Client
}

func NewGeminiClientCache() *GeminiClientCache {
	return &GeminiClientCache{
		clients: make(map[string]*gemini.Client),
	}
}

// Get returns a cached Gemini client for the given API key, creating one lazily
// on the first call. The returned client is safe for concurrent use.
func (c *GeminiClientCache) Get(ctx context.Context, apiKey string) (*gemini.Client, error) {
	key := hashKey(apiKey)

	c.mu.RLock()
	if client, ok := c.clients[key]; ok {
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if client, ok := c.clients[key]; ok {
		return client, nil
	}

	client, err := gemini.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	c.clients[key] = client
	return client, nil
}

// Close shuts down all cached Gemini clients.
func (c *GeminiClientCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var firstErr error
	for key, client := range c.clients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(c.clients, key)
	}
	return firstErr
}

// VertexClientKey identifies a unique Vertex AI client configuration.
type VertexClientKey struct {
	ProjectID       string
	Region          string
	CredentialsJSON string // optional; empty means use ADC
	CredentialsFile string // optional; empty means use ADC
}

func (k VertexClientKey) cacheKey() string {
	h := sha256.New()
	h.Write([]byte(k.ProjectID))
	h.Write([]byte{0})
	h.Write([]byte(k.Region))
	h.Write([]byte{0})
	h.Write([]byte(k.CredentialsJSON))
	h.Write([]byte{0})
	h.Write([]byte(k.CredentialsFile))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// VertexClientCache is a thread-safe cache of Vertex AI clients keyed by
// the composite of ProjectID + Region + Credentials.
type VertexClientCache struct {
	mu      sync.RWMutex
	clients map[string]*vertex.Client
}

func NewVertexClientCache() *VertexClientCache {
	return &VertexClientCache{
		clients: make(map[string]*vertex.Client),
	}
}

// Get returns a cached Vertex AI client for the given configuration, creating
// one lazily on the first call. The returned client is safe for concurrent use.
func (c *VertexClientCache) Get(ctx context.Context, key VertexClientKey) (*vertex.Client, error) {
	cacheKey := key.cacheKey()

	c.mu.RLock()
	if client, ok := c.clients[cacheKey]; ok {
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if client, ok := c.clients[cacheKey]; ok {
		return client, nil
	}

	opts := buildVertexOpts(key)
	client, err := vertex.NewClient(ctx, key.ProjectID, key.Region, opts...)
	if err != nil {
		return nil, fmt.Errorf("create vertex client (project=%s, region=%s): %w",
			key.ProjectID, key.Region, err)
	}

	c.clients[cacheKey] = client
	return client, nil
}

// Close shuts down all cached Vertex AI clients.
func (c *VertexClientCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var firstErr error
	for key, client := range c.clients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(c.clients, key)
	}
	return firstErr
}

func buildVertexOpts(key VertexClientKey) []option.ClientOption {
	var opts []option.ClientOption
	if key.CredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(key.CredentialsJSON)))
	}
	if key.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(key.CredentialsFile))
	}
	return opts
}

// hashKey produces a hex-encoded SHA-256 of the input to avoid storing
// raw API keys as map keys in memory.
func hashKey(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:])
}
