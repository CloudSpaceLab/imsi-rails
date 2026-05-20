# UI/UX Refactor Plan

Review date: 2026-05-20

Repository: `CloudSpaceLab/imsi-rails`

Reviewed evidence:

- GitHub issues #1-#24 via `gh issue list` on 2026-05-20.
- Local implementation docs in `docs/implementation-notes/`.
- Vue app code in `apps/web/src/`.
- Build check: `npm run web:build`.
- Browser screenshots:
  - `imsi-rails-desktop-audit.png`
  - `imsi-rails-mobile-audit.png`

GitHub tracking:

- Epic issue: https://github.com/CloudSpaceLab/imsi-rails/issues/37

## Executive Verdict

The current UI is a working static prototype, not a bank-grade product experience.

It has many of the right nouns: control room, route board, policy simulator, FX board, latency waterfall, provider scorecards, transaction timeline. But the experience does not yet feel like premium operational infrastructure for serious fintech and banking users. It reads as a generic light admin dashboard with stacked panels, weak workflow hierarchy, shallow controls, static data, and little of the "calm command center" quality described in the product brief.

The issue is not a single visual tweak. The product needs an end-to-end UI/UX refactor across:

- design-system foundations
- information architecture
- screen composition
- workflow safety
- status and recommendation copy
- data visualization
- responsive behavior
- accessibility
- frontend architecture and state
- QA gates and visual regression

The refactor should be tracked as a dedicated UI/UX revamp program, not scattered across feature issues.

## Highest-Risk Mismatches

### 1. Product direction and current styling conflict

The design language calls for a premium operational control room with a dark operational canvas:

- `docs/users-brand-design-language.md` defines "premium operational intelligence" and "dark operational canvas".
- `docs/prd.md` requires a calm dark operational theme for MVP.
- `docs/masterplan.md` repeats the dark, high-signal visual direction.

The current working tree moves the app to a light admin palette:

- `apps/web/src/styles.css:2` sets `color-scheme: light`.
- `apps/web/src/styles.css:5-8` sets light background and surface tokens.

This does not automatically make the UI wrong, but it is unapproved drift from the documented product direction and still does not achieve a Fortune-500 fintech standard.

Decision needed: recommit to a dark-first operations theme, or explicitly update the PRD/design language to define a premium light theme. Recommendation: dark-first operations shell, with a light executive/read-only mode later.

### 2. Existing GitHub issues are feature-slice issues, not UX-quality issues

Issues #8, #9, #11, #12, #16, #17, #18, #23, and #24 ask for important surfaces, but their acceptance criteria are mostly existence checks. They do not require the product to pass a modern UX audit.

Example:

- #8 asks for a control room shell with summary, grid, incidents, stuck count, freshness, and responsive layout.
- The code has those artifacts, but the screen still does not answer "what is broken, who owns it, and what should I do next?" in under 5 seconds with premium clarity.

The backlog needs design-quality acceptance criteria, not only component-presence criteria.

### 3. The app is a single long static demo

`apps/web/src/App.vue:84-138` renders every major product area in one long page:

- route board
- recommendation panel
- route configuration
- policy simulator
- FX board
- latency waterfall
- downtime timeline
- provider scorecards
- transaction trail

This creates a scroll-heavy demo, not an operations product. The control room should prioritize the live decision surface and move deep work into corridor, transaction, policy, FX, incident, and audit views.

### 4. Workflow controls are not bank-grade

The UI shows buttons and inputs, but the critical safety workflow is not represented.

Examples:

- `apps/web/src/components/RouteConfigurationPanel.vue:27` uses a generic `Check` button instead of a policy preview and approval flow.
- `apps/web/src/components/RouteConfigurationPanel.vue:35` uses uncontrolled checkboxes for provider state.
- `apps/web/src/components/RouteConfigurationPanel.vue:56-57` uses text `^` and `v` controls instead of a proper ranked fallback editor.
- `apps/web/src/components/RouteConfigurationPanel.vue:102` has a reason field, but no validation, review step, approval path, or rollback surface.
- `apps/web/src/components/UiButton.vue` has no `disabled`, `loading`, `danger`, confirmation, or icon/state contract.

For a switching platform, these are not small UI details. They are trust and risk controls.

### 5. Static data blocks real UX validation

`apps/web/src/data.ts` holds all current product data. The implementation notes repeatedly confirm static prototype scope and no live API:

- `docs/implementation-notes/vue-control-room-shell.md`
- `docs/implementation-notes/route-configuration-ui.md`
- `docs/implementation-notes/latency-downtime-drilldowns.md`
- `docs/implementation-notes/policy-simulator-shadow-routing.md`
- `docs/implementation-notes/fx-cost-board.md`

