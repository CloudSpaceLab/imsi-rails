# imsi-rails Product Requirements Document

## 1. Product Summary

`imsi-rails` is a bank-facing IMTO reliability, routing, and switching platform.

Banks connect their international money transfer providers to `imsi-rails`, monitor live provider and corridor health, compare speed/cost/FX/reliability, and route every transaction to the best eligible provider according to bank policy and live rail performance.

The product is not a consumer remittance app. It is infrastructure for banks.

## 2. Problem

Banks typically work with many IMTO providers, but each provider has separate integrations, portals, SLAs, reconciliation files, support processes, and performance characteristics.

This creates operational pain:

- no single view of provider health
- no objective provider scorecard
- late discovery of provider downtime or latency spikes
- manual switching decisions during incidents
- unclear failure ownership
- expensive or slow routes used when better routes are available
- fragmented FX/cost visibility
- difficult reconciliation and settlement follow-up
- weak audit trail for why traffic was routed a certain way

The bank gets blamed when international transfers fail, even when the root cause is a provider, payout network, local rail, or settlement issue.

## 3. Goals

1. Give banks one control room across all connected IMTO providers.
2. Route every transaction through the best eligible provider based on bank policy and live performance.
3. Detect degraded providers, corridors, destination banks, and rails early.
4. Let operators safely switch or split traffic during incidents.
5. Provide clear FX, cost, speed, and reliability comparison.
6. Provide deep latency and downtime drilldowns.
7. Explain every route decision with an audit trail.
8. Reduce failed/stuck transactions and manual operations effort.

## 4. Non-Goals

The first product will not:

- become a consumer remittance app
- hold customer funds directly unless a future regulated model requires it
- replace licensed IMTOs
- become a public per-transaction auction
- build every African payout method in the MVP
- automate all compliance investigations
- replace the bank's CRM or core banking system
- provide advanced treasury optimization in the first release

## 5. Users and Personas

### Payments Operations Analyst

Needs:

- see route health quickly
- find stuck transfers
- know who owns the next action
- disable degraded providers safely
- export evidence for provider follow-up

### Head of Remittances / Diaspora Banking

Needs:

- increase successful volume
- improve customer experience
- compare providers objectively
- protect revenue during provider incidents
- negotiate with providers using performance data

### Technology / Integration Lead

Needs:

- reduce one-off integrations
- use clean APIs and event contracts
- monitor adapter health
- enforce idempotency and reliability patterns

### Risk and Compliance Lead

Needs:

- approved providers only
- auditable policy decisions
- immutable change history
- clear exception ownership

### Treasury / Finance

Needs:

- FX rate visibility
- fees and spread comparison
- settlement aging
- reconciliation exceptions
- prefund exposure where applicable

### Executive Sponsor

Needs:

- single view of operational health
- proof that reliability improved
- provider accountability
- expansion path across corridors and countries

## 6. Product Scope

### MVP Scope

1. Bank onboarding
   - bank profile
   - users and roles
   - providers and corridors
   - bank routing policy

2. Provider and route registry
   - provider capabilities
   - supported corridors
   - payout methods
   - limits
   - fees and FX configuration
   - live status

3. Transaction intake
   - REST API
   - idempotency key
   - required field validation
   - canonical transaction state
   - route decision record

4. Routing decision engine
   - eligibility filtering
   - simple weighted scoring
   - fallback list
   - rejected route reasons
   - policy version tracking

5. Provider adapter framework
   - sandbox adapter
   - one real provider adapter for pilot
   - callback/webhook handling
   - provider status normalization

6. Health and latency engine
   - active probes
   - passive health from transaction outcomes
   - P50/P95/P99 latency
   - downtime event detection
   - provider/corridor state

7. Control room UI
   - global health
   - active degraded routes
   - stuck transactions
   - route recommendations
   - traffic split

8. Corridor detail UI
   - provider ranking
   - success rate
   - latency
   - FX/cost comparison
   - current policy

9. Transaction trace UI
   - lifecycle timeline
   - selected route
   - rejected routes
   - current state
   - owner of next action

10. Route configuration UI
   - enable/disable provider
   - fallback order
   - traffic split presets
   - preview impact
   - audit reason

11. FX and cost board
   - current rates
   - stale-rate detection
   - fee/spread display
   - effective cost comparison

12. Basic reconciliation
   - file import
   - reference matching
   - exception grouping
   - settlement aging

## 7. User Stories

### Control Room

As a payments operations analyst, I want to see all degraded providers and corridors on one screen so I know where to act first.

Acceptance criteria:

- screen shows global health score
- degraded routes are ranked by severity
- each route shows provider, corridor, error rate, latency, volume impact, and recommendation
- data freshness is visible

### Route Decision

As a bank, I want the platform to select the best eligible route so transactions avoid providers that are slow, expensive, unhealthy, or out of policy.

Acceptance criteria:

- system filters out ineligible routes before scoring
- selected route is stored with score and policy version
- rejected routes include rejection reasons
- decision can be replayed for audit

### Manual Switch

As an operator, I want to disable a provider for a corridor so new traffic stops going to a degraded route.

Acceptance criteria:

- operator can disable by provider, corridor, payout method, destination bank, or amount band
- UI shows estimated affected traffic
- operator must enter a reason
- change is versioned and auditable
- new traffic routes to next eligible provider

### Circuit Breaker

As a bank, I want the platform to automatically pause new traffic to a route when its reliability drops below threshold.

Acceptance criteria:

