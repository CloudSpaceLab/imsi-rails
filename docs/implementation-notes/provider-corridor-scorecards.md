# Provider and Corridor Scorecards Implementation Notes

Branch: `feat/provider-corridor-scorecards`

This slice upgrades the control-room scorecards from generic provider scores to objective corridor performance metrics.

## What It Implements

- success rate by provider and corridor
- P50, P95, and P99 latency by provider and corridor
- stuck transaction rate
- settlement/reconciliation exception count
- scorecard time-window selector

## Scope Boundaries

- Static prototype data only.
- No live analytics aggregation endpoint yet.
- No export workflow yet.

Those belong after the backend health and reconciliation slices land.

## Verification

Run:

```powershell
npm run web:build
```
