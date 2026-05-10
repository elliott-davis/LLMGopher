# Implementation Plan: UI Model Rate Limit Controls

**Branch**: `[33-ui-model-rate-limits]` | **Date**: 2026-05-03 | **Spec**: `specs/33-ui-model-rate-limits/spec.md`
**Input**: Feature specification from `specs/33-ui-model-rate-limits/spec.md`

## Summary

Expose the existing per-model `rate_limit_rps` admin API field in the Next.js model management UI so administrators can set, update, and inspect model-level request limits without raw API calls. The implementation is UI-focused: extend the TypeScript model contract, add non-negative rate limit parsing to model create/update server actions, add create/edit form controls and explanatory copy, display the model-level policy in the model inventory, and cover invalid input plus gateway failure behavior with focused Vitest tests.

## Technical Context

**Language/Version**: Go 1.22+ backend; TypeScript with Next.js 15, React 19, and server/client components in `ui/`  
**Primary Dependencies**: Existing Go `net/http` admin model routes, PostgreSQL-backed `models.rate_limit_rps`, Next.js App Router, React, existing local UI primitives, Sonner toasts  
**Storage**: Existing PostgreSQL `models.rate_limit_rps`; no new migrations or persistence planned  
**Testing**: Focused `vitest run` tests for `ui/src/lib/actions.ts`, model forms, row/table rendering, and failure states; `go test ./internal/api/... -run Model -v` only if backend contract drift is found  
**Target Platform**: Linux gateway plus Docker Compose admin UI service  
**Project Type**: Web service with companion Next.js admin UI  
**Performance Goals**: No gateway request hot-path changes; UI-only form/table additions should not add extra model list round trips beyond existing refresh behavior  
**Constraints**: Preserve existing model create/edit/delete/provider assignment behavior; never expose credentials or raw API keys; treat `0` as no model-level limit; reject negative values before or during save with clear feedback  
**Scale/Scope**: One model management surface in `ui/` covering create modal, edit modal, model inventory table, server actions, TypeScript types, and tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Gate

- **Upstream parity**: PASS. The UI exposes existing LLMGopher/LiteLLM-style model-level request throttling behavior from spec 09; no OpenAI-compatible client API changes.
- **High-throughput runtime**: PASS. No request hot-path, streaming path, proxy, middleware, goroutine, or async worker changes are planned.
- **Typed contracts**: PASS. Frontend `Model` types and action payloads will mirror the typed Go `ModelConfig`/admin JSON field `rate_limit_rps`.
- **Routing reliability**: PASS. The feature surfaces a routing/rate-limit policy field already consumed by the state cache and rate limiter without changing aliases, retries, fallback, or provider selection.
- **Multi-tenant spend governance**: PASS. Model-level throttling is an operator spend/governance control; the UI distinguishes it from API key rate limits.
- **Observability**: PASS. Existing gateway rate-limit errors and audit behavior remain unchanged; UI save failures surface gateway error messages without leaking secrets.
- **API capability UX parity**: PASS. The target `ui/` surface is `ui/src/app/(dashboard)/models`, `CreateModelModal`, `EditModelModal`, and model row/table components.
- **Security and config**: PASS. No credential/config changes; existing admin UI gateway access remains unchanged.
- **Test and lint discipline**: PASS. Add Vitest coverage for parsing, payloads, invalid input, form persistence on failure, and no-limit display; run UI tests/lint where available.
- **Linter-first enforcement**: PASS. No new deterministic Go lint rule is needed; existing TypeScript/ESLint/Vitest checks cover this UI parity change.

### Post-Design Gate

- **Upstream parity**: PASS. Contracts define only the existing admin model `rate_limit_rps` field and keep runtime model enforcement behavior unchanged.
- **High-throughput runtime**: PASS. Implementation remains in admin UI rendering/actions and existing cache refresh behavior.
- **Typed contracts**: PASS. `data-model.md` and `contracts/model-rate-limit-controls.md` define typed UI/admin payload shapes.
- **Routing reliability**: PASS. The UI treats `rate_limit_rps` as the source-of-truth model policy value that backend cache sync applies.
- **Multi-tenant spend governance**: PASS. Copy and display states clarify model limits are separate from key-level limits.
- **Observability**: PASS. Save failure handling keeps form state and displays backend reasons.
- **API capability UX parity**: PASS. Model create, edit, and inventory surfaces close the documented API/UI gap.
- **Security and config**: PASS. No secrets or credentials are introduced.
- **Test and lint discipline**: PASS. Quickstart lists targeted UI tests plus lint/build verification.
- **Linter-first enforcement**: PASS. No additional lint enforcement is introduced.

## Project Structure

### Documentation (this feature)

```text
specs/33-ui-model-rate-limits/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── model-rate-limit-controls.md
└── tasks.md              # Created later by /speckit-tasks
```

### Source Code (repository root)

```text
internal/
└── api/
    ├── admin.go          # Existing POST/PUT /v1/admin/models rate_limit_rps contract
    └── admin_test.go     # Backend tests only if contract drift is discovered

pkg/
└── llm/
    └── types.go          # Existing ModelConfig rate_limit_rps JSON shape

ui/
└── src/
    ├── app/(dashboard)/
    │   └── models/
    │       └── page.tsx
    ├── components/
    │   ├── CreateModelModal.tsx
    │   ├── EditModelModal.tsx
    │   ├── ModelRowActions.tsx
    │   └── model rate-limit focused tests as needed
    └── lib/
        ├── actions.ts
        ├── actions.test.ts
        └── types.ts
```

**Structure Decision**: Use the existing Next.js dashboard model management layout, local modal/action patterns, and backend admin model contract. Keep backend Go files read-only unless implementation reveals the existing admin API no longer returns or accepts `rate_limit_rps`.

## Complexity Tracking

No constitution violations or additional complexity exceptions are required.
