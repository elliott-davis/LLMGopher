# Spec 02: Vision / Image Inputs

## Status
complete

## Goal
Enable image inputs in chat completion requests, allowing clients to send messages containing both text and image URLs (or base64-encoded images). Required for GPT-4o vision, Claude 3 vision, and Gemini multimodal.

## Background
`Message.Content` is currently a `string`. Spec 01 changes it to `json.RawMessage` to support both strings and content part arrays — this spec depends on that change being in place.

Each provider already handles structured content to some degree:
- `internal/proxy/provider_openai.go` — pass-through, needs no translation
- `internal/proxy/provider_anthropic.go` — already translates message structure; needs image block support
- `internal/providers/google/translate_gemini.go` — already translates content parts; Gemini SDK has `ImageData`/`FileData` types

## Requirements

### 1. Content part types (`pkg/llm/types.go`)

```go
type ContentPart struct {
    Type     string        `json:"type"` // "text" | "image_url"
    Text     string        `json:"text,omitempty"`
    ImageURL *ImageURLPart `json:"image_url,omitempty"`
}

type ImageURLPart struct {
    URL    string `json:"url"`    // https:// URL or data:image/...;base64,... URI
    Detail string `json:"detail,omitempty"` // "auto" | "low" | "high"
}
```

Add helper to `Message`:
```go
// ContentParts returns the content as a slice of ContentPart.
// If content is a plain string, returns a single text part.
// If content is a JSON array, unmarshals as []ContentPart.
func (m *Message) ContentParts() ([]ContentPart, error)

// ContentString returns content as a plain string.
// If content is an array, concatenates all text parts.
func (m *Message) ContentString() string
```

### 2. OpenAI provider
No changes needed. The request body is forwarded as-is and the new `json.RawMessage` content type serializes correctly.

### 3. Anthropic provider (`internal/proxy/provider_anthropic.go`)

When translating a message with array content:
- Text parts → `{type: "text", text: "..."}`
- Image URL parts with `https://` URL → `{type: "image", source: {type: "url", url: "..."}}`
- Image URL parts with `data:image/...;base64,...` → `{type: "image", source: {type: "base64", media_type: "image/jpeg", data: "<base64>"}}`

Extract `media_type` from the data URI prefix.

### 4. Gemini provider (`internal/providers/google/translate_gemini.go`)

When translating messages with array content:
- Text parts → `genai.Text("...")`
- Image URL parts with `https://` → `genai.FileData{URI: url, MIMEType: inferMIMEFromURL(url)}`
- Image URL parts with data URI → `genai.Blob{MIMEType: mediaType, Data: base64Decoded}`

### 5. Token counting (`internal/proxy/tokencount.go`)
Image tokens are provider-specific and complex to estimate. Use a fixed conservative estimate per image: 1000 tokens per image in the prompt. Log a warning that image token counts are approximate.

## Out of Scope
- Image generation endpoint (spec 25)
- Video inputs
- File uploads (base64 only, or URL — no multipart/form-data)
- Per-image `detail` field enforcement (pass through but don't adjust token estimates)

## Acceptance Criteria
- [x] A message with `content: [{type: "text", text: "..."}, {type: "image_url", image_url: {url: "..."}}]` is accepted and decoded correctly
- [x] A message with `content: "plain string"` still works (backward compat)
- [x] OpenAI provider forwards content array unchanged
- [x] Anthropic provider translates image URL to `{type: "image", source: {type: "url", ...}}`
- [x] Anthropic provider translates base64 data URI to `{type: "image", source: {type: "base64", ...}}`
- [x] Gemini provider translates content parts to SDK types
- [x] Unit tests cover text-only, image-only, and mixed content messages for each provider

## Key Files
- `pkg/llm/types.go` — `ContentPart`, `ImageURLPart`, helpers on `Message`
- `internal/proxy/provider_anthropic.go` — content part translation
- `internal/proxy/provider_anthropic_test.go` — tests
- `internal/providers/google/translate_gemini.go` — Gemini part translation
- `internal/proxy/tokencount.go` — image token estimate
