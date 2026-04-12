package proxy

import (
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestCountPromptTokens_AddsFixedEstimatePerImage(t *testing.T) {
	tc := NewTokenCounter(slog.New(slog.NewTextHandler(io.Discard, nil)))

	baseMessages := []llm.Message{
		{
			Role:    "user",
			Content: json.RawMessage(`[{"type":"text","text":"describe this image"}]`),
		},
	}
	withImageMessages := []llm.Message{
		{
			Role:    "user",
			Content: json.RawMessage(`[{"type":"text","text":"describe this image"},{"type":"image_url","image_url":{"url":"https://example.com/cat.png"}}]`),
		},
	}

	base := tc.CountPromptTokens("gpt-4o", baseMessages)
	withImage := tc.CountPromptTokens("gpt-4o", withImageMessages)

	if got, want := withImage-base, estimatedTokensPerPromptImage; got != want {
		t.Fatalf("image token delta = %d, want %d", got, want)
	}
}

func TestCountPromptTokens_AddsEstimateForEachImage(t *testing.T) {
	tc := NewTokenCounter(slog.New(slog.NewTextHandler(io.Discard, nil)))

	oneImage := []llm.Message{
		{
			Role:    "user",
			Content: json.RawMessage(`[{"type":"text","text":"two photos"},{"type":"image_url","image_url":{"url":"https://example.com/one.png"}}]`),
		},
	}
	twoImages := []llm.Message{
		{
			Role:    "user",
			Content: json.RawMessage(`[{"type":"text","text":"two photos"},{"type":"image_url","image_url":{"url":"https://example.com/one.png"}},{"type":"image_url","image_url":{"url":"https://example.com/two.png"}}]`),
		},
	}

	withOne := tc.CountPromptTokens("gpt-4o", oneImage)
	withTwo := tc.CountPromptTokens("gpt-4o", twoImages)

	if got, want := withTwo-withOne, estimatedTokensPerPromptImage; got != want {
		t.Fatalf("second image token delta = %d, want %d", got, want)
	}
}

func TestCountImagesInMessage_OnlyCountsImageURLParts(t *testing.T) {
	msg := llm.Message{
		Role: "user",
		Content: json.RawMessage(
			`[{"type":"text","text":"hello"},{"type":"image_url","image_url":{"url":"https://example.com/cat.png"}},{"type":"image_url","image_url":{"url":"data:image/png;base64,aGVsbG8="}}]`,
		),
	}

	if got, want := countImagesInMessage(msg), 2; got != want {
		t.Fatalf("countImagesInMessage = %d, want %d", got, want)
	}
}
