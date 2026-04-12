# Spec 23: Teams & Organizations

## Status
pending

## Goal
Add a team and organization model so that API keys can be grouped, team-level budgets and rate limits can be enforced, and usage can be attributed at the team level. This is the foundation for multi-tenant operation.

## Background
All keys are currently in a flat namespace. Adding teams requires new DB tables, foreign keys on existing tables, and enforcement logic in the request path.

This spec establishes the data model. Spec 24 adds RBAC and JWT auth on top of it.

## Requirements

### 1. Migration (`internal/storage/migrations/00010_teams.sql`)

```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, name)
);

CREATE TABLE team_budgets (
    team_id UUID PRIMARY KEY REFERENCES teams(id) ON DELETE CASCADE,
    budget_usd NUMERIC(12,6) NOT NULL,
    spent_usd NUMERIC(12,6) NOT NULL DEFAULT 0,
    budget_duration TEXT,
    budget_reset_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add team FK to api_keys (nullable — keys without a team are "standalone")
ALTER TABLE api_keys
    ADD COLUMN team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    ADD COLUMN rate_limit_team_rps INTEGER;
```

### 2. New types (`pkg/llm/types.go`)

```go
type Organization struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Metadata  map[string]string `json:"metadata,omitempty"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
}

type Team struct {
    ID        string            `json:"id"`
    OrgID     string            `json:"org_id"`
    Name      string            `json:"name"`
    Metadata  map[string]string `json:"metadata,omitempty"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
}

type TeamBudget struct {
    TeamID       string     `json:"team_id"`
    BudgetUSD    float64    `json:"budget_usd"`
    SpentUSD     float64    `json:"spent_usd"`
    RemainingUSD float64    `json:"remaining_usd"`
    ResetAt      *time.Time `json:"reset_at,omitempty"`
}
```

Add `TeamID string` to `APIKeyConfig`.

### 3. State cache update (`internal/storage/cache.go`)

Load teams and organizations into `GatewayState`:
```go
type GatewayState struct {
    APIKeys       map[string]*APIKey          // keyed by hash
    Models        map[string]*Model           // keyed by alias
    Providers     map[uuid.UUID]*ProviderConfig
    Teams         map[uuid.UUID]*Team         // new
    Organizations map[uuid.UUID]*Organization  // new
}
```

Teams and organizations are relatively static — polling every 30 seconds is sufficient.

### 4. Team budget enforcement (`internal/proxy/cost_worker.go`)

After deducting from the key-level budget, also deduct from the team budget if the key has a `team_id`:

```go
if key.TeamID != "" {
    if err := w.budgetTracker.DeductTeam(ctx, key.TeamID, cost); err != nil {
        w.logger.Warn("team budget exceeded or deduction failed", "team_id", key.TeamID, "error", err)
        // Log but don't block the request (deduction happens after response is sent)
    }
}
```

Add `DeductTeam(ctx, teamID, amount)` and `RemainingTeamBudget(ctx, teamID)` to `BudgetTracker` interface and `PgBudgetTracker` implementation.

### 5. Pre-request team budget check

Before dispatching to the provider (in `handler.go`), check team budget:
```go
if key.TeamID != "" {
    remaining, err := h.budgetTracker.RemainingTeamBudget(ctx, key.TeamID)
    if err == nil && remaining <= 0 {
        return 429 "team budget exhausted"
    }
}
```

### 6. Admin API for organizations and teams

New routes (all require admin auth):

**Organizations:**
- `GET /v1/admin/orgs` — list all organizations
- `POST /v1/admin/orgs` — create: `{"name": "acme", "metadata": {}}`
- `GET /v1/admin/orgs/{id}` — get single org
- `PUT /v1/admin/orgs/{id}` — update name/metadata
- `DELETE /v1/admin/orgs/{id}` — delete (cascades to teams and nullifies key team_ids)

**Teams:**
- `GET /v1/admin/orgs/{org_id}/teams` — list teams in org
- `POST /v1/admin/orgs/{org_id}/teams` — create team
- `GET /v1/admin/orgs/{org_id}/teams/{id}` — get team
- `PUT /v1/admin/orgs/{org_id}/teams/{id}` — update
- `DELETE /v1/admin/orgs/{org_id}/teams/{id}` — delete

**Team budgets:**
- `GET /v1/admin/teams/{id}/budget` — get team budget
- `PUT /v1/admin/teams/{id}/budget` — set team budget

**Key-to-team assignment:**
- Handled via `PUT /v1/admin/keys/{id}` accepting `team_id` field (spec 05 extension)

### 7. Usage API team grouping (spec 08 extension)

Add `group_by=team` to `GET /v1/admin/usage`. Join `audit_log` → `api_keys` → `teams`.

## Out of Scope
- Team member management (users within teams — that's part of RBAC in spec 24)
- Per-team model access lists (add after spec 24 RBAC is in place)
- Team-level rate limits (the `rate_limit_team_rps` column is reserved for a future spec)

## Acceptance Criteria
- [ ] Organization and team can be created via admin API
- [ ] An API key can be assigned to a team
- [ ] A team with an exhausted budget returns 429 on the next request
- [ ] Team budget deduction is atomic (concurrent requests don't over-spend)
- [ ] `GET /v1/admin/usage?group_by=team` returns usage broken down by team
- [ ] Keys without a `team_id` continue to work unchanged
- [ ] Migration runs cleanly on existing schema

## Key Files
- `internal/storage/migrations/00010_teams.sql` — new migration
- `pkg/llm/types.go` — `Organization`, `Team`, `TeamBudget`, extend `APIKeyConfig`
- `internal/storage/cache.go` — load teams/orgs into state
- `internal/storage/budget_tracker.go` — team deduct/remaining methods
- `internal/proxy/handler.go` — team budget pre-check
- `internal/proxy/cost_worker.go` — team budget deduction
- `internal/api/admin.go` — org/team CRUD handlers
- `internal/api/router.go` — new routes