Static data is acceptable for early mockups, but the UI cannot pass an operational UX audit until states are designed for loading, stale data, empty data, errors, permissions, optimistic updates, and failed actions.

### 6. Frontend architecture is not settled

The architecture docs say SvelteKit, uPlot, and TanStack Table:

- `docs/masterplan.md:113-114`
- `docs/research-and-architecture.md:141`

The repo implements Vue 3 with no router, no table library, no chart library, no state layer, and no test runner:

- `apps/web/package.json`
- `docs/implementation-notes/vue-control-room-shell.md`

This needs an explicit ADR. Either Vue becomes the accepted product framework, or the app should be reset before the UI investment deepens.

## GitHub Issue Comparison

All GitHub issues #1-#24 are currently open.

| Issue | Current code status | UX gap |
| --- | --- | --- |
| #1 Create architecture ADRs and performance budgets | Not satisfied for frontend. Docs still reference SvelteKit while repo uses Vue. | Must decide framework, charting, table, testing, visual regression, and browser support before deep UI work. |
| #8 Build control room UI shell | Partially implemented in `App.vue` and `CorridorMatrix.vue`. | Exists as a dashboard shell, but lacks command-center hierarchy, incident prioritization, drill-in model, and decision clarity. |
| #9 Build transaction trace UI | Partially implemented in `TransactionTimeline.vue`. | Timeline is too thin. It lacks search, selected/rejected routes, owner details, system references, blocker state, and audit drawer. |
| #11 Build provider/corridor scorecards | Partially implemented in `ProviderScorecards.vue`. | Scorecards are static and card-heavy. They need ranking, trend, SLA context, export, and corridor/provider drilldown. |
| #12 Build latency waterfall and downtime timeline | Partially implemented in `LatencyWaterfall.vue` and `DowntimeTimeline.vue`. | Waterfall is a bar list, not a diagnostic console. Needs root-cause comparison, time-series context, event overlays, and filter behavior. |
| #16 Build route configuration UI | Partially implemented in `RouteConfigurationPanel.vue`. | Controls exist but the safe policy-change workflow does not. Needs current/proposed split, affected traffic, validation, approval, rollback, and permissions. |
| #17 Add policy simulator and shadow routing | Partially implemented in `PolicySimulatorPanel.vue`. | Simulator is demo-only. Needs scenario builder, explainable eligibility, historical replay, report export, and "cannot affect live" safety posture. |
| #18 Build FX and cost board | Partially implemented in `FxCostBoard.vue`. | Data exists, but economics hierarchy is weak. Needs reliability vs cost tradeoff visual, stale-rate blocking, route recommendation rationale, and rate age treatment. |
| #20 Add RBAC, audit exports, and maker-checker approvals | Not implemented in UI. | Must shape critical UI actions now, not after visual redesign. |
| #23 Build design-system primitives from brand language | Partially implemented. | Primitives exist, but no design tokens contract, states matrix, theme guidance, keyboard/focus audit, or component QA. |
| #24 Implement operational copy and state taxonomy | Partially implemented in `copy.ts`. | Copy has been shortened to generic labels like `Shift`, `Test`, and `Export`; it should use precise banking operations language. |

## UX North Star

The redesigned product should feel like a premium bank operations control room for cross-border money movement.

Primary operator question:

> What is broken, who owns it, what traffic is at risk, and what safe action should I take next?

Primary executive/economic buyer question:

> Which providers and routes deserve more traffic, and what proof do we have?

Primary risk/compliance question:

> Can we prove every routing and policy decision was eligible, approved, versioned, and reversible?

## Refactor Strategy

Do not polish the current page in place. Refactor around product workflows.

### Phase 0: UX Governance and Architecture Freeze

Goal: stop design drift before expensive refactoring.

Deliverables:

- frontend ADR: Vue vs SvelteKit decision
- design direction decision: dark-first operational shell or documented light theme
- component/token naming standard
- route and screen IA
- UX quality gate checklist added to PR workflow
- screenshot-based visual regression baseline

Exit criteria:

- framework decision is documented
- design-system contract exists
- every UI issue has UX acceptance criteria beyond "component exists"

### Phase 1: Design-System Foundation

Goal: create a premium visual and interaction foundation before rebuilding screens.

Deliverables:

- semantic color tokens for app canvas, panels, text, status, risk, policy, latency, FX, and audit
- typography scale with tabular numerals
- spacing and density scale for operations screens
- focus, hover, active, disabled, loading, skeleton, empty, stale, error, and blocked states
- `UiButton` rebuilt with variants: primary, secondary, ghost, danger, icon, segmented, loading, disabled
- icon system decision and implementation
- badges, chips, metric tiles, table cells, timelines, alert banners, drawers, tabs, segmented controls
- component examples for all health states: healthy, watch, degraded, blocked, recovery, unknown, stale

