# Frontend Framework Review

Date: 2026-05-19

## Decision

Use Vue 3 + Vite + TypeScript for the authenticated product UI.

Keep the landing page as dependency-free static HTML/CSS/JS.

## Why Vue

Vue is the better delivery choice for `imsi-rails`:

- lower hiring and handoff risk than Svelte
- mature enterprise ecosystem
- strong TypeScript support through official tooling
- Vite-native speed and first-class templates
- official answers for routing and state through Vue Router and Pinia
- enough performance for a dense control-room UI when paired with rollups, virtualization, and canvas charting

## Why Not Svelte First

Svelte is excellent, but the product does not gain enough from Svelte's compile-time model to justify the ecosystem risk.

The important UI performance work is framework-independent:

- avoid sending raw event firehoses to the browser
- downsample chart data
- virtualize large tables
- use server-side rollups
- update live dashboards by deltas
- keep status and route configuration components small

Vue can meet these requirements while being easier to hire for and easier to explain to bank technology teams.

## Recommended UI Stack

- Vue 3
- Vite
- TypeScript
- Vue Router
- Pinia only when shared state becomes necessary
- uPlot for dense latency/time-series charts
- TanStack Table for large operational tables
- ECharts only for visualizations that need richer interaction than uPlot

## Guardrails

- Do not adopt a large component framework before building core design-system primitives.
- Do not let Vue become a generic admin dashboard.
- Keep the first authenticated shell under the JavaScript performance budget.
- Use canvas-first charting for dense latency and downtime views.
- Use pagination or virtualization above 1,000 table rows.
- Keep the landing page static.

## Sources

- Vue TypeScript guide: https://vuejs.org/guide/typescript/overview
- Vue FAQ: https://vuejs.org/about/faq.html
- Vite guide: https://vite.dev/guide/
- Vue Router: https://router.vuejs.org/
- Pinia: https://pinia.vuejs.org/

