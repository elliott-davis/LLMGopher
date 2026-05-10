# Surface Spec: Teams

## Page Goal

Give administrators a tenant overview that connects teams to member counts and governance status, with budget health visible enough to support spend management workflows.

## Source Context

- Current route: `/teams`
- Current state: `Coming soon.`
- Mock contract: `specs/34-ui-e2e-testing-suite/contracts/admin-teams.md`
- Mock fixtures: `ui/tests/fixtures/teams.ts`
- Related specs: `23-teams-organizations`, `31-ui-key-budgets`, `34-ui-e2e-testing-suite`

## Primary Users

- Administrator managing tenant or team-level gateway access.
- Operator reviewing tenant spend posture and ownership during an incident.

## Required Capabilities

- Show a grid or table of teams with display name, member count, budget utilization, and budget health state.
- Clearly mark teams at or above their alert threshold.
- Link or guide administrators from a team to its relevant budget controls.
- Show empty and unavailable states separately.
- Keep first implementation read-only unless the real teams API supports safe mutations.
- Reconcile mock-only fields with the production teams contract before enabling production data.

## Acceptance Criteria

1. Given seeded teams Research and Platform, when an administrator opens `/teams`, then both teams render with display name and member count.
2. Given Research has `budget_utilization` of `0.86`, when `/teams` renders, then Research shows a warning state because it is above the 85% alert threshold.
3. Given Platform has `budget_utilization` of `0.40`, when `/teams` renders, then Platform does not show a warning state.
4. Given no teams exist, when the page renders, then the empty state explains that no teams are configured.
5. Given the teams endpoint is unavailable, when the page renders, then the unavailable state does not imply teams are absent.

## Data Contract

The first UI version consumes `/v1/admin/teams`. The mock contract is intentionally minimal and must be reconciled with `23-teams-organizations` when the real API ships.

Required fields:

- `id`
- `display_name`
- `member_count`
- `budget_utilization`

## Test Hooks

- `team-{id}-row`
- `team-research-warn`

## Out Of Scope

- Creating, deleting, or renaming teams until the production teams API supports it.
- Member management.
- Role assignment; RBAC behavior belongs to `24-rbac-jwt-auth`.
