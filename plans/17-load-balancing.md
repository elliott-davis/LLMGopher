# Spec 17: Load Balancing Across Provider Deployments

## Status
pending

## Goal
Allow a model alias to be backed by multiple provider deployments, with traffic distributed across them using configurable routing strategies. This enables handling higher throughput, active-active provider configs, and cost optimization.

## Background
Currently, a model alias maps 1:1 to a single provider configuration. The `models` table has a single `provider_id`. Load balancing requires a many-to-many relationship between models and providers, with weight and priority configuration.

## Requirements

### 1. Migration (`internal/storage/migrations/00007_model_deployments.sql`)

Add a `model_deployments` table:
```sql
CREATE TABLE model_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES providers(id),
    weight INTEGER NOT NULL DEFAULT 1,  -- relative weight for weighted routing
    priority INTEGER NOT NULL DEFAULT 0, -- higher priority = preferred; equal priority = round-robin
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(model_id, provider_id)
);
```

The existing `models.provider_id` column becomes a "primary" deployment for backward compatibility. Do not remove it.

### 2. Update `ModelConfig` (`pkg/llm/types.go`)

```go
type ModelDeployment struct {
    ID         string `json:"id"`
    ProviderID string `json:"provider_id"`
    Weight     int    `json:"weight"`
    Priority   int    `json:"priority"`
    IsActive   bool   `json:"is_active"`
}
```

Add `Deployments []ModelDeployment` to `ModelConfig`. If `Deployments` is empty, fall back to the single `ProviderID` for backward compatibility.

### 3. State cache update (`internal/storage/cache.go`)

Load deployments alongside models in the state cache poll:
```sql
SELECT md.id, md.model_id, md.provider_id, md.weight, md.priority, md.is_active
FROM model_deployments md
WHERE md.is_active = TRUE
```

Attach deployments to their `ModelConfig` by `model_id`.

### 4. Load balancer (`internal/proxy/loadbalancer.go`)

```go
type LoadBalancer struct {
    strategy string // "round_robin" | "weighted" | "least_latency"
    // internal state for round-robin counter, latency tracking
}

// Select picks a deployment from the list based on the configured strategy.
func (lb *LoadBalancer) Select(deployments []ModelDeployment) *ModelDeployment
```

**Round-robin:** Atomic counter mod len(active deployments). Only considers deployments at the highest priority level.

**Weighted:** Random selection weighted by `weight` field. Among equal-priority deployments.

**Least-latency:** Track a rolling average of response latency per deployment (in-memory). Select the deployment with the lowest p50 latency. Update after each request completes.

Default strategy: `round_robin`.

### 5. Integration in `handler.go`

In `resolveModel`, when the model has multiple active deployments, use the load balancer to select one:
```go
deployment := h.loadBalancer.Select(modelCfg.Deployments)
providerCfg := state.Providers[deployment.ProviderID]
```

If the selected deployment fails (after retries per spec 15), mark it as temporarily unhealthy (see requirement 6) and re-select.

### 6. Health tracking (simple cooldown)

If a deployment returns 5xx errors for all retries, put it in a cooldown for a configurable duration (default 60 seconds):
```go
type deploymentHealth struct {
    cooledDownUntil time.Time
}
```

During selection, skip deployments in cooldown. If all deployments are in cooldown, skip the cooldown check and try them anyway (fail-open).

Log when a deployment enters or exits cooldown.

### 7. Admin API

New routes:
- `GET /v1/admin/models/{id}/deployments` — list deployments
- `POST /v1/admin/models/{id}/deployments` — add a deployment: `{"provider_id": "...", "weight": 1, "priority": 0}`
- `PUT /v1/admin/models/{id}/deployments/{deployment_id}` — update weight/priority/is_active
- `DELETE /v1/admin/models/{id}/deployments/{deployment_id}` — remove deployment

### 8. Config

Add to `pkg/config/config.go`:
```go
LoadBalancing struct {
    Strategy string `mapstructure:"strategy"` // default "round_robin"
    CooldownSeconds int `mapstructure:"cooldown_seconds"` // default 60
} `mapstructure:"load_balancing"`
```

## Out of Scope
- Cross-instance coordination of load balancer state (each gateway instance has independent health tracking)
- Adaptive TPM-aware routing (track tokens-per-minute to stay under provider rate limits)

## Acceptance Criteria
- [ ] A model with two deployments distributes requests in round-robin
- [ ] A deployment that returns 5xx is put into cooldown and traffic shifts to the other deployment
- [ ] After cooldown expires, the deployment receives traffic again
- [ ] Models with a single deployment (no `model_deployments` rows) continue to work unchanged
- [ ] Weighted strategy biases traffic proportionally to `weight` values
- [ ] Admin API adds/removes deployments and changes take effect within cache poll interval

## Key Files
- `internal/storage/migrations/00007_model_deployments.sql` — new migration
- `pkg/llm/types.go` — `ModelDeployment`, extend `ModelConfig`
- `internal/storage/cache.go` — load deployments
- `internal/proxy/loadbalancer.go` — new file
- `internal/proxy/handler.go` — integrate load balancer
- `internal/api/admin.go` — deployment CRUD
- `internal/api/router.go` — new routes
- `pkg/config/config.go` — load balancing config
