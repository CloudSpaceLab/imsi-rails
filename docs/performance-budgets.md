# imsi-rails Performance Budgets

These budgets define the first engineering quality bar for `imsi-rails`.

The numbers are intentionally strict because the product promise is reliability infrastructure for banks, not a decorative dashboard.

## Backend Hot Path

| Area | Budget |
| --- | --- |
| Route decision p95 | under 20 ms, excluding external provider/bank calls |
| Transaction intake p95 | under 100 ms, excluding external provider/bank calls |
| Route eligibility/scoring benchmark | under 1 ms for 100 candidate routes on a development machine |
| Circuit-breaker state visibility | new state affects routing within 5 seconds |
| Idempotency lookup/write | one durable transaction per intake |
| Pilot service memory | target under 256 MB per service under pilot load |
| Provider adapter isolation | adapter timeout must not block route decisions |

## Dashboard and UI

| Area | Budget |
| --- | --- |
| Landing page | dependency-free static page |
| Landing first viewport | usable without network dependencies |
| Product app initial JS | target under 200 KB gzip for authenticated shell |
| Dashboard initial load | under 3 seconds on a typical corporate network |
| Live health freshness | under 5 seconds for pilot |
| Local UI interaction | under 100 ms for visible response |
| Large tables | virtualized or paginated above 1,000 rows |
| Charts | render downsampled/rollup data by default |

## Data and Analytics

| Area | Budget |
| --- | --- |
| Transaction events | append-only durable lifecycle events |
| Analytics queries | never block transaction routing |
| Pilot analytics | PostgreSQL rollups first |
| ClickHouse adoption | only after measured PostgreSQL rollup/query pressure |
| Event replay | lifecycle events replayable for audit and dashboard rebuilds |

## Compatibility

- Chrome, Edge, Safari, and Firefox.
- Websocket fallback to polling for live dashboard updates.
- Critical actions must not depend only on drag-and-drop.
- Every real-time screen must show last updated timestamp and stale-data state.
- All traffic-changing actions must support keyboard access, preview, audit reason, and rollback path.

## Test Gates

Before a pilot release:

- run Go route scoring microbenchmarks
- run k6 intake and dashboard smoke tests
- run browser checks for landing and product shell
- inspect bundle size for the authenticated app shell
- verify no analytics query is required in route-decision tests

