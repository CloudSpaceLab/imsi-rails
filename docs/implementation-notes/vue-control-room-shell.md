# Vue Control Room Shell

This slice introduces the first operational web surface for imsi-rails: a Vue 3, Vite, and TypeScript control-room shell focused on bank operations teams, treasury/FX teams, integration teams, and product owners.

## Delivered

- Control-room layout with persistent navigation, operational summary, corridor matrix, recommended action, provider scorecards, and transaction trace.
- Design primitives for health states, live data freshness, route scores, provider scorecards, and transaction timeline steps.
- Operational copy taxonomy mapped into reusable UI labels for route states, incident actions, and rejection reasons.
- Responsive corridor matrix that becomes a compact card stack on mobile while retaining the same decision-critical fields.

## Scope Boundaries

- Static prototype data only. Live intake, policy, and audit APIs will connect in later slices.
- No charting dependency yet. This keeps the shell lightweight until the first real latency and downtime drilldown requirements land.
- No router dependency yet. The initial experience is a single control room to avoid navigation bloat before the backend contracts stabilize.

## Verification

- `npm run web:build`
- Browser check at desktop and mobile widths with no console errors.

