# Implementation Plan: UI API Key Budget Controls

**Branch**: `[31-ui-key-budgets]` | **Date**: 2026-05-02 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/31-ui-key-budgets/spec.md`

## Summary

Surface the existing per-key budget API from spec 07 in the admin UI so operators can view, create, update, reset, and remove budgets without raw API calls. The implementation will extend the current API key management surface with budget status, a budget form/action flow, server actions that call the gateway budget endpoints, and tests for validation, auth failure, no-budget state, and refresh behavior.

## Technical Context

**Language/Version**: TypeScript with Next.js 15 / React 19 for `ui/`; Go 1.22+ for existing gateway admin API contracts  
**Primary Dependencies**: Existing Next.js server actions, shadcn/base UI components, `sonner` toasts, Vitest + Testing Library, gateway admin budget handlers in `internal/api`  
**Storage**: Existing PostgreSQL-backed `api_key_budgets` state through the gateway API; no new database schema planned  
**Testing**: `cd ui && npm test`, focused Vitest component/action tests, `cd ui && npm run lint`; Go budget API tests remain covered by `go test ./internal/api/...` if gateway auth behavior changes  
**Target Platform**: Admin UI served by the Docker Compose `ui` service and calling the `gateway` service over the internal Compose network  
**Project Type**: Web admin UI backed by an OpenAI-compatible API gateway  
**Performance Goals**: Preserve gateway request hot-path behavior; avoid unbounded client or server-side fan-out when loading budget state; budget mutations should refresh visible state within the existing gateway cache window  
**Constraints**: Budget routes currently require Bearer auth while existing key list routes do not; UI must not expose admin tokens to client components; validation must match gateway request rules; reset/remove actions require explicit confirmation  
**Scale/Scope**: One admin UI surface for key-scoped budgets, covering view, set/update, reset, and delete for API keys already visible on `/keys`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - aligns with existing LLMGopher spec 07 budget endpoints rather than introducing a separate UI-only contract.
- **High-throughput runtime**: PASS - feature is admin UI only and does not change proxy, streaming, routing, cost, or audit hot paths.
- **Typed contracts**: PASS - adds typed UI request/response models mirroring `upsertBudgetRequest` and `budgetResponse`; no generated upstream schema is available for this internal admin surface.
- **Routing reliability**: N/A - no provider routing, retry, fallback, load-balancing, cooldown, timeout, or health-aware routing behavior changes.
- **Multi-tenant spend governance**: PASS - improves key-scoped spend governance visibility and lifecycle control; team/org/per-model budgets remain out of scope per spec.
- **Observability**: PASS - preserves existing gateway audit/cost behavior and uses UI success/error feedback; no new request logging of secrets or budget tokens.
- **API capability UX parity**: PASS - directly closes the documented UI gap for the existing budget API in the `/keys` admin UI.
- **Security and config**: PASS - budget route auth is handled server-side by the UI using a server-only admin API key; no plaintext client API key or provider credential is displayed.
- **Test and lint discipline**: PASS - plan requires Vitest coverage for typed actions and UI flows plus lint; Go tests are required only if backend auth or handler behavior changes.
- **Linter-first enforcement**: PASS - no new repeatable lint rule is needed; existing TypeScript/ESLint and Go lint/vet checks cover the touched surfaces.

## Project Structure

### Documentation (this feature)

```text
specs/31-ui-key-budgets/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── ui_key_budget_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
ui/
├── src/app/(dashboard)/keys/
│   ├── page.tsx
│   └── page.test.tsx
├── src/components/
│   ├── APIKeyRowActions.tsx
│   ├── APIKeyBudgetModal.tsx
│   ├── APIKeyBudgetForm.tsx
│   └── APIKeyBudgetStatus.tsx
├── src/lib/
│   ├── actions.ts
│   ├── actions.test.ts
│   ├── budget.ts
│   ├── budget.test.ts
│   └── types.ts
├── README.md
└── package.json

docker-compose.yaml
```

**Structure Decision**: Keep budget management in the existing API key UI rather than adding a separate navigation route, because the budget API is scoped to one key and operators already manage key lifecycle actions from `/keys`.

## Phase 0: Research

Research is captured in `research.md`. All technical unknowns from this plan have been resolved:

- Reuse the existing gateway budget endpoints from spec 07.
- Treat a budget GET `404` as a valid "no budget configured" UI state.
- Add a server-only UI admin token configuration for protected budget endpoints.
- Keep the primary UI entry point on `/keys` using a modal/details flow and existing server action patterns.

## Phase 1: Design & Contracts

Design artifacts:

- `data-model.md` defines UI-facing budget state, form input, status, and lifecycle action entities.
- `contracts/ui_key_budget_contract.md` documents the UI to gateway budget API contract and auth handling.
- `quickstart.md` documents local setup and validation steps.

Post-design constitution check:

- **Upstream parity**: PASS - contract mirrors `GET|PUT|DELETE /v1/admin/keys/{id}/budget` and `POST /v1/admin/keys/{id}/budget/reset`.
- **High-throughput runtime**: PASS - no gateway hot path changes; budget state is fetched only for admin UI needs.
- **Typed contracts**: PASS - UI types and parser helpers are planned for all request/response shapes.
- **Routing reliability**: N/A.
- **Multi-tenant spend governance**: PASS - key-scoped budget management is exposed with explicit reset/remove confirmations.
- **Observability**: PASS - existing gateway responses are surfaced through actionable UI errors without logging or exposing secrets.
- **API capability UX parity**: PASS - all spec 31 budget lifecycle capabilities have a UI design.
- **Security and config**: PASS - Bearer token remains server-only and Compose/local docs identify the dev key setup.
- **Test and lint discipline**: PASS - tests cover actions, validation, rendering, and mutation flows; lint remains mandatory for UI changes.
- **Linter-first enforcement**: PASS - no new linter rule required.

## Complexity Tracking

No constitution gate violations are planned. The only notable complexity is the server-side admin token used by the UI for budget endpoints; this is required because those gateway routes are already protected by auth while the current key listing UI is not.

## Verification Strategy

- `cd ui && npm test`
- `cd ui && npm run lint`
- Manual quickstart flow from `quickstart.md` against `make dev`
- If backend auth behavior changes during implementation: `go test ./internal/api/... -run 'Test.*Budget|TestAdminBudget' -v`
