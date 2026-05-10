# Contract: Admin Audit

Mirrors `/v1/admin/audit` (already shipped — see `internal/api/admin_audit*.go`). Mock returns a subset of fields the audit page renders.

## `GET /v1/admin/audit?actor=&action=&from=&to=&limit=&cursor=`

**Response 200**:
```json
{
  "data": [
    {
      "id": "audit_001",
      "actor": "user_admin_test",
      "action": "key.rotate",
      "resource_kind": "api_key",
      "resource_id": "key_checkout_service",
      "at": "2026-05-09T11:30:00Z",
      "outcome": "success"
    }
  ],
  "next_cursor": null
}
```

The mock only needs to support the filters the audit page exposes; sorting is server-side desc-by-`at`.
