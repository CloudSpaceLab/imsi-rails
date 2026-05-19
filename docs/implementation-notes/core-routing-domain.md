# Core Routing Domain Implementation Notes

Branch: `feat/core-routing-domain`

This slice introduces the first Rust workspace and the `imsi-core` crate.

## What It Implements

- canonical transaction states
- transaction event schema
- lifecycle transition rules
- default unsafe failover boundary
- provider and route registry models
- route capability fields for corridor, payout method, amount range, currency, destination bank, liquidity, FX age, and route status
- bank routing policy model
- eligibility filtering with explicit rejection reasons
- weighted route scoring
- route decision audit model
- in-memory route decision store for sandbox/pilot flow
- sandbox provider adapter with configurable latency, failure, rejection, timeout, and duplicate callbacks
- sample lifecycle fixture

## What It Intentionally Does Not Implement Yet

- transaction intake REST API
- PostgreSQL persistence
- NATS/JetStream event publishing
- real provider adapters
- RBAC/maker-checker controls
- product UI screens

Those belong in later slices so the routing domain can stay small, testable, and free of infrastructure coupling.

## Verification

Run:

```powershell
cargo test
```

Current result:

- 9 unit tests passing

