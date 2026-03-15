package llm

import (
	"math"
	"testing"
)

func TestModelPricing_CostUSD(t *testing.T) {
	tests := []struct {
		name       string
		pricing    ModelPricing
		prompt     int
		completion int
		wantCost   float64
	}{
		{
			name:       "gpt-4o 1k prompt 500 completion",
			pricing:    ModelPricing{PromptPer1K: 0.0025, CompletionPer1K: 0.0100},
			prompt:     1000,
			completion: 500,
			wantCost:   0.0025 + 0.0050,
		},
		{
			name:       "zero tokens",
			pricing:    ModelPricing{PromptPer1K: 0.01, CompletionPer1K: 0.03},
			prompt:     0,
			completion: 0,
			wantCost:   0,
		},
		{
			name:       "fractional tokens",
			pricing:    ModelPricing{PromptPer1K: 0.0005, CompletionPer1K: 0.0015},
			prompt:     150,
			completion: 50,
			wantCost:   (150.0/1000)*0.0005 + (50.0/1000)*0.0015,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pricing.CostUSD(tc.prompt, tc.completion)
			if math.Abs(got-tc.wantCost) > 1e-12 {
				t.Errorf("CostUSD(%d, %d) = %.12f, want %.12f", tc.prompt, tc.completion, got, tc.wantCost)
			}
		})
	}
}
