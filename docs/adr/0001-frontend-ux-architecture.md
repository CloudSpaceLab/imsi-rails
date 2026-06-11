# ADR 0001: Frontend UX Architecture

Date: 2026-05-20

Status: Accepted

## Context

The product brief and PRD require `imsi-rails` to feel like a premium bank-grade IMTO control room, not a generic admin dashboard. The current implementation is a Vue 3/Vite prototype and the older architecture notes referenced SvelteKit, creating drift at the exact moment the UI needs deeper investment.

## Decision

Keep the current Vue 3/Vite app for the UI/UX revamp and make it the accepted MVP frontend stack.

Frontend choices:

- Framework: Vue 3 with Vite.
- Navigation: screen-state navigation inside the MVP shell before adopting a full router.
- Design direction: light-first bank operations surface with dense, restrained instrumentation.
- Data model: typed UI contracts and mock services first, then backend API adapters.
- Icons: `@lucide/vue`.
- Charting: Chart.js for bank-facing dashboard line/bar/doughnut charts, with uPlot retained for dense diagnostic time-series and latency drilldowns.
- Tables: native semantic tables/grids for MVP; add virtualization when rows exceed 1,000.
- Realtime model: websocket-first data shape with polling fallback and explicit stale/unavailable states.
- QA: build/typecheck on every change, then add Playwright screenshots, accessibility checks, and visual regression.

## Performance Budgets

- Initial dashboard load under 3 seconds on a typical corporate network.
- Dashboard updates visible within 5 seconds of health-state change.
- Critical route-policy screens render without horizontal scrolling at desktop/tablet sizes.
- Critical numbers must show units, measurement windows, and freshness.
- Large operational tables must paginate or virtualize before exceeding 1,000 rows.

## Consequences

- No SvelteKit migration during the UI revamp.
- Product screens are rebuilt around workflows rather than polishing the old one-page demo.
- Static fixtures are organized as typed scenarios, so loading, stale, empty, permission-denied, and API-failure states can be designed before backend APIs are complete.
- Dashboard charts must map to backend rollups described in `docs/adr/0002-settlement-data-architecture.md`.
- All traffic-changing workflows must show preview, reason capture, approval state, audit context, and rollback target before activation.
