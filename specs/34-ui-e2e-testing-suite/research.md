# Phase 0 Research — UI E2E Testing Suite

Resolves the open questions in `spec.md` and pins concrete tooling choices.

---

## Decision 1: Visual regression vendor — **Applitools Eyes**

**Rationale**: `TESTING.md` already specifies Applitools' Playwright SDK and the suppression API (`ignoreRegion`, `layoutRegion`). The SDK runs inside the existing `@playwright/test` runner — no second harness, no separate CI step. Applitools' AI-assisted diffing reduces false positives on antialiasing/font-rendering deltas across runners, which is the most common visual-flake source for non-Chromium-controlled CI.

**Alternatives considered**:
- *Percy* — solid product, but the SDK requires a separate `percy exec` wrapper around the Playwright invocation, which complicates retry/trace artifacts.
- *Chromatic* — strongest for Storybook-driven snapshots; we don't have Storybook and adopting it just for Chromatic is out of scope.
- *Playwright built-in `toHaveScreenshot()`* — free, but pixel-exact diffs flake on font hinting between macOS-local and Linux-CI. No managed reviewer workflow.

**Open follow-up (Story 3 only)**: requires `APPLITOOLS_API_KEY` provisioned as a repo secret before the visual suite is enabled in CI.

---

## Decision 2: Mock backend approach — **In-process Hono server, Node-side**

**Rationale**: The UI uses Next.js **server actions** (`ui/src/lib/actions.ts`) that issue `fetch` from the Node runtime to `http://gateway:8080`. Browser-side MSW cannot intercept these. Two viable options:

1. **In-process Hono mock** started by Playwright's `webServer` chain alongside `next dev`, listening on `127.0.0.1:8080` (the hostname `gateway` resolves via a `/etc/hosts`-style override or by setting `NEXT_PUBLIC_GATEWAY_BASE` and refactoring `actions.ts` to read it).
2. **`gateway --mock` flag** on the Go binary — contract-true but adds a Go build step and CI dependency.

Picking **(1)**. Drift risk is real but manageable because the contract files in `contracts/` are derived from existing endpoint behavior and are reviewed alongside backend changes. The cost of adding a Go build to UI CI runs (per PR, on Ubuntu, with caching to manage) is higher than the cost of contract review.

**Required prerequisite refactor**: `ui/src/lib/actions.ts` currently hardcodes `const GATEWAY_BASE = "http://gateway:8080"`. Replace with `process.env.LLMGOPHER_GATEWAY_BASE ?? "http://gateway:8080"` so tests can point it at `http://127.0.0.1:{mockPort}`. This is a one-line change with no behavior impact in production (env var unset → existing default).

**Alternatives considered**:
- *Playwright `route()` interception* — only works for browser-originated requests; server actions bypass it.
- *Undici mock interceptor* — works at Node fetch layer, but is per-test setup and doesn't survive Next.js server-component rendering across requests.
- *MSW in Node mode* — viable, but Next.js 15's server-component fetch wrapping has had compatibility issues with MSW; Hono with explicit handlers is more predictable and gives us a real HTTP server we can also hit from non-Playwright tools (curl, browser).

---

## Decision 3: Selector policy — **role/label first, `data-testid` second**

**Rationale**: Playwright's `getByRole` and `getByLabel` queries double as accessibility assertions — if the test can't find the role, the a11y tree is broken. `data-testid` is reserved for cases where role/label is genuinely ambiguous: table rows, drawer triggers, multi-instance widgets, and animated-state markers (`data-failed`, `data-flow-strip`).

**Enforcement**: Add an ESLint rule `playwright/no-raw-locators` (custom or via `eslint-plugin-playwright`) to fail builds when E2E specs use raw CSS selectors. Pin testids in the same PR as the test that needs them — do not pre-seed.

**Alternatives considered**:
- *Pure `data-testid`* — fast to write, but bypasses the a11y tree and the testids drift from product naming.
- *Role-only* — strictest, but tables and drawers don't have unique roles.

---

## Decision 4: Fixture format — **TypeScript modules, not JSON**

**Rationale**: Fixtures co-locate with the mock handlers and need to reference shared TypeScript types from `ui/src/lib/types.ts`. JSON files would lose type safety and require parallel `.d.ts` definitions. TypeScript fixtures also let us compute derived state (e.g., a key whose `usage_usd / budget_usd = 0.86` for the budget warning test) without hand-maintaining magic numbers.

**Convention**: One file per entity under `ui/tests/fixtures/`. Each exports a `seed()` function that produces a deterministic snapshot. The default Playwright project loads `seed.ts` which composes all entities. Per-test seeds may opt into a different composition.

---

## Decision 5: CI runner — **`ubuntu-latest`, cache npm + Playwright browsers**

**Rationale**: macOS runners cost ~10× ubuntu and the suite has no macOS-specific code path. Cache `~/.cache/ms-playwright` keyed by `package-lock.json` to avoid the ~100MB Chromium download per run. Cache `~/.npm` keyed by the same.

**Alternatives considered**:
- *Self-hosted runner* — overkill for current PR volume; revisit if CI minutes become a bottleneck.

---

## Decision 6: Visual fixture freezing — **`Date.now` patched, sparklines seeded with PRNG**

**Rationale**: The mock backend freezes its notion of "now" to `2026-05-09T12:00:00.000Z`. Client-side time formatters that call `Date.now()` directly are patched in the Playwright fixture via `await page.addInitScript(() => { Date.now = () => 1762689600000; })`. Random-looking sparklines are generated via a seeded PRNG (mulberry32 with seed = `0xDEADBEEF`) so each run produces identical SVG.

**Open follow-up**: A subset of components may use `performance.now()` for animations; those must be wrapped in `ignoreRegion` rather than mocked.

---

## Decision 7: Density (compact) — **deferred**

`TESTING.md` references a `light-compact` Playwright project. The UI does not yet have a density toggle. Story 3 ships with two projects (`light-comfy`, `dark-comfy`); the third project lands when the density feature lands.

---

## Decision 8: Accessibility scope — **WCAG 2.1 AA, axe default rules**

**Rationale**: Run axe with default WCAG 2.1 AA configuration on every primary route. Disable the `color-contrast` rule **only** if the dark-theme palette legitimately fails it and the failure is tracked as design debt — do not silently exclude.

---

## Resolved NEEDS CLARIFICATION

All open questions in `spec.md` are now resolved:

| Spec question | Resolved by |
|---|---|
| Visual regression vendor (Applitools/Percy/Chromatic) | Decision 1 — Applitools |
| Mock backend approach (in-process vs `gateway --mock`) | Decision 2 — In-process Hono |
| CI runner | Decision 5 — `ubuntu-latest` |
