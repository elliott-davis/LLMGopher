# Implementation Plan: Batch API

**Branch**: `[29-batch-api]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/29-batch-api/spec.md`

## Summary

Implement an OpenAI-compatible Batch API for asynchronous bulk processing of chat completion requests. Clients submit a batch of requests, receive a batch ID, poll for status, and retrieve results when complete. This enables high-throughput, cost-efficient processing of large workloads (evals, document processing, data enrichment). The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: PostgreSQL-backed gateway state and migrations where required  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Advanced Features roadmap feature converted from `plans/29-batch-api.md`

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
specs/29-batch-api/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── batch_api_contract.md
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

OpenAI's Batch API accepts a JSONL file of requests, processes them asynchronously, and returns a JSONL file of responses. Pricing is typically 50% lower. The gateway implementation processes batches using background workers against the configured providers.

### Dependencies

- Spec 01 (function calling) - batch requests may include tool calls
- Spec 06 (audit log query) - batch results stored alongside regular audit entries

### Key Files From Original Plan

- `internal/storage/migrations/00013_batch_api.sql - new migration`
- `internal/batch/worker.go - batch processing worker (new package)`
- `internal/api/handler_files.go - file CRUD handlers (new file)`
- `internal/api/handler_batches.go - batch CRUD handlers (new file)`
- `internal/api/router.go - new routes`
- `pkg/config/config.go - batch config`
- `cmd/gateway/main.go - start batch worker`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
