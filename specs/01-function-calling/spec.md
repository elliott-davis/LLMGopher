# Feature Specification: Function Calling / Tool Use

**Feature Branch**: `[01-function-calling]`  
**Created**: 2026-04-26  
**Status**: Draft - implemented, functional verification needed  
**Input**: Converted from `plans/01-function-calling.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Submit Tool-Aware Chat Requests (Priority: P1)

An API client sends OpenAI-compatible chat completion requests that include tool definitions, tool choice instructions, assistant tool calls, and tool result messages.

**Why this priority**: Tool use is required for agentic workflows in current OpenAI, Anthropic, and Gemini model families.

**Independent Test**: Send a chat completion request with one function tool and verify the gateway accepts the OpenAI-compatible schema without breaking existing text-only messages.

**Acceptance Scenarios**:

1. **Given** a valid API key and a configured chat model, **When** the client sends `tools` and `tool_choice`, **Then** the request is accepted and routed without schema loss.
2. **Given** an assistant message containing a tool call, **When** the next request includes a tool result message, **Then** the gateway preserves the tool call identifier and content.

---

### User Story 2 - Receive Provider Tool Calls (Priority: P2)

An API client receives tool call responses from upstream providers in OpenAI-compatible response and streaming formats.

**Why this priority**: Clients expect a single response contract regardless of which provider serves the model.

**Independent Test**: Use provider test doubles or recorded fixtures to verify non-streaming and streaming tool call responses normalize to OpenAI-compatible `tool_calls`.

**Acceptance Scenarios**:

1. **Given** an upstream provider returns a tool-use response, **When** the gateway responds to the client, **Then** `finish_reason` is `tool_calls` and tool arguments are JSON strings.
2. **Given** an upstream provider streams partial tool arguments, **When** the gateway emits server-sent events, **Then** deltas preserve tool call index, id, name, and argument fragments.

---

### User Story 3 - Preserve Existing Chat Behavior (Priority: P3)

Existing clients continue sending plain text chat completion requests without changing request bodies or response handling.

**Why this priority**: Tool support must not regress the primary chat completion path.

**Independent Test**: Run existing chat completion tests and a text-only smoke request with no tool fields.

**Acceptance Scenarios**:

1. **Given** a request without tool fields, **When** it is proxied through any supported provider, **Then** the response shape remains unchanged.
2. **Given** a message with string content, **When** the gateway decodes and forwards it, **Then** it remains usable as a plain text message.

### Edge Cases

- Legacy `functions` and `function_call` fields are accepted for OpenAI compatibility but are not translated for Anthropic or Gemini.
- Tool result streaming from the client response side is outside the feature boundary.
- Invalid tool argument JSON from an upstream provider must not crash the gateway; it should return a compatible error or preserve the provider error semantics.
- Requests without tools must not pay a token-counting or translation penalty beyond normal chat handling.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST accept OpenAI-compatible tool definitions in chat completion requests.
- **FR-002**: The system MUST accept both string and object forms of tool choice where supported by the OpenAI-compatible schema.
- **FR-003**: The system MUST preserve assistant tool call identifiers, tool names, and JSON argument strings in canonical messages.
- **FR-004**: The system MUST accept tool result messages that reference the originating tool call.
- **FR-005**: The system MUST translate supported provider-native tool use responses into OpenAI-compatible `tool_calls`.
- **FR-006**: The system MUST emit OpenAI-compatible streaming deltas for provider tool call streams.
- **FR-007**: The system MUST preserve existing text-only chat completion behavior for clients that do not use tools.
- **FR-008**: The system MUST include tool definitions in prompt usage estimation with a conservative, documented approach.
- **FR-009**: The system MUST remain aligned with OpenAI chat completion request, response, and streaming behavior for tool use where provider capabilities allow it.
- **FR-010**: The system MUST preserve typed Go contracts for externally visible request and response behavior.

### Key Entities

- **Tool Definition**: A client-provided callable function description, including name, description, and JSON Schema parameters.
- **Tool Choice**: A client instruction that controls whether the model may, must, or must not call a tool.
- **Tool Call**: A model response that asks the client to invoke a named function with JSON arguments.
- **Tool Result Message**: A client message containing the result of a previous tool call.
- **Streaming Tool Delta**: A partial tool call response emitted during server-sent event streaming.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: OpenAI-compatible clients can complete a single tool call round trip through the gateway without client-side schema changes.
- **SC-002**: Existing text-only chat completion tests continue to pass with no request or response contract changes.
- **SC-003**: Streaming tool call responses preserve ordered argument fragments across the full response stream.
- **SC-004**: Tool-aware provider translation is covered by unit tests for OpenAI pass-through, Anthropic translation, and Gemini translation.

### Compatibility & Operational Criteria

- **CC-001**: Existing OpenAI SDK chat clients can send `tools` and read `tool_calls` using standard SDK types.
- **CC-002**: Provider-specific translation failures return OpenAI-compatible errors and do not expose secrets or raw credentials.
- **CC-003**: Token usage estimates include tool definitions so budget and audit records remain conservative.
- **CC-004**: Request IDs, structured logging, audit context, and async cost tracking behavior remain unchanged for the chat path.

## Assumptions

- The first conversion records the feature as implemented because the original plan marks all acceptance criteria complete.
- Functional verification still needs a running-gateway smoke test or SDK-level validation before the feature is considered verified.
- Anthropic and Gemini support enough provider-native function calling behavior to normalize common tool use workflows.
- Parallel tool call UI and response-side tool result streaming are intentionally out of scope.
