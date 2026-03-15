package storage

import (
	"testing"

	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestPgPricingStore_LookupPricing_ExactMatch(t *testing.T) {
	ps := &PgPricingStore{
		cache: map[string]llm.ModelPricing{
			"gpt-4o":           {PromptPer1K: 0.0025, CompletionPer1K: 0.0100},
			"gpt-4o-mini":      {PromptPer1K: 0.00015, CompletionPer1K: 0.0006},
			"claude-3-5-sonnet": {PromptPer1K: 0.0030, CompletionPer1K: 0.0150},
		},
	}

	tests := []struct {
		model      string
		wantPrompt float64
		wantCompl  float64
	}{
		{"gpt-4o", 0.0025, 0.0100},
		{"gpt-4o-mini", 0.00015, 0.0006},
		{"claude-3-5-sonnet", 0.0030, 0.0150},
	}

	for _, tc := range tests {
		t.Run(tc.model, func(t *testing.T) {
			p := ps.LookupPricing(tc.model)
			if p.PromptPer1K != tc.wantPrompt {
				t.Errorf("PromptPer1K = %v, want %v", p.PromptPer1K, tc.wantPrompt)
			}
			if p.CompletionPer1K != tc.wantCompl {
				t.Errorf("CompletionPer1K = %v, want %v", p.CompletionPer1K, tc.wantCompl)
			}
		})
	}
}

func TestPgPricingStore_LookupPricing_PrefixMatch(t *testing.T) {
	ps := &PgPricingStore{
		cache: map[string]llm.ModelPricing{
			"gpt-4o":            {PromptPer1K: 0.0025, CompletionPer1K: 0.0100},
			"gpt-4-turbo":       {PromptPer1K: 0.0100, CompletionPer1K: 0.0300},
			"claude-3-5-sonnet": {PromptPer1K: 0.0030, CompletionPer1K: 0.0150},
		},
	}

	tests := []struct {
		model      string
		wantPrefix string
	}{
		{"gpt-4o-2024-08-06", "gpt-4o"},
		{"claude-3-5-sonnet-20241022", "claude-3-5-sonnet"},
		{"gpt-4-turbo-preview", "gpt-4-turbo"},
	}

	for _, tc := range tests {
		t.Run(tc.model, func(t *testing.T) {
			got := ps.LookupPricing(tc.model)
			want := ps.LookupPricing(tc.wantPrefix)
			if got.PromptPer1K != want.PromptPer1K || got.CompletionPer1K != want.CompletionPer1K {
				t.Errorf("LookupPricing(%q) = %+v, want same as %q = %+v",
					tc.model, got, tc.wantPrefix, want)
			}
		})
	}
}

func TestPgPricingStore_LookupPricing_UnknownModel_ReturnsDefault(t *testing.T) {
	ps := &PgPricingStore{
		cache: map[string]llm.ModelPricing{
			"gpt-4o": {PromptPer1K: 0.0025, CompletionPer1K: 0.0100},
		},
	}

	p := ps.LookupPricing("totally-unknown-model-xyz")
	if p.PromptPer1K != defaultPricing.PromptPer1K {
		t.Errorf("PromptPer1K = %v, want default %v", p.PromptPer1K, defaultPricing.PromptPer1K)
	}
	if p.CompletionPer1K != defaultPricing.CompletionPer1K {
		t.Errorf("CompletionPer1K = %v, want default %v", p.CompletionPer1K, defaultPricing.CompletionPer1K)
	}
}

func TestPgPricingStore_LookupPricing_CaseInsensitive(t *testing.T) {
	ps := &PgPricingStore{
		cache: map[string]llm.ModelPricing{
			"gpt-4o": {PromptPer1K: 0.0025, CompletionPer1K: 0.0100},
		},
	}

	lower := ps.LookupPricing("gpt-4o")
	upper := ps.LookupPricing("GPT-4O")
	if lower != upper {
		t.Errorf("case sensitivity mismatch: %+v vs %+v", lower, upper)
	}
}

func TestPgPricingStore_LookupPricing_LongestPrefixWins(t *testing.T) {
	ps := &PgPricingStore{
		cache: map[string]llm.ModelPricing{
			"gpt-4":       {PromptPer1K: 0.0300, CompletionPer1K: 0.0600},
			"gpt-4o":      {PromptPer1K: 0.0025, CompletionPer1K: 0.0100},
			"gpt-4o-mini": {PromptPer1K: 0.00015, CompletionPer1K: 0.0006},
		},
	}

	p := ps.LookupPricing("gpt-4o-mini-2024-07-18")
	if p.PromptPer1K != 0.00015 {
		t.Errorf("expected gpt-4o-mini pricing, got PromptPer1K = %v", p.PromptPer1K)
	}
}
