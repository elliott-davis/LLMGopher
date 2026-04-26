<!--
SYNC IMPACT REPORT
==================
Version change: 1.0.0 -> 1.1.0
New version: 1.1.0

Modified principles:
- None

Added sections:
- X. API Capability UX Parity

Removed sections:
- None

Templates requiring updates:
- .specify/templates/plan-template.md ✅ updated with API capability UX parity gate
- .specify/templates/spec-template.md ✅ updated with UI surfacing requirements and success criteria
- .specify/templates/tasks-template.md ✅ updated with UI surface and UI test task guidance

Follow-up TODOs: None.
-->

# LLMGopher Constitution

## Core Principles

### I. Upstream API & Behavioral Parity (NON-NEGOTIABLE)

LLMGopher exists to provide a Go implementation of LiteLLM-style gateway behavior, not to
invent a new LLM gateway API. The default direction for every public API, request/response
schema, streaming protocol, error envelope, admin operation, and operational behavior MUST be
upstream alignment with LiteLLM and the provider APIs it normalizes.

OpenAI-compatible endpoints remain the highest-priority compatibility surface because they
maximize SDK adoption. Existing OpenAI clients MUST work without code changes where the
corresponding feature is implemented. Divergence from LiteLLM or OpenAI behavior requires an
explicit rationale in the spec or PR and MUST be limited to cases where Go runtime constraints,
security, correctness, or maintainability justify it.

All OpenAI-compatible errors MUST return `{"error": {"message": "...", "type": "...",
"code": "..."}}` with established types such as `authentication_error`, `rate_limit_error`,
and `invalid_request_error`.

**Rationale**: Maximum adoptability comes from being a faster, safer implementation of familiar
gateway behavior, not from asking users to learn a project-specific API.

### II. High-Throughput Go Runtime

LLMGopher SHOULD outperform Python gateway implementations under high concurrency by using
Go's type system, goroutines, streaming primitives, and predictable resource management.
Latency-sensitive request paths MUST avoid avoidable blocking database writes, unbounded
goroutine creation, unbounded buffering, global locks, and external observability calls.

Cost tracking, audit logging, budget deductions, metrics callbacks, and post-response
bookkeeping MUST be asynchronous unless the operation is required to make an authorization,
budget, rate-limit, or routing decision before the provider call.

Performance-sensitive features MUST document expected concurrency impact and include a focused
benchmark, load test, or regression test when the change touches the hot path, streaming path,
rate limiter, router, provider client, or storage cache.

**Rationale**: Higher concurrency and throughput are core reasons for this project to exist,
but performance claims must remain grounded in measurable behavior.

### III. Typed Contracts & Generated Interfaces

Canonical gateway behavior MUST be represented with typed Go structs, interfaces, and explicit
adapter boundaries. Unstructured `map[string]any`, raw JSON passthrough, and provider-specific
dynamic fields are permitted only at explicit provider, plugin, or compatibility boundaries and
MUST be converted into typed domain contracts as early as practical.

When upstream OpenAPI schemas, JSON Schemas, protocol definitions, or provider SDK contracts
exist, new features SHOULD evaluate automated consumption or generation before adding large
hand-maintained struct sets. Generated code MUST be reproducible, reviewed for compatibility
impact, and isolated from hand-written business logic.

**Rationale**: Type safety is a primary advantage of the Go implementation. Generated contracts
reduce drift from rapidly changing provider and LiteLLM-compatible APIs.

### IV. Gateway Routing & Reliability

Routing behavior is core gateway functionality. Model aliases, provider selection, retries,
fallback chains, load balancing, per-deployment rate limits, cooldowns, timeouts, and health-aware
routing MUST be designed as first-class gateway capabilities rather than ad hoc provider logic.

Routing decisions MUST be observable, testable, and scoped by the relevant identity and policy:
API key, team, organization, model group, provider, deployment, budget, and rate-limit state.
Failures MUST preserve OpenAI-compatible error semantics unless an upstream-aligned LiteLLM
behavior specifies otherwise.

