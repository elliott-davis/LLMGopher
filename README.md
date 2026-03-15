# LLMGopher — AI API Gateway

A high-performance, low-latency AI API Gateway written in Go. Exposes an OpenAI-compatible API and proxies requests to multiple LLM providers (OpenAI, Anthropic, Vertex AI) with middleware for rate limiting, guardrails, cost tracking, and audit logging.



## Architecture

```
Client (OpenAI SDK) ──▶ Gateway ──▶ Middleware Chain ──▶ Provider Proxy ──▶ LLM Backend
                         │              │
                         │         Auth → Rate Limit → Guardrails
                         │
                    PostgreSQL (audit, budgets)
                    Redis (rate limiting, optional)
```

## Project Structure

```
cmd/gateway/          Entry point (cobra CLI)
internal/
  api/                HTTP router and handlers
  middleware/         Auth, rate limiting, guardrails, logging
  proxy/              Reverse proxy and SSE streaming
  storage/            PostgreSQL implementations
pkg/
  llm/                Core interfaces and domain types
  config/             12-factor configuration (viper + cobra)
docker/               Docker-compose support files
k8s/local/            Kubernetes manifests (kind, production-like)
```

## Quick Start (docker-compose)

```bash
# Start everything — postgres, redis, and the gateway
make dev

# Gateway is live at http://localhost:8080
curl http://localhost:8080/health

# Test with the pre-configured API key
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-test-key-1" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "hello"}]}'

# View logs
make dev-logs

# Stop everything
make dev-down
```

To use real API keys, copy `.env.example` to `.env` and fill in your keys:

```bash
cp .env.example .env
# edit .env with your real keys
make dev-restart
```

## Running Locally (no containers)

If you prefer running the Go binary directly (requires postgres and redis running separately):

```bash
# Uses config.yaml in the current directory
make run

# Or with explicit flags
go run ./cmd/gateway --postgres-dsn "postgres://..." --redis-enabled=false

# Or with environment variables
LLMGOPHER_POSTGRES_DSN="postgres://..." LLMGOPHER_REDIS_ENABLED=false go run ./cmd/gateway
```

## Configuration

