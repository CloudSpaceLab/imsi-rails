# imsi-rails GitHub Issues and Milestones

This backlog is written so it can be copied into GitHub issues or created with the GitHub CLI.

Created in GitHub on 2026-05-19 for `CloudSpaceLab/imsi-rails` as issues #1-#22 across milestones M0-M5.

Updated on 2026-05-19 with design-language implementation issues #23-#24.

Updated on 2026-05-20 with UI/UX revamp epic #37 and child issues #38-#51 across milestone `M1.5: UI/UX Revamp`.

Updated on 2026-05-21 after a deeper backend review with transfer orchestration milestone `M1.1: Transfer Orchestration Backbone` and issues #52-#70.

## Milestones

### M0: Foundation and Landing Page

Goal:

- establish product narrative, architecture direction, PRD, and landing page

Exit:

- docs and landing page are committed
- stakeholders can understand the product and pilot path

### M1: Pilot Core

Goal:

- accept transactions, make route decisions, and trace lifecycle events

Exit:

- sandbox transaction can be routed and inspected in UI

### M1.1: Transfer Orchestration Backbone

Goal:

- turn route selection into a durable transfer orchestration layer with attempts, provider operations, callbacks, outbox events, exceptions, reconciliation states, and evidence reports

Exit:

- provider attempts are durable and state transitions are owned by an orchestrator
- provider callbacks are ingested through a durable inbox
- unknown outcomes and duplicate payout risk are modeled explicitly
- reports and reconciliation work from immutable events rather than manual evidence gathering

### M1.5: UI/UX Revamp

Goal:

- rebuild the product UI from a static light admin prototype into a premium dark-first bank operations control room

Exit:

- primary screens are workflow-based, audit-safe, responsive, keyboard reachable, and screenshot-verified
- issue #37 tracks the epic and issues #38-#51 track the implementation slices

### M2: Reliability Intelligence

Goal:

- monitor provider/corridor health and expose latency/downtime drilldowns

Exit:

- degradation is detected and visible with root-cause context

### M3: Switching and Configuration

Goal:

- safely change route behavior through policy, traffic splits, and circuit breakers

Exit:

- bank operations can preview, approve, apply, and roll back route changes

### M4: FX, Cost, and Reconciliation

Goal:

- compare FX/cost by provider and detect settlement/reconciliation exceptions

Exit:

- bank can see effective cost and unresolved settlement exceptions by route

### M5: Bank Pilot Hardening

Goal:

- prepare for controlled bank pilot

Exit:

- security, performance, deployment, and operating runbooks are ready

## Issues

### 52. ORCH-00 Backend orchestration epic: turn route selection into transfer rails control layer

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `architecture`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/52

Child issues:

- #53 ORCH-01 Persist transfers, attempts, provider operations, timeline spans, and outbox schema
- #54 ORCH-02 Implement transfer orchestrator state machine with attempts and finality boundaries
- #55 ORCH-03 Define provider adapter interface and runtime execution policy
- #56 ORCH-04 Expand sandbox adapter to simulate real IMTO failure modes
- #57 ORCH-05 Add durable webhook inbox and callback normalization pipeline
- #58 ORCH-06 Add provider status poller and unknown-outcome resolver
- #59 ORCH-07 Align OpenAPI with canonical transfer rails contract
- #60 ORCH-08 Integrate capability, liquidity, FX quote, and settlement snapshots into route policy
- #61 ORCH-09 Replace pilot in-memory stores with MariaDB repositories
- #62 ORCH-10 Implement transactional outbox and NATS JetStream publication
- #63 REL-01 Feed dashboard analytics from orchestration events instead of seeded metrics
- #64 EXC-01 Implement exception engine with ownership, SLA clocks, and recommended actions
- #65 INC-01 Implement incident lifecycle and postmortem evidence model
- #66 REC-01 Implement automatic reconciliation state machine and evidence matching
- #67 REP-01 Implement report generation and evidence packs
- #68 ADAPT-01 Build Thunes-style adapter simulation
- #69 ADAPT-02 Build MoneyGram payout-partner simulation
- #70 ADAPT-03 Build PAPSS/ISO 20022-style rail simulation

### 53. ORCH-01 Persist transfers, attempts, provider operations, timeline spans, and outbox schema

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `data-model`, `persistence`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/53

### 54. ORCH-02 Implement transfer orchestrator state machine with attempts and finality boundaries

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `payments`, `orchestration`, `reliability`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/54

### 55. ORCH-03 Define provider adapter interface and runtime execution policy

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `adapter`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/55

### 56. ORCH-04 Expand sandbox adapter to simulate real IMTO failure modes

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `adapter`, `testing`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/56

