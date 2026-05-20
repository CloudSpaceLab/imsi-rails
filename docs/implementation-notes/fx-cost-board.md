# FX and Cost Board Implementation Notes

Branch: `feat/fx-cost-board`

This slice adds a focused FX and cost board to the bank operations view.

## What It Implements

- FX rate by provider, corridor, and currency pair
- rate timestamp per provider
- stale-rate alert
- fee, spread, and effective cost per route
- cheapest route vs route to use
- short operator explanation for the route call

## Scope Boundaries

- Static prototype data only.
- No rate feed integration yet.
- No treasury approval flow yet.
- No reconciliation matching yet.

The goal is to make routing cost visible beside reliability, not to build a full treasury workstation.

## Verification

Run:

```powershell
npm run web:build
```
