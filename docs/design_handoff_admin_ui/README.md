# Handoff: LLMGopher Admin UI

## Overview

A high-fidelity admin/control-plane UI for **LLMGopher**, an open-source OpenAI-compatible LLM router. The UI is the operator surface for a platform/infra admin: managing upstream providers, routing rules, virtual API keys, teams, budgets, rate limits, request logs, guardrails, audit history, and org settings.

Target persona: **platform admin** at an org running LLMGopher in production. Future role-aware views (developer self-serve) compose on top of the same shell.

## About the Design Files

The files in this bundle are **design references created in HTML** — prototypes showing intended look and behavior, **not production code to copy directly**.

Your task is to **recreate these designs in the existing LLMGopher codebase** at https://github.com/elliott-davis/LLMGopher (the `ui/` folder, currently a rudimentary React/TypeScript app), using whatever framework patterns already live there. The HTML/JSX in this bundle uses inline Babel + a single `window.LG_DATA` object for mocks; your real implementation should be conventional React modules wired to the Go backend's `/admin/*` endpoints.

If you want to start clean, React + TypeScript + Vite (or whatever the existing `ui/` already uses) is the recommended target. Keep the design tokens (`tokens.css`) verbatim — they're already CSS custom properties and provide both themes for free.

## Fidelity

**High-fidelity (hifi).** Pixel-perfect mockups with final colors, typography, spacing, interactions, and copy. Recreate exactly using the codebase's libraries and patterns. The visual language was deliberately chosen — don't substitute another design system unless the user asks.

## Visual System

- **Type**: Inter (UI) + JetBrains Mono (IDs, model names, hex codes, numbers, code blocks). Numbers use `font-variant-numeric: tabular-nums`.
- **Color**: oklch-tuned warm-cool neutrals; Go-gopher cyan **#00ADD8** as the single accent. Used sparingly — primary buttons, active nav row, brand pills, success-state ring on the budget gauge.
- **Density**: 14px base, comfortable row height 44px (36px in compact mode).
- **Radii**: 6/10/14px (`--r-sm/-md/-lg`).
- **Borders**: hairline `1px solid var(--border)` everywhere; never use box-shadow as a divider.
- **Status colors**: success (green oklch 0.62/0.13/155), warn (amber 0.72/0.14/75), danger (red 0.58/0.20/27). Always paired with a soft tinted background and a matching border at 25% mix.

## Information Architecture

Sidebar (4 groups, 11 destinations):

```
Operate    → Overview · Logs (live) · Audit log
Configure  → Providers · Routes · Guardrails
Govern     → API Keys · Teams · Budgets · Rate limits
Org        → Settings
```

The sidebar shows a brand mark + version, an environment switcher (production / staging), nav with badges (counts, "live" indicator), and a user/role footer. Topbar shows breadcrumbs (group / page), a `⌘K` jump-to search, notifications bell, help. Search is non-functional in the prototype but should be wired to a global command palette in production.

## Screens

### 1. Overview
**Purpose**: At-a-glance health and traffic across the gateway.

**Layout**:
- Page head with title, subtitle, time-range selector, refresh button, "Export" primary CTA.
- 4-column KPI grid: Requests / Spend / Error rate / p95 latency. Each KPI has a label, value (28px tnum), delta with hint, and a sparkline (160×40, brand cyan).
- **Live request flow strip** (the standout interaction): 6 stages — Ingress → Auth → Rate-limit → Guardrails → Route → Provider — rendered as numbered nodes connected by a hairline rail. Animated pulses (8px dots with 22% halo) traverse the rail every ~380ms in 3 lanes. Pulse color: brand cyan = ok, amber = fallback used, red = error. Match these stages to the actual middleware chain in the Go gateway; rename if your chain differs.
- 2-column split: Top routes table (route → strategy → rpm → p95 → errors) and Provider health (logo + name + region/deployments + p95 + status pill).
- 2-column split: Recent requests (streaming, max-height 320, click row → drawer) and Org budget glance (large $ value + ring gauge + per-team progress bars).

**Tweakable layouts**: `cards` (default), `charts` (adds stacked traffic chart + spend-by-team donut), `tables` (hides KPI grid).