### 57. ORCH-05 Add durable webhook inbox and callback normalization pipeline

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `webhooks`, `adapter`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/57

### 58. ORCH-06 Add provider status poller and unknown-outcome resolver

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `adapter`, `reliability`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/58

### 59. ORCH-07 Align OpenAPI with canonical transfer rails contract

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `api`, `documentation`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/59

### 60. ORCH-08 Integrate capability, liquidity, FX quote, and settlement snapshots into route policy

Milestone: M3: Switching and Configuration

Labels: `backend`, `routing`, `policy`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/60

### 61. ORCH-09 Replace pilot in-memory stores with MariaDB repositories

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `persistence`, `data-model`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/61

### 62. ORCH-10 Implement transactional outbox and NATS JetStream publication

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `reliability`, `persistence`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/62

### 63. REL-01 Feed dashboard analytics from orchestration events instead of seeded metrics

Milestone: M2: Reliability Intelligence

Labels: `backend`, `analytics`, `observability`, `reliability`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/63

### 64. EXC-01 Implement exception engine with ownership, SLA clocks, and recommended actions

Milestone: M2: Reliability Intelligence

Labels: `backend`, `operations`, `reliability`, `orchestration`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/64

### 65. INC-01 Implement incident lifecycle and postmortem evidence model

Milestone: M2: Reliability Intelligence

Labels: `backend`, `incident`, `operations`, `reliability`, `P1`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/65

### 66. REC-01 Implement automatic reconciliation state machine and evidence matching

Milestone: M4: FX, Cost, and Reconciliation

Labels: `backend`, `reconciliation`, `operations`, `P0`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/66

### 67. REP-01 Implement report generation and evidence packs

Milestone: M4: FX, Cost, and Reconciliation

Labels: `backend`, `reports`, `audit`, `operations`, `P1`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/67

### 68. ADAPT-01 Build Thunes-style adapter simulation

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `adapter`, `testing`, `P1`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/68

### 69. ADAPT-02 Build MoneyGram payout-partner simulation

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `adapter`, `testing`, `P1`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/69

### 70. ADAPT-03 Build PAPSS/ISO 20022-style rail simulation

Milestone: M1.1: Transfer Orchestration Backbone

Labels: `backend`, `adapter`, `testing`, `reconciliation`, `P1`

GitHub: https://github.com/CloudSpaceLab/imsi-rails/issues/70

### 37. UI/UX revamp epic: rebuild bank-grade control room experience

Milestone: M1.5: UI/UX Revamp

Labels: `frontend`, `ui`, `design`, `P0`

Child issues:

- #38 UXR-01 Frontend ADR and UX architecture freeze
- #39 UXR-02 Dark-first product design tokens
- #40 UXR-03 Core UI primitive rebuild
- #41 UXR-04 App shell and route-based navigation
- #42 UXR-05 Control Room redesign
- #43 UXR-06 Corridor matrix and provider ranking redesign
- #44 UXR-07 Safe route configuration workflow
- #45 UXR-08 Policy simulator and shadow report redesign
- #46 UXR-09 FX and cost board redesign
- #47 UXR-10 Latency and downtime diagnostics redesign
- #48 UXR-11 Transaction trace and audit drawer redesign
- #49 UXR-12 Incident detail and evidence export UX
- #50 UXR-13 UI data contracts and scenario fixtures
- #51 UXR-14 Accessibility, visual regression, and performance QA

### 1. Create architecture ADRs and performance budgets

Milestone: M0: Foundation and Landing Page

Labels: `architecture`, `performance`, `documentation`

Priority: P0

Description:

Document the key technology decisions and the performance budgets that all implementation must respect.

Acceptance criteria:

- ADR exists for backend language/runtime
- ADR exists for eventing/message bus
- ADR exists for data stores
- ADR exists for frontend framework and charting
- backend and frontend performance budgets are documented
- hot path constraints are explicit

### 2. Build static landing page

Milestone: M0: Foundation and Landing Page

Labels: `frontend`, `marketing`, `design`

Priority: P0

Description:

Create a static landing page that positions imsi-rails as bank-facing IMTO reliability infrastructure.

Acceptance criteria:

- page opens directly from `landing/index.html`
- no build step required
- hero clearly says imsi-rails is not a remittance app
- visual shows routing/control-room concept
- page is responsive
- page avoids external dependencies

### 3. Define canonical transaction lifecycle model

Milestone: M1: Pilot Core

Labels: `backend`, `data-model`, `payments`

Priority: P0

Description:

Define transaction states and events used across providers, banks, UI, analytics, and reconciliation.

