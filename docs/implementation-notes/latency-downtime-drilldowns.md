# Latency and Downtime Drilldowns Implementation Notes

Branch: `feat/latency-downtime-drilldowns`

This slice adds the first operator-facing reliability drilldowns to the Vue control room.

## What It Implements

- end-to-end latency summary
- step-level latency waterfall
- provider, corridor, destination bank, and time-window filters
- downtime timeline with provider events, breaker recovery, and operator actions
- responsive mobile card layout for step latency rows

## Scope Boundaries

- Static prototype data only.
- No charting dependency yet.
- No live health API connection yet.

Those belong after the backend reliability endpoints are merged.

## Verification

Run:

```powershell
npm run web:build
```