Configuration follows [12-factor app](https://12factor.net/config) principles with three layers of precedence:

1. **CLI flags** (highest priority) — `--postgres-dsn`, `--redis-addr`, etc.
2. **Environment variables** — `LLMGOPHER_POSTGRES_DSN`, `LLMGOPHER_REDIS_ADDR`, etc.
3. **Config file** (lowest priority) — `config.yaml` in the working directory or `/etc/llmgopher/`

Un-prefixed env vars (`POSTGRES_DSN`, `REDIS_ADDR`, etc.) are also supported for backward compatibility.

### All Settings

| Flag | Env Var | Default | Description |
|---|---|---|---|
| `--listen-addr` | `LLMGOPHER_LISTEN_ADDR` | `:8080` | Server listen address |
| `--postgres-dsn` | `LLMGOPHER_POSTGRES_DSN` | (required) | PostgreSQL connection string |
| `--postgres-max-open-conns` | `LLMGOPHER_POSTGRES_MAX_OPEN_CONNS` | `25` | Max open DB connections |
| `--postgres-max-idle-conns` | `LLMGOPHER_POSTGRES_MAX_IDLE_CONNS` | `5` | Max idle DB connections |
| `--postgres-conn-max-lifetime` | `LLMGOPHER_POSTGRES_CONN_MAX_LIFETIME` | `5m` | Max connection lifetime |
| `--redis-enabled` | `LLMGOPHER_REDIS_ENABLED` | `false` | Enable Redis rate limiting |
| `--redis-addr` | `LLMGOPHER_REDIS_ADDR` | `localhost:6379` | Redis address |
| `--redis-password` | `LLMGOPHER_REDIS_PASSWORD` | (empty) | Redis password |
| `--redis-db` | `LLMGOPHER_REDIS_DB` | `0` | Redis database number |
| `--api-keys` | `LLMGOPHER_API_KEYS` | (empty) | Comma-separated `KEY:ID` pairs |
| `--guardrail-endpoint` | `LLMGOPHER_GUARDRAIL_ENDPOINT` | (empty) | Guardrail service URL |
| `--guardrail-timeout` | `LLMGOPHER_GUARDRAIL_TIMEOUT` | `5s` | Guardrail request timeout |
| `--rate-limit-rps` | `LLMGOPHER_RATE_LIMIT_RPS` | `60` | Requests/sec per key |
| `--rate-limit-burst` | `LLMGOPHER_RATE_LIMIT_BURST` | `10` | Token bucket burst |
| `--openai-api-key` | `LLMGOPHER_OPENAI_API_KEY` | (empty) | OpenAI API key |
| `--openai-base-url` | `LLMGOPHER_OPENAI_BASE_URL` | `https://api.openai.com` | OpenAI base URL |
| `--anthropic-api-key` | `LLMGOPHER_ANTHROPIC_API_KEY` | (empty) | Anthropic API key |
| `--anthropic-base-url` | `LLMGOPHER_ANTHROPIC_BASE_URL` | `https://api.anthropic.com` | Anthropic base URL |
| `--vertex-project-id` | `LLMGOPHER_VERTEX_PROJECT_ID` | (empty) | Google Cloud project ID used for Vertex OpenAI endpoint |
| `--vertex-region` | `LLMGOPHER_VERTEX_REGION` | `us-central1` | Google Cloud region used for Vertex OpenAI endpoint |
| `--config` | `LLMGOPHER_CONFIG` | (auto-detect) | Explicit path to config file |

## Testing Vertex AI

Vertex support uses Google Cloud IAM (ADC), not a static provider API key.

1. Authenticate ADC (pick one):

```bash
# Local user credentials
gcloud auth application-default login

# Optional: confirm ADC project (or set one)
gcloud auth application-default set-quota-project YOUR_GCP_PROJECT_ID
```

Or use a service account key:

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
```

2. Start the gateway with Vertex configured:

```bash
LLMGOPHER_POSTGRES_DSN="postgres://..." \
LLMGOPHER_API_KEYS="sk-test-key-1:local-dev" \
LLMGOPHER_VERTEX_PROJECT_ID="YOUR_GCP_PROJECT_ID" \
LLMGOPHER_VERTEX_REGION="us-central1" \
go run ./cmd/gateway
```

3. Send a chat completion request using a `vertex/` model prefix:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-test-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "vertex/gemini-2.5-flash",
    "messages": [{"role": "user", "content": "Write a one-line haiku about gateways."}]
  }'
```

If ADC is missing or invalid, the gateway will start but skip Vertex provider registration and log a warning.

<details>
<summary><b>Google Cloud Service account creation</b></summary>

1. Create the service account (Cloud Console)

Open Google Cloud Console and select your project (e.g. gen-lang-client-0088272616).
Go to IAM & Admin → Service accounts.
Click Create service account.
Service account name: e.g. llmgopher-vertex.
Click Create and continue.
Grant access: click Add another role, choose Vertex AI User (or Vertex AI Administrator if you want broader access). Click Continue → Done.
Open the new service account → Keys tab → Add key → Create new key → JSON → Create.

2. Use it with the gateway

The gateway already uses Application Default Credentials. When GOOGLE_APPLICATION_CREDENTIALS is set to the path of that JSON file, it will use the service account.

Local (no Docker):

export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your-service-account.json"
# then start gateway as usual (e.g. make run or go run ./cmd/gateway)
Docker Compose:
Mount the key and set the env var in the gateway service, e.g.:

# docker-compose.yaml (snippet for gateway service)
gateway:
  environment:
    GOOGLE_APPLICATION_CREDENTIALS: /secrets/gcp-sa.json
  volumes:
    - ./path/to/your-service-account.json:/secrets/gcp-sa.json:ro
(Use a path that exists on your machine for ./path/to/...; avoid committing the file.)

After that, set LLMGOPHER_VERTEX_PROJECT_ID (and optionally LLMGOPHER_VERTEX_REGION) as you already do; the gateway will use the service account for Vertex and you can keep using your existing curl with vertex/gemini-2.5-flash

</details>

## Kubernetes

The `k8s/local/` directory contains manifests for deploying to a kind cluster, useful for production-like testing:

```bash
make k8s-up        # Create kind cluster
make k8s-deploy    # Build, load image, deploy all manifests
make k8s-logs      # Tail gateway logs
make k8s-status    # Show pod status
make k8s-undeploy  # Remove all resources
make k8s-down      # Delete the kind cluster
```

## Make Targets

```
  dev               Start all services via docker-compose (build + up)
  dev-down          Stop and remove all containers
  dev-logs          Tail logs from all services
  dev-ps            Show running service status
  dev-restart       Rebuild and restart the gateway only
  build             Build the gateway binary
  run               Build and run the gateway locally (reads config.yaml)
  docker-build      Build the gateway Docker image
  test              Run the Go unit test suite
  test-v            Run tests with verbose output
  clean             Stop compose, delete kind cluster, remove build artifacts
```