- threshold can be configured
- route state changes to degraded or blocked
- traffic shifts only for new transactions unless safe failover applies
- alert/event is emitted
- recovery state is tracked

### Latency Drilldown

As an operations manager, I want to see where transaction latency is happening so I can identify whether the issue is provider, bank, local rail, compliance, webhook, or settlement.

Acceptance criteria:

- end-to-end latency is visible
- step-level latency waterfall is visible
- drilldown supports provider, corridor, destination bank, payout method, and time window
- downtime timeline shows incidents and operator actions

### FX and Cost

As a remittance product owner, I want to compare FX and fees across providers so routing decisions balance cost with reliability and speed.

Acceptance criteria:

- provider rate, fee, spread, and effective cost are visible
- stale rates are flagged
- UI explains when the cheapest route was not selected due to risk or SLA

## 8. Functional Requirements

### Transaction Intake

- Accept transaction creation requests through REST API.
- Require bank ID, sender country, receiver country, amount, currency, payout method, beneficiary destination, provider constraints, and idempotency key.
- Return switch reference and selected route.
- Prevent duplicate processing for repeated idempotency keys.
- Emit transaction lifecycle event.

### Eligibility Engine

Must reject routes when:

- provider is disabled
- corridor unsupported
- payout method unsupported
- amount outside limits
- FX rate stale beyond policy
- provider/license/policy status blocks route
- prefund/liquidity state blocks route
- circuit breaker blocks route
- compliance state requires manual review

### Scoring Engine

Must support weighted scoring using:

- success rate
- P95 time-to-credit
- cost
- FX quality
- provider uptime
- settlement reliability
- manual intervention rate
- support/resolution score

### Provider Health

Must collect:

- API availability
- timeout rate
- provider callback lag
- transaction success/failure
- time-to-credit
- stuck transactions
- settlement/reconciliation exceptions

### Configuration

Must support:

- provider enable/disable
- route priority
- traffic split
- fallback order
- circuit breaker threshold
- scoring weights
- policy preview
- policy version history
- rollback

### Audit

Must record:

- route decision inputs
- selected route
- rejected routes
- scores
- actor/user
- policy version
- config changes
- approvals
- circuit breaker events

## 9. Non-Functional Requirements

### Performance

- route decision p95 under 20 ms excluding external systems
- transaction intake p95 under 100 ms excluding external systems
- health state update visible within 5 seconds
- dashboard initial load under 3 seconds on corporate network
- charts remain usable with 30 days of rollup data
- tables virtualized above 1,000 rows

### Reliability

- all transaction writes idempotent
- provider adapter failures isolated from routing core
- no duplicate payouts from unsafe failover
- durable event replay for lifecycle events
- clear degraded/stale-data UI state

### Security

- SSO/OIDC integration path
- RBAC
- maker-checker approvals
- signed webhooks where supported
- audit exports
- encryption in transit
- secrets not stored in source control

### Compatibility

- REST/JSON
- webhooks
- SFTP/CSV
- SOAP/XML adapter path
- CSV/XLSX exports
- Chrome, Edge, Safari, Firefox
- websocket fallback to polling

## 10. UI Requirements

### Visual Direction

Premium fintech control room:

- calm dark operational theme for MVP
- high contrast text
- meaningful status colors only
- dense but readable layouts
- charts that feel alive without noise
- no decoration that competes with incident signals

### MVP Screens

1. Control Room
2. Corridor Detail
3. Transaction Trace
4. Route Configuration
5. FX and Cost Board
6. Latency Drilldown
7. Reconciliation Exceptions
8. Audit Log

### Key Components

- health badge
- route score chip
- latency sparkline
- corridor matrix
- provider scorecard
- transaction timeline
- traffic split control
- policy diff viewer
- FX comparison table
- incident banner
- audit drawer

## 11. Data Model Summary

Core entities:

- Bank
- User
- Role
- Provider
- Corridor
- Route
- RoutePolicy
- PolicyVersion
- Transaction
- TransactionEvent
- RouteDecision
- ProviderHealthSample
- CircuitBreakerState
- FxRate
- ReconciliationFile
- ReconciliationItem
- AuditEvent

## 12. Success Metrics

Product success:

- failed/stuck transactions reduced
- average and P95 time-to-credit improved
- provider downtime detected faster
- route changes completed safely faster
- manual follow-up reduced
- reconciliation exceptions reduced or resolved faster
- cost per successful transaction reduced where SLA allows

Pilot report metrics:

- transaction volume
- transaction value
- success rate by provider/corridor
- latency by provider/corridor/step
- traffic shifted due to degradation
- failures avoided
- route recommendations accepted/ignored
- FX/cost impact
- incident time-to-detect and time-to-resolve

## 13. Launch Plan

### Internal Alpha

- sandbox providers
- simulated latency/failure conditions
- policy simulator
- UI with generated data

### Bank Pilot

- one anchor bank
- 3-5 routes
- one corridor cluster
- account payout first
- manual switching before full automation
- shadow routing before production routing changes

### Expansion

- add real provider adapters
- add ClickHouse analytics when justified
- add cash pickup and wallets
- add multi-bank deployment profile
- add Africa-to-Africa corridors

## 14. Open Questions

- Which anchor bank will provide the first transaction and reconciliation samples?
- Which 3-5 providers should be included in the first real pilot?
- Will the bank allow active provider probes, or only passive health from transactions?
- What compliance decisions must happen before route selection?
- Which FX source is authoritative for the bank?
- What is the safe failover boundary for each provider?
- What maker-checker approvals does the bank require?

