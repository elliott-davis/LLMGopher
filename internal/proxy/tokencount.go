package proxy

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/ed007183/llmgopher/pkg/llm"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

// TokenCounter estimates token counts for prompts and completions.
// It caches tiktoken encodings per model family to avoid repeated init cost.
type TokenCounter struct {
	mu     sync.RWMutex
	cache  map[string]*tiktoken.Tiktoken
	logger *slog.Logger
}

func NewTokenCounter(logger *slog.Logger) *TokenCounter {
	return &TokenCounter{
		cache:  make(map[string]*tiktoken.Tiktoken),
		logger: logger,
	}
}

// CountPromptTokens estimates the token count for the prompt messages.
func (tc *TokenCounter) CountPromptTokens(model string, messages []llm.Message) int {
	enc := tc.getEncoding(model)
	if enc == nil {
		return tc.estimateFallback(messages)
	}

	// OpenAI's token counting includes per-message overhead.
	// ~4 tokens per message for role/name framing, +2 for the reply priming.
	tokensPerMessage := 4
	total := 0
	for _, m := range messages {
		total += tokensPerMessage
		total += len(enc.Encode(m.Content, nil, nil))
		total += len(enc.Encode(m.Role, nil, nil))
		if m.Name != "" {
			total += len(enc.Encode(m.Name, nil, nil))
		}
	}
	total += 2 // reply priming
	return total
}

// CountTextTokens counts tokens in a plain text string.
func (tc *TokenCounter) CountTextTokens(model string, text string) int {
	enc := tc.getEncoding(model)
	if enc == nil {
		return len(text) / 4 // rough heuristic: ~4 chars per token
	}
	return len(enc.Encode(text, nil, nil))
}

func (tc *TokenCounter) getEncoding(model string) *tiktoken.Tiktoken {
	key := tc.encodingKey(model)

	tc.mu.RLock()
	if enc, ok := tc.cache[key]; ok {
		tc.mu.RUnlock()
		return enc
	}
	tc.mu.RUnlock()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Double-check after acquiring write lock.
	if enc, ok := tc.cache[key]; ok {
		return enc
	}

	enc, err := tiktoken.EncodingForModel(key)
	if err != nil {
		tc.logger.Debug("tiktoken encoding not found, trying cl100k_base",
			"model", model, "key", key, "error", err,
		)
		enc, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			tc.logger.Warn("failed to load any tiktoken encoding", "error", err)
			return nil
		}
	}

	tc.cache[key] = enc
	return enc
}

// encodingKey normalizes model names to tiktoken-compatible keys.
func (tc *TokenCounter) encodingKey(model string) string {
	m := strings.ToLower(model)
	switch {
	case strings.HasPrefix(m, "gpt-4"):
		return "gpt-4"
	case strings.HasPrefix(m, "gpt-3.5"):
		return "gpt-3.5-turbo"
	case strings.HasPrefix(m, "claude"):
		return "cl100k_base" // Anthropic models use a similar BPE vocabulary
	case strings.HasPrefix(m, "o1"), strings.HasPrefix(m, "o3"):
		return "gpt-4"
	default:
		return model
	}
}

func (tc *TokenCounter) estimateFallback(messages []llm.Message) int {
	total := 0
	for _, m := range messages {
		total += len(m.Content)/4 + 4
	}
	return total + 2
}
