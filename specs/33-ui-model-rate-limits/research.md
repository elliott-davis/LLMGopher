# Research: UI Model Rate Limit Controls

## Decision: Treat `rate_limit_rps` as an existing backend contract

**Rationale**: `pkg/llm.ModelConfig` includes `RateLimitRPS int json:"rate_limit_rps"`, and `internal/api.HandleCreateModel`/`HandleUpdateModel` already validate non-negative values and persist them to `models.rate_limit_rps`. The UI gap is omission from TypeScript types, server action payloads, form controls, and inventory display.

**Alternatives considered**:
- Add a new backend endpoint for rate-limit-only updates. Rejected because full model update already owns `rate_limit_rps`, and adding another endpoint would duplicate validation and contract surface.
- Store a UI-only policy value. Rejected because the database/admin API is already the source of truth.

## Decision: Use requests per second, with `0` meaning no model-level limit

**Rationale**: Backend fields and validation use `rate_limit_rps`, and spec 09 defines `0` as no model-level limit. The UI should keep the same unit and semantics to avoid translation bugs.

**Alternatives considered**:
- Show requests per minute for readability. Rejected because it would require conversion and could drift from backend validation/error messages.
- Hide zero-valued limits as blank. Rejected because administrators need an explicit "no model-level limit" state.

## Decision: Validate non-negative input in the UI action before sending payloads

**Rationale**: Backend remains authoritative, but early validation gives clearer feedback and avoids avoidable failing requests. The action should still surface gateway errors so concurrent edits, provider/model validation, or backend failures keep form state and display the failure reason.

**Alternatives considered**:
- Rely only on `<input min={0}>`. Rejected because HTML constraints can be bypassed and server actions must validate submitted data.
- Rely only on backend validation. Rejected because the UI already validates context window and required fields in actions.

## Decision: Display model policy as a dedicated inventory column

**Rationale**: A dedicated column makes limited and unrestricted models scannable, satisfying the operator at-a-glance requirement without hiding policy behind row actions.

**Alternatives considered**:
- Put the value only inside edit forms. Rejected because operators would need to inspect rows one by one.
- Add a separate detail page. Rejected because existing model management is table/modal based and the feature scope is narrow.

## Decision: Keep implementation in the existing model management UI

**Rationale**: The current model workflow lives in `ui/src/app/(dashboard)/models/page.tsx`, `CreateModelModal`, `EditModelModal`, `ModelRowActions`, and `ui/src/lib/actions.ts`. Extending those modules preserves local patterns and avoids introducing a new model settings area.

**Alternatives considered**:
- Create a new "Rate Limits" dashboard section. Rejected because this feature is specifically per-model configuration and should sit with model configuration.
