# imsi-rails Delivery Masterplan

## Product Mission

Make international money transfer reliability measurable, configurable, and self-correcting for banks.

Banks should onboard their IMTO providers into `imsi-rails`, see provider performance in one place, and let the best eligible provider win each transaction based on reliability, speed, cost, FX quality, compliance readiness, liquidity, and live rail health.

## Delivery Strategy

Ship the smallest bank-grade control plane that proves value:

1. Monitor every connected route.
2. Trace every transaction.
3. Compare providers by corridor.
4. Configure route priorities safely.
5. Switch traffic away from degraded routes.
6. Prove the impact with data.

Do not build a consumer remittance product. Do not build a provider auction marketplace. Do not build broad BI before the routing and reliability loop works.

## Masterplan Review Status

Reviewed against the PRD and project goals on 2026-05-19.

Status: complete and intact, with the following additions required and now incorporated:

- exact user segmentation beyond generic bank personas
- brand and design language for an internationally competitive product
- quality gates for UI, copy, and design-system decisions
- stronger connection between buyer value, operator workflows, and implementation milestones

### Alignment Matrix

| PRD goal | Masterplan coverage | Status |
| --- | --- | --- |
| One control room across IMTO providers | Control Room, Web Application, Reliability Intelligence | Covered |
| Best eligible route per transaction | Routing API, Policy Engine, Switching milestone | Covered |
| Detect degraded providers/corridors early | Health Engine, latency/downtime drilldowns, circuit breakers | Covered |
| Safe manual and automatic switching | Policy Engine, traffic split, maker-checker, rollback | Covered |
| FX, cost, speed, reliability comparison | FX/Cost milestone and corridor detail screens | Covered |
| Deep latency and downtime drilldowns | Reliability Intelligence and UI delivery plan | Covered |
| Explain every route decision | Route decision audit, transaction trace, audit log | Covered |
| Reduce stuck/failed transactions | Pilot report, success metrics, health engine, switching | Covered |
| Lightweight/fast architecture | MVP architecture and engineering principles | Covered |
| Premium, internationally competitive UX | Brand/design language and UI quality gates | Added |

### Non-Negotiables

The masterplan should be rejected or revised if future work violates these principles:

1. The product is bank infrastructure, not a consumer remittance app.
2. The bank controls routing policy; providers win traffic by measured performance.
3. The routing hot path stays small, fast, and explainable.
4. Every important decision is auditable.
5. The UI must help operators act correctly under pressure.
6. Design polish must improve clarity, not hide operational truth.

## System Shape

### Core Services

1. Routing API
   - transaction intake
   - idempotency
   - route eligibility
   - route scoring
   - route decision audit
   - status lookup

2. Adapter Worker
   - provider submissions
   - provider callbacks/webhooks
   - SFTP/CSV polling
   - provider status normalization

3. Health Engine
   - active probes
   - passive health from transaction outcomes
   - latency sampling
   - downtime event detection
   - circuit breaker state

4. Policy Engine
   - bank rules
   - provider enablement
   - corridor controls
   - scoring weights
   - traffic split
   - maker-checker approval

5. Reconciliation Worker
   - provider file import
   - bank posting file import
   - matching and exception detection

6. Web Application
   - control room
   - corridors
   - transactions
   - FX and cost
   - route configuration
   - latency drilldowns
   - audit

## MVP Architecture

Keep the MVP compact:

- Rust + Tokio + Axum for backend services
- PostgreSQL for system of record and pilot analytics rollups
- NATS for event flow and live dashboard updates
- SvelteKit for app UI
- uPlot for dense latency charts
- TanStack Table for large operational tables
- OpenTelemetry instrumentation from day one

Add ClickHouse only after the pilot proves event volume and drilldown needs exceed PostgreSQL rollups.

## Product Milestones

### Milestone 0: Foundation and Narrative

Outcome:

- repo has product docs, architecture, PRD, issue plan, and landing page
- stakeholders understand what is being built and what is not being built

Deliverables:

- architecture document
- PRD
- user, brand, and design-language guide
- landing page
- milestone/issue backlog
- performance budgets

Exit criteria:

- docs are reviewed
- landing page opens locally
- GitHub milestones/issues are ready or created
- brand/design language is accepted as the standard for copy, UI, and components

### Milestone 1: Pilot Core

Outcome:

- the bank can send transactions into the platform and see lifecycle traces

Deliverables:

- transaction intake API
- canonical transaction model
- provider/route registry
- route decision audit log
- first sandbox provider adapter
- control room UI shell
- transaction trace UI
- first design-system primitives for health state, route score, transaction timeline, and data freshness

Exit criteria:

- transaction can be accepted, routed, traced, and inspected
- every route decision has an explainable audit record
- UI shows live transaction states from event stream
- UI shell follows the design-language quality gates

### Milestone 2: Reliability Intelligence

Outcome:

- the platform detects route degradation and shows latency/downtime root cause

Deliverables:

- health engine
- provider/corridor scorecards
- latency waterfall
- downtime timeline
- circuit breaker state model
- provider health dashboard
- alert rules
- latency and downtime visual language

Exit criteria:

- provider degradation is visible within 5 seconds of threshold breach
- latency can be split by provider, corridor, destination bank, step, and payout method
- circuit breaker can mark a route degraded without changing code
- each degradation view explains time window, freshness, severity, and recommended action

### Milestone 3: Switching and Configuration

Outcome:

- operations teams can safely change traffic allocation and route priorities

Deliverables:

- eligibility engine
- weighted route scoring
- fallback route list
- provider enable/disable controls
- traffic split controls
- policy simulator
- shadow routing
- maker-checker approval for sensitive changes
- traffic split and policy diff UI

Exit criteria:

- operator can preview the impact of a route change before saving
- route changes are versioned and auditable
- degraded route can be bypassed for new traffic
- fallback routing works with idempotency protection
- every traffic-changing action has preview, reason capture, approval rules, and rollback path

### Milestone 4: FX, Cost, and Reconciliation

Outcome:

- bank can compare cost, speed, FX, and settlement quality across providers

Deliverables:

- FX rate ingestion
- stale-rate detection
- cost comparison view
- provider fee/spread model
- reconciliation import
- unmatched item view
- duplicate/mismatch detection
- settlement aging
- FX comparison design system component

Exit criteria:

- UI shows effective cost by provider/corridor
- route engine can avoid stale FX data
- reconciliation exceptions are grouped by provider and reason
- UI explains when the cheapest route is not selected because risk, latency, or policy is worse

### Milestone 5: Bank Pilot Hardening

Outcome:

- platform is ready for a controlled bank pilot

Deliverables:

- SSO/OIDC integration path
- RBAC roles
- audit exports
- load test suite
- dashboard performance audit
- disaster recovery runbook
- deployment guide
- incident workflow
- design QA and accessibility pass

Exit criteria:

- pilot performance budgets pass
- critical actions require correct permissions
- all sensitive config changes are audited
- rollback path is documented and tested
- critical operator workflows are keyboard-accessible and screen states are readable under incident pressure

## User Model

Full reference: [Users, Brand, and Design Language](users-brand-design-language.md).

MVP user priority:

1. Daily Operator
2. Head of Remittances / Economic Buyer
3. Technical Integrator
4. Risk and Compliance
5. Executive Buyer

### Primary Users

| User | Core question | Must-have product surfaces |
| --- | --- | --- |
| Payments Operations Analyst | What is broken, who owns it, and what should I do next? | Control Room, Transaction Trace, Incident Detail, Route Configuration |
| Head of Remittances / Diaspora Banking | Which providers deserve more traffic and why? | Corridor Detail, Provider Scorecards, FX/Cost Board, Pilot Report |
| Technical Integrator | Can this integrate and run reliably in our environment? | API Docs, Adapter Health, Deployment Guide, Observability |
| Risk/Compliance Lead | Can we prove every route and policy decision was approved? | Audit Log, Route Decision Detail, Policy Diff, Exports |
| Executive Sponsor | Did reliability, customer experience, and revenue improve? | Executive Summary, Pilot Report, Provider Scorecard Summary |

### Design Implication

The product must support two modes without becoming two products:

- operational mode for users who act during incidents
- executive/evidence mode for users who review performance and approve expansion

