# Implementation Plan: Semantic Caching

**Branch**: `[27-semantic-caching]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/27-semantic-caching/spec.md`

## Summary

Cache LLM responses and retrieve them for semantically similar (not just identical) queries. Reduces cost and latency for use cases where users ask the same question in different phrasings. The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: PostgreSQL-backed gateway state and migrations where required  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Advanced Features roadmap feature converted from `plans/27-semantic-caching.md`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - converted plan identifies the OpenAI, LiteLLM, provider, admin, or operational behavior being matched.
- **High-throughput runtime**: PASS - hot-path impact must stay bounded and async cost/audit behavior must remain asynchronous unless explicitly required for policy decisions.
- **Typed contracts**: PASS - public request, response, provider, storage, and policy contracts should use typed Go structures or documented generated contracts.
- **Routing reliability**: PASS - routing, retry, fallback, load balancing, and provider selection behavior are central to this feature.
- **Multi-tenant spend governance**: PASS - existing key, budget, usage, and audit behavior must remain intact.
- **Observability**: PASS - request IDs, structured logs, audit context, metrics, traces, callbacks, or smoke evidence must cover meaningful runtime paths.
- **Security and config**: PASS - credentials remain encrypted, hashed, and redacted; config follows established precedence and startup validation.
- **Test discipline**: PASS - success, failure, compatibility, and smoke verification are required before verified status.

## Project Structure

### Documentation (this feature)

```text
specs/27-semantic-caching/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── semantic_caching_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
internal/api/
internal/proxy/
internal/middleware/
internal/storage/
pkg/llm/
```

**Structure Decision**: Follow the existing gateway layer boundaries and keep public contracts, provider adapters, middleware, storage, and API handlers in their established packages.

## Converted Plan Details

### Background

Spec 19 implements exact-match caching keyed on a SHA-256 hash of the request. Semantic caching requires embedding the request's messages, performing a similarity search against cached embeddings, and returning the cached response if similarity exceeds a threshold. Redis supports vector similarity search via the RediSearch module (Redis Stack). As an alternative, qdrant is a purpose-built vector database with a Go client.

### Dependencies

- Spec 19 (exact-match response caching) - ResponseCache interface and cache config foundation
- An embedding model must be configured in the gateway (via spec 11 generic provider or direct config)

### Key Files From Original Plan

- `pkg/llm/cache.go - extend with SemanticCache interface`
- `internal/storage/cache_semantic_redis.go - Redis Stack implementation (new file)`
- `internal/storage/cache_semantic_qdrant.go - Qdrant implementation (new file)`
- `internal/proxy/handler.go - semantic cache lookup/store`
- `pkg/config/config.go - semantic cache config`
- `cmd/gateway/main.go - init semantic cache if configured`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