### 2. Logs / Request inspector
**Purpose**: Find and inspect any request that traversed the gateway.

**Layout**:
- Filter chips row: 200 / 4xx / 5xx / fallback used / guardrail-blocked. Search input on the right.
- Table columns: Time · Status · Method · Route → Model · Key · Tokens (in/out) · Latency · Cost. Status pills colored by class. Fallback rows show a small "fb" warn pill alongside the status.
- Click a row → **Request inspector drawer** (720px, slides from right):
  - Tabs: trace · prompt · response · headers
  - **Trace tab**: 4-KPI strip (status, latency, tokens, cost) + **Latency timeline** (the second standout) — a per-stage waterfall: auth, rate-limit, guardrails, model call, post-process. Rows are 14px-tall colored bars on a hairline grid, with a stage label, a width-proportional bar, and a right-aligned ms value. On fallback requests, the failed primary stage renders with diagonal stripes in `--danger` and a leading `✗`. Below: a "Routing decision" card showing the chain: route pill → (failed pill if fallback) → provider logo + model.
  - **Prompt / Response / Headers tabs**: monospace code blocks with light syntax highlighting (keys/strings/numbers in muted hue tokens). Copy-as-cURL and Replay buttons in the drawer header.

### 3. Providers
**Purpose**: Manage upstream LLM endpoints (OpenAI, Anthropic, Vertex, Bedrock, Mistral, vLLM, etc.).

**Layout**:
- Page head with "Add provider" primary CTA.
- 4-column condensed KPI row: total providers, healthy count, degraded/offline count, models exposed.
- Single table: Provider (logo + name + id) · Base URL (mono) · Region pill · Deployments (num) · p50 / p95 · Errors 24h · Status pill · ⋯ menu.

**Add Provider drawer** (3-step wizard):
1. Pick provider type — 6 cards in a 2-col grid, each with logo + name + description; selected card gets a cyan border and 20% halo box-shadow.
2. Credentials & endpoint — display name, base URL, API key (password input — store in vault, never echo), region, timeout. After paste, show a dashed "connection ok" card with model count from `GET /v1/models`.
3. Pick models to expose — checkbox table of upstream models with pricing pulled from the provider.

### 4. Routes
**Purpose**: Logical model names mapped to one or more upstream deployments via fallback / weighted / latency / single strategies.

**Layout**: 320px route list on the left, detail panel on the right.
- Selected route shows an inset cyan left-border on its row.
- Detail panel: route name (mono) + description + active pill + YAML/Edit buttons. Stats strip (Strategy, RPM, p95, Error rate, Attached keys), then the **Routing diagram** — a left-side route box → curved bezier paths → right-side member cards. Stroke width scales with weight. Fallback paths render dashed in `--warn`. Member cards have a 3px left border (cyan for active members, amber for fallback role) and a weight pill on the right.
- Below the diagram: a YAML configuration block showing the actual route definition with key/string/number coloring.

### 5. API Keys
**Purpose**: Virtual keys with per-key budget, rate-limit, and route allowlist.

**Layout**: Single table — Name (with id + created date subline) · Key prefix (mono, masked: `sk-lg-prod-7Q2mF…`) · Team pill · Budget bar (90×5px) · RPM cap · Status · Last used · ⋯.

**Click row** → 720px drawer with tabs: overview · routes · limits · activity. Header has Rotate and Disable (red) buttons. Limits tab is a form (RPM, TPM, monthly cap, alert %, hard-cap toggle).

### 6. Teams
**Purpose**: Group members; assign budgets, route access, default rate limits.

**Layout**: 3-column card grid. Each team card: header (name + status pill + id/owner subline), budget block (label + value + bar), 3-column metric row (members / keys / routes), footer with route pills.

### 7. Budgets
**Purpose**: Track spend by team / key / model with hard + soft caps.

**Layout**:
- 4 KPIs: Spend MTD, Org cap, Forecast EOM, Top spender.
- 2-col split: "By team" (progress bars with a vertical dashed line at the 85% alert threshold) and "By model" table (logo + model · requests · tokens · cost).

### 8. Rate limits
**Purpose**: RPM/TPM ceilings at org / team / key / model / route scopes.