Exit criteria:

- no one-off colors in product components
- all controls have keyboard-visible focus
- status colors are paired with text and/or icons
- component states are documented with screenshots

### Phase 2: Information Architecture

Goal: turn the one-page demo into an operations app.

Recommended navigation:

- Control Room
- Corridors
- Transactions
- Incidents
- Routing Policy
- FX and Costs
- Reconciliation
- Providers
- Audit

Screen model:

- Control Room: live triage, not every feature.
- Corridor Detail: route ranking, provider comparison, cost/speed/reliability tradeoffs.
- Transaction Trace: forensic lifecycle and audit.
- Incident Detail: timeline, owner, impact, recommended action, evidence export.
- Routing Policy: safe policy editor with preview, approval, activation, rollback.
- FX and Costs: economics lens by corridor.
- Reconciliation: exception queues and settlement aging.
- Audit: route decisions, policy versions, approvals, exports.

Exit criteria:

- Control Room first viewport answers current health and next action.
- Deep workflows are reachable without turning the home page into a mega-dashboard.
- Mobile/tablet mode becomes executive read-only summary, not a compressed desktop control surface.

### Phase 3: Control Room Redesign

Goal: make the daily operator fast and calm.

Above-the-fold structure:

- global live state and freshness
- current incident severity strip
- at-risk volume/value
- top degraded corridors
- next recommended action with reason and safety status
- route board focused on exceptions first

Required interactions:

- filter by corridor/provider/payout method/window
- open incident detail
- preview traffic shift
- export evidence
- pin or acknowledge recommendation
- see stale/live data state

Exit criteria:

- no more than one primary action competes for attention
- degraded/blocked routes sort above healthy routes
- each recommendation explains trigger, affected traffic, suggested route, risk, and approval requirement

### Phase 4: Workflow-Safe Configuration

Goal: make route changes feel powerful but protected.

Policy change flow:

1. Select scope: corridor, payout method, destination bank, amount band.
2. View current policy and live traffic impact.
3. Make proposed changes: provider enablement, fallback order, traffic split, scoring weights.
4. Run validation and shadow preview.
5. Capture reason and evidence.
6. Request approval where needed.
7. Activate staged rollout.
8. Monitor rollback conditions.
9. Export audit trail.

Required UI patterns:

- current vs proposed diff
- affected transaction/value estimate
- risk level badge
- maker-checker state
- rollback target
- scheduled activation
- post-change monitoring banner

Exit criteria:

- no traffic-changing action can happen without preview, reason, and audit context
- sensitive actions show approval state and rollback route
- keyboard users can complete the workflow

### Phase 5: Data Visualization Refactor

Goal: replace static bars and cards with diagnostic visuals that explain performance.

Latency:

- end-to-end trend with P50/P95/P99
- step-level waterfall
- provider/corridor comparison
- downtime timeline with traffic-shift overlays
- "current degraded route vs next eligible route" comparison

FX and cost:

- effective cost table
- stale rate state
- cheapest vs recommended explanation
- cost/speed/reliability triad
- rate age and source display

Provider scorecards:

- sortable provider ranking
- SLA compliance
- stuck rate and reconciliation exception trend
- support/incident response score
- exportable provider evidence report

Exit criteria:

- each chart has units, window, freshness, and accessible summary
- dense tables virtualize or paginate when needed
- charts do not become decorative analytics noise

### Phase 6: Transaction, Incident, and Audit Evidence

Goal: make every decision traceable.

Transaction Trace:

- searchable by switch reference, provider reference, bank reference
- lifecycle timeline with owner, timestamp, source system, state, and references
- selected route and score
- rejected routes and reasons
- current blocker and safe/manual action
- audit drawer

Incident Detail:

- severity, scope, start time, current owner
- affected transaction/value
- root-cause hypothesis
- traffic-shift history
- provider/support notes
- resolution notes
- evidence export

Audit:

- route decision log
- policy version diff
- actor, approval, reason, timestamp
- export package

Exit criteria:

- every critical screen can produce evidence for operations, provider dispute, or regulator review

### Phase 7: Frontend Data and State

Goal: move from static demo to product-grade UI state.

Deliverables:

- typed API contracts or mock service layer
- state/query layer
- polling/websocket fallback model
- loading, stale, error, empty, and permission states
- route-scoped data modules
- generated fixtures for demos/tests
- no direct import of production screen data from one static `data.ts`