Acceptance criteria:

- canonical state enum exists
- lifecycle event schema exists
- state transition rules are documented
- unsafe failover states are identified
- sample lifecycle fixtures exist

### 4. Implement transaction intake API

Milestone: M1: Pilot Core

Labels: `backend`, `api`, `payments`

Priority: P0

Description:

Build the API endpoint banks use to submit transactions for routing.

Acceptance criteria:

- OpenAPI spec exists
- API accepts required transaction fields
- idempotency key is required
- duplicate idempotency requests do not duplicate processing
- transaction is persisted
- lifecycle event is emitted

### 5. Build provider and route registry

Milestone: M1: Pilot Core

Labels: `backend`, `data-model`, `routing`

Priority: P0

Description:

Create the registry for providers, corridors, payout methods, limits, cost inputs, FX freshness, and policy status.

Acceptance criteria:

- provider model exists
- corridor model exists
- route model exists
- limits and payout methods are represented
- route can be enabled/disabled by policy
- registry can be cached for hot path decisions

### 6. Implement route decision audit log

Milestone: M1: Pilot Core

Labels: `backend`, `audit`, `routing`

Priority: P0

Description:

Persist every route decision with selected route, rejected routes, scores, policy version, and inputs.

Acceptance criteria:

- route decision record is stored
- selected route and score are stored
- rejected routes and reasons are stored
- policy version is stored
- decision can be retrieved by transaction ID

### 7. Create sandbox provider adapter

Milestone: M1: Pilot Core

Labels: `backend`, `adapter`, `testing`

Priority: P0

Description:

Build a fake provider adapter that can simulate success, delay, timeout, rejection, duplicate callback, and stale status.

Acceptance criteria:

- adapter supports configurable latency
- adapter supports configurable failure rate
- adapter emits callbacks
- adapter supports duplicate callback simulation
- adapter can be used in demos and tests

### 8. Build control room UI shell

Milestone: M1: Pilot Core

Labels: `frontend`, `ui`, `operations`

Priority: P0

Description:

Create the first authenticated product UI shell for live reliability monitoring.

Acceptance criteria:

- global health summary exists
- corridor/provider status grid exists
- active incident panel exists
- stuck transaction counter exists
- data freshness indicator exists
- responsive desktop/tablet layout exists

### 9. Build transaction trace UI

Milestone: M1: Pilot Core

Labels: `frontend`, `ui`, `payments`

Priority: P0

Description:

Create a transaction timeline view that shows routing, state changes, provider events, and current owner.

Acceptance criteria:

- searchable by switch reference
- lifecycle timeline is visible
- selected route is visible
- rejected routes are visible
- current owner/action is visible
- audit drawer is accessible

### 10. Implement health event ingestion

Milestone: M2: Reliability Intelligence

Labels: `backend`, `observability`, `reliability`

Priority: P0

Description:

Ingest active and passive health signals from providers, adapters, transactions, and internal services.

Acceptance criteria:

- health sample schema exists
- provider API status can be recorded
- timeout/error rate can be recorded
- callback lag can be recorded
- transaction outcome signals feed health state
- events are emitted for state changes

### 11. Build provider/corridor scorecards

Milestone: M2: Reliability Intelligence

Labels: `analytics`, `frontend`, `reliability`

Priority: P1

Description:

Show objective provider performance by corridor.

Acceptance criteria:

- success rate by provider/corridor
- P50/P95/P99 latency by provider/corridor
- stuck rate by provider/corridor
- settlement/reconciliation exception count
- scorecard time window selector

### 12. Build latency waterfall and downtime timeline

Milestone: M2: Reliability Intelligence

Labels: `frontend`, `analytics`, `observability`

Priority: P0

Description:

Create drilldowns that show where delays happen and when downtime begins/ends.

Acceptance criteria:

- end-to-end latency visible
- step-level latency visible
- downtime events visible on timeline
- operator actions visible on timeline
- filter by provider, corridor, destination bank, and time window

### 13. Implement circuit breaker state machine

Milestone: M2: Reliability Intelligence

Labels: `backend`, `reliability`, `routing`

Priority: P0

Description:

Create circuit breaker states and transitions for route health.

Acceptance criteria:

- states include healthy, degraded, blocked, recovery testing
- thresholds can be configured
- state changes are persisted
- state changes emit events
- routing engine can read breaker state from cache

### 14. Implement eligibility engine

Milestone: M3: Switching and Configuration

Labels: `backend`, `routing`, `policy`

Priority: P0

Description:

Filter all possible routes before scoring.

Acceptance criteria:

