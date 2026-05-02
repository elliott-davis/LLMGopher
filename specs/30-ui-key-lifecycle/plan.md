# Implementation Plan: UI API Key Lifecycle Controls

**Branch**: `[030-ui-key-lifecycle]` | **Date**: 2026-05-02 | **Spec**: `spec.md`
**Input**: Feature specification from `specs/30-ui-key-lifecycle/spec.md`

## Summary

Close the UI/API gap for API key lifecycle management by extending the existing Next.js admin key inventory and create modal to use the backend key lifecycle fields already delivered by spec 05. The implementation will add lifecycle field visibility, create/edit/delete/reactivate/deactivate flows, model allowlist selection sourced from the model inventory, metadata editing, destructive confirmation, and post-action refresh feedback without changing gateway authentication or enforcement semantics.

## Technical Context

**Language/Version**: TypeScript 5, React 19, Next.js 15 App Router for `ui/`; Go 1.22+ backend contracts already exist  
**Primary Dependencies**: Next.js server actions, existing shadcn/base UI components, `sonner` toasts, gateway admin endpoints under `/v1/admin/keys` and `/v1/admin/models`  
**Storage**: PostgreSQL-backed gateway state via existing backend APIs; UI stores no secrets or lifecycle state beyond form state  
**Testing**: `npm run lint` and `npm run build` in `ui/`; focused UI component/server action tests if a test harness exists or is added; existing Go admin tests remain the backend contract guard  
**Target Platform**: Admin web UI served in the Docker Compose gateway environment  
**Project Type**: Web application frontend for an OpenAI-compatible Go API gateway  
**Performance Goals**: Key list and lifecycle actions remain operator-scale UI interactions; no request hot-path, streaming, routing, or async cost/audit path changes  
**Constraints**: Preserve one-time raw key display, never expose secret material after create, keep OpenAI-compatible backend error envelopes intact, reflect the 5-second state cache sync expectation after mutations  
**Scale/Scope**: One admin UI surface (`/keys`) plus shared UI action/type helpers for API key lifecycle parity

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - the UI surfaces existing LLMGopher admin key APIs from spec 05; no OpenAI client API behavior changes are introduced.
- **High-throughput runtime**: PASS - implementation is UI-only over admin endpoints and does not touch provider, middleware, streaming, routing, or async cost/audit hot paths.
- **Typed contracts**: PASS - TypeScript interfaces will mirror existing typed Go admin request/response shapes for keys and models.
- **Routing reliability**: N/A - no routing, retry, fallback, load-balancing, cooldown, timeout, or health-aware routing behavior changes.
- **Multi-tenant spend governance**: PASS - key-scoped active status, model allowlists, metadata, and rate limits remain governed by backend enforcement; UI makes those controls discoverable.
- **Observability**: PASS - backend audit/logging behavior remains unchanged; UI error feedback must preserve redaction and identify failed lifecycle actions.
- **API capability UX parity**: PASS - this feature exists specifically to expose existing key lifecycle APIs through `ui/src/app/(dashboard)/keys/page.tsx` and related components.
- **Security and config**: PASS - raw key material is only displayed immediately after creation; update/delete flows never render or log secrets and require no new runtime config.
- **Test and lint discipline**: PASS - plan requires UI lint/build plus focused tests or documented test harness gap; existing Go tests remain applicable to backend contracts.
- **Linter-first enforcement**: PASS - no new deterministic rule is needed beyond existing TypeScript/Next linting for this UI parity work.

## Project Structure

### Documentation (this feature)

```text
specs/30-ui-key-lifecycle/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── ui_key_lifecycle_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
ui/
├── package.json
└── src/
    ├── app/(dashboard)/keys/page.tsx
    ├── components/
    │   ├── CreateAPIKeyModal.tsx
    │   ├── EditAPIKeyModal.tsx
    │   ├── APIKeyRowActions.tsx
    │   └── ui/
    └── lib/
        ├── actions.ts
        └── types.ts

internal/api/
├── admin.go
├── admin_test.go
└── router.go
```

**Structure Decision**: Keep the feature in the existing `ui/` Next.js admin app and reuse the existing backend admin API handlers. Backend files are contract references and test anchors, not expected implementation targets unless the UI exposes a contract gap during development.

## Complexity Tracking

No constitution violations are expected. Any later implementation that changes backend enforcement, stores secret material in UI state beyond the create dialog, or adds blocking work to gateway request paths must document the exception here before implementation proceeds.

## Phase 0 Research

See `research.md`.

## Phase 1 Design

See `data-model.md`, `contracts/ui_key_lifecycle_contract.md`, and `quickstart.md`.

## Post-Design Constitution Check

- **Upstream parity**: PASS - contracts preserve existing `/v1/admin/keys` request/response semantics.
- **High-throughput runtime**: PASS - no hot-path Go runtime changes are planned.
- **Typed contracts**: PASS - UI types explicitly cover `expires_at`, `metadata`, `allowed_models`, and `is_active`.
- **Routing reliability**: N/A.
- **Multi-tenant spend governance**: PASS - model allowlists and metadata remain key-scoped controls.
- **Observability**: PASS - mutation failures surface action-specific messages without exposing secrets.
- **API capability UX parity**: PASS - all key lifecycle fields in the spec have planned UI controls; key rotation remains out of scope per spec assumptions.
- **Security and config**: PASS - one-time key display behavior is preserved and no new config is introduced.
- **Test and lint discipline**: PASS - verification includes `ui` lint/build and lifecycle smoke checks.
- **Linter-first enforcement**: PASS - no additional lint rule is warranted for this scoped UI change.

## Verification Strategy

- Run `npm run lint` and `npm run build` from `ui/`.
- Run focused UI tests if available after implementation; if a test harness is absent, document the gap and rely on lint/build plus quickstart smoke verification.
- Use `make dev` or the existing Docker Compose stack to smoke test create, edit, deactivate, reactivate, delete, gateway-unavailable, and validation-error flows.
- Run `go test ./internal/api/... -run 'Test.*APIKey' -v` only if backend contract changes become necessary.

## Implementation Verification Notes

- Added a Vitest, React Testing Library, jest-dom, and jsdom harness under `ui/` rather than relying only on lint/build.
- `npm test`, `npm run lint`, and `npm run build` passed after implementation.
- Backend key contracts were not changed, so the focused Go API key test command was not required.
