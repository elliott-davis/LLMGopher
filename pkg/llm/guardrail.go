package llm

import "context"

// GuardrailVerdict is the outcome of a guardrail check.
type GuardrailVerdict struct {
	Allowed bool
	Reason  string // human-readable explanation when blocked
}

// Guardrail validates prompts against safety policies before they reach the LLM.
// Implementations may call an external service (NeMo Guardrails, custom model, etc.).
type Guardrail interface {
	// Check evaluates the prompt content and returns a verdict.
	// Implementations must respect ctx deadlines to avoid blocking the hot path.
	Check(ctx context.Context, req *ChatCompletionRequest) (*GuardrailVerdict, error)
}
