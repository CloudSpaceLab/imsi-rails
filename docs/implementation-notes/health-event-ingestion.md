# Health Event Ingestion Implementation Notes

Branch: `feat/health-event-ingestion`

This slice adds the first reliability-intelligence backend surface for imsi-rails.

## What It Implements

- `POST /v1/health/samples`
- `GET /v1/health/routes/{route_id}`
- health sample schema for provider API status, timeout rate, error rate, callback lag, and transaction outcome signals
- route health snapshots that feed the same `RouteHealthBook` shape used by the routing engine
- a small `RouteHealthProvider` interface so the routing layer can read live health snapshots from the ingestion service
- provider/route health states: `unknown`, `healthy`, `watch`, and `degraded`
- state-change events when a provider or route moves between health states
- circuit breaker states: `unknown`, `healthy`, `degraded`, `blocked`, and `recovery_testing`
- configurable circuit breaker thresholds for success rate, uptime, latency, and recovery
- persisted circuit breaker records and circuit breaker state-change events
- routing intake reads blocked routes from the health service before route selection
- in-memory health sample store and in-memory state-change event sink for pilot/dev usage
- OpenAPI contract updates for health ingestion and route health lookup
- demo API wiring in `cmd/imsi-api`

## What It Intentionally Does Not Implement Yet

- persistent Timescale/Postgres storage
- NATS/JetStream publication
- circuit-breaker transitions
- active provider polling workers
- latency waterfall analytics
- dashboard API aggregation

Those are later M2/M3 slices so the ingestion contract can stay small and easy to validate first.

## Verification

Run:

```powershell
go test ./...
go vet ./...
```
