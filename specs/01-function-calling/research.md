# Research: Function Calling / Tool Use

## Decision: Use OpenAI Tool Calling as the Canonical Public Contract

**Rationale**: The gateway exposes an OpenAI-compatible API surface, and current SDKs model agentic workflows around `tools`, `tool_choice`, `tool_calls`, and streaming tool call deltas.

**Alternatives considered**: Provider-specific tool schemas were rejected for the public contract because they would force clients to branch by provider.

## Decision: Translate Provider-Native Tool Use at Adapter Boundaries

**Rationale**: Anthropic and Gemini represent tool use differently, but the provider adapters already own request and response translation. Keeping translation there preserves `pkg/llm` as the provider-neutral domain layer.

**Alternatives considered**: Passing raw provider tool blocks through the canonical API was rejected because it would violate OpenAI compatibility and typed contract expectations.

## Decision: Count Serialized Tool Definitions Conservatively

**Rationale**: Tool definitions contribute to prompt size and therefore cost. A conservative serialized estimate prevents budget and audit undercounting.

**Alternatives considered**: Ignoring tools in token counts was rejected because it understates spend. Exact provider-specific counting can be revisited if needed.

## Decision: Treat Existing Plan Status as Implementation Evidence Only

**Rationale**: The original plan marks all acceptance criteria complete, but there is no converted artifact documenting a running-gateway or SDK smoke test.

**Alternatives considered**: Marking the Spec Kit feature as verified was rejected because it would blur implemented code with functional validation.