Exit criteria:

- screens can render realistic states without editing component code
- stale data is visually obvious
- failed mutations show recovery paths

### Phase 8: QA and Audit Gates

Goal: make "premium" enforceable.

Deliverables:

- responsive screenshots for desktop, tablet, mobile
- accessibility check pass for keyboard focus, labels, contrast, and landmark structure
- visual regression suite
- component interaction tests
- build, typecheck, lint, test scripts
- performance budget check for dashboard bundle and render path

Exit criteria:

- every UI pull request includes screenshots for changed screens
- no critical workflow ships without keyboard and stale/error state verification

## Proposed Issue Tracker

Use these as a dedicated revamp program. They can be created as GitHub issues under a new milestone such as `M1.5: UI/UX Revamp`.

### UXR-00: UI/UX Revamp Epic

Labels: `frontend`, `ui`, `design`, `P0`

Scope:

- Own the end-to-end redesign program.
- Track links to existing issues #8, #9, #11, #12, #16, #17, #18, #23, #24.
- Define the target screen map and quality gates.

Acceptance criteria:

- epic includes phased checklist
- every surface has an owner issue
- design quality gates are visible
- screenshots are attached before and after refactor

### UXR-01: Decide Frontend Architecture and UX ADR

Labels: `architecture`, `frontend`, `documentation`, `P0`

Related: #1

Acceptance criteria:

- decide Vue vs SvelteKit
- decide charting/table libraries
- decide state/query layer
- decide icon library
- define visual regression approach
- document performance budgets for dashboard screens

### UXR-02: Build Product Design-System Tokens

Labels: `frontend`, `ui`, `design`, `P0`

Related: #23, #24

Acceptance criteria:

- semantic tokens exist for canvas, surfaces, text, lines, focus, and status
- status palette covers healthy, watch, degraded, blocked, recovery, unknown, stale
- tokens support dark-first operational theme
- no screen uses ad hoc color values outside the token layer
- typography, spacing, radius, and shadow rules are documented

### UXR-03: Rebuild Core UI Primitives

Labels: `frontend`, `ui`, `design`, `P0`

Related: #23

Acceptance criteria:

- `UiButton` supports icon, disabled, loading, danger, segmented, and confirmation-ready states
- health badges include state, trigger, window, and updated time
- route score chip includes score, status, reason, confidence, and policy version
- metric tiles support trend, units, and freshness
- focus states pass keyboard audit

### UXR-04: Refactor App Shell and Navigation

Labels: `frontend`, `ui`, `operations`, `P0`

Related: #8

Acceptance criteria:

- app has route-based screens or equivalent screen state
- primary navigation matches product IA
- control room is not a mega-scroll page
- mobile/tablet layout is designed as read-only summary where appropriate
- navigation labels use operational language, not demo labels like `Test`

### UXR-05: Redesign Control Room

Labels: `frontend`, `ui`, `operations`, `P0`

Related: #8, #24

Acceptance criteria:

- first viewport shows live state, top risk, affected value, and recommended action
- degraded/blocked corridors sort above healthy corridors
- recommendation panel explains trigger, affected traffic, next route, risk, and approval state
- stale/live data state is obvious
- no decorative or redundant panels compete with incident signals

### UXR-06: Redesign Corridor Matrix and Provider Ranking

Labels: `frontend`, `ui`, `analytics`, `P0`

Related: #8, #11

Acceptance criteria:

- route board can be sorted and filtered
- each corridor shows health, selected route, eligible alternatives, P95, cost, stuck count, and action
- provider scorecards rank by success, latency, cost, exceptions, and incident history
- window selector changes visible state
- rows support drill-in to corridor detail

### UXR-07: Redesign Route Configuration as a Safe Policy Workflow

Labels: `frontend`, `ui`, `policy`, `operations`, `P0`

Related: #16, #20

Acceptance criteria:

- policy editor shows current vs proposed state
- provider toggles, fallback order, traffic split, and scoring weights are scoped by corridor/payout/bank/amount
- preview shows affected transaction/value estimate
- reason is validated
- approval and rollback states are visible
- dangerous actions cannot be triggered accidentally

### UXR-08: Redesign Policy Simulator and Shadow Report

Labels: `frontend`, `ui`, `routing`, `P1`

Related: #17

Acceptance criteria:

- simulator clearly distinguishes live, proposed, and shadow-only behavior
- route eligibility and rejection reasons are visible
- historical replay summary is exportable
- proposed policy impact is shown by volume, latency, cost, and risk
- "can affect live" state requires explicit permission/approval

### UXR-09: Redesign FX and Cost Board

