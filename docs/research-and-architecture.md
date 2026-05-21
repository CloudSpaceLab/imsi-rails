# imsi-rails Research and Architecture

Research date: 2026-05-19

## Architecture Goal

Build a super efficient, fast, lightweight switching platform that banks can trust as the operating layer for international money transfer reliability.

The architecture should:

- keep the transaction routing path very small and predictable
- avoid analytics queries in the hot path
- support adapters for messy IMTO realities: REST, SOAP, webhooks, SFTP, CSV, portals, and batch files
- provide excellent latency, downtime, FX, and cost drilldowns
- run well in conservative bank environments
- support a lean pilot without forcing the full scale stack on day one

## Recommended Stack

### Backend Hot Path

Use Go for the routing, switching, circuit breaker, adapter runtime, authentication, RBAC, LDAP/AD onboarding, and dashboard API.

Why:

- Go produces small operationally simple services: compiled binaries, easy containerization, strong standard-library HTTP support, and mature libraries for Postgres, NATS, LDAP, OIDC, SFTP, XML/SOAP, and observability.
- The existing implementation already has Go transaction intake, route selection, health ingestion, and circuit breaker services with tests.
- Bank integration work is more likely to be limited by provider/bank systems and security review than raw routing compute.
- Go is easier to hire for and maintain in bank-integration teams while keeping lower operational overhead than heavier enterprise stacks.
- A compiled single-binary service is easier to deploy in controlled bank infrastructure than a large runtime-heavy stack.

Alternative:

- Rust remains a future option for extremely performance-sensitive components, but it is no longer the default backend path for the pilot.
- Node.js/TypeScript is acceptable for UI-adjacent BFF work, but not for the core switching runtime.

Sources:

- Go: https://go.dev/
- Go FAQ: https://go.dev/doc/faq
- OWASP Authorization Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
- go-ldap: https://pkg.go.dev/github.com/go-ldap/ldap/v3
- go-oidc: https://github.com/coreos/go-oidc

### Eventing and Streaming

Use NATS Core for low-latency status/event fanout and NATS JetStream for durable transaction lifecycle events.

Why:

- NATS is lightweight, single-binary infrastructure with pub/sub, request/reply, and streaming.
- Core NATS is strong for live health events and dashboard updates.
- JetStream adds persistence and replay for transaction events, audit, recovery, and shadow routing tests.
- JetStream replication can be increased only where durability matters, avoiding overpaying for resilience on every signal.

Use cases:

- route.health.changed
- transaction.state.changed
- provider.latency.sampled
- circuit_breaker.opened
- fx.rate.updated
- reconciliation.exception.created

Sources:

- NATS overview: https://nats.io/about/
- Core NATS: https://docs.nats.io/nats-concepts/core-nats
- JetStream: https://docs.nats.io/nats-concepts/jetstream
- NATS monitoring: https://docs.nats.io/running-a-nats-service/nats_admin/monitoring

### Transactional Store

Use PostgreSQL as the system of record.

Why:

- Banks trust PostgreSQL-style relational durability, constraints, indexes, and transaction semantics.
- The business domain needs strong relational integrity across banks, providers, corridors, policy versions, transactions, route decisions, and audit logs.
- Declarative partitioning supports large event and transaction tables while keeping recent/high-use partitions fast.

Initial tables:

- banks
- providers
- corridors
- routes
- route_policies
- transactions
- transaction_events
- route_decisions
- circuit_breaker_events
- fx_rates
- reconciliation_files
- reconciliation_items
- audit_events

Sources:

- PostgreSQL partitioning: https://www.postgresql.org/docs/current/ddl-partitioning.html
- PostgreSQL monitoring stats: https://www.postgresql.org/docs/current/monitoring-stats.html

### Analytical Store

Start with PostgreSQL rollups for the pilot. Add ClickHouse when the bank needs high-cardinality latency, downtime, and provider comparison drilldowns over large volumes.

Why:

- ClickHouse is built for fast analytical queries, inserts, compression, and vectorized execution.
- It is excellent for latency waterfalls, route comparison, high-cardinality event analysis, and dashboard queries.
- It should not be in the transaction hot path.

Recommended approach:

- Pilot: PostgreSQL partitions plus materialized rollups.
- Scale: stream immutable transaction/health events from NATS JetStream into ClickHouse.
- UI: query pre-aggregated views first, raw events only when drilling down.

Sources:

- ClickHouse real-time analytics: https://clickhouse.com/use-cases/real-time-analytics
- Why ClickHouse is fast: https://clickhouse.com/docs/concepts/why-clickhouse-is-so-fast
- ClickHouse observability: https://clickhouse.com/docs/use-cases/observability

### Observability

Use OpenTelemetry for traces, metrics, and logs from day one. Use Prometheus/Grafana for engineering observability, but build the bank-facing reliability dashboard as a first-class product UI.

Why:

- OpenTelemetry gives a vendor-neutral instrumentation model.
- The transaction trace UI can be powered by domain events and OTel spans.
- Prometheus/Grafana are useful internally, but banks need a purpose-built operations experience, not generic charts.

Sources:

- OpenTelemetry: https://opentelemetry.io/docs/what-is-opentelemetry/
- Prometheus overview: https://prometheus.io/docs/introduction/overview/
- Prometheus alerting: https://prometheus.io/docs/alerting/latest/overview/
- Grafana alerting: https://grafana.com/docs/grafana/latest/alerting/

### Frontend

