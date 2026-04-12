# Spec 24: RBAC & JWT Authentication

## Status
pending

## Goal
Add role-based access control to the admin API and provide JWT as an alternative authentication method. This enables safe multi-operator deployments where different users have different levels of access.

## Background
Currently all API keys have identical admin access — anyone with a valid bearer token can read and write all admin endpoints. Spec 23 adds organizations and teams. This spec layers roles onto that model and adds JWT auth.

## Requirements

### 1. Migration (`internal/storage/migrations/00011_rbac.sql`)

```sql
CREATE TYPE admin_role AS ENUM (
    'proxy_admin',     -- full access to everything
    'proxy_viewer',    -- read-only access to all resources
    'org_admin',       -- manage teams and keys within their org
    'team_admin',      -- manage keys within their team
    'key_owner'        -- read their own key's usage/budget only
);

ALTER TABLE api_keys ADD COLUMN role admin_role NOT NULL DEFAULT 'key_owner';
```

### 2. Role definitions

| Role | Admin API Access |
|------|-----------------|
| `proxy_admin` | Full read/write on all endpoints |
| `proxy_viewer` | GET on all admin endpoints, no write |
| `org_admin` | Full CRUD on teams and keys within their org only |
| `team_admin` | Full CRUD on keys within their team only; read team budget |
| `key_owner` | Read their own key's usage and budget; no admin access |

### 3. Role enforcement middleware (`internal/middleware/authz.go`)

```go
// RequireRole returns a middleware that checks the API key's role meets the minimum.
func RequireRole(minRole Role, cache *storage.StateCache) func(http.Handler) http.Handler
```

Apply to admin routes:
```go
// In router.go:
adminMux.Handle("DELETE /v1/admin/keys/{id}",
    RequireRole(RoleProxyAdmin, cache)(HandleDeleteKey(db)))
adminMux.Handle("GET /v1/admin/audit",
    RequireRole(RoleProxyViewer, cache)(HandleGetAudit(db)))
```

For `org_admin` and `team_admin`, also check that the resource being accessed belongs to the key's org/team.

### 4. Resource scoping helpers

Add to `internal/middleware/authz.go`:
```go
// RequireOrgScope ensures the key's org_id matches the org_id of the resource being accessed.
func RequireOrgScope(cache *storage.StateCache) func(http.Handler) http.Handler

// RequireTeamScope ensures the key's team_id matches the team_id of the resource.
func RequireTeamScope(cache *storage.StateCache) func(http.Handler) http.Handler
```

These are applied to org-scoped and team-scoped admin routes.

### 5. JWT authentication (`internal/middleware/jwt_auth.go`)

Support Bearer tokens that are JWTs in addition to raw API keys. JWT auth is used for the admin API only (not for LLM requests).

```go
type JWTConfig struct {
    Enabled    bool
    SecretKey  string // HMAC-SHA256 secret, or path to RSA public key
    Algorithm  string // "HS256" | "RS256"
    Issuer     string // optional issuer validation
    Audience   string // optional audience validation
}
```

JWT claims expected:
```json
{
  "sub": "user-id-or-email",
  "role": "proxy_admin",
  "org_id": "...",  // optional, for org_admin
  "team_id": "...", // optional, for team_admin
  "exp": 1234567890
}
```

The JWT middleware:
1. Checks `Authorization: Bearer <token>`
2. If token parses as a JWT (has 3 dot-separated base64 segments), validates as JWT
3. Otherwise, falls back to API key lookup (existing behavior)
4. On valid JWT, creates a synthetic `APIKeyConfig` with the role and org/team IDs from claims, stores in context

### 6. Master key

Add a `master_key` to config:
```go
Auth struct {
    MasterKey string `mapstructure:"master_key"` // always has proxy_admin role
} `mapstructure:"auth"`
```

The master key bypasses role checks. Used for initial setup and emergency access.

### 7. Admin API: key role assignment

`POST /v1/admin/keys` and `PUT /v1/admin/keys/{id}` accept `role` field. Only a `proxy_admin` can assign roles. An `org_admin` can create keys with roles <= `team_admin` within their org.

### 8. Update `APIKeyConfig`

```go
Role   string `json:"role"`   // one of the role values
OrgID  string `json:"org_id,omitempty"`
TeamID string `json:"team_id,omitempty"` // already added in spec 23
```

Include `role`, `org_id` in state cache load and scan.

## Out of Scope
- SSO / OIDC integration
- UI-based role management (update the Next.js UI in a separate task)
- Per-endpoint permission tables (role enum is sufficient)
- Token rotation

## Acceptance Criteria
- [ ] A `proxy_viewer` key can call `GET /v1/admin/audit` but receives 403 on `POST /v1/admin/keys`
- [ ] An `org_admin` key can create teams within their org but not in another org
- [ ] A `key_owner` key receives 403 on all admin endpoints except their own budget/usage
- [ ] A valid JWT with `role: "proxy_admin"` claim has full admin access
- [ ] An expired JWT returns 401
- [ ] The master key always has full access regardless of role checks
- [ ] LLM request endpoints (`/v1/chat/completions`, etc.) are unaffected by RBAC — only admin endpoints are scoped
- [ ] Migration runs cleanly

## Key Files
- `internal/storage/migrations/00011_rbac.sql` — new migration
- `pkg/llm/types.go` — add `Role`, `OrgID` to `APIKeyConfig`
- `internal/storage/cache.go` — include role/org_id in scan
- `internal/middleware/authz.go` — `RequireRole`, scoping middleware (new file)
- `internal/middleware/jwt_auth.go` — JWT validation (new file)
- `pkg/config/config.go` — JWT and master key config
- `internal/api/router.go` — apply role middleware to admin routes
- `cmd/gateway/main.go` — init JWT config
