# Phase 1 Data Model — Mock Backend Fixtures

The mock backend is stateless on disk; it boots from a TypeScript seed module and keeps an in-memory store keyed by entity id. Reused TypeScript types live in `ui/src/lib/types.ts` — fixtures import those, never redefine them.

## Entities

### Provider
**Source type**: `Provider` from `ui/src/lib/types.ts`.

| Field | Notes |
|---|---|
| `id` | Stable test id, e.g. `prov_openai_prod`, `prov_anthropic_prod`, `prov_vertex_degraded`. |
| `kind` | `openai` \| `anthropic` \| `vertex` \| `bedrock` \| `cohere` \| `generic`. |
| `display_name` | Human label shown in `data-testid="provider-row"`. |
| `health` | `healthy` \| `degraded` \| `offline` — drives badge color in tests. |

**Seed defaults**: 3 providers — one healthy, one degraded, one offline.

---

### Model
**Source type**: `Model`.

| Field | Notes |
|---|---|
| `id` | e.g. `gpt-4o`, `claude-3.5-sonnet`. |
| `provider_id` | FK to Provider. |
| `enabled` | `true` for at least 4 of the seed set. |
| `rate_limit` | Optional; populated for at least one model to support rate-limit tests. |

**Seed defaults**: 6 models across the 3 providers.

---

### APIKey
**Source type**: `APIKey` + `APIKeyBudget`.

| Field | Notes |
|---|---|
| `id` | e.g. `key_checkout_service`, `key_research_team`, `key_near_cap`, `key_over_cap`. |
| `name` | Display name. |
| `prefix` | Last 4 chars only — never a real-looking key. |
| `is_active` | `true` for most; one inactive for lifecycle tests. |
| `budget` | One key at 86% utilization, one at 100% (over cap), one with no budget. |
| `model_allowlist` | One key restricted to a single model for the model-restriction test. |

**Computed invariants**:
- `key_near_cap.budget.usage_usd / key_near_cap.budget.limit_usd === 0.86` (drives "near cap" pill).
- `key_over_cap.budget.usage_usd >= key_over_cap.budget.limit_usd` (drives 429 contract test).

**Seed defaults**: 4 API keys.

---

### Team
| Field | Notes |
|---|---|
| `id` | `team_research`, `team_platform`. |
| `display_name` | Shown in `data-testid="team-{id}-row"`. |
| `member_count` | Integer; surfaces in the teams grid. |
| `budget_utilization` | Float 0–1. `team_research = 0.86` to drive `data-testid="team-research-warn"`. |

**Seed defaults**: 2 teams.

---

### Budget
Owned per-key (see APIKey) and per-team (see Team). No standalone entity; covered by the parent.

---

### RateLimitRule
| Field | Notes |
|---|---|
| `id` | `rl_chat_default`, `rl_tripped`. |
| `scope` | `key` \| `model` \| `team`. |
| `rps` | Integer. |
| `tripped` | `true` for exactly one rule (`rl_tripped`) so the "tripped" pill test has a deterministic target. |

**Seed defaults**: 3 rules.

---

### Guardrail
| Field | Notes |
|---|---|
| `id` | `gr_jail`, `gr_pii`, `gr_secrets`. |
| `enabled` | Mix: `gr_jail=false` (test toggles it on), `gr_pii=true`, `gr_secrets=false`. |

**Seed defaults**: 3 guardrails.

---

### LogRow
**Source type**: To be added (`LogRow` does not exist yet in `types.ts`; introduce alongside Story 2).

| Field | Notes |
|---|---|
| `id` | `log_001`–`log_020`. |
| `request_id` | Stable per-row, used for `ignoreRegion` masking in visual tests. |
| `status_code` | Mix: ~70% 2xx, ~10% 4xx, ~10% 5xx, ~10% fallback-handled. |
| `provider_chain` | Array of `{provider_id, status, latency_ms}` — at least one row has 2+ stages with `status="failed"` on the primary for the trace-tab test. |
| `timestamp` | All rows pinned to the frozen `now` minus deterministic offsets. |

**Seed defaults**: 20 log rows. One row has `id="log_fallback"` so `data-testid="log-row-fallback"` resolves.

---

### AuditEntry
**Source type**: `AuditRecord`.

| Field | Notes |
|---|---|
| `id` | `audit_001`–`audit_010`. |
| `actor` | Test user id. |
| `action` | Mix of `key.create`, `key.rotate`, `provider.add`, `budget.update`. |
| `at` | Frozen-now offsets. |

**Seed defaults**: 10 audit entries.

---

## State transitions

The mock backend supports these mutations (one entry per entity that has user-visible writes in the UI):

| Entity | Mutation | Effect |
|---|---|---|
| Provider | `POST /v1/admin/providers` | Append. New row appears in providers table. |
| APIKey | `POST /v1/admin/keys/{id}/rotate` | Rotates `prefix`; one-time reveal returned in response. |
| APIKey | `PATCH /v1/admin/keys/{id}` | Updates editable fields (name, rate limit, expiration, allowlist, is_active). |
| APIKey | `DELETE /v1/admin/keys/{id}/budget` (alias `/budget/reset`) | Resets `usage_usd` to 0. |
| Guardrail | `PATCH /v1/admin/guardrails/{id}` | Toggles `enabled`. |
| RateLimitRule | `POST /v1/admin/rate-limits` | Append. |

All other endpoints are read-only.

## Reset semantics

- The mock state resets between Playwright **test files** (not per-test) by default — restoring the seed snapshot. Tests that need cross-test isolation use `test.describe.serial` with explicit `beforeEach(reset)`.
- Reset is implemented as a structuredClone of the seed snapshot; no DB, no async.
