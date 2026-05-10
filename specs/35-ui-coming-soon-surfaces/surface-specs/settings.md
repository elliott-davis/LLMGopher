# Surface Spec: Settings

## Page Goal

Provide a bounded organization settings surface for gateway identity, security posture, notifications, and display preferences without becoming a catch-all for unrelated administration.

## Source Context

- Current route: `/settings`
- Current state: `Coming soon.`
- UI test expectation: `ui/tests/e2e/settings.spec.ts`
- Visual testing expectation: `TESTING.md` snapshot matrix says "all four cards"
- Related specs: `24-rbac-jwt-auth`, `34-ui-e2e-testing-suite`

## Primary Users

- Administrator responsible for gateway-level UI and operational settings.
- Operator reviewing read-only gateway configuration during support or incident response.

## Required Capabilities

- Render four cards: Gateway Profile, Security, Notifications, and Display.
- Gateway Profile shows gateway name/environment and any safe read-only deployment identifiers available to the UI.
- Security shows authentication and RBAC posture, with unavailable controls clearly disabled until backend support exists.
- Notifications shows alert destinations and categories only when a backing contract exists; otherwise it must be marked unavailable.
- Display shows UI preferences such as density or theme only when those preferences can be persisted or safely local.
- Each editable card must have clear dirty, saving, success, validation error, and unavailable states; production-facing saves require a reconciled backend contract unless the preference is explicitly local-only.
- The page must not expose secrets, provider credential keys, admin API keys, or raw config file values.

## Acceptance Criteria

1. Given an administrator opens `/settings`, when the page renders, then Gateway Profile, Security, Notifications, and Display cards are visible.
2. Given a card has no backing API, when it renders, then it is visibly disabled or read-only with copy explaining why.
3. Given a reconciled production mutation contract or local-only display preference supports editing, when an editable setting is changed and saved successfully, then the persisted value is visible and the save confirmation is announced.
4. Given validation fails, when the administrator tries to save, then the card identifies the invalid field and preserves the entered value.
5. Given a save request fails, when the gateway returns an error, then the card displays the failure reason and keeps the form open.
6. Given any setting value contains secret-like data, when the page renders, then the value is redacted or omitted.

## Data Contract

No production settings contract is defined yet. First implementation should begin with safe read-only cards and local-only display preferences unless a backend settings API is introduced.

Minimum card model:

- card id
- title
- description
- availability state
- fields
- save capability
- last saved or read-only state

## Test Hooks

- Prefer card headings and accessible labels.
- Add stable card identifiers only if visual or E2E tests cannot locate the four cards by role and name.

## Out Of Scope

- Provider credential management.
- Admin API key rotation.
- Full RBAC role and permission editing.
- Notification delivery backend implementation.