Labels: `frontend`, `ui`, `fx`, `analytics`, `P0`

Related: #18

Acceptance criteria:

- cheapest vs recommended route explanation is visually dominant
- rate age and source are visible
- stale rates are blocked or downgraded consistently
- effective cost, speed, reliability, and eligibility are compared together
- table is responsive without becoming a vertical dump of labels

### UXR-10: Redesign Latency and Downtime Drilldowns

Labels: `frontend`, `ui`, `observability`, `analytics`, `P0`

Related: #12

Acceptance criteria:

- latency view shows trend, waterfall, and root-cause comparison
- downtime timeline overlays provider events, circuit breaker changes, and operator actions
- filters are functional and preserve context
- each chart/table includes units, time window, and freshness
- root-cause view answers whether delay is provider, bank rail, compliance, webhook, or settlement

### UXR-11: Redesign Transaction Trace and Audit Drawer

Labels: `frontend`, `ui`, `payments`, `audit`, `P0`

Related: #9, #6, #20

Acceptance criteria:

- searchable transaction trace exists
- selected route, rejected routes, score components, policy version, and fallback list are visible
- each lifecycle step shows timestamp, owner, source system, state, and references
- current blocker is visually dominant
- audit drawer can export evidence

### UXR-12: Add Incident Detail and Evidence Export UX

Labels: `frontend`, `ui`, `operations`, `audit`, `P1`

Related: #12, #20, #22

Acceptance criteria:

- incident detail screen exists
- affected traffic/value is visible
- timeline includes system and operator events
- provider escalation evidence can be exported
- resolution and owner fields are represented

### UXR-13: Replace Static Screen Data with UI Data Contracts

Labels: `frontend`, `api`, `testing`, `P0`

Related: #4, #5, #6, #10, #14, #15

Acceptance criteria:

- screen data comes from typed service contracts or mock adapters
- loading, empty, error, stale, and permission states exist
- static fixtures are organized by scenario
- components no longer depend on one global `data.ts` for product behavior

### UXR-14: Add Frontend QA, Accessibility, and Visual Regression

Labels: `frontend`, `testing`, `accessibility`, `performance`, `P0`

Related: #21

Acceptance criteria:

- scripts exist for lint, typecheck, unit/component tests, and screenshot tests
- desktop/tablet/mobile screenshots are generated in CI or local verification
- keyboard workflows pass for control room, policy edit, simulator, and trace
- contrast and landmark checks pass
- bundle and render budgets are documented and checked

## Existing Issue Updates Needed

The current GitHub issues can remain, but their acceptance criteria should be tightened.

Recommended additions:

- Add "passes UI quality gates from `docs/users-brand-design-language.md`" to all UI issues.
- Add screenshots as required evidence for #8, #9, #11, #12, #16, #17, #18, #23.
- Add keyboard and stale/error state requirements to every interactive UI issue.
- Add "not static-only" follow-up issues for each screen once backend contracts exist.
- Add an explicit issue for reconciliation UI, currently absent from the frontend implementation.

## Implementation Order

Recommended order:

1. UXR-01, UXR-02, UXR-03
2. UXR-04, UXR-05
3. UXR-06, UXR-07
4. UXR-08, UXR-09, UXR-10
5. UXR-11, UXR-12
6. UXR-13
7. UXR-14

Do not start by restyling every existing panel. Start by locking the design system and screen model, then rebuild the control room.

## 30/60/90-Day UI Revamp Plan

### First 30 Days

- Architecture/design ADR.
- Dark-first design-system tokens.
- Core component rebuild.
- New app shell/navigation.
- Redesigned control room first viewport.
- Screenshot and accessibility baseline.

### Days 31-60

- Corridor detail and provider ranking.
- Route configuration workflow.
- Policy simulator redesign.
- Transaction trace redesign.
- UI data contracts and fixtures.

### Days 61-90

- FX/cost board redesign.
- Latency/downtime diagnostics.
- Incident detail and evidence export.
- Audit drawer.
- Full responsive/accessibility/performance QA pass.
- Pilot-ready UX review.

## Definition of Done for the Revamp

The UI revamp is complete when:

- a bank operator can identify top risk and recommended action in under 5 seconds
- live, stale, estimated, and unavailable data states are obvious
- every traffic-changing action has preview, reason, approval, and rollback context
- every route decision can be inspected and exported
- every critical number has unit, time window, and source/freshness
- screen hierarchy is calm, premium, and operational
- desktop and tablet views are polished
- mobile does not break and works as executive read-only summary
- keyboard navigation works for critical workflows
- visual regression screenshots are part of the verification loop
- the app no longer feels like a generic admin dashboard
