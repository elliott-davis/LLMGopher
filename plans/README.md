# LLMGopher — Implementation Plans

Spec-driven feature roadmap ordered by implementation complexity. Each spec is self-contained: it describes the goal, what exists to build on, functional requirements, explicit exclusions, acceptance criteria, and the key files to touch.

## How to use these specs

Each spec is designed to be handed directly to an LLM for implementation. The LLM should:
1. Read the spec
2. Read the referenced key files
3. Implement requirements in order
4. Verify against acceptance criteria

Specs within the same tier are mostly independent and can be implemented in any order. Cross-spec dependencies are noted in each file.

---

## Tier 1 — Core API completeness

These fill critical gaps in the gateway's OpenAI compatibility. Start here.

| # | Spec | Status | Notes |
|---|------|--------|-------|
| 01 | [Function Calling / Tool Use](01-function-calling.md) | complete | Highest priority. Needed for GPT-4o, Claude 3.5+, Gemini 1.5+ |
| 02 | [Vision / Image Inputs](02-vision-inputs.md) | pending | Depends on 01 (content part types) |
| 03 | [GET /v1/models Endpoint](03-models-list-endpoint.md) | pending | Required for SDK compatibility |
| 04 | [POST /v1/completions Endpoint](04-completions-endpoint.md) | pending | Legacy text completion support |

---

## Tier 2 — Admin & Operations

Expose data that already exists in the database but has no API surface.

| # | Spec | Status | Notes |
|---|------|--------|-------|
| 05 | [API Key Management Improvements](05-admin-key-management.md) | pending | Delete, expiry, metadata, model allowlist |
| 06 | [Audit Log Query API](06-admin-audit-api.md) | pending | Data exists in `audit_log` table |
| 07 | [Budget Management API](07-admin-budget-api.md) | pending | Data exists in `api_key_budgets` table |
| 08 | [Usage & Spend Summary API](08-admin-usage-api.md) | pending | Aggregate queries on `audit_log` |
| 09 | [Per-Model Rate Limits](09-per-model-rate-limits.md) | pending | Extends existing rate limiter |
| 10 | [Budget Lifecycle (Alerts & Resets)](10-budget-lifecycle.md) | pending | Soft alerts + periodic resets |

---

## Tier 3 — Provider Expansion

New provider implementations following the existing `llm.Provider` interface.

| # | Spec | Status | Notes |
|---|------|--------|-------|
| 11 | [Generic OpenAI-Compatible Provider](11-generic-openai-provider.md) | pending | Covers Ollama, Groq, Mistral, vLLM via base_url config |
| 12 | [AWS Bedrock Provider](12-bedrock-provider.md) | pending | Requires AWS SDK v2, SigV4 signing |
| 13 | [Cohere Provider](13-cohere-provider.md) | pending | Chat + embeddings + rerank |
| 14 | [Additional Compatible Providers](14-additional-providers.md) | pending | Together AI, Fireworks, DeepSeek, Perplexity |

---

## Tier 4 — Reliability

New logic in the proxy layer for resilient request handling.

| # | Spec | Status | Notes |
|---|------|--------|-------|
| 15 | [Provider Retry with Backoff](15-provider-retry.md) | pending | `sethvargo/go-retry` already in go.sum |
| 16 | [Provider Fallback Chains](16-provider-fallback.md) | pending | Depends on 15 |
| 17 | [Load Balancing Across Deployments](17-load-balancing.md) | pending | Depends on 11 (multiple deployments) |
| 18 | [Context-Window Fallback](18-context-window-fallback.md) | pending | Depends on 16 |

---

## Tier 5 — Performance

| # | Spec | Status | Notes |
|---|------|--------|-------|
| 19 | [Exact-Match Response Caching](19-response-caching.md) | pending | Redis already wired for rate limiting |

---

## Tier 6 — Observability

| # | Spec | Status | Notes |
|---|------|--------|-------|
| 20 | [Prometheus Metrics Endpoint](20-prometheus-metrics.md) | pending | No external infra needed |
| 21 | [OpenTelemetry Tracing](21-opentelemetry.md) | pending | OTEL already an indirect dep |
| 22 | [Observability Callbacks](22-observability-callbacks.md) | pending | Webhook + Langfuse/LangSmith/Helicone |

---

## Tier 7 — Multi-tenancy

| # | Spec | Status | Notes |
|---|------|--------|-------|
| 23 | [Teams & Organizations](23-teams-organizations.md) | pending | New DB tables required |
| 24 | [RBAC & JWT Authentication](24-rbac-jwt-auth.md) | pending | Depends on 23 |

---

## Tier 8 — Advanced Features

Significant new subsystems. Implement after tiers 1–6 are stable.

| # | Spec | Status | Notes |
|---|------|--------|-------|
| 25 | [Image, Audio & Rerank Endpoints](25-new-endpoints.md) | pending | New endpoint types |
| 26 | [Guardrail Integrations](26-guardrail-integrations.md) | pending | Presidio, Azure, Lakera + response-side filtering |
| 27 | [Semantic Caching](27-semantic-caching.md) | pending | Depends on 19, requires vector store |
| 28 | [Traffic Mirroring / Shadow Mode](28-traffic-mirroring.md) | pending | |
| 29 | [Batch API](29-batch-api.md) | pending | Largest single-feature lift |
