# Spec 12: AWS Bedrock Provider

## Status
completed

## Goal
Add an AWS Bedrock provider that exposes models hosted on Amazon Bedrock (Claude via Bedrock, Llama, Titan, Mistral, Command R, etc.) through the gateway's OpenAI-compatible API. Bedrock is a critical provider for customers using AWS infrastructure.

## Background
Unlike the other providers, Bedrock uses AWS SigV4 request signing instead of a Bearer token. It does not have a single base URL — requests go to `https://bedrock-runtime.{region}.amazonaws.com`.

AWS Bedrock supports a Converse API (`POST /model/{modelId}/converse` and `/converse-stream`) that provides a unified interface across all Bedrock models. This is the right abstraction to build on rather than per-model APIs.

Go AWS SDK v2: `github.com/aws/aws-sdk-go-v2`. This is not currently in `go.mod` and needs to be added.

## Requirements

### 1. Add AWS SDK dependency

```
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime
```

### 2. New provider (`internal/proxy/provider_bedrock.go`)

```go
type BedrockProvider struct {
    client *bedrockruntime.Client
    region string
}

func NewBedrockProvider(region, accessKeyID, secretAccessKey, sessionToken string) (*BedrockProvider, error)
func (p *BedrockProvider) Name() string // returns "bedrock"
func (p *BedrockProvider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error)
func (p *BedrockProvider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error)
```

If `accessKeyID` is empty, use the default AWS credential chain (env vars → `~/.aws/credentials` → EC2 instance metadata → ECS task role).

### 3. Converse API request translation

`llm.ChatCompletionRequest` → Bedrock `ConverseInput`:

- `Messages`: translate each `llm.Message` to `types.Message`
  - `role: "user"` → `types.ConversationRoleUser`
  - `role: "assistant"` → `types.ConversationRoleAssistant`
  - `role: "system"` → not a message; extract to `System: []types.SystemContentBlock`
  - content string → `types.ContentBlockMemberText{Value: content}`
- `model` → `ModelId` (Bedrock model ID format, e.g., `anthropic.claude-3-5-sonnet-20241022-v2:0`)
- `max_tokens` → `InferenceConfig.MaxTokens` (required by Bedrock; default to 4096 if nil)
- `temperature` → `InferenceConfig.Temperature`
- `top_p` → `InferenceConfig.TopP`
- `stop` → `InferenceConfig.StopSequences`
- `tools` → `ToolConfig.Tools` (Bedrock has a tool spec format; translate from `llm.Tool`)

### 4. Converse API response translation

`ConverseOutput` → `llm.ChatCompletionResponse`:
- `Output.Message.Content` → `choices[0].message.content`
- `StopReason: "end_turn"` → `finish_reason: "stop"`
- `StopReason: "tool_use"` → `finish_reason: "tool_calls"`
- `Usage.InputTokens` → `usage.prompt_tokens`
- `Usage.OutputTokens` → `usage.completion_tokens`

### 5. Streaming translation

Use `ConverseStreamInput` / `ConverseStreamOutput`. Translate the event stream to OpenAI SSE format:
- `ContentBlockDeltaEvent` with `DeltaMemberText` → chunk with `choices[0].delta.content`
- `ContentBlockDeltaEvent` with `DeltaMemberToolUse` → tool call delta
- `MessageStopEvent` → `[DONE]`
- Wrap the event stream as an `io.ReadCloser` that emits SSE-formatted bytes

### 6. Model ID mapping

Bedrock model IDs differ from common names. Add a mapping table:
```go
var BedrockModelAliases = map[string]string{
    "claude-3-5-sonnet":        "anthropic.claude-3-5-sonnet-20241022-v2:0",
    "claude-3-5-haiku":         "anthropic.claude-3-5-haiku-20241022-v1:0",
    "claude-3-opus":            "anthropic.claude-3-opus-20240229-v1:0",
    "llama-3-70b":              "meta.llama3-70b-instruct-v1:0",
    "mistral-large":            "mistral.mistral-large-2402-v1:0",
    "command-r-plus":           "cohere.command-r-plus-v1:0",
    "titan-text-express":       "amazon.titan-text-express-v1",
}
```

If the requested model name matches an alias, use the Bedrock model ID. Otherwise, pass the model name through as-is (allowing direct Bedrock model ID usage).

### 7. Provider configuration

Add to `pkg/config/config.go`:
```go
Bedrock struct {
    Region          string `mapstructure:"region"`
    AccessKeyID     string `mapstructure:"access_key_id"`
    SecretAccessKey string `mapstructure:"secret_access_key"`
}
```

Register the provider in `cmd/gateway/main.go` if `cfg.Bedrock.Region` is non-empty.

### 8. Credential storage

For DB-configured providers, store `access_key_id` and `secret_access_key` as encrypted credentials (same encryption as other provider credentials). The `auth_type` for Bedrock providers is `"aws_bedrock"`.

## Out of Scope
- Cross-region inference profiles
- Bedrock Agents
- Fine-tuning / model customization
- Bedrock Knowledge Bases (RAG)

## Acceptance Criteria
- [x] A configured Bedrock provider successfully routes a chat completion to Claude via Bedrock
- [x] Token counts and cost are recorded (Bedrock Converse API returns usage)
- [x] Streaming works via `ConverseStream`
- [x] Default AWS credential chain works (no explicit key required in config)
- [x] Model alias mapping resolves `claude-3-5-sonnet` to the correct Bedrock model ID
- [x] Tool calls translate correctly for Claude via Bedrock
- [x] Unit tests mock the Bedrock client interface

## Implementation Notes
- Added `internal/proxy/provider_bedrock.go` with Bedrock Converse and ConverseStream translation:
  - OpenAI request -> Bedrock messages/system/inference config/tool config translation
  - Bedrock response -> OpenAI choice/finish_reason/usage translation
  - Bedrock stream events -> OpenAI-compatible SSE chunks with tool-call deltas and `[DONE]`
- Added Bedrock model alias mapping with passthrough support for direct Bedrock model IDs.
- Added static Bedrock provider config wiring in `pkg/config/config.go` and registration in `cmd/gateway/main.go` when a region is configured.
- Extended dynamic provider sync to support DB providers with `auth_type = "aws_bedrock"` and encrypted JSON credentials (`access_key_id`, `secret_access_key`, optional `session_token`).
- Extended admin provider validation so Bedrock provider credentials are required and validated as JSON on create (and validated on update when provided).
- Extended provider credential loading query to include `aws_bedrock`.
- Added/updated tests for Bedrock provider translation and streaming, dynamic registration, model provider inference, credential loading query, and admin validation.

## Key Files
- `internal/proxy/provider_bedrock.go` — new file
- `pkg/config/config.go` — Bedrock config section
- `cmd/gateway/main.go` — register provider
- `go.mod` / `go.sum` — add AWS SDK dependency