**Layout**: 4 KPIs (active rules, throttle events 24h, top throttled key, backend) then a single table — Scope pill · Target (mono) · RPM · TPM · Burst · Window · Action (throttle/reject/queue) · Status (active or "tripped" warn pill).

### 9. Guardrails
**Purpose**: Pre/post-processing rules that redact, warn, or block.

**Layout**: 4 KPIs then a 2-col card grid. Each guardrail card has name + id/owner, a toggle (cyan when on), and a body row showing Action pill, Scope pill, hits 24h, and a Configure button.

### 10. Audit log
**Purpose**: Immutable record of admin actions.

**Layout**: Single table — Time · Actor (avatar + email or "system" pill) · Action pill · Target (mono) · IP. "tamper-evident" pill in the card header.

### 11. Settings
**Purpose**: Org-wide configuration.

**Layout**: 2-col card grid:
- Authentication (SSO, default role, MFA toggle, session lifetime)
- Secrets (vault backend, encryption key, rotation policy)
- Storage (Postgres / Redis / S3 endpoints)
- Integrations (Slack, PagerDuty, Datadog, OpenTelemetry)

Each setting row is a label / value / action triplet separated by hairlines.

## Standout Interactions (preserve these)

1. **Live request flow strip** on the Overview. The pulse animation (3 lanes, color-coded by outcome) gives the operator a visceral sense of system health. Implementation: a `setInterval` spawning pulses with a `requestAnimationFrame`-driven `left: %` over 2.2s. Cap concurrent pulses at ~12 to keep CPU low.

2. **Latency waterfall** in the Request inspector. Per-stage horizontal bars on a shared time axis, with diagonal-stripe failure state for fallback primaries. Timing data should come from Go's middleware chain — emit `x-llmgopher-trace` headers with stage durations.

## Component Inventory

Reusable pieces (from `components.jsx` and inline in screen files):

- `<Logo kind="openai|anthropic|google|mistral|bedrock|vllm">` — square 24/32px chip with mono initial
- `<Pill kind="success|warn|danger|brand"> + dot` — status/category chip
- `<StatusPill status="healthy|degraded|offline|active|throttled|warning|disabled">` — preset mapping
- `<Spark data={[...]} w h color>` — SVG sparkline with optional filled area
- `<Toggle on onChange>` — 32×18 cyan-when-on switch
- `<Icon d={...} size>` — single-path stroked icon (lucide-style)
- `.tbl` — table with sticky uppercase header, hover row, hairline rows, `.num` right-align tnum cells
- `.kpi` — metric card (label / large tnum value / delta hint)
- `.card / .card-head / .card-body / .card-body.tight` — container with optional flush body
- `.btn / .btn.primary / .btn.ghost / .btn.sm / .btn.danger` — button family
- `.drawer-mask + .drawer` — right-side slide-over (animated 0.22s cubic-bezier)
- `.tabs > .tab.active` — underlined tab strip with cyan accent
- `.code` — mono pre block with `.k` (keyword), `.s` (string), `.n` (number) coloring tokens

## Design Tokens

See `tokens.css` (drop-in CSS custom properties). Highlights:

| Token | Light | Dark |
|---|---|---|
| `--bg` | oklch(0.985 0.003 240) | oklch(0.18 0.012 240) |
| `--bg-elev` | oklch(1 0 0) | oklch(0.21 0.013 240) |
| `--bg-sunken` | oklch(0.965 0.004 240) | oklch(0.15 0.011 240) |
| `--border` | oklch(0.91 0.006 240) | oklch(0.30 0.014 240) |
| `--fg` | oklch(0.22 0.01 240) | oklch(0.96 0.005 240) |
| `--fg-muted` | oklch(0.48 0.01 240) | oklch(0.72 0.01 240) |
| `--brand` | #00ADD8 | #00ADD8 |
| `--success` | oklch(0.62 0.13 155) | same |
| `--warn` | oklch(0.72 0.14 75) | same |
| `--danger` | oklch(0.58 0.20 27) | same |

Spacing: rows are `--row-h` (44px / 36px compact), card padding 18px, page padding 24px 28px.