**Rationale**: LiteLLM adoption is driven as much by operational routing behavior as by endpoint
compatibility. Treating routing as a core contract prevents provider-specific drift.

### V. Multi-Tenant Spend Governance

LLMGopher MUST support multi-tenant gateway operation as a project identity, including teams or
organizations, key-scoped access, model allowlists, RBAC/JWT-capable admin access, budgets, usage
summaries, and audit APIs.

Spend and access controls MUST be enforceable before provider calls when they affect permission,
budget availability, or rate-limit eligibility. Post-response accounting MAY remain asynchronous
when exact token and cost data are only known after the provider response.

**Rationale**: A production LLM gateway controls shared credentials, shared spend, and shared
policy. Multi-tenancy and spend governance cannot be bolted on safely after the API surface
stabilizes.

### VI. Observable & Auditable Operation

Every request path MUST preserve request IDs or call IDs, structured logs, redaction of secrets
and sensitive payloads, and enough audit context to explain routing, authorization, rate-limit,
budget, and provider outcomes.

Metrics, tracing, and callback/webhook integrations SHOULD be added for meaningful runtime paths:
provider calls, routing decisions, streaming lifecycle, rate limiting, budget enforcement, cache
hits, guardrails, and async worker failures. Observability integrations MUST NOT leak secrets and
MUST NOT add blocking external calls to the hot path.

**Rationale**: Gateway operators need to debug spend, latency, routing, and policy decisions
without compromising user data or provider credentials.

### VII. Security by Default (NON-NEGOTIABLE)

- Provider API keys MUST be stored AES-256-GCM encrypted; the encryption key is supplied at
  runtime via `--provider-credentials-key` and MUST never be committed to source control.
- Client API keys MUST be stored as SHA-256 hashes only; plaintext keys MUST NOT be persisted.
- Credentials, secrets, and key material MUST NOT appear in logs, error messages, audit trails,
  or HTTP responses under any circumstances.
- All changes to authentication or credential handling code require an explicit security review
  comment in the PR.

**Rationale**: The gateway sits in front of all LLM spend and holds provider credentials. A
credential leak has outsized blast radius compared to a typical service.

### VIII. Test Discipline

All new production code MUST target 80% test coverage. The following conventions are mandatory:

- HTTP handler tests MUST use `httptest.NewRecorder()`.
- Interface dependencies MUST be exercised via mocks from `internal/mocks/`.
- Integration paths (middleware chain, provider proxy) MUST have at least one end-to-end test
  exercising the full request lifecycle.
- Tests MUST be written before or alongside the implementation — no "tests later" PRs.
- Compatibility changes MUST include contract tests or golden fixtures for upstream-aligned
  request/response behavior where practical.
- Routing, budget, rate-limit, and multi-tenant policy changes MUST include failure-path tests.

**Rationale**: The gateway's correctness directly impacts LLM spend and availability. Untested
code paths become production incidents.

### IX. 12-Factor Configuration

All runtime configuration MUST follow 12-factor app principles:

- Precedence: CLI flags > `LLMGOPHER_*` env vars > unprefixed env vars > config file > defaults.
- No hardcoded configuration values in source code.
- All secrets and credentials supplied via environment or CLI flags — never embedded in config
  files committed to the repository.
- Configuration MUST be validated at startup; the process MUST fail fast with a clear error
  if required values are missing.

**Rationale**: The gateway runs in Docker Compose, Kubernetes, and bare-metal environments. A
single configuration model prevents environment-specific drift bugs.

### X. API Capability UX Parity

New or materially changed API capabilities MUST include corresponding updates to the admin UI
or operator-facing UI surfaces so users can discover, configure, monitor, and validate the
capability without relying only on direct API calls. Specs and plans MUST identify the UI surface
for each API capability, including forms, tables, detail views, status indicators, usage/audit
views, documentation links, and error feedback as applicable.

