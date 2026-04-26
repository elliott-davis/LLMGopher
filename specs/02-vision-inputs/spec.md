# Feature Specification: Vision / Image Inputs

**Feature Branch**: `[02-vision-inputs]`  
**Created**: 2026-04-26  
**Status**: Draft - implemented, functional verification needed  
**Input**: Converted from `plans/02-vision-inputs.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Submit Mixed Text and Image Chat Messages (Priority: P1)

An API client sends a chat completion message containing text plus image URL content parts using the OpenAI-compatible content array schema.

**Why this priority**: Modern multimodal models require mixed content messages for GPT-4o, Claude 3, and Gemini vision workflows.

**Independent Test**: Send a chat completion request with one text part and one image URL part and verify the gateway accepts and routes the message.

**Acceptance Scenarios**:

1. **Given** a valid API key and a configured vision-capable model, **When** the client sends text and image URL content parts, **Then** the gateway accepts the request and preserves both parts.
2. **Given** an image URL uses HTTPS, **When** the request is translated for a provider, **Then** the provider receives an image reference in its expected format.

---

### User Story 2 - Submit Base64 Image Data (Priority: P2)

An API client sends base64-encoded image data in an OpenAI-compatible data URI content part.

**Why this priority**: Some clients cannot host images publicly and need inline image payload support.

**Independent Test**: Send a chat completion request with a `data:image/...;base64,...` image URL and verify provider translation extracts media type and image bytes.

**Acceptance Scenarios**:

1. **Given** a supported image data URI, **When** the gateway translates the message, **Then** the provider receives the media type and decoded image content.
2. **Given** a mixed message with text and base64 image content, **When** the gateway estimates prompt usage, **Then** it applies a conservative image token estimate.

---

### User Story 3 - Preserve Text-Only Chat Compatibility (Priority: P3)

Existing clients continue sending plain string message content without adopting content arrays.

**Why this priority**: Vision support depends on structured content but must not break the main chat path.

**Independent Test**: Send a text-only request and verify it decodes, routes, and responds as before.

**Acceptance Scenarios**:

1. **Given** a message with string content, **When** it is processed by the gateway, **Then** it remains equivalent to a single text content part.
2. **Given** a content array containing only text parts, **When** a plain text representation is needed, **Then** the text parts are concatenated in order.

### Edge Cases

- Unsupported video inputs are rejected or passed through only when a provider already supports them outside this feature boundary.
- Multipart file uploads are outside the feature boundary; clients must use image URLs or data URIs.
- The `detail` field is preserved where present but does not change token estimates in this feature.
- Invalid data URI media types or malformed base64 should fail with a clear OpenAI-compatible request error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST accept OpenAI-compatible content arrays containing text and image URL parts.
- **FR-002**: The system MUST continue accepting plain string message content.
- **FR-003**: The system MUST expose helper behavior that treats string content as a single text part.
- **FR-004**: The system MUST translate HTTPS image URL parts for providers that support remote image references.
- **FR-005**: The system MUST translate base64 data URI image parts for providers that require inline image data.
- **FR-006**: The system MUST infer or extract image media types when provider translation requires them.
- **FR-007**: The system MUST apply a conservative prompt token estimate for each image input.
- **FR-008**: The system MUST preserve OpenAI-compatible request shape for pass-through providers.
- **FR-009**: The system MUST remain aligned with OpenAI-compatible multimodal chat behavior for text and image content parts.
- **FR-010**: The system MUST preserve typed Go contracts for externally visible content part behavior.

### Key Entities

- **Content Part**: A typed segment of message content, such as text or image URL.
- **Image URL Part**: An image reference that may be an HTTPS URL or a base64 data URI, with optional detail preference.
- **Data URI Image**: Inline image content that includes media type and base64 data.
- **Vision-Capable Model**: A configured model that can receive image input through its provider.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Clients can send a mixed text plus image chat request using the OpenAI-compatible schema without client-side changes.
- **SC-002**: Text-only requests continue to pass existing chat completion tests.
- **SC-003**: Provider translation tests cover HTTPS image URLs, base64 image data, and mixed text/image messages.
- **SC-004**: Image-containing requests produce conservative usage estimates that are documented in audit or cost behavior.

### Compatibility & Operational Criteria

- **CC-001**: Existing OpenAI SDK clients can send multimodal chat content arrays against the gateway.
- **CC-002**: Anthropic and Gemini translations preserve image content without exposing raw payloads in logs.
- **CC-003**: Prompt token accounting is conservative for image requests so budget checks do not undercount usage.
- **CC-004**: Vision support reuses the chat completion request path without changing authentication, rate limiting, guardrail, audit, or cost worker behavior.

## Assumptions

- This feature depends on `01-function-calling` changing canonical message content to support structured JSON content.
- The first conversion records the feature as implemented because the original plan marks all acceptance criteria complete.
- Functional verification still needs a running-gateway smoke test against at least one vision-capable model or provider fixture.
- Image generation, video inputs, and multipart file upload workflows are deferred to later endpoint work.
