package google

import (
	"testing"
)

func TestHashKey(t *testing.T) {
	h1 := hashKey("sk-abc123")
	h2 := hashKey("sk-abc123")
	h3 := hashKey("sk-different")

	if h1 != h2 {
		t.Errorf("same input produced different hashes: %s vs %s", h1, h2)
	}
	if h1 == h3 {
		t.Error("different inputs produced the same hash")
	}
	if len(h1) != 64 {
		t.Errorf("expected 64-char hex SHA-256, got %d chars", len(h1))
	}
}

func TestVertexClientKey_CacheKey(t *testing.T) {
	k1 := VertexClientKey{ProjectID: "proj-1", Region: "us-central1"}
	k2 := VertexClientKey{ProjectID: "proj-1", Region: "us-central1"}
	k3 := VertexClientKey{ProjectID: "proj-1", Region: "europe-west1"}
	k4 := VertexClientKey{ProjectID: "proj-2", Region: "us-central1"}
	k5 := VertexClientKey{ProjectID: "proj-1", Region: "us-central1", CredentialsJSON: `{"type":"service_account"}`}

	if k1.cacheKey() != k2.cacheKey() {
		t.Error("identical configs produced different cache keys")
	}
	if k1.cacheKey() == k3.cacheKey() {
		t.Error("different regions produced the same cache key")
	}
	if k1.cacheKey() == k4.cacheKey() {
		t.Error("different projects produced the same cache key")
	}
	if k1.cacheKey() == k5.cacheKey() {
		t.Error("different credentials produced the same cache key")
	}
}

func TestGeminiClientCache_NewAndClose(t *testing.T) {
	cache := NewGeminiClientCache()
	if cache == nil {
		t.Fatal("NewGeminiClientCache returned nil")
	}
	if len(cache.clients) != 0 {
		t.Errorf("new cache should be empty, got %d entries", len(cache.clients))
	}
	if err := cache.Close(); err != nil {
		t.Errorf("closing empty cache: %v", err)
	}
}

func TestVertexClientCache_NewAndClose(t *testing.T) {
	cache := NewVertexClientCache()
	if cache == nil {
		t.Fatal("NewVertexClientCache returned nil")
	}
	if len(cache.clients) != 0 {
		t.Errorf("new cache should be empty, got %d entries", len(cache.clients))
	}
	if err := cache.Close(); err != nil {
		t.Errorf("closing empty cache: %v", err)
	}
}

func TestBuildVertexOpts(t *testing.T) {
	tests := []struct {
		name     string
		key      VertexClientKey
		wantOpts int
	}{
		{
			name:     "no credentials (ADC)",
			key:      VertexClientKey{ProjectID: "p", Region: "r"},
			wantOpts: 0,
		},
		{
			name:     "credentials JSON only",
			key:      VertexClientKey{ProjectID: "p", Region: "r", CredentialsJSON: `{}`},
			wantOpts: 1,
		},
		{
			name:     "credentials file only",
			key:      VertexClientKey{ProjectID: "p", Region: "r", CredentialsFile: "/path/to/sa.json"},
			wantOpts: 1,
		},
		{
			name:     "both credentials JSON and file",
			key:      VertexClientKey{ProjectID: "p", Region: "r", CredentialsJSON: `{}`, CredentialsFile: "/path"},
			wantOpts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildVertexOpts(tt.key)
			if len(opts) != tt.wantOpts {
				t.Errorf("buildVertexOpts() returned %d opts, want %d", len(opts), tt.wantOpts)
			}
		})
	}
}
