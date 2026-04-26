# Implementation Plan: Cohere Provider

**Branch**: `[13-cohere-provider]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/13-cohere-provider/spec.md`

## Summary

Add a Cohere provider supporting chat completions, embeddings, and reranking. Cohere's Command R+ is a strong RAG-optimized model and its embed-english-v3.0 embeddings are widely used; its rerank endpoint has no equivalent in other providers. The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: Existing gateway state where applicable  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Provider Expansion roadmap feature converted from `plans/13-cohere-provider.md`

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
specs/13-cohere-provider/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── cohere_provider_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
pkg/llm/
internal/proxy/
internal/providers/
pkg/config/
```

**Structure Decision**: Follow the existing gateway layer boundaries and keep public contracts, provider adapters, middleware, storage, and API handlers in their established packages.

## Converted Plan Details

### Background

Cohere has its own API format (not OpenAI-compatible), but offers clear documentation. The provider implements llm.Provider for chat and optionally llm.EmbeddingProvider for embeddings. The rerank endpoint will be handled as part of spec 25. Cohere API base URL: https://api.cohere.com/v2 Relevant endpoints: - POST /v2/chat - chat completions - POST /v2/embed - embeddings - POST /v2/rerank - rerank (out of scope for this spec)

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `internal/proxy/provider_cohere.go - new file`
- `pkg/config/config.go - Cohere config section`
- `cmd/gateway/main.go - register provider`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