Use Vue 3 with Vite for the product application, Vue Router for screen/query state, uPlot for dense latency/time-series charts, and native semantic tables until row volume requires virtualization.

Why:

- The existing product shell is Vue 3/Vite and already has tested control-room, transaction, provider, policy, FX, reconciliation, and audit surfaces.
- Vue Router gives route-backed drilldowns without a framework migration.
- uPlot is small and optimized for time-series charts.
- Native semantic tables keep the pilot bundle small; add virtualization when operational tables exceed 1,000 rows.

Dashboard guidance:

- charts must be canvas-first for dense time-series
- tables must be virtualized for large result sets
- use server-side aggregation/downsampling
- update dashboards by delta, not full-page refresh
- cap live dashboard refresh to 1-3 seconds unless a screen truly needs faster

Sources:

- Vue: https://vuejs.org/
- Vue Router: https://router.vuejs.org/
- uPlot: https://github.com/leeoniya/uPlot
- TanStack Table: https://tanstack.com/table/v8/docs/overview
- Apache ECharts canvas/SVG guidance: https://echarts.apache.org/handbook/en/best-practices/canvas-vs-svg/

### API Contracts

Use OpenAPI for REST APIs and AsyncAPI-style event documentation for event streams.

Bank-facing API groups:

- transaction intake
- transaction status
- route decision preview
- provider health
- FX rates
- route policy
- reconciliation imports
- audit exports

Source:

- OpenAPI Initiative: https://www.openapis.org/

## Target Architecture

```text
Bank channels / operations
        |
        v
API gateway / bank edge connector
        |
        v
imsi-rails routing core
  - auth and idempotency
  - transaction intake
  - eligibility engine
  - route scoring
  - circuit breaker checks
  - route decision audit
        |
        +--> Provider adapters
        |      - legacy IMTOs
        |      - digital IMTOs
        |      - B2B payout networks
        |      - bank-owned rails
        |
        +--> NATS Core / JetStream
        |      - live health
        |      - durable lifecycle events
        |      - replay and shadow routing
        |
        +--> PostgreSQL
        |      - system of record
        |      - policy versions
        |      - audit and reconciliation
        |
        +--> Analytics store
               - PostgreSQL rollups first
               - ClickHouse at scale
```

## Hot Path Rules

The route decision path should never depend on heavy analytics queries.

Hot path should use:

- in-memory provider capability cache
- in-memory bank policy cache
- in-memory circuit breaker state
- in-memory FX/cost snapshot with freshness markers
- one transactional write for intake/decision durability
- async event publication after durable state is established

Avoid in hot path:

- raw ClickHouse queries
- wide joins across historical events
- synchronous reconciliation matching
- expensive fraud/compliance analytics unless required by bank policy
- blocking provider health probes

## Deployment Profiles

### Lean Pilot

Run:

- one routing API service
- one worker/adapters service
- PostgreSQL
- NATS
- static frontend served by the API or CDN
- OpenTelemetry collector optional if the bank environment supports it

Goal:

- minimize CPU/memory
- reduce procurement friction
- prove routing reliability with real transactions

### Bank Production

Run:

- horizontally scaled routing API
- separate adapter workers
- NATS cluster with JetStream replication
- PostgreSQL HA
- ClickHouse for analytical drilldowns if event volume warrants it
- OTel collector, Prometheus, alerting
- SSO/OIDC integration

### Multi-Bank SaaS / Managed

Run:

- strict tenant isolation
- per-bank policy and encryption boundaries
- regional data residency controls
- provider adapters shared where contractually allowed
- analytics separated by tenant and role

## Performance Budgets

Backend:

- route decision p95 under 20 ms excluding external provider/bank calls
- transaction intake p95 under 100 ms excluding external provider/bank calls
- circuit breaker state update visible in routing within 5 seconds
- dashboard health data freshness under 5 seconds for pilot
- no service should require more than 256 MB memory under pilot load without a clear reason

Frontend:

- landing page dependency-free and usable from a static host
- product app initial JS target under 200 KB gzip for the authenticated shell
- dashboard interaction response under 100 ms for local UI state changes
- large tables virtualized
- charts downsampled before rendering
- no live screen should render every raw event when aggregate deltas are enough

Testing:

- k6 thresholds for API p95 and error rate
- Criterion benchmarks for route scoring and policy evaluation
- browser performance budgets for dashboard screens

Sources:

- k6 thresholds: https://grafana.com/using-k6/thresholds
- Criterion.rs: https://bheisler.github.io/criterion.rs/book/index.html

## Security and Compatibility

Bank-grade baseline:

- OIDC/SAML SSO integration path
- role-based access control
- maker-checker approvals for sensitive policy changes
- immutable audit log for route decisions and configuration changes
- encryption in transit
- encrypted secrets
- tenant and bank-level data boundaries
- idempotency keys on transaction intake
- replay protection on provider callbacks
- signed webhooks where supported

Compatibility:

- REST/JSON for modern integrations
- SOAP/XML adapter capability for legacy providers
- SFTP/CSV support for reconciliation and batch status
- webhooks with retry and signature validation
- export CSV/XLSX/PDF for operations and audit
- graceful dashboard fallback from websockets to polling

## Key Architecture Decision

The platform should be event-driven and analytics-rich, but the transaction routing core must stay small.

The winning architecture is not "more services." It is a clean separation:

- route fast
- record every decision
- stream events
- analyze outside the hot path
- expose simple configuration controls
- let provider performance determine traffic allocation

