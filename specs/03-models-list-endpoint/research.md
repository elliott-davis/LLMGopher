# Research: GET /v1/models Endpoint

## Decision: Return Configured Gateway Models Only

**Rationale**: The endpoint should list the aliases clients can actually use through the gateway. The state cache already contains that configured model list.

**Alternatives considered**: Dynamic upstream provider discovery was rejected because it adds latency, exposes provider-specific inventory, and may list models that are not configured or authorized in the gateway.

## Decision: Require Existing API Key Authentication

**Rationale**: Model availability can reveal operational configuration and should follow the same access policy as other `/v1` endpoints.

**Alternatives considered**: Public unauthenticated discovery was rejected because it broadens information exposure without clear user value.

## Decision: Return Empty List for Empty State

**Rationale**: An empty configuration is a valid deployment state and should not look like a server failure.

**Alternatives considered**: Returning an error for empty state was rejected because OpenAI-compatible clients expect a list response.

## Decision: Treat Existing Plan Status as Implementation Evidence Only

**Rationale**: The original plan marks all acceptance criteria complete, but SDK smoke verification is only noted as optional.

**Alternatives considered**: Marking the Spec Kit feature as verified was rejected until the SDK smoke test is documented.
