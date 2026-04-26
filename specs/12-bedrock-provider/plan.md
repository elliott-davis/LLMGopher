# Implementation Plan: AWS Bedrock Provider

**Branch**: `[12-bedrock-provider]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/12-bedrock-provider/spec.md`

## Summary

Add an AWS Bedrock provider that exposes models hosted on Amazon Bedrock (Claude via Bedrock, Llama, Titan, Mistral, Command R, etc.) through the gateway's OpenAI-compatible API. Bedrock is a critical provider for customers using AWS infrastructure. The original plan marks this feature complete; functional smoke verification is still required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: Existing gateway state where applicable  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Provider Expansion roadmap feature converted from `plans/12-bedrock-provider.md`

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
specs/12-bedrock-provider/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── bedrock_provider_contract.md
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

Unlike the other providers, Bedrock uses AWS SigV4 request signing instead of a Bearer token. It does not have a single base URL - requests go to https://bedrock-runtime.{region}.amazonaws.com. AWS Bedrock supports a Converse API (POST /model/{modelId}/converse and /converse-stream) that provides a unified interface across all Bedrock models. This is the right abstraction to build on rather than per-model APIs. Go AWS SDK v2: github.com/aws/aws-sdk-go-v2. This is not currently in go.mod and needs to be added.

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `internal/proxy/provider_bedrock.go - new file`
- `pkg/config/config.go - Bedrock config section`
- `cmd/gateway/main.go - register provider`
- `go.mod / go.sum - add AWS SDK dependency`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
