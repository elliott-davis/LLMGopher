package api

import (
	"testing"
)

func TestDeriveOutcome(t *testing.T) {
	tests := []struct {
		statusCode   int
		errorMessage string
		want         string
	}{
		{200, "", "success"},
		{201, "", "success"},
		{302, "", "success"},
		{399, "", "success"},
		{401, "", "unauthorized"},
		{403, "", "unauthorized"},
		{429, "budget exceeded for this key", "budget_denied"},
		{429, "BUDGET limit reached", "budget_denied"},
		{429, "rate limit exceeded", "rate_limited"},
		{429, "", "rate_limited"},
		{400, "", "client_error"},
		{404, "", "client_error"},
		{422, "", "client_error"},
		{499, "", "client_error"},
		{500, "", "failure"},
		{502, "", "failure"},
		{503, "", "failure"},
	}

	for _, tc := range tests {
		got := deriveOutcome(tc.statusCode, tc.errorMessage)
		if got != tc.want {
			t.Errorf("deriveOutcome(%d, %q) = %q, want %q", tc.statusCode, tc.errorMessage, got, tc.want)
		}
	}
}

func TestBuildReferenceSummary(t *testing.T) {
	refs := buildReferenceSummary("", "", "")
	if len(refs) != 3 {
		t.Fatalf("expected 3 refs for all-empty, got %d", len(refs))
	}

	refs = buildReferenceSummary("key-001", "gpt-4o", "openai")
	if len(refs) != 0 {
		t.Fatalf("expected 0 refs for intact row, got %d", len(refs))
	}

	refs = buildReferenceSummary("", "gpt-4o", "openai")
	if len(refs) != 1 || refs[0].Field != "actor_id" || refs[0].State != "missing" {
		t.Fatalf("unexpected refs: %+v", refs)
	}

	refs = buildReferenceSummary("key-001", "", "openai")
	if len(refs) != 1 || refs[0].Field != "model" {
		t.Fatalf("unexpected refs: %+v", refs)
	}

	refs = buildReferenceSummary("key-001", "gpt-4o", "")
	if len(refs) != 1 || refs[0].Field != "provider" {
		t.Fatalf("unexpected refs: %+v", refs)
	}
}
