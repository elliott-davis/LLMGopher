# Implementation Plan: Vision / Image Inputs

**Branch**: `[02-vision-inputs]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/02-vision-inputs/spec.md`

## Summary

Enable OpenAI-compatible multimodal chat messages with text and image URL content parts, including HTTPS image references and base64 data URIs. This feature depends on structured message content from `01-function-calling` and needs functional verification against a running gateway or provider fixture.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Standard library JSON and base64 handling, existing provider clients  
**Storage**: N/A  
**Testing**: `go test ./pkg/llm ./internal/proxy ./internal/providers/google/...` plus image request smoke tests  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Text-only requests must not regress; image requests use conservative usage estimation  
**Constraints**: Preserve redaction of image payloads in logs and maintain OpenAI-compatible errors for malformed content  
**Scale/Scope**: Chat completion message content for OpenAI pass-through, Anthropic translation, and Gemini/Vertex translation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - follows OpenAI-compatible multimodal chat content parts.
- **High-throughput runtime**: PASS - translation is request-local; no new synchronous storage or external observability calls.
- **Typed contracts**: PASS - content parts and image URL parts are represented as typed contracts.
- **Routing reliability**: N/A - no routing selection behavior changes.
- **Multi-tenant spend governance**: PASS - image prompt usage is conservatively estimated.
- **Observability**: PASS - logs and audit data must avoid leaking raw image payloads.
- **Security and config**: PASS - no new credentials or config secrets.
- **Test discipline**: PASS - requires text-only, image URL, base64 image, mixed content, and provider translation tests.

## Project Structure

### Documentation (this feature)

```text
specs/02-vision-inputs/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── openai-content-parts.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
pkg/llm/
└── types.go

internal/proxy/
├── provider_openai.go
├── provider_anthropic.go
├── provider_anthropic_test.go
└── tokencount.go

internal/providers/google/
└── translate_gemini.go
```

**Structure Decision**: Keep multimodal request contracts in `pkg/llm` and provider-specific image conversion inside provider adapters.

## Complexity Tracking

No constitution violations are expected.

## Verification Strategy

- Run unit tests for content helper behavior and provider translation.
- Verify plain string content still works.
- Run a functional smoke test with an HTTPS image URL.
- Run a functional or fixture-backed smoke test with a base64 data URI.
- Record smoke-test results before changing feature status from "functional verification needed" to verified.
