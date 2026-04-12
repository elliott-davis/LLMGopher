# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

LLMGopher is an OpenAI-compatible API gateway written in Go that proxies requests to multiple LLM providers (OpenAI, Anthropic, Google Vertex AI). It handles auth, rate limiting, guardrail checks, cost tracking, and audit logging, with PostgreSQL for persistence and optional Redis for rate limiting.

## Commands

```bash
make build        # Build ./bin/gateway
make run          # Run locally (requires postgres + redis)
make test         # Run full test suite
make test-v       # Run tests with verbose output
make dev          # Start full local env via Docker Compose (gateway + postgres + redis + UI)
make dev-logs     # Tail Docker Compose logs
make dev-down     # Stop Docker Compose
make dev-restart  # Rebuild gateway + UI
make k8s-up       # Start kind cluster
make k8s-deploy   # Deploy to kind
```

Run a single test:
```bash
go test ./internal/api/... -run TestFunctionName -v
```

There is no lint target in the Makefile. Use `go vet ./...` directly.

Local dev test key: `sk-test-key-1:key-001` (when using `make dev`).

## Architecture

The gateway is a single binary (`cmd/gateway/main.go`) with these layers:

**`pkg/llm`** — Domain layer. All canonical types (mirroring OpenAI's API schema), the `Provider` interface, and the model registry (prefix-based routing, e.g. `"gpt-4*"` → OpenAI provider). All other packages depend on this; it has no internal dependencies.

**`pkg/config`** — Config loading via Viper + Cobra. Precedence: CLI flags > `LLMGOPHER_*` env vars > unprefixed env vars > config file > defaults.

**`internal/api`** — HTTP router and handlers (Go 1.22+ `http.ServeMux`). Wires middleware, handler dependencies, and routes (`/health`, `/ready`, `/v1/chat/completions`, `/v1/embeddings`, `/v1/admin/*`).

**`internal/middleware`** — Middleware chain (applied bottom-to-top via `middleware.Chain()`): RequestID → Logging → Recovery → Auth → RateLimit → Guardrail.

**`internal/proxy`** — Provider implementations translate canonical `llm.ChatCompletionRequest` to/from provider-native formats. Cost calculation and audit logging happen asynchronously via `cost_worker.go` after the response is sent — this is intentional to avoid latency impact.

**`internal/storage`** — PostgreSQL (goose migrations in `migrations/`), in-memory state cache (hot-swapped every 5s without restart), pricing cache (refreshed every 1 min from DB), audit logger, and budget tracker.

**`internal/providers/google`** — Vertex AI auth and stream handling.

**`internal/mocks`** — Mock implementations of all key interfaces for unit tests.

## Key Patterns

**State cache hot-swapping:** `internal/storage/cache.go` polls the DB every 5 seconds. API keys can be created/revoked without restarting the gateway.

**Non-blocking async cost/audit:** After sending the response to the client, the proxy sends token counts and cost data to a channel consumed by `cost_worker.go`. Audit log and budget deduction happen there.

**Middleware ordering:** `middleware.Chain()` reverses the slice — last element in the list wraps the handler first. Recovery and logging are outermost; guardrail check is innermost.

**Provider credential encryption:** API keys for providers are stored AES-256-GCM encrypted. Requires `--provider-credentials-key` (base64 of 32-byte key) to be set.

**OpenAI-compatible error format:** All errors return `{"error": {"message": "...", "type": "...", "code": "..."}}` with types like `authentication_error`, `rate_limit_error`, etc.

## Database

Migrations are in `migrations/` and run automatically at startup via goose. Key tables: `api_keys` (SHA-256 hashed), `audit_log`, `api_key_budgets`, `model_pricing`, `models`, `providers`.

## Test Coverage Target

New code should target 80% test coverage. Use `httptest.NewRecorder()` for HTTP handler tests and the mocks in `internal/mocks/` for interface dependencies.
