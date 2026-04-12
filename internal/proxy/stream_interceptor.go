package proxy

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"sync"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// StreamResult holds the accumulated metrics from an intercepted stream.
type StreamResult struct {
	OutputText   string
	OutputTokens int
	FinishReason string
}

// StreamInterceptor wraps a provider's SSE stream (io.ReadCloser) and
// intercepts each chunk to:
//   - Extract text deltas and accumulate the full output
//   - Forward SSE data to the client without adding latency
//   - Report accumulated results when the stream completes
//
// It implements io.ReadCloser so it can be used as a drop-in replacement
// for the raw provider stream.
type StreamInterceptor struct {
	src    io.ReadCloser
	model  string
	tc     *TokenCounter
	pr     *io.PipeReader
	pw     *io.PipeWriter
	once   sync.Once
	result StreamResult
	done   chan struct{}
}

// NewStreamInterceptor creates an interceptor that reads from src,
// extracts text deltas, and makes the SSE data available for reading.
func NewStreamInterceptor(src io.ReadCloser, model string, tc *TokenCounter) *StreamInterceptor {
	pr, pw := io.Pipe()
	si := &StreamInterceptor{
		src:   src,
		model: model,
		tc:    tc,
		pr:    pr,
		pw:    pw,
		done:  make(chan struct{}),
	}
	go si.process()
	return si
}

// Read returns intercepted SSE data for the client. Blocks until data is
// available or the stream ends.
func (si *StreamInterceptor) Read(p []byte) (int, error) {
	return si.pr.Read(p)
}

// Close closes the interceptor and the underlying source.
func (si *StreamInterceptor) Close() error {
	si.once.Do(func() {
		si.src.Close()
		si.pr.Close()
	})
	return nil
}

// Result blocks until the stream is fully consumed and returns the
// accumulated metrics. Safe to call from any goroutine.
func (si *StreamInterceptor) Result() StreamResult {
	<-si.done
	return si.result
}

// process reads SSE lines from the source, extracts text deltas,
// and forwards the raw SSE data to the pipe writer.
func (si *StreamInterceptor) process() {
	defer close(si.done)
	defer si.pw.Close()
	defer si.src.Close()

	var outputBuf strings.Builder
	scanner := bufio.NewScanner(si.src)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Forward every line (including blank lines that delimit SSE events).
		si.pw.Write([]byte(line + "\n"))

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			continue
		}

		// Try to parse as an OpenAI-compatible chunk to extract the text delta.
		var chunk llm.ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		for _, choice := range chunk.Choices {
			if choice.Delta != nil {
				if s := choice.Delta.ContentString(); s != "" {
					outputBuf.WriteString(s)
				}
			}
			if choice.FinishReason != "" {
				si.result.FinishReason = choice.FinishReason
			}
		}

		// If the final chunk includes usage from the provider, capture it.
		if chunk.Usage != nil && chunk.Usage.CompletionTokens > 0 {
			si.result.OutputTokens = chunk.Usage.CompletionTokens
		}
	}

	si.result.OutputText = outputBuf.String()

	// If the provider didn't report token counts, estimate them.
	if si.result.OutputTokens == 0 && si.result.OutputText != "" {
		si.result.OutputTokens = si.tc.CountTextTokens(si.model, si.result.OutputText)
	}
}
