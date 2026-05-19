# Circuit Breaker State Machine Implementation Notes

Branch: `feat/circuit-breaker-state-machine`

This slice turns route health samples into route safety decisions the switch can use before selecting a provider.

## What It Implements

- circuit breaker states: `unknown`, `healthy`, `degraded`, `blocked`, and `recovery_testing`
- configurable thresholds for success rate, uptime, p95 latency, and recovery
- deterministic state evaluation from route health snapshots
- blocked-to-recovery-testing-to-healthy transition flow
- persisted in-memory circuit breaker records per route
- circuit breaker state-change events
- route health API responses that include the current circuit breaker record
- routing intake that reads blocked routes from the health source and rejects them with `circuit_breaker_open`

## Scope Boundaries

- No manual operator override yet.
- No maker-checker approval flow yet.
- No persistent database cache yet.
- No active probing worker yet.

Those belong in later switching/configuration and pilot-hardening slices.

## Verification

Run:

```powershell
go test ./...
go vet ./...
```
