# Policy Simulator and Shadow Routing Implementation Notes

Branch: `feat/policy-simulator-shadow-routing`

This slice adds the bank-facing simulator surface for testing routing policy changes before activation.

## What It Implements

- sample transaction selector
- shadow-only simulation control
- current vs proposed route comparison
- rejected route explanations
- reportable shadow result metrics
- shadow result table for export workflows

## Scope Boundaries

- Static prototype data only.
- No production policy writes.
- No historical transaction import yet.
- No maker-checker approval flow yet.
- No backend simulator endpoint in this UI branch.

The backend execution endpoint should stack on the Go routing branch, then wire into this surface after the UI and backend stacks are integrated.

## Verification

Run:

```powershell
npm run web:build
```
