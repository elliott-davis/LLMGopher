# Feature Specification: GET /v1/models Endpoint

**Feature Branch**: `[03-models-list-endpoint]`  
**Created**: 2026-04-26  
**Status**: Draft - implemented, functional verification needed  
**Input**: Converted from `plans/03-models-list-endpoint.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List Available Models with OpenAI SDKs (Priority: P1)

An API client calls the OpenAI-compatible models list endpoint to discover configured model aliases before sending inference requests.

**Why this priority**: Many SDKs and tools call the models endpoint during setup and fail if it is missing.

**Independent Test**: Call `GET /v1/models` with a valid API key and verify the response follows the OpenAI list schema.

**Acceptance Scenarios**:

1. **Given** active models are configured, **When** a client calls `GET /v1/models`, **Then** the response contains a list object with model entries.
2. **Given** a model alias can be used for chat completions, **When** the models endpoint returns it, **Then** the entry `id` matches that alias.

---

### User Story 2 - Handle Empty Model Configuration (Priority: P2)

An authenticated client calls the models endpoint when no active models are configured.

**Why this priority**: Empty deployments should be discoverable without presenting a server error.

**Independent Test**: Run the handler against an empty model cache and verify it returns an empty `data` array.

**Acceptance Scenarios**:

1. **Given** no models are configured, **When** a client calls `GET /v1/models`, **Then** the gateway returns HTTP 200 with an empty `data` array.
2. **Given** the state cache is available but contains no active models, **When** the endpoint responds, **Then** it does not expose internal cache details.

---

### User Story 3 - Enforce Existing Authentication (Priority: P3)

Unauthenticated clients cannot use the models endpoint to discover gateway configuration.

**Why this priority**: Model availability is tenant-sensitive operational information.

**Independent Test**: Call `GET /v1/models` without a valid API key and verify existing authentication middleware denies the request.

**Acceptance Scenarios**:

1. **Given** a missing API key, **When** a client calls `GET /v1/models`, **Then** the gateway returns the standard authentication error.
2. **Given** an invalid API key, **When** a client calls `GET /v1/models`, **Then** no model data is returned.

### Edge Cases

- Single model detail lookup is out of scope for this feature.
- Dynamic upstream provider model discovery is out of scope; only gateway-configured models are listed.
- Provider or capability filtering is out of scope.
- Empty or nil model state returns an empty list instead of an internal server error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose an authenticated `GET /v1/models` endpoint.
- **FR-002**: The system MUST return an OpenAI-compatible list response with `object` set to `list`.
- **FR-003**: The system MUST include one model entry per active configured model alias.
- **FR-004**: Each model entry MUST include `id`, `object`, `created`, and `owned_by`.
- **FR-005**: The model entry `id` MUST match the alias clients use for inference requests.
- **FR-006**: The system MUST return an empty `data` array when no models are configured.
- **FR-007**: The endpoint MUST use the existing API key authentication behavior for `/v1` endpoints.
- **FR-008**: The endpoint MUST not call upstream providers to discover models.
- **FR-009**: The system MUST remain aligned with OpenAI models list response behavior for SDK compatibility.
- **FR-010**: The system MUST preserve typed Go contracts or explicit response structs for the models list response.

### Key Entities

- **Model List Response**: An OpenAI-compatible list wrapper containing model entries.
- **Model Entry**: A configured model alias exposed to clients, with creation timestamp and owner metadata.
- **Configured Model Alias**: The client-facing model identifier used in inference requests.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `GET /v1/models` returns HTTP 200 and an OpenAI-compatible list object for authenticated clients.
- **SC-002**: Every returned `id` can be used as a model value for a compatible inference endpoint.
- **SC-003**: Empty model configuration returns HTTP 200 with `data: []`.
- **SC-004**: Authentication failures return the existing OpenAI-compatible authentication error shape.

### Compatibility & Operational Criteria

- **CC-001**: OpenAI SDK `models.list()` can parse the gateway response without client changes.
- **CC-002**: The endpoint reads gateway configuration state only and does not add provider discovery latency.
- **CC-003**: The endpoint preserves request IDs and structured logging behavior from the existing middleware chain.
- **CC-004**: No provider credentials, internal deployment names, or inactive model records are exposed.

## Assumptions

- The gateway state cache is the source of truth for configured active model aliases.
- The first conversion records the feature as implemented because the original plan marks all acceptance criteria complete.
- Functional verification still needs a running-gateway smoke test or SDK-level `models.list()` call.
- The owner value can remain the project-level owner string unless a later multi-tenant feature changes ownership semantics.