The UI should default to operational clarity, then provide executive summaries through rollups and exports.

## Brand and Design Language

Full reference: [Users, Brand, and Design Language](users-brand-design-language.md).

### Brand Idea

`imsi-rails` is the reliability layer beneath international money movement.

### Product Phrase

IMTO reliability infrastructure for banks.

### Brand Promise

Every connected IMTO rail becomes observable, comparable, and safely switchable.

### Design North Star

Premium operational intelligence.

The product should feel like a high-end control room for international money transfer reliability: alive, calm, precise, and safe.

### Copy Principles

- lead with operational outcomes
- use evidence-led language
- distinguish selected, recommended, eligible, rejected, degraded, and stale states
- avoid hype and generic fintech claims
- explain uncertainty instead of hiding it

### Visual Principles

- dark operational canvas for MVP
- color used for meaning, not decoration
- dense but readable layouts
- strong numerals and visible time windows
- crisp status states and route scores
- motion only for meaningful state change

### Component Priorities

1. Health Badge
2. Route Score Chip
3. Corridor Matrix
4. Transaction Timeline
5. FX Comparison Table
6. Traffic Split Control
7. Policy Diff Viewer
8. Audit Drawer

### UI Quality Gates

A screen is not done until:

- the primary user can answer the screen's main question in under 5 seconds
- live/stale data state is obvious
- all critical numbers have units and time windows
- every recommended action explains why
- every destructive or traffic-changing action has preview and rollback path
- keyboard use works for critical controls
- mobile/tablet executive read-only view does not break
- performance budget is respected

## UI Delivery Plan

### Landing Page

Purpose:

- make the product legible to bank executives, remittance teams, and technical sponsors

Requirements:

- static and dependency-free
- premium control-room aesthetic
- clear value proposition
- no generic remittance app messaging
- first viewport signals the actual product

### Product UI

Build in this order:

1. Control Room
   - live health
   - active incidents
   - traffic split
   - recommended action

2. Transaction Trace
   - lifecycle timeline
   - selected route
   - rejected routes
   - current owner

3. Corridor Detail
   - provider ranking
   - latency
   - failure rate
   - FX/cost comparison

4. Route Configuration
   - provider toggles
   - fallback order
   - traffic split
   - preview and approval

5. Latency Drilldown
   - step-level waterfall
   - downtime timeline
   - heatmaps

6. Reconciliation
   - imports
   - exceptions
   - aging

## Engineering Principles

1. Hot path stays small.
   - Route decisions must use cached policy, cached capability, cached health, and fresh enough FX snapshots.

2. Everything important emits an event.
   - The dashboard, audit trail, scorecards, and incident analysis depend on clean event discipline.

3. Analytics are outside the hot path.
   - ClickHouse or rollup queries must never block transaction routing.

4. Adapters are isolated.
   - A slow or broken provider adapter must not degrade the routing core.

5. Configuration is versioned.
   - No route policy change should happen without a policy version, actor, timestamp, reason, and rollback path.

6. The UI is operational software.
   - It must help people make correct decisions under pressure.

## Delivery Risks

| Risk | Mitigation |
| --- | --- |
| Provider APIs are inconsistent | Normalize through adapter contracts and canonical states |
| Bank integration takes too long | Start with sandbox, file samples, and one posting path |
| UI becomes overloaded | Build control room first, deep drilldowns second |
| Routing becomes too clever too early | Start with eligibility plus simple weighted scoring |
| Analytics stack becomes heavy | Use PostgreSQL rollups first, add ClickHouse when justified |
| Neutrality is questioned | Make routing policy bank-controlled and audit every decision |
| Auto-switching causes duplicate payouts | Use idempotency, pre-finality failover rules, and clear safe/unsafe states |

## Pilot Success Report

At the end of the pilot, produce a report showing:

- transactions routed
- success rate by provider and corridor
- P50/P95/P99 time-to-credit
- stuck transaction reduction
- failures avoided by switching
- provider downtime detected
- traffic shifted due to degradation
- cost/FX impact
- reconciliation exceptions by provider
- operator actions and time-to-resolution

This report becomes the bank expansion sales asset.
