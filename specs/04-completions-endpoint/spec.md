# Feature Specification: POST /v1/completions Endpoint

**Feature Branch**: `[04-completions-endpoint]`  
**Created**: 2026-04-26  
**Status**: Draft - implemented, functional verification needed  
**Input**: Converted from `plans/04-completions-endpoint.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Legacy Text Completions (Priority: P1)

An API client calls the OpenAI-compatible legacy completions endpoint with a single string prompt and receives a text completion response.

**Why this priority**: Older SDK integrations, eval frameworks, and LangChain legacy chains still depend on this endpoint.

**Independent Test**: Send `POST /v1/completions` with a string prompt and verify the response has the `text_completion` object shape.

**Acceptance Scenarios**:

1. **Given** a valid API key and configured model, **When** a client sends a string prompt, **Then** the gateway returns a valid completions response.
2. **Given** a provider returns chat-style content internally, **When** the gateway responds to the client, **Then** the first choice text contains the generated completion text.

---

### User Story 2 - Stream Legacy Text Completions (Priority: P2)

An API client requests streaming completions and receives server-sent events in the completions chunk format.

**Why this priority**: Streaming compatibility is required for clients that render completions incrementally.

**Independent Test**: Send a streaming completions request and verify each event uses the completions object shape and text delta field.

**Acceptance Scenarios**:

1. **Given** `stream` is true, **When** the provider emits chat deltas, **Then** the gateway maps content deltas to completions `text` chunks.
2. **Given** a streaming response completes, **When** the gateway ends the stream, **Then** clients receive the expected terminal event semantics.

---

### User Story 3 - Reject Unsupported Prompt Shapes (Priority: P3)

An API client sends an unsupported completions prompt shape and receives a clear request error.

**Why this priority**: The first implementation supports single string prompts only and must fail unsupported inputs predictably.

**Independent Test**: Send a prompt array and verify the gateway returns HTTP 400 with an OpenAI-compatible error.

**Acceptance Scenarios**:

1. **Given** a prompt array, **When** a client calls `POST /v1/completions`, **Then** the gateway rejects it with a clear invalid request error.
2. **Given** compatibility-only parameters such as `echo` or `logprobs`, **When** the gateway decodes the request, **Then** unsupported behavior is ignored without breaking the request.

### Edge Cases

- Prompt arrays are out of scope and must not be silently coerced.
- `echo`, `logprobs`, `best_of`, and `suffix` are accepted only for request compatibility and are not implemented in this feature.
- Direct provider completions APIs are out of scope; the gateway may translate to its existing chat completion dispatch path.
- Cost and audit behavior should match chat completions even when the external endpoint is legacy completions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose `POST /v1/completions` using the same authentication and policy middleware as chat completions.
- **FR-002**: The system MUST accept a single string prompt and reject prompt arrays with a clear request error.
- **FR-003**: The system MUST translate a valid completions request into an equivalent single-turn chat request for provider dispatch.
- **FR-004**: The system MUST translate provider chat responses back into OpenAI-compatible completions responses.
- **FR-005**: The system MUST set the completions response `object` field to `text_completion`.
- **FR-006**: The system MUST support streaming completions responses using server-sent events.
- **FR-007**: The system MUST preserve audit logging and cost tracking behavior for completions requests.
- **FR-008**: The system MUST allow guardrail checks to evaluate the prompt as user content.
- **FR-009**: The system MUST remain aligned with OpenAI legacy completions behavior for the supported request and response subset.
- **FR-010**: The system MUST preserve typed Go contracts for externally visible completions request and response behavior.

### Key Entities

- **Completion Request**: A legacy OpenAI-compatible request containing model, string prompt, generation options, and streaming preference.
- **Completion Response**: A legacy OpenAI-compatible response containing generated text choices and usage.
- **Completion Stream Chunk**: A server-sent event chunk containing incremental generated text.
- **Prompt**: The user-provided single string converted into a chat user message for provider dispatch.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `POST /v1/completions` with a string prompt returns HTTP 200 and a valid completions response.
- **SC-002**: Streaming completions responses can be parsed by OpenAI-compatible clients.
- **SC-003**: Prompt arrays return HTTP 400 with a clear OpenAI-compatible error message.
- **SC-004**: Cost and audit records are produced for successful completions requests.

### Compatibility & Operational Criteria

- **CC-001**: OpenAI SDK `completions.create()` can call the gateway for the supported prompt subset.
- **CC-002**: The completions endpoint reuses existing chat routing, authentication, rate limit, guardrail, audit, and cost worker behavior.
- **CC-003**: Unsupported compatibility parameters do not change provider behavior or create misleading response data.
- **CC-004**: Errors use the established OpenAI-compatible error envelope.

## Assumptions

- The first conversion records the feature as implemented because the original plan marks all acceptance criteria complete and includes implementation notes.
- Functional verification still needs a running-gateway smoke test or SDK-level completions call.
- Single string prompts cover the initial compatibility need; prompt arrays and advanced legacy completions features can be added later if demand appears.
- Provider-native completions APIs are not required because chat translation provides a consistent cross-provider behavior.