- rejects unsupported corridor
- rejects unsupported payout method
- rejects amount outside limits
- rejects disabled provider
- rejects stale FX where policy requires fresh rate
- rejects circuit-breaker-blocked route
- returns rejection reasons

### 15. Implement weighted route scoring

Milestone: M3: Switching and Configuration

Labels: `backend`, `routing`, `optimization`

Priority: P0

Description:

Score eligible routes by reliability, speed, cost, FX, liquidity, and operations burden.

Acceptance criteria:

- scoring weights are configurable
- score output is explainable
- selected route is deterministic for same inputs
- score components are stored in audit log
- unit tests cover scoring edge cases

### 16. Build route configuration UI

Milestone: M3: Switching and Configuration

Labels: `frontend`, `policy`, `operations`

Priority: P0

Description:

Allow bank operations to safely configure provider toggles, fallback order, traffic split, and scoring weights.

Acceptance criteria:

- provider enable/disable control exists
- fallback order editor exists
- traffic split presets exist
- preview impact exists
- required change reason exists
- change history is visible

### 17. Add policy simulator and shadow routing

Milestone: M3: Switching and Configuration

Labels: `backend`, `frontend`, `routing`

Priority: P1

Description:

Let users test proposed policy changes against sample or historical transactions before activation.

Acceptance criteria:

- sample transaction can be simulated
- current vs proposed route is compared
- rejected routes are explained
- shadow policy can run without affecting production
- shadow results are reportable

### 18. Build FX and cost board

Milestone: M4: FX, Cost, and Reconciliation

Labels: `frontend`, `fx`, `analytics`

Priority: P0

Description:

Show provider rates, fees, spreads, effective cost, and stale-rate alerts.

Acceptance criteria:

- FX rates visible by corridor/currency pair
- rate timestamp visible
- stale rate warning exists
- fee/spread/effective cost visible
- cheapest route vs recommended route explanation exists

### 19. Implement reconciliation import and exception matching

Milestone: M4: FX, Cost, and Reconciliation

Labels: `backend`, `reconciliation`, `operations`

Priority: P0

Description:

Import provider/bank files and identify unmatched, delayed, duplicate, and mismatched records.

Acceptance criteria:

- CSV import supported
- matching by reference, amount, currency, beneficiary
- unmatched items visible
- duplicates visible
- aging buckets visible
- exceptions grouped by provider and reason

### 20. Add RBAC, audit exports, and maker-checker approvals

Milestone: M5: Bank Pilot Hardening

Labels: `security`, `audit`, `enterprise`

Priority: P0

Description:

Implement bank-grade permissions and approval controls for sensitive actions.

Acceptance criteria:

- roles defined
- sensitive actions require permission
- policy changes can require maker-checker
- audit log export exists
- unauthorized actions are blocked server-side

### 21. Add performance and load test suite

Milestone: M5: Bank Pilot Hardening

Labels: `performance`, `testing`, `backend`

Priority: P0

Description:

Create repeatable performance tests for transaction intake, route decision, dashboard APIs, and health event ingestion.

Acceptance criteria:

- k6 scripts exist
- route scoring microbenchmark exists
- thresholds are defined
- CI can run a lightweight performance smoke test
- failure creates actionable output

### 22. Create bank pilot deployment guide and runbooks

Milestone: M5: Bank Pilot Hardening

Labels: `devops`, `documentation`, `pilot`

Priority: P0

Description:

Document how to deploy, operate, monitor, back up, and recover the pilot environment.

Acceptance criteria:

- deployment guide exists
- environment variables documented
- backup/restore plan exists
- incident runbook exists
- rollback plan exists
- provider escalation workflow documented

### 23. Build design-system primitives from brand language

Milestone: M1: Pilot Core

Labels: `frontend`, `ui`, `design`

Priority: P0

Description:

Create the first product design-system primitives from the users, brand, and design-language guide.

Acceptance criteria:

- health badge component exists
- route score chip component exists
- data freshness indicator exists
- transaction timeline primitives exist
- corridor status cell exists
- status colors and labels match brand language
- components support keyboard focus and accessible labels

### 24. Implement operational copy and state taxonomy

Milestone: M1: Pilot Core

Labels: `frontend`, `ui`, `documentation`

Priority: P1

Description:

Translate the brand voice and operational state taxonomy into reusable UI copy patterns.

Acceptance criteria:

- status taxonomy includes healthy, watch, degraded, blocked, recovery, unknown, and stale data
- action labels are standardized for shift traffic, pause new traffic, preview policy, rollback, and export evidence
- route rejection reasons use consistent copy
- recommendation copy explains why an action is suggested
- copy avoids generic remittance app language
