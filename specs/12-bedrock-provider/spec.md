# Feature Specification: AWS Bedrock Provider

**Feature Branch**: `[12-bedrock-provider]`  
**Created**: 2026-04-26  
**Status**: Draft - implemented, functional verification needed  
**Input**: Converted from `plans/12-bedrock-provider.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Route to AWS Bedrock (Priority: P1)

Gateway operators route chat and embedding requests to AWS Bedrock models.

**Why this priority**: Add an AWS Bedrock provider that exposes models hosted on Amazon Bedrock (Claude via Bedrock, Llama, Titan, Mistral, Command R, etc.) through the gateway's OpenAI-compatible API. Bedrock is a critical provider for customers using AWS infrastructure.

**Independent Test**: Exercise the primary success path for AWS Bedrock Provider through the gateway and verify the public behavior matches the converted acceptance criteria.

**Acceptance Scenarios**:

1. **Given** the gateway is configured for AWS Bedrock Provider, **When** A configured Bedrock provider successfully routes a chat completion to Claude via Bedrock, **Then** the behavior is accepted for this feature.
2. **Given** the gateway is configured for AWS Bedrock Provider, **When** Token counts and cost are recorded (Bedrock Converse API returns usage), **Then** the behavior is accepted for this feature.
3. **Given** the gateway is configured for AWS Bedrock Provider, **When** Streaming works via ConverseStream, **Then** the behavior is accepted for this feature.

---

### User Story 2 - Handle Policy and Failure Cases (Priority: P2)

Operators and API clients receive predictable behavior when requests are invalid, providers fail, policy limits apply, or required configuration is absent.

**Why this priority**: Gateway features affect shared spend, availability, and provider credentials, so failure behavior must be explicit and safe.

**Independent Test**: Trigger the relevant negative paths from the converted plan and verify compatible errors, redaction, and audit context.

**Acceptance Scenarios**:

1. **Given** an invalid or unsupported request for AWS Bedrock Provider, **When** the gateway rejects it, **Then** the response uses the expected public error contract.
2. **Given** required configuration is missing or disabled, **When** the feature is invoked, **Then** the gateway fails safely without leaking credentials or internal state.

---

### User Story 3 - Verify Operational Readiness (Priority: P3)

Gateway operators can test, observe, and roll out AWS Bedrock Provider without losing request traceability or spend accountability.

**Why this priority**: A gateway feature is only production-ready when it can be verified, audited, and debugged under normal operating conditions.

**Independent Test**: Run the automated tests and quickstart smoke checks, then inspect logs, metrics, traces, or audit records relevant to this feature.

**Acceptance Scenarios**:

1. **Given** the feature is exercised successfully, **When** operators inspect runtime evidence, **Then** request IDs and audit or observability context are preserved.
2. **Given** the feature affects usage or provider calls, **When** a request completes, **Then** cost and audit behavior remains consistent with existing gateway rules.

### Edge Cases

- Cross-region inference profiles
- Bedrock Agents
- Fine-tuning / model customization
- Bedrock Knowledge Bases (RAG)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST deliver AWS Bedrock Provider for the gateway behavior described in the converted roadmap plan.
- **FR-002**: The system MUST preserve OpenAI-compatible request, response, and error semantics where the feature touches public APIs.
- **FR-003**: The system MUST preserve existing authentication, rate limiting, guardrail, audit, and cost behavior unless the feature explicitly changes those policies.
- **FR-004**: The system MUST provide tests for success, failure, and compatibility behavior described by the acceptance criteria.
- **FR-005**: The system MUST satisfy this acceptance behavior: A configured Bedrock provider successfully routes a chat completion to Claude via Bedrock.
- **FR-006**: The system MUST satisfy this acceptance behavior: Token counts and cost are recorded (Bedrock Converse API returns usage).
- **FR-007**: The system MUST satisfy this acceptance behavior: Streaming works via ConverseStream.
- **FR-008**: The system MUST satisfy this acceptance behavior: Default AWS credential chain works (no explicit key required in config).
- **FR-009**: The system MUST satisfy this acceptance behavior: Model alias mapping resolves claude-3-5-sonnet to the correct Bedrock model ID.
- **FR-010**: The system MUST satisfy this acceptance behavior: Tool calls translate correctly for Claude via Bedrock.

### Key Entities *(include if feature involves data)*

- **Gateway Feature Configuration**: Runtime settings, database records, or provider metadata that enable AWS Bedrock Provider.
- **Client Request**: The user-facing request or admin operation that exercises the feature.
- **Gateway Decision**: The routing, policy, provider, cache, audit, or observability decision made by the gateway.
- **Verification Evidence**: Tests, smoke checks, audit records, metrics, or traces proving the behavior works.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The primary AWS Bedrock Provider workflow succeeds through the gateway using the public API or admin API described by the converted plan.
- **SC-002**: Negative and unsupported cases return predictable compatible errors without leaking secrets.
- **SC-003**: Automated tests cover the converted acceptance criteria for success and failure paths.
- **SC-004**: Functional smoke verification is recorded before the feature is marked verified.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing clients that do not use AWS Bedrock Provider continue to work without request or response contract changes.
- **CC-002**: Request IDs, structured logs, audit context, and redaction are preserved for runtime paths touched by the feature.
- **CC-003**: Any provider call, budget change, routing decision, cache event, or observability event introduced by the feature remains testable and auditable.
- **CC-004**: Configuration follows the established precedence and startup validation rules.

## Assumptions

- The original plan marks this feature complete; functional smoke verification is still required.
- Converted dependencies: none beyond existing gateway infrastructure.
- Out-of-scope boundaries from the original plan remain unchanged: Cross-region inference profiles; Bedrock Agents; Fine-tuning / model customization
- The detailed implementation notes from the Claude plan are preserved in `plan.md`, while this spec focuses on user-visible and operational outcomes.