An API-only delivery is permitted only when the capability has no meaningful UI surface or is
strictly internal. The exception MUST be documented in the spec or PR with the affected user
role, reason UI exposure is not useful, and any follow-up needed when the capability becomes
operator-facing.

**Rationale**: LLMGopher is operated through both APIs and the admin UI. Shipping backend
capabilities without UI exposure makes the feature hard to adopt, validate, and support.

## Development Workflow

Feature development follows the layer order defined by the architecture:

1. **Upstream contract first**: Identify the LiteLLM, OpenAI, or provider behavior being matched
   and document any deliberate divergence before implementation.
2. **Generated or typed contract second** (`pkg/llm`): Consume upstream schemas when practical,
   then define or update canonical types and the `Provider` interface.
3. **Storage and policy third** (`internal/storage`): Add migrations, cache behavior, budget,
   tenant, and policy storage before wiring handlers or proxy logic.
4. **Routing, middleware, and proxy fourth** (`internal/middleware`, `internal/proxy`): Implement
   routing, middleware, provider translation, and streaming after contracts are stable.
5. **Handlers last** (`internal/api`): Wire routes and handlers once all dependencies are ready.
6. **UI exposure with API delivery** (`ui/`): Surface new or changed API capabilities in the
   admin UI in the same delivery unless an API-only exception is documented.

Database migrations in `migrations/` run automatically at startup via goose. Migrations MUST
be backward compatible (no destructive schema changes without a multi-step migration plan).

The state cache (`internal/storage/cache.go`) hot-swaps every 5 seconds without restart.
Features that depend on runtime-configurable state MUST use the cache, not in-process globals.

## Quality Gates

All pull requests MUST satisfy the following before merge:

- `go vet ./...` passes with zero warnings.
- `go test ./...` passes with 80%+ coverage on changed packages.
- No new hardcoded credentials, secrets, or API keys in any committed file.
- Public API changes identify the upstream LiteLLM, OpenAI, or provider behavior being matched
  and include compatibility tests or a documented reason tests are not practical.
- Performance-sensitive changes include a benchmark, load test, or explicit no-regression
  rationale appropriate to the affected path.
- Multi-tenant, budget, rate-limit, and routing changes include negative tests for denied,
  exhausted, over-limit, fallback, and provider-failure scenarios as applicable.
- Observability changes preserve redaction and avoid blocking hot-path external calls.
- New or materially changed API capabilities include corresponding `ui/` changes or a documented
  API-only exception explaining why no operator-facing UI exists.
- Middleware ordering preserved: RequestID → Logging → Recovery → Auth → RateLimit → Guardrail.
  (`middleware.Chain()` reverses the slice — last element wraps the handler first.)
- Async cost/audit path MUST NOT be converted to synchronous without explicit architecture
  decision documented in the PR.
- OpenAI-compatible error format verified for any new error condition.

## Governance

This constitution supersedes all other informal practices and conventions in LLMGopher. It
is the authoritative statement of non-negotiable project rules.

**Amendment procedure**: Any principle may be amended by opening a PR that modifies this file,
stating the rationale, and obtaining approval from the project maintainer. During early project
formation, the version field is informational and MUST NOT block amendments. Release-grade
versioning policy may be added later when the project needs it.

**Versioning policy**:
- The constitution version is retained for traceability only.
- Amendments SHOULD update the Sync Impact Report with the nature of the change.
- Semantic version constraints are intentionally deferred until release governance is useful.

**Compliance review**: Every PR description MUST include a brief statement of how the change
complies with (or is exempt from) the Core Principles above. The Quality Gates section
defines the automated checks; principle compliance is a human reviewer responsibility.

**Version**: 1.1.0 | **Ratified**: 2026-03-15 | **Last Amended**: 2026-04-26
