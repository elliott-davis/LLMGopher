# Feature Specification: Generic OpenAI-Compatible Provider

**Feature Branch**: `[11-generic-openai-provider]`  
**Created**: 2026-04-26  
**Status**: Draft - implemented, functional verification needed  
**Input**: Converted from `plans/11-generic-openai-provider.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Route to OpenAI-Compatible Providers (Priority: P1)

Gateway operators configure arbitrary OpenAI-compatible upstream providers.

**Why this priority**: Implement a provider that forwards requests to any OpenAI-compatible endpoint by configuring only a base_url and API key. This single implementation covers Groq, Mistral AI, Ollama, vLLM, LM Studio, Together AI, Fireworks AI, DeepSeek, Perplexity, and any other server implementing the OpenAI chat completions API.

**Independent Test**: Exercise the primary success path for Generic OpenAI-Compatible Provider through the gateway and verify the public behavior matches the converted acceptance criteria.

**Acceptance Scenarios**:

1. **Given** the gateway is configured for Generic OpenAI-Compatible Provider, **When** A provider with base_url: "https://api.groq.com/openai/v1" and a Groq API key successfully completes a chat request, **Then** the behavior is accepted for this feature.
2. **Given** the gateway is configured for Generic OpenAI-Compatible Provider, **When** An Ollama provider with no API key (auth_type: "none") works with a locally running Ollama server, **Then** the behavior is accepted for this feature.
3. **Given** the gateway is configured for Generic OpenAI-Compatible Provider, **When** Provider registered via DB (not static config) is picked up at startup and after state cache refresh, **Then** the behavior is accepted for this feature.

---

### User Story 2 - Handle Policy and Failure Cases (Priority: P2)

Operators and API clients receive predictable behavior when requests are invalid, providers fail, policy limits apply, or required configuration is absent.

**Why this priority**: Gateway features affect shared spend, availability, and provider credentials, so failure behavior must be explicit and safe.

**Independent Test**: Trigger the relevant negative paths from the converted plan and verify compatible errors, redaction, and audit context.

**Acceptance Scenarios**:

1. **Given** an invalid or unsupported request for Generic OpenAI-Compatible Provider, **When** the gateway rejects it, **Then** the response uses the expected public error contract.
2. **Given** required configuration is missing or disabled, **When** the feature is invoked, **Then** the gateway fails safely without leaking credentials or internal state.

---

### User Story 3 - Verify Operational Readiness (Priority: P3)

Gateway operators can test, observe, and roll out Generic OpenAI-Compatible Provider without losing request traceability or spend accountability.

**Why this priority**: A gateway feature is only production-ready when it can be verified, audited, and debugged under normal operating conditions.

**Independent Test**: Run the automated tests and quickstart smoke checks, then inspect logs, metrics, traces, or audit records relevant to this feature.

**Acceptance Scenarios**:

1. **Given** the feature is exercised successfully, **When** operators inspect runtime evidence, **Then** request IDs and audit or observability context are preserved.
2. **Given** the feature affects usage or provider calls, **When** a request completes, **Then** cost and audit behavior remains consistent with existing gateway rules.

### Edge Cases

- Embeddings support via the generic provider (add if needed; OpenAI compat servers often support /embeddings too, but keep this spec focused)
- Streaming format differences (some compat servers have minor SSE deviations - handle in a follow-up)
- Tool/function call translation (generic compat servers use the same format as OpenAI)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST deliver Generic OpenAI-Compatible Provider for the gateway behavior described in the converted roadmap plan.
- **FR-002**: The system MUST preserve OpenAI-compatible request, response, and error semantics where the feature touches public APIs.
- **FR-003**: The system MUST preserve existing authentication, rate limiting, guardrail, audit, and cost behavior unless the feature explicitly changes those policies.
- **FR-004**: The system MUST provide tests for success, failure, and compatibility behavior described by the acceptance criteria.
- **FR-005**: The system MUST satisfy this acceptance behavior: A provider with base_url: "https://api.groq.com/openai/v1" and a Groq API key successfully completes a chat request.
- **FR-006**: The system MUST satisfy this acceptance behavior: An Ollama provider with no API key (auth_type: "none") works with a locally running Ollama server.
- **FR-007**: The system MUST satisfy this acceptance behavior: Provider registered via DB (not static config) is picked up at startup and after state cache refresh.
- **FR-008**: The system MUST satisfy this acceptance behavior: Well-known base URL defaulting works for groq, mistral, together.
- **FR-009**: The system MUST satisfy this acceptance behavior: Model routing routes to the correct generic provider by provider name.
- **FR-010**: The system MUST satisfy this acceptance behavior: Unit test covers the provider with a mock HTTP server.

### Key Entities *(include if feature involves data)*

- **Gateway Feature Configuration**: Runtime settings, database records, or provider metadata that enable Generic OpenAI-Compatible Provider.
- **Client Request**: The user-facing request or admin operation that exercises the feature.
- **Gateway Decision**: The routing, policy, provider, cache, audit, or observability decision made by the gateway.
- **Verification Evidence**: Tests, smoke checks, audit records, metrics, or traces proving the behavior works.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The primary Generic OpenAI-Compatible Provider workflow succeeds through the gateway using the public API or admin API described by the converted plan.
- **SC-002**: Negative and unsupported cases return predictable compatible errors without leaking secrets.
- **SC-003**: Automated tests cover the converted acceptance criteria for success and failure paths.
- **SC-004**: Functional smoke verification is recorded before the feature is marked verified.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing clients that do not use Generic OpenAI-Compatible Provider continue to work without request or response contract changes.
- **CC-002**: Request IDs, structured logs, audit context, and redaction are preserved for runtime paths touched by the feature.
- **CC-003**: Any provider call, budget change, routing decision, cache event, or observability event introduced by the feature remains testable and auditable.
- **CC-004**: Configuration follows the established precedence and startup validation rules.

## Assumptions

- The original plan marks this feature complete; functional smoke verification is still required.
- Converted dependencies: none beyond existing gateway infrastructure.
- Out-of-scope boundaries from the original plan remain unchanged: Embeddings support via the generic provider (add if needed; OpenAI compat servers often support /embeddings too, but keep this spec focused); Streaming format differences (some compat servers have minor SSE deviations - handle in a follow-up); Tool/function call translation (generic compat servers use the same format as OpenAI)
- The detailed implementation notes from the Claude plan are preserved in `plan.md`, while this spec focuses on user-visible and operational outcomes.
