# Spec 01: Function Calling / Tool Use

## Status
complete

## Goal
Add function calling and tool use to the gateway's canonical request/response types and implement the necessary translation for each provider. This is the most critical missing feature — Claude 3.5+, GPT-4o, and Gemini 1.5+ all require it for agentic use cases.

## Background
`pkg/llm/types.go` defines the canonical types. `ChatCompletionRequest` has no `tools`, `tool_choice`, `functions`, or `function_call` fields. `Message` has no `tool_calls` or `tool_call_id` fields. `Choice` has no `tool_calls` field. `ChatCompletionChunk` delta has no streaming tool call support.

Each provider has a dedicated translation layer:
- `internal/proxy/provider_openai.go` — already OpenAI-compatible, needs no translation; currently passes the full struct as-is
- `internal/proxy/provider_anthropic.go` — translates to/from Anthropic Messages API; needs tool block translation
- `internal/providers/google/provider_gemini.go` and `translate_gemini.go` — translates to Gemini `FunctionDeclaration` format

## Requirements

### 1. Canonical type additions (`pkg/llm/types.go`)

Add the following types:

```go
type Tool struct {
    Type     string   `json:"type"` // "function"
    Function Function `json:"function"`
}

type Function struct {
    Name        string          `json:"name"`
    Description string          `json:"description,omitempty"`
    Parameters  json.RawMessage `json:"parameters,omitempty"` // JSON Schema object
}

type ToolCall struct {
    ID       string       `json:"id"`
    Type     string       `json:"type"` // "function"
    Function FunctionCall `json:"function"`
}

type FunctionCall struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"` // JSON string
}

type ToolChoice struct {
    Type     string    `json:"type"`               // "none" | "auto" | "required" | "function"
    Function *struct { Name string `json:"name"` } `json:"function,omitempty"`
}
```

Extend `ChatCompletionRequest`:
- `Tools []Tool` `json:"tools,omitempty"`
- `ToolChoice json.RawMessage` `json:"tool_choice,omitempty"` — can be string or object
- `Functions []Function` `json:"functions,omitempty"` — legacy field
- `FunctionCall json.RawMessage` `json:"function_call,omitempty"` — legacy field

Extend `Message`:
- `ToolCalls []ToolCall` `json:"tool_calls,omitempty"`
- `ToolCallID string` `json:"tool_call_id,omitempty"` — for role="tool" messages
- Change `Content` from `string` to `json.RawMessage` — required because tool result messages may have structured content, and this is a prerequisite for spec 02 (vision). Provide a `ContentString() string` helper that returns the string value for backward compatibility.

Extend `Choice`:
- `Message` already uses `*Message`; the `ToolCalls` field on `Message` covers non-streaming
- Delta in streaming: `ToolCalls []ToolCallDelta` where `ToolCallDelta` mirrors OpenAI's streaming format (index, id, type, function.name/arguments partial)

### 2. OpenAI provider (`internal/proxy/provider_openai.go`)
No translation needed. The provider forwards the request body as-is. Verify the struct serializes correctly with the new fields present (omitempty ensures backward compatibility).

### 3. Anthropic provider (`internal/proxy/provider_anthropic.go`)

**Request translation:**
- `tools` → Anthropic `tools` array: `{name, description, input_schema: parameters}`
- `tool_choice` string `"auto"` → `{"type": "auto"}`, `"none"` → `{"type": "none"}`, `"required"` → `{"type": "any"}`, object with function name → `{"type": "tool", "name": "..."}`
- Message role `"tool"` → Anthropic role `"user"` with content `[{type: "tool_result", tool_use_id: tool_call_id, content: content}]`
- Assistant messages with `tool_calls` → content array with `{type: "tool_use", id, name, input: parsed_arguments}`

**Response translation:**
- Anthropic `stop_reason: "tool_use"` → `finish_reason: "tool_calls"`
- Anthropic content block `{type: "tool_use", id, name, input}` → `ToolCall{id, type: "function", function: {name, arguments: json.Marshal(input)}}`

**Streaming:**
- Anthropic `content_block_start` with `{type: "tool_use"}` → emit `tool_calls[n].id` and `tool_calls[n].function.name`
- Anthropic `content_block_delta` with `{type: "input_json_delta", partial_json}` → emit `tool_calls[n].function.arguments` partial

### 4. Gemini provider (`internal/providers/google/translate_gemini.go`)

**Request translation:**
- `tools` → Gemini `tools: [{functionDeclarations: [{name, description, parameters}]}]`
- `tool_choice` → Gemini `toolConfig.functionCallingConfig.mode` (`AUTO`, `NONE`, `ANY`, or `{mode: "ANY", allowedFunctionNames: [...]}`)

**Response translation:**
- Gemini `FunctionCall` part → `ToolCall`
- `finish_reason: STOP` with function call present → `finish_reason: "tool_calls"`

### 5. Token counting (`internal/proxy/tokencount.go`)
Include tool definitions in prompt token count. Count tool definition JSON as part of the system/prompt tokens (conservative estimate: serialize tools to JSON, count tokens).

### 6. Backward compatibility
All new fields use `omitempty`. Existing requests without tools must continue to work unchanged. The `Content` field change from `string` to `json.RawMessage` must not break existing JSON unmarshaling — test with both `"content": "hello"` and `"content": [...]`.

## Out of Scope
- Parallel tool calls UI in the control plane
- Tool result streaming (response-side)
- Legacy `functions`/`function_call` translation for Anthropic/Gemini (accept the fields, pass through for OpenAI only)

## Acceptance Criteria
- [x] `ChatCompletionRequest` serializes/deserializes with `tools` and `tool_choice` fields
- [x] OpenAI provider forwards tool calls unchanged
- [x] Anthropic provider correctly translates a single tool call round-trip (request + response)
- [x] Anthropic streaming emits partial tool call arguments correctly
- [x] Gemini provider translates tool definitions and responses
- [x] Requests without tools still work (backward compat)
- [x] `Message.Content` accepts both string and array JSON values
- [x] Unit tests cover OpenAI pass-through, Anthropic translation, and Gemini translation

## Key Files
- `pkg/llm/types.go` — add new types, extend existing structs
- `internal/proxy/provider_anthropic.go` — request/response translation
- `internal/proxy/provider_anthropic_test.go` — extend tests
- `internal/providers/google/translate_gemini.go` — Gemini translation
- `internal/providers/google/translate_vertex.go` — Vertex (same as Gemini)
- `internal/proxy/tokencount.go` — include tool definitions in prompt count
