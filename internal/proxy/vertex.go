package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ed007183/llmgopher/pkg/llm"
	"golang.org/x/oauth2"
)

const (
	vertexModelPrefix   = "vertex/"
	vertexDefaultPublisher = "google" // Vertex openapi endpoint expects publisher/model (e.g. google/gemini-2.5-flash)
)

// VertexProvider forwards OpenAI-compatible payloads to Vertex AI's
// OpenAI-compatible endpoint without payload translation.
type VertexProvider struct {
	projectID   string
	region      string
	endpointURL string
	tokenSource oauth2.TokenSource
	client      *http.Client
}

func NewVertexProvider(projectID, region string, tokenSource oauth2.TokenSource) *VertexProvider {
	if region == "" {
		region = "us-central1"
	}

	endpointURL := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1beta1/projects/%s/locations/%s/endpoints/openapi/chat/completions",
		region,
		projectID,
		region,
	)

	return &VertexProvider{
		projectID:   projectID,
		region:      region,
		endpointURL: endpointURL,
		tokenSource: tokenSource,
		client:      &http.Client{},
	}
}

func (p *VertexProvider) Name() string { return "vertex" }

func (p *VertexProvider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	reqCopy := p.sanitizeRequestModel(req, false)

	body, err := json.Marshal(reqCopy)
	if err != nil {
		return nil, fmt.Errorf("marshal vertex request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create vertex request: %w", err)
	}
	if err := p.setHeaders(httpReq); err != nil {
		return nil, err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("vertex request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readUpstreamError(resp)
	}

	var result llm.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode vertex response: %w", err)
	}
	return &result, nil
}

func (p *VertexProvider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error) {
	reqCopy := p.sanitizeRequestModel(req, true)

	body, err := json.Marshal(reqCopy)
	if err != nil {
		return nil, fmt.Errorf("marshal vertex stream request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create vertex stream request: %w", err)
	}
	if err := p.setHeaders(httpReq); err != nil {
		return nil, err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("vertex stream request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, readUpstreamError(resp)
	}

	return resp.Body, nil
}

func (p *VertexProvider) sanitizeRequestModel(req *llm.ChatCompletionRequest, stream bool) *llm.ChatCompletionRequest {
	reqCopy := *req
	reqCopy.Stream = stream
	model := strings.TrimPrefix(req.Model, vertexModelPrefix)
	// Vertex openapi endpoint requires publisher/model (e.g. google/gemini-2.5-flash).
	if model != "" && !strings.Contains(model, "/") {
		model = vertexDefaultPublisher + "/" + model
	}
	reqCopy.Model = model
	return &reqCopy
}

func (p *VertexProvider) setHeaders(r *http.Request) error {
	token, err := p.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("fetch gcp access token: %w", err)
	}

	// Remove caller credentials and replace with GCP IAM bearer token.
	r.Header.Del("Authorization")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+token.AccessToken)
	return nil
}