Type scale: 22px page h1 / 14-15px card h3 / 13-14px body / 11-12px subtitle / 10-11px section labels (uppercase, 0.06em tracking).

## State Management

Suggested (React + TanStack Query):

- `useProviders()`, `useRoutes()`, `useKeys()`, `useTeams()`, `useGuardrails()`, `useAudit()` — list endpoints with 30s polling
- `useRequestStream()` — SSE or WebSocket for the live logs table and Overview flow strip
- `useDashboard()` — single endpoint returning the 4 KPIs + sparklines pre-aggregated server-side
- Mutations: `createProvider`, `updateRoute`, `rotateKey`, `disableKey`, `toggleGuardrail`, `updateBudget`, `updateRateLimitRule`
- Drawer open state is local component state; deep-link via `?inspect=req_<id>` query param so request URLs are shareable.

## Backend Contracts (suggested)

`data.js` is a faithful sketch of the response shapes — point your Go handlers at these:

- `GET /admin/providers` → `Provider[]` with status, latencyP50/P95, errors24h
- `GET /admin/routes` → `Route[]` with members[], strategy, live rpm/p95/errRate
- `GET /admin/keys` → `Key[]` with prefix (masked), budget, budgetCap, rpm, status, lastUsed
- `GET /admin/teams` → `Team[]` with members, keys, budget, cap, models[]
- `GET /admin/requests?limit=50&filter=...` → `Request[]` and SSE stream at `/admin/requests/stream`
- `GET /admin/dash` → aggregated KPIs + 48-bucket sparklines
- `GET /admin/audit` → append-only log with tamper-evident hash chain
- `POST/PATCH/DELETE` mutations return the updated resource

## Animations

- Drawer slide-in: 0.22s cubic-bezier(.2,.7,.2,1), translateX(20px → 0) + opacity
- Drawer mask fade: 0.18s ease-out
- Toggle: 0.15s linear on background-color and thumb left
- Pulse animation: 2.2s linear left: 0% → 100%, 3 lanes, pulses spawned every 380ms

## Files Included

- `LLMGopher Admin.html` — entry HTML; loads React/Babel and the JSX modules
- `tokens.css` — design tokens only (extracted from `styles.css`)
- `styles.css` — full stylesheet (tokens + layout + components)
- `components.jsx` — shared primitives (Pill, StatusPill, Logo, Spark, Toggle, Icon)
- `data.js` — mock data shaped like the suggested API contracts
- `screens-overview.jsx` — Overview + live flow strip + cards
- `screens-providers-routes.jsx` — Providers list, Add-Provider drawer, Routes list + detail + diagram
- `screens-keys-teams.jsx` — Keys table + drawer, Teams grid, Rate limits, Budgets
- `screens-logs-rest.jsx` — Logs table, Request inspector + waterfall, Guardrails, Audit, Settings
- `app.jsx` — sidebar/topbar shell, route state, drawer wiring
- `tweaks-panel.jsx` — internal tweaks UI (theme/density/dashboard layout); not needed in production

## Implementation Order (recommended)

1. Drop `tokens.css` into `ui/src/styles/` and import globally. Verify both themes work via `data-theme="dark"` on `<html>`.
2. Build the app shell: sidebar, topbar, content area, drawer primitive. Match spacing/borders exactly.
3. **Overview** without the live flow (static KPI cards + tables). Wire to `GET /admin/dash`.
4. **Providers** + Add-Provider wizard.
5. **Routes** + diagram.
6. **API Keys** + drawer.
7. **Logs** + Request inspector with static waterfall.
8. **Live flow strip** on Overview (spike feature — wire to SSE last).
9. Teams · Budgets · Rate limits · Guardrails · Audit · Settings (largely table-driven).

## Notes

- Don't include the tweaks panel in production. Theme can be a real user setting; density and dashboard layout are nice-to-have prefs but not on the critical path.
- Key prefixes must always be masked server-side. The full key is shown exactly once at creation time (a one-time-reveal modal — not designed yet; ask before shipping).
- The "tamper-evident" pill on Audit implies a hash-chained log. If you don't have that, drop the pill — don't claim a guarantee you don't provide.
- All cost figures are USD; the prototype hardcodes `$` — internationalize when you have non-US customers.
