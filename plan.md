# UI/UX Reset Plan - Inbound Settlement Control Tower

Planning date: 2026-06-01

## Executive Decision

The product should be refocused as a simple, premium control tower for inbound IMTO settlement into Nigerian bank accounts.

The current experience is trying to expose too many domains at once: provider scorecards, FX/cost analytics, route policy, incidents, audits, routing contracts, incoming credits, SLA, reconciliation, and several debit/outflow-adjacent surfaces. The real product center is simpler:

> Monitor incoming IMTO settlement requests, select the safest eligible final-leg route, prove whether the beneficiary was credited, and resolve/reconcile anything that breaches SLA or exhausts automatic requery.

Everything that does not help an operator answer "where is this incoming transaction, is customer value delivered, and what safe action should happen next?" should either disappear from primary navigation or sit behind a feature switch.

## 2026-06-03 Architecture Correction

The dashboard cannot look credible to banks if it is only a polished demo. It must be backed by a core operating database model and chart rollups.

The primary product tables should capture:

- partner contracts and standing accounts
- incoming settlement instructions
- route decisions and rejected-route reasons
- final-leg attempts across Fidelity core, NIP, Moniepoint, Interswitch, and other enabled local rails
- provider/API calls, latency, timeout, and callback outcomes
- SLA deadlines, cooldown windows, requery attempts, and late-success observation
- outcome evidence: provider callback, rail/session evidence, core ledger posting, settlement batch, operator note, and audit event
- remediation cases, maker-checker status, duplicate-risk level, and safe closure path
- reconciliation matches and breaks
- immutable audit events and policy versions

The primary dashboard should read from four backend rollups, not hand-written UI numbers:

| Dashboard chart | Backend rollup | Operating question |
| --- | --- | --- |
| SLA completion trend | `partner_sla_windows` | Are international banking partner SLAs failing? |
| Local provider performance | `route_health_windows` | Which local payment providers are working and which performs best? |
| Exception mix | `exception_windows` | Which failed/reversed transfers or middleware calls need repair? |
| API telemetry | `api_health_windows` | Is the switching API telemetry fresh enough for operations decisions? |

Detailed database architecture is tracked in `docs/adr/0002-settlement-data-architecture.md`, with the initial MariaDB schema draft in `db/settlement_core.sql`.

## Current UI Diagnosis

The Vue app currently has twelve top-level navigation areas:

- Control Room
- Transactions
- Routes
- Policy
- Incidents
- Rates & costs
- Reconcile
- Providers
- Audit
- Routing contracts
- Incoming credits
- Inbound SLA

That is the noise. The app contains many useful pieces, but they are exposed as separate dashboards before the user has a clear mental model. The result feels broad instead of powerful.

The code shape mirrors the product noise:

- `apps/web/src/App.vue` is a large multi-screen container with many unrelated workflows in one file.
- `apps/web/src/types.ts` carries broad product concepts instead of a tight inbound settlement model.
- `apps/web/src/router.ts` exposes too many top-level screens.
- `apps/web/src/styles.css` has a strong dark visual direction, but the density of panels, tables, badges, and dashboards makes the experience feel busy.

## Refocused Product Model

### Core Story

A foreign bank or IMTO has an SLA with Fidelity Bank in Nigeria.

The foreign partner operates a standing account with Fidelity Bank. The account has:

- available prefund balance
- optional credit line or overdraft rules
- SLA deadline for beneficiary credit
- settlement and reconciliation requirements
- permitted corridors, currencies, destination banks, and rails

An incoming transaction arrives:

1. Foreign bank asks Fidelity Bank to credit a beneficiary in naira.
2. The platform validates contract, balance/credit availability, beneficiary destination, and policy.
3. If the beneficiary bank is Fidelity, the platform uses intra-bank routing.
4. If the beneficiary bank is external, the platform uses an external route such as NIP, Moniepoint, Interswitch, or another configured rail.
5. The platform monitors the final credit leg against the SLA.
6. If the transaction breaches SLA, the platform does not blindly mark it failed. It enters a controlled timeout, cooldown, requery, and reconciliation workflow.
7. If automatic checks exhaust, the transaction moves into manual remediation where an operator can requery, attach evidence, mark completed outside platform, approve reversal, or escalate.
8. Route performance is penalized when the route produces timeouts, uncertain outcomes, callback inconsistency, or manual remediation burden.

### Product Promise

One dashboard and API for the final leg of inbound IMTO settlement:

- observe incoming transaction health
- route to the best eligible intra-bank or external rail
- track SLA and late-success risk
- avoid duplicate credit and unsafe reroute
- reconcile provider, rail, and bank-ledger outcomes
- remediate unresolved outcomes with evidence
- improve future route selection based on operational truth

## Target Information Architecture

The app should have a maximum of five top-level dashboards.

### 1. Command Center

Purpose: live inbound settlement control tower.

Primary question:

> What inbound value is at risk right now, which routes are under pressure, and what should operations do next?

Above the fold:

- live inbound value and count
- credited within SLA
- open SLA breaches
- unknown final outcomes
- value at risk
- standing account availability
- top recommended action
- route health strip for intra-bank, NIP, Moniepoint, Interswitch, and any configured partner route

Visual structure:

- central "inflow spine" from foreign bank to Fidelity standing account to beneficiary bank
- SLA pressure lane grouped by `within SLA`, `cooldown`, `requerying`, `manual remediation`, `reconciled`
- right-side action stack with only the next safest actions
- compact route health cards with P95 credit time, timeout rate, late-success rate, and open cases

### 2. Inflows

Purpose: searchable incoming transaction workspace.

Primary question:

> Where is this incoming transaction, who owns the next action, and is customer value delivered?

Core surfaces:

- transaction search by switch reference, partner reference, bank reference, NIP/session reference, beneficiary account, amount, and settlement batch
- transaction lifecycle timeline
- final-leg route chosen and why
- current SLA state and deadline
- callback, posting, ledger, and settlement evidence
- late-success guard status
- current owner and safe next action

This dashboard absorbs the current `Transactions`, `Incoming credits`, and credit detail surfaces.

### 3. Routes

Purpose: route selection, route health, and SLA performance by endpoint.

Primary question:

> Which final-leg route should receive new inbound traffic right now?

Core surfaces:

- intra-bank route performance
- external route performance: NIP, Moniepoint, Interswitch, Paystack transfer, MFBs, wallets if enabled
- route eligibility matrix
- destination-bank health
- SLA breach trend
- timeout and late-success trend
- automatic route penalty score
- traffic split and fallback order

This dashboard absorbs the current `Routes`, parts of `Providers`, and the useful parts of `Rates & costs`.

### 4. Exceptions

Purpose: settlement, reconciliation, and remediation center.

Primary question:

> Which transactions are not safe to close yet, and what evidence is needed to close them?

Core queues:

- SLA breached, still within cooldown
- automatic requery in progress
- automatic requery exhausted
- uncertain outcome / late-success risk
- provider says paid, bank ledger missing
- bank ledger posted, provider callback missing
- duplicate-risk watch
- reversal pending
- completed outside platform awaiting maker-checker

Manual actions:

- requery status by route/provider
- request provider evidence
- attach posting/session evidence
- mark completed outside platform
- approve reversal
- close reconciliation break
- escalate to route owner

This dashboard absorbs the current `Reconcile`, `Incidents`, and unresolved credit workflows.

### 5. Settings

Purpose: low-frequency configuration and advanced features.

Primary question:

> What contracts, SLAs, routes, approvals, and feature switches govern inbound settlement?

Core surfaces:

- partner contracts and standing accounts
- credit line rules
- SLA definitions
- route eligibility rules
- provider/rail credentials and endpoints
- maker-checker approval rules
- circuit breaker thresholds
- feature switches
- audit log and exports

This dashboard absorbs the current `Policy`, `Routing contracts`, `Inbound SLA`, `Audit`, and admin configuration.

## What Gets Tucked Behind Feature Switches

The primary app should not look like a debit/outflow operations suite. Debit-adjacent features only appear when enabled and only inside the workflow where they matter.

Suggested feature switches:

- `externalDebitRailOps`: exposes debit/refund/return handling for payment gateways and outbound rail mechanics.
- `walletAndCashPickup`: exposes wallet, cash pickup, and agent-network endpoints.
- `advancedFxCosts`: exposes full FX/rate economics beyond basic cost and stale-rate eligibility.
- `providerCommercialScorecards`: exposes commercial/provider negotiation scorecards.
- `policySimulator`: exposes shadow-routing and historical replay tools.
- `multiBankPortfolio`: exposes group-level executive views across multiple banks.
- `advancedComplianceCases`: exposes deeper AML/RFI/case-management workflows.
- `settlementFileImports`: exposes bulk settlement-file upload and matching tools.

Default MVP feature state:

| Feature | Default | Reason |
| --- | --- | --- |
| Intra-bank routing | On | Core final-leg route |
| NIP / external bank routing | On | Core final-leg route |
| SLA monitoring | On | Core promise |
| Cooldown/requery/backoff | On | Prevents unsafe failure handling |
| Manual remediation | On | Needed after automation exhausts |
| Reconciliation evidence | On | Needed to prove customer value |
| Debit/refund operations | Off | Advanced exception mode |
| Wallet/cash pickup | Off | Not needed for account-payout pilot |
| Full FX board | Off | Useful, but not core to final-leg credit monitoring |
| Provider commercial scorecards | Off | Executive/provider management, not daily operations |

## SLA Timeout, Cooldown, and Late-Success Model

The hardest part of the product is not showing failed transactions. It is knowing when a transaction is safe to call failed.

### State Model

Recommended final-leg states:

- `received`: inbound instruction accepted
- `validated`: contract, account, policy, and route eligibility checked
- `routing`: final-leg route being selected
- `submitted`: posted to intra-bank or external rail
- `accepted`: rail/provider accepted the request
- `credit_pending`: beneficiary credit not yet proven
- `credited`: customer value delivered and evidence attached
- `sla_breached`: SLA deadline missed, but outcome not final
- `cooldown`: waiting before requery because late success is possible
- `requerying`: automatic status check in progress
- `outcome_uncertain`: route cannot prove paid or failed
- `manual_remediation`: automation exhausted and operator action required
- `completed_outside_platform`: operator confirmed credit with external evidence
- `reversal_pending`: safe reversal/return path initiated
- `failed_safe`: no credit occurred and reroute/reversal is safe
- `failed_unsafe`: failure claimed but duplicate-credit risk remains
- `reconciled`: settlement, ledger, and provider evidence matched

### Backoff Rules

Backoff should be route-specific and SLA-specific. Example default:

- T+0 at SLA breach: mark `sla_breached`, freeze unsafe duplicate actions, open cooldown clock.
- T+30s: first automatic requery.
- T+2m: second requery plus provider callback check.
- T+5m: third requery plus ledger/NIP/session evidence check.
- T+15m: final automatic requery.
- Exhausted: move to `manual_remediation`.

The timings above are placeholders. Settings should allow each route to define:

- SLA duration
- cooldown windows
- max requery attempts
- requery provider/API method
- late-success observation window
- whether reroute is ever allowed after acceptance
- evidence required before manual completion

### Route Penalty

Routes should lose score when they create operational uncertainty, not only when they hard fail.

Penalty inputs:

- SLA breach rate
- timeout rate
- late success after timeout
- callback lag
- inconsistent status response
- missing session/posting evidence
- manual remediation rate
- duplicate-risk events
- reconciliation break rate
- operator override frequency

Penalty should decay after recovery, so a route can earn traffic back through recovery testing.

## Visual Design Direction

The target feel is Fortune-500 bank infrastructure: premium, quiet, dense, and decisive.

The app should feel expensive because it removes doubt, not because it adds decoration.

### Visual Principles

- Calm dark operational canvas.
- One strong primary visual per dashboard.
- Tables are compact and purposeful, not wall-to-wall.
- Status colors are reserved for operational state.
- No nested cards inside cards.
- No decorative gradients competing with risk signals.
- Motion only for live updates, SLA clocks, and state transitions.
- Every number shows unit, freshness, and time window.

### Signature Components

- `InflowSpine`: visualizes foreign bank -> Fidelity standing account -> beneficiary bank/rail.
- `SlaClock`: compact circular or horizontal SLA countdown with breached/cooldown/requery states.
- `StandingAccountMeter`: prefund, credit line, available limit, and projected exhaustion.
- `RouteHealthStrip`: intra-bank, NIP, Moniepoint, Interswitch, and enabled rails.
- `BackoffLadder`: shows automatic requery attempts, next attempt time, and exhausted state.
- `OutcomeConfidenceBadge`: `proven credited`, `probably pending`, `uncertain`, `safe failed`, `unsafe failed`.
- `RemediationQueue`: prioritized by value at risk, SLA age, duplicate risk, and evidence gaps.
- `EvidenceDrawer`: ledger posting, provider callback, rail session, settlement batch, operator note, and audit trail.
- `RoutePenaltyMeter`: shows why a route is losing future traffic.

## Screen-Level Refactor Plan

### Command Center Layout

Top row:

- Inbound value today
- Credited within SLA
- Open SLA breaches
- Unknown outcomes
- Standing account availability

Main left:

- live inflow spine
- route health strip
- SLA pressure lanes

Main right:

- recommended action
- highest-risk transactions
- automatic backoff currently running
- routes being penalized

Bottom:

- compact route table sorted by operational risk
- no broad provider analytics unless opened as a drilldown

### Inflows Layout

Top:

- universal search
- filters: SLA state, route, destination bank, partner, value band, outcome confidence

Main:

- transaction work queue
- selected transaction trace side panel

Trace:

- route decision
- final-leg attempt timeline
- SLA clock and backoff ladder
- evidence checklist
- safe actions only

### Routes Layout

Top:

- route health by final-leg route
- active circuit breakers
- route penalty changes

Main:

- route matrix by destination bank and rail
- route detail drawer with success, P95, timeout, late-success, callback lag, recon breaks

Controls:

- enable/disable by corridor, destination bank, amount band, and partner contract
- traffic split
- fallback order
- recovery testing

All traffic-changing controls require preview, reason, approval state, and rollback target.

### Exceptions Layout

Top:

- manual queue size
- value at risk
- oldest unresolved
- duplicate-risk count

Main:

- queue tabs: cooldown, requerying, exhausted, recon breaks, completed outside platform, reversals

Case detail:

- current theory of outcome
- evidence status
- requery action
- mark completed outside platform
- approve reversal
- close break
- audit trail

### Settings Layout

Top:

- partner/contract selector
- environment and feature switches

Sections:

- standing account and credit line
- SLA and backoff policy
- route eligibility and thresholds
- maker-checker and roles
- integrations
- audit exports

## Frontend Implementation Plan

### Phase 1 - Product Scope Reset

Goal: make the app simpler before making it prettier.

Tasks:

- Replace twelve top-level screens with five: `Command Center`, `Inflows`, `Routes`, `Exceptions`, `Settings`.
- Move `FX`, `Providers`, `Audit`, `Contracts`, `Credits`, and `Inbound SLA` into tabs/drawers under the five-screen model.
- Add feature switches for debit/outflow, wallets, cash pickup, advanced FX, provider scorecards, and policy simulator.
- Rename product copy around inbound settlement, final-leg credit, SLA, outcome confidence, and remediation.

Acceptance criteria:

- top-level navigation has no more than five items
- no debit/outflow feature appears in primary navigation by default
- first viewport answers live inbound risk and next action

### Phase 2 - Domain Model Reset

Goal: align types and mock data with the actual bank workflow.

Tasks:

- Create inbound-focused domain types:
  - `InboundContract`
  - `StandingAccount`
  - `CreditFacility`
  - `IncomingInstruction`
  - `FinalLegAttempt`
  - `SlaPolicy`
  - `BackoffPolicy`
  - `RequeryAttempt`
  - `OutcomeEvidence`
  - `RemediationCase`
  - `RoutePenalty`
- Convert current `CreditLeg`, `RoutingContract`, `InboundSlaRow`, and `ReconciliationItem` into the new model.
- Add realistic examples for Fidelity Bank:
  - Bank X in England -> Fidelity standing account -> Fidelity beneficiary
  - Bank X in England -> Fidelity standing account -> external beneficiary via NIP
  - NIP timeout followed by late success
  - Moniepoint callback missing but ledger posted
  - Interswitch failed-safe with reversal pending

Acceptance criteria:

- every UI state maps to the inbound settlement model
- timeout/cooldown/requery states are represented in mock data and tests

### Phase 3 - App Architecture Cleanup

Goal: stop `App.vue` from being the whole product.

Tasks:

- Split `App.vue` into a shell and route views:
  - `views/CommandCenterView.vue`
  - `views/InflowsView.vue`
  - `views/RoutesView.vue`
  - `views/ExceptionsView.vue`
  - `views/SettingsView.vue`
- Move screen computations into composables:
  - `useInboundTower`
  - `useRouteHealth`
  - `useRemediationQueue`
  - `useFeatureSwitches`
- Create small domain components for the signature components listed above.
- Keep shared primitives in `components/`.

Acceptance criteria:

- `App.vue` is mostly shell, auth, and layout
- each view owns one primary workflow
- tests mount views independently

### Phase 4 - Premium Visual System

Goal: make the product visually rich without returning to noise.

Tasks:

- Tighten tokens for canvas, surface, line, text, live signal, SLA, risk, recovery, and policy.
- Reduce panel count and use full-width bands or split views instead of stacked card grids.
- Add a strong first-viewport visual to Command Center: the inflow spine plus SLA pressure lanes.
- Replace generic KPI cards with outcome-specific instruments:
  - SLA clock
  - standing account meter
  - route penalty meter
  - backoff ladder
- Use icon buttons and compact tooltips for secondary actions.

Acceptance criteria:

- one visual hierarchy per screen
- no generic dashboard cards that do not drive an action
- status color only communicates status
- text fits on desktop and mobile without overlap

### Phase 5 - Remediation and Requery Workflow

Goal: make the hardest operational workflow first-class.

Tasks:

- Build manual requery by route/provider.
- Show automatic attempts and next attempt time.
- Show when automation has exhausted.
- Require evidence and reason for manual completion.
- Add maker-checker state for `completed outside platform` and reversal approval.
- Persist audit event shape in mocks/API contract.

Acceptance criteria:

- operator can see why a transaction is not safely failed
- operator can manually requery after automatic backoff exhausts
- operator can mark completed outside platform with evidence
- duplicate-credit risk is explicit before any closure action

### Phase 6 - Verification

Goal: make the redesign shippable, not just attractive.

Tasks:

- Update unit/component tests for five-screen navigation.
- Add tests for timeout -> cooldown -> requery -> manual remediation.
- Add tests for feature switches hiding debit/outflow/advanced modules.
- Run `npm run web:build` and `npm run web:test`.
- Use browser screenshots at desktop and mobile for Command Center, Inflows, and Exceptions.

Acceptance criteria:

- build passes
- tests pass
- screenshots show no overlap or noisy first viewport
- critical workflows are keyboard reachable

## Suggested Build Order

1. Navigation and copy reset.
2. Inbound domain types and mock data.
3. Command Center first viewport.
4. Inflows search and transaction trace.
5. Exceptions/remediation workflow.
6. Routes health and route penalty view.
7. Settings with feature switches and SLA/backoff policy.
8. Visual polish and responsive QA.

## Definition of Done

The refactor is successful when:

- a Fidelity operations user can understand inbound settlement risk in under five seconds
- the app has four or five main dashboards, not twelve
- incoming transactions are the center of every workflow
- debit/outflow and advanced analytics are hidden unless enabled
- SLA breach does not equal naive failure
- cooldown, backoff, requery, late success, and reconciliation are visible and actionable
- every manual completion or reversal has evidence and maker-checker/audit context
- route penalties affect future recommendations
- the interface feels like a premium bank control tower, not a noisy admin dashboard

---

# imsi-rails - IMTO Switch and Monitoring Dashboard Brief

Research date: 2026-05-19

## Working Thesis

Top Nigerian and African banks do not rely on one IMTO rail. They combine:

- Legacy global IMTOs: Western Union, MoneyGram, Ria.
- Digital-first IMTOs: WorldRemit, Sendwave, Remitly, Small World, Boss Money/Revolution, NALA, TapTap Send, ACE, LemFi, Mukuru, etc.
- B2B payout networks/aggregators: Thunes, TerraPay, Onafriq/MFS Africa, Mastercard/Transfast.
- Pan-African/local bank rails: PAPSS, AccessAfrica, UBA AfriCash, Ecobank Rapidtransfer, GTBank GTMT, FirstBank First Global Transfer.

The opportunity is not another remittance app, FX exchange product, or consumer money-transfer brand. It is an orchestration, reliability, compliance, and observability layer that helps banks choose the best route per corridor and detect failures before customers feel them.

## Deeper Product Definition

The platform should be positioned as a neutral reliability switch for banks. Each bank gets one operating layer across many IMTOs, aggregators, and settlement rails:

- one integration surface into multiple IMTO and payout partners
- one reliability dashboard across all routes and corridors
- one transaction lifecycle trace from sender initiation to beneficiary credit
- one routing policy engine that selects the best eligible rail per transaction
- one exception/reconciliation workspace for failed, stuck, delayed, or mismatched transactions
- one compliance and audit trail for regulator, bank, and provider review

The strongest promise is not "use our rail." It is "your bank will always use the best available eligible rail for this transaction, based on origin, destination, amount, payout method, compliance rules, cost, speed, reliability, and current rail health."

Important nuance: the platform should not claim that the absolute best rail is always knowable. It can guarantee that it selects the best eligible route according to observable data, bank policy, regulatory constraints, and real-time availability.

## Core Product Question

For every transaction, the switch must answer:

> Given this sender, receiver, origin country, destination country, currency, payout method, amount, compliance profile, bank preference, and current market/rail conditions, which route has the best probability of completing quickly, cheaply, compliantly, and without manual intervention?

This turns the platform into a decision engine and control plane, not only a message router.

## Bank Sales Positioning

### Category

International money transfer reliability infrastructure for banks.

The bank-facing description:

> One dashboard, one switching layer, and one routing intelligence engine for every IMTO provider connected to the bank.

The platform is not trying to replace IMTOs. It helps banks onboard, monitor, compare, and route across IMTOs so the best eligible provider wins each transaction.

### Why Banks Buy

Banks already work with many IMTO providers, but the operating model is fragmented:

- each IMTO has its own integration, portal, file format, SLA, support path, and reconciliation process
- operations teams often discover failures after customers complain
- provider performance is hard to compare objectively
- manual route decisions depend on tribal knowledge
- failed or delayed transfers hurt the bank's brand even when the provider caused the issue
- banks lack a unified view of cost, speed, reliability, and settlement exposure across routes

The platform gives the bank a neutral operating layer across all providers, while preserving the bank's existing commercial/provider relationships.

### Core Promise

For the bank:

- every IMTO connected through one control layer
- every transaction routed through the best eligible provider based on bank policy and live performance
- every route continuously measured on reliability, speed, cost, compliance readiness, and operational quality
- every failure visible early, with automatic traffic shifting where safe
- every routing decision explainable to operations, management, auditors, and regulators

This is the bank's IMTO control tower.

### Best Provider Wins Model

"Best provider wins" should mean measured, policy-governed competition, not a public bidding marketplace.

The bank defines the rules:

- which providers are approved
- which corridors each provider can serve
- minimum SLA thresholds
- maximum acceptable cost
- compliance and KYC requirements
- liquidity and settlement limits
- fallback rules
- when human approval is required

The platform measures performance:

- success rate
- time-to-credit
- failed/stuck transaction rate
- cost and FX quality
- settlement reliability
- support responsiveness
- reconciliation breaks
- incident history

The routing engine allocates traffic:

- more volume to providers meeting bank goals
- less or no new volume to degraded providers
- fallback traffic to the next eligible provider
- shadow-mode testing before changing production allocation

This creates a fair performance loop. IMTOs can win more bank volume by being faster, cheaper, more reliable, and easier to reconcile.

### What The Platform Is Not

- not a consumer remittance app
- not the bank's FX marketplace
- not a replacement for licensed IMTOs
- not a new wallet or cash pickup network
- not a provider of last-mile payout liquidity unless explicitly added later
- not a public auction where IMTOs bid for each transaction

It is the bank's switching, monitoring, policy, and reliability layer across existing and future IMTO relationships.

### Buyer Personas

| Buyer | What they care about |
| --- | --- |
| Head of Remittances / Diaspora Banking | higher successful volume, better customer experience, stronger IMTO relationships |
| Payments Operations | fewer stuck transfers, faster incident detection, clear ownership of failures |
| CIO / CTO | fewer one-off integrations, cleaner API layer, observable transaction flow |
| Risk / Compliance | auditable routing decisions, approved providers only, clear exception handling |
| Treasury / Finance | settlement visibility, prefund exposure, cost and FX performance |
| Executive Management | reliable remittance revenue, customer confidence, provider accountability |

### Sales Narrative

Lead with reliability, not technology:

1. Banks lose confidence and revenue when international transfers fail or stall.
2. Most banks already have many IMTO providers, but no single truth about which route is best at any moment.
3. The platform gives banks one control tower for IMTO health, transaction routing, failures, reconciliation, and provider scorecards.
4. Banks keep their providers. The switch makes those providers compete on measurable performance.
5. The first pilot proves value by reducing stuck transactions, improving time-to-credit, and showing where money should have been routed.

### Commercial Model

Keep incentives clean: the bank should be the primary customer so the routing engine optimizes for the bank's policy and customers.

Possible pricing:

- implementation/onboarding fee
- monthly platform fee per bank or bank group
- per-successful-transaction switching fee
- optional premium modules for advanced reconciliation, provider scorecards, or multi-country rollout

Avoid early pricing structures that make neutrality questionable, such as taking hidden incentives from providers to influence routing.

### Bank Onboarding Requirements

For a pilot, the bank should provide:

- list of active IMTO providers and corridors
- current transaction status codes and reconciliation files
- provider API/SFTP/portal access where available
- settlement file samples
- bank posting/account validation integration path
- historical transaction data for baseline performance
- SLA and commercial rules per provider
- compliance and approval rules
- named operations owners for incident escalation

The first sale should feel like operational relief, not a massive transformation project.

## Practical Delivery Lens

The first version should prove three high-value outcomes:

1. Banks can see the real-time health of every connected IMTO/rail in one place.
2. Banks stop sending new transactions to routes that are failing, delayed, too expensive, or out of policy.
3. Banks can explain every routing decision with an audit trail.

Everything else is secondary until those outcomes are working in production.

### Highest-Value Wedge

Start with inbound remittances into Nigerian bank accounts, because this is painful enough to matter, narrow enough to ship, and rich enough to prove routing value.

Initial product:

- one anchor bank or bank group
- direct-to-account payout first
- 3-5 connected routes, not every IMTO
- one corridor cluster first, such as US/UK/EU -> Nigeria
- routing based on rules plus measured rail health
- operations dashboard for live monitoring and manual override
- reconciliation exception view, not a full finance suite

This avoids feature bloat while still proving the core platform thesis.

### MVP Success Metrics

The MVP should be judged by operational and commercial outcomes:

- reduce failed or stuck transactions
- reduce average time-to-credit
- reduce manual follow-up by operations teams
- improve provider/bank SLA visibility
- reduce cost per successful transfer where cheaper rails meet SLA
- shorten incident detection and response time
- prove that auto-switching prevents avoidable failures

If a feature does not help one of these metrics, it should wait.

### What Not To Build First

Defer these until the core switch is proven in production:

- consumer remittance app
- mobile wallet payout across every African country
- full ML/AI route optimization
- public provider bidding/auction marketplace
- advanced revenue-sharing simulations
- broad treasury/liquidity optimization
- multi-tenant white-label portals
- every possible IMTO adapter
- full case-management suite
- custom CRM or customer support tooling

The platform should integrate with bank/provider systems where possible instead of recreating them.

## Nigeria Bank Evidence Snapshot

Public bank pages show heavy overlap across the same providers:

| Bank / Group | Publicly listed IMTO or cross-border transfer options |
| --- | --- |
| Access Bank | AccessAfrica, Ria, Western Union, MoneyGram, Transfast, LeadRemit, PAPSS |
| UBA Nigeria | AfriCash, Western Union, MoneyGram, Ria, PAPSS, Juba Express, BNB IMTO, TerraPay |
| Zenith Bank | Western Union, MoneyGram, WorldRemit, Flutterwave, Nairagram, Boss Revolution, Cashpot, LeadRemit, Sendwave, SmallWorld, Venture Garden, Simplify, ACE Money Transfer, Funtech |
| FirstBank | First Global Transfer, Ria, MoneyGram, Western Union, TransFast, Boss Money, Flutterwave, WorldRemit, Thunes, Sendwave, SmallWorld, Venture Garden, Funtech |
| GTBank | Western Union, MoneyGram, Transfast, GTMT, Remitly, Thunes ADMT; GTBank also lists WorldRemit in IMT services |
| Ecobank | Western Union, MoneyGram, Ria, MTN, Rapidtransfer; 2025 partnership with Thunes for instant cross-border payments across Sub-Saharan Africa |
| Fidelity / Union Bank | Fidelity lists Western Union, MoneyGram, Ria, WorldRemit, Transfast, Boss Money, Small World, Cashpot; Union lists Western Union, MoneyGram, Ria, WorldRemit, SmallWorld, Wari |

## Africa Bank Evidence Snapshot

| Bank / Group | Publicly listed options |
| --- | --- |
| KCB Kenya | Western Union, MoneyGram, WorldRemit, Ria, Transfast, Small World, Dahabshiil, TerraPay, Thunes, Onafriq/MFS Africa, Global Money Exchange, Gulf Exchange |
| Equity Bank Kenya | Transfast, Sendwave, NALA, Remitly, Ria, Funtech, Al Fardan Exchange, SWIFT, MoneyGram, Western Union, WorldRemit, Taptap Send |
| Standard Bank South Africa | MoneyGram; also SWIFT/international payments and Shyft for FX transfers |
| Absa South Africa | Western Union through Absa app, online banking, and branches |
| Ecobank Group | Rapidtransfer, Western Union, MoneyGram, Ria, plus Thunes partnership across Sub-Saharan Africa |

## Platform Priority

### Tier 1 - Must support early

- Western Union
- MoneyGram
- Ria
- WorldRemit / Zepz
- Sendwave / Zepz
- Remitly
- Transfast / Mastercard Cross-Border Services
- PAPSS
- TerraPay
- Thunes
- Onafriq / MFS Africa

### Tier 2 - Corridor and bank demand dependent

- Small World
- Boss Money / Boss Revolution / IDT
- NALA
- TapTap Send
- ACE Money Transfer
- Juba Express
- Dahabshiil
- Mukuru
- Flutterwave Send
- Afriex
- Paga Remit
- LemFi / RightCard
- Verto
- Fincra
- Raenest
- Interswitch and eTranzact where licensed/available

### Tier 3 - Bank proprietary or intra-group rails

- AccessAfrica
- UBA AfriCash
- Ecobank Rapidtransfer
- GTBank GTMT
- FirstBank First Global Transfer

## Product Architecture

### Control Plane vs Data Plane

The product should be designed around two layers:

1. Control plane
   - provider registry
   - corridor/routing configuration
   - bank-specific policy
   - compliance rules
   - cost and SLA models
   - real-time rail health
   - circuit breakers
   - risk and operational limits

2. Data plane
   - transaction intake
   - account validation
   - provider submission
   - payment posting
   - webhook/event handling
   - retry/failover execution
   - settlement/reconciliation matching
   - customer/bank status updates

This separation matters because banks need to change routing rules without redeploying transaction processing code.

### Core Modules

1. Provider registry
   - IMTO/provider, license status, corridors, payout methods, currencies, limits, KYC documents, cut-off windows, settlement model, support contacts, webhook/API details.

2. Normalization layer
   - One canonical transaction state model across providers:
     - created
     - accepted
     - compliance_pending
     - prefund_pending
     - sent_to_provider
     - received_by_bank
     - account_validated
     - credited
     - paid_cash
     - failed
     - reversed
     - disputed
     - reconciled

3. Bank adapters
   - Core banking posting
   - NIBSS/NIP or local instant payment rail
   - name enquiry/account validation
   - BVN/NIN/KYC checks
   - sanctions/PEP/AML screening
   - ledger and suspense accounts
   - customer notification

4. Provider adapters
   - API, webhook, SFTP, batch, or operator-console integration.
   - Standardize reference IDs: provider reference, bank reference, switch reference, settlement batch ID.

5. Routing engine
   - Route by corridor, payout method, currency, amount, license/compliance constraints, provider availability, cost, FX rate, expected settlement time, historical success rate, and bank preference.

6. Observability and control room
   - Real-time corridor health
   - provider SLA scorecards
   - stuck transaction queues
   - reconciliation exceptions
   - liquidity/prefund status
   - compliance queues
   - incident and escalation workflows

7. Settlement and reconciliation
   - Automated matching across provider files, bank postings, customer notifications, and settlement accounts.
   - Exception categories: duplicate, missing credit, late settlement, FX mismatch, name mismatch, reversal pending, unresolved payout.

## Routing and Auto-Switching Model

### Route Eligibility First

Before scoring a route, the switch should eliminate routes that are not eligible:

- provider is not licensed/approved for the destination country
- provider does not support the origin-destination corridor
- provider does not support the payout method: account, wallet, cash pickup, card, or branch
- amount is outside provider, bank, or regulatory limits
- currency pair is unsupported
- recipient bank is unsupported or degraded
- sanctions/AML/KYC rules require manual review
- prefunding or settlement balance is insufficient
- provider or local rail is under an active circuit breaker
- bank-specific policy blocks the provider or corridor

Only eligible routes should enter scoring.

### Route Score

A practical first version can use a weighted scoring model:

```text
route_score =
  reliability_weight * route_success_probability
+ speed_weight       * normalized_expected_time_to_credit
+ cost_weight        * normalized_total_cost
+ fx_weight          * normalized_fx_quality
+ liquidity_weight   * liquidity_confidence
+ ops_weight         * low_manual_intervention_score
+ support_weight     * provider_resolution_score
- risk_penalty
- incident_penalty
```

Banks should be able to tune weights by corridor and product:

- payroll remittance may prioritize speed and certainty
- retail diaspora remittance may prioritize cost and convenience
- high-value transfers may prioritize compliance confidence and settlement reliability
- bank-owned channels may prioritize margin after minimum SLA thresholds are met

### Scoring Inputs

The routing engine should combine static configuration with live operational signals:

| Input | Examples |
| --- | --- |
| Corridor | UK-Nigeria, US-Ghana, Kenya-Nigeria, South Africa-Zimbabwe |
| Payout method | bank account, cash pickup, wallet, card, branch |
| Destination bank | Access, UBA, Zenith, FirstBank, GTBank, Ecobank, KCB, Equity |
| Provider capability | supported countries, banks, currencies, limits, channels |
| Price | provider fee, bank fee, FX spread, revenue share, settlement cost |
| Speed | expected time to credit, P50/P95/P99 latency, cut-off windows |
| Reliability | success rate, reversal rate, retry rate, stuck rate |
| Current health | API uptime, webhook delay, settlement delay, posting failures |
| Liquidity | prefund balance, settlement balance, available limits |
| Compliance | license status, KYC completeness, sanctions status, transaction purpose |
| Operations | support SLA, unresolved incident count, manual intervention rate |

### Auto-Switching Behaviors

The platform should support more than one switching mode:

1. Pre-transaction smart routing
   - Select the best route before sending the transaction.

2. In-flight failover
   - If provider acceptance fails before money movement is final, retry with the next eligible route using idempotency protection.

3. Rail degradation switching
   - If a provider, bank API, local payment rail, or corridor breaches thresholds, new traffic is shifted away automatically.

4. Partial traffic shifting
   - Move 10%, 25%, 50%, then 100% of traffic to another route when testing recovery or onboarding a new provider.

5. Bank-policy override
   - Operations users can pin a corridor to a route, block a provider, pause auto-switching, or force manual approval.

6. Shadow routing
   - The engine records what it would have selected under a new policy without actually changing production routing.

7. Post-failure recommendation
   - If a transaction cannot safely be re-routed automatically, the dashboard recommends the next best manual action.

### Circuit Breakers

Every rail should have automated circuit breakers. Examples:

- success rate below threshold over 5/15/60 minutes
- P95 time-to-credit above SLA
- provider API timeout/error rate above threshold
- webhook delay above threshold
- reconciliation mismatch spike
- duplicate credit risk detected
- prefund balance below threshold
- compliance response unavailable
- downstream bank posting failure spike

Circuit breakers should have states:

- healthy
- degraded
- blocked for new traffic
- manual review only
- recovery testing

### Route Selection Output

For every routed transaction, the engine should persist:

- selected route
- route score
- eligible routes considered
- rejected routes and rejection reason
- scoring inputs at the time of decision
- bank policy version
- compliance decision version
- fallback route list
- whether auto-switching was allowed

This audit trail is essential for bank confidence, provider disputes, and regulatory review.

## Dashboard KPIs

The dashboard should not only report what happened. It should tell operations teams what to do next.

### Executive View

- total volume and value by corridor, bank, and provider
- success rate and SLA compliance
- top degraded corridors
- revenue, cost, and FX margin trends
- settlement exposure
- unresolved high-value exceptions
- provider scorecard ranking

### Reliability

- Provider uptime by corridor and payout method
- End-to-end success rate
- P50/P95/P99 time to credit
- Provider acceptance rate
- Bank posting success rate
- webhook/SFTP lag
- failed transaction volume and value
- retry and failover success rate

### Customer Experience

- time to customer notification
- cash pickup availability
- direct-to-account availability
- unresolved customer complaints
- refund/reversal time
- provider support SLA performance

### Commercials

- fee and FX margin by route
- total cost to customer
- bank revenue share
- liquidity/prefund utilization
- settlement float
- cost per successful transfer

### Compliance

- CBN/license status
- amount-limit breaches
- KYC/BVN/NIN validation failures
- sanctions/PEP/AML hits
- suspicious transaction reports pending
- missing transaction purpose or documentation

### Operations

- transactions stuck by state
- reconciliation breaks
- settlement aging
- exceptions by provider and bank branch/channel
- manual intervention rate
- incident count and mean time to resolve

### Routing Intelligence

- percentage of transactions routed by best-score route
- savings from optimized routing
- time saved from optimized routing
- failovers attempted and failovers successful
- route score distribution by corridor
- volume shifted due to circuit breakers
- route recommendations ignored by operators
- shadow-routing delta: what would have changed under new policy

### Corridor Command Center

Each corridor should have its own operating screen:

- active providers and payout rails
- live health by provider
- cost and speed comparison
- route score ranking
- current traffic split
- open incidents
- stuck transaction count
- settlement exposure
- recommended action

Example:

| Corridor | Best route now | Reason | Avoid |
| --- | --- | --- | --- |
| UK -> Nigeria account | Remitly -> Thunes -> NIP | best P95 speed and low failure rate | Ria due webhook delay |
| US -> Nigeria cash pickup | Western Union | broad branch coverage | MoneyGram if settlement queue delayed |
| Kenya -> Nigeria account | PAPSS or Onafriq | local-currency Africa corridor | SWIFT due cost/speed |

### Transaction Trace View

Every transaction should be inspectable as a timeline:

1. received from bank/channel
2. validated recipient
3. screened compliance
4. eligible routes calculated
5. route selected
6. sent to provider
7. provider accepted
8. payout rail submitted
9. beneficiary credited or cash available
10. customer/bank notified
11. settlement matched
12. reconciled

Operators should see the current owner of the next action: switch, provider, bank, compliance, settlement, or support.

## UI/UX Strategy

The experience should feel modern, mesmerizing, and deeply capable without becoming visually noisy. The product is a mission-critical control room, not a decorative analytics site.

The UI should make complex remittance infrastructure feel simple:

- what is healthy
- what is degrading
- what is failing
- what route is best right now
- what changed recently
- what action should an operator take next

### Experience Principles

1. Calm by default, deep on demand
   - The first view should show operational truth clearly.
   - Advanced latency, downtime, routing, FX, and reconciliation details should be one click away, not always on screen.

2. Beautiful through clarity
   - Use clean hierarchy, strong spacing, restrained color, sharp typography, and elegant motion.
   - Avoid visual decoration that competes with incident signals.
   - Make the product feel premium through precision, not ornament.

3. Real-time confidence
   - Health states should update live.
   - Operators should immediately see whether data is fresh, stale, estimated, or delayed.
   - Every chart or score should show its timestamp and measurement window.

4. Explain every decision
   - Route selection must be inspectable.
   - The UI should show why one provider won and why others were rejected or ranked lower.
   - Operators should never need to guess why traffic moved.

5. Safe configuration
   - Route and provider configuration should support preview, validation, approval, shadow mode, staged rollout, and rollback.
   - Dangerous changes should be difficult to make accidentally.

6. Maximum compatibility
   - Desktop-first for operations rooms, but responsive enough for tablets and executive mobile views.
   - Works on modern Chrome, Edge, Safari, and Firefox.
   - Performs well on lower bandwidth and older corporate machines.
   - Supports keyboard navigation, accessible contrast, screen-reader labels, and exportable data.
   - Gracefully degrades when live streams/websockets fail by falling back to polling.

### Information Architecture

Primary navigation should be simple:

| Area | Purpose |
| --- | --- |
| Control Room | Live health, active incidents, route recommendations, traffic shifts |
| Corridors | Origin-destination performance, provider rankings, cost/speed/reliability |
| Transactions | Search, trace, stuck items, failure reasons, ownership |
| Providers | SLA scorecards, uptime, latency, settlement, reconciliation quality |
| FX & Costs | provider rates, spreads, fees, effective cost, historical trends |
| Routing Policy | route rules, scoring weights, fallbacks, circuit breakers, traffic split |
| Reconciliation | settlement files, unmatched items, duplicates, delayed credits |
| Incidents | degraded rails, timeline, ownership, resolution notes |
| Audit | routing decisions, config changes, approvals, policy versions |

### Control Room Experience

This should be the main screen for daily operations.

Above the fold:

- global health score
- live transaction volume and value
- active degraded routes
- top corridors by risk
- stuck transaction count
- average and P95 time-to-credit
- provider ranking for current window
- recommended action panel

Visual style:

- status map or corridor matrix for quick scanning
- compact provider health cards
- sparkline trends for latency and failures
- clear severity colors: healthy, warning, degraded, blocked
- subtle live-update motion when values change

The control room should feel alive, but not frantic.

### Latency and Downtime Drilldowns

Latency analysis should be one of the product's strongest differentiators.

Drilldowns should answer:

- where exactly did the delay happen?
- is the issue provider API, bank posting, local rail, webhook/status callback, compliance, or settlement?
- is the delay isolated to a provider, corridor, destination bank, amount band, or payout method?
- when did degradation start?
- did traffic shift automatically?
- what route would have performed better?

Required latency views:

1. End-to-end latency
   - transaction received to beneficiary credited
   - P50/P95/P99 by provider, corridor, bank, and payout method

2. Step latency waterfall
   - intake
   - validation
   - compliance
   - route decision
   - provider acceptance
   - bank/local rail posting
   - beneficiary credit
   - notification
   - settlement match

3. Downtime/event timeline
   - provider API errors
   - timeout spikes
   - webhook lag
   - posting failures
   - settlement delay
   - operator actions
   - circuit breaker state changes

4. Heatmaps
   - corridor vs provider
   - hour-of-day vs failure rate
   - destination bank vs latency
   - amount band vs manual review rate

5. Root-cause comparison
   - current degraded route vs next-best route
   - before/after circuit breaker activation
   - provider SLA vs observed performance

### FX and Cost Experience

FX should be easy to understand without turning the product into a trading terminal.

Core FX views:

- live provider rates by corridor/currency pair
- bank reference rate
- effective customer rate
- provider fee
- bank fee
- FX spread
- total cost to customer
- bank margin
- historical rate movement
- stale-rate alerts
- rate source and timestamp

The UI should make comparison obvious:

| Provider | Rate | Fee | Effective Cost | Speed | Reliability | Recommendation |
| --- | --- | --- | --- | --- | --- | --- |
| Provider A | best | medium | lowest | fast | healthy | route more |
| Provider B | good | low | medium | slow | degraded | avoid |
| Provider C | weak | high | high | fastest | healthy | use for urgent transfers |

FX should feed routing, but operators should see when the cheapest route is not selected because reliability or speed risk is too high.

### Swap and Configuration UI

Route configuration must be extremely easy, because bank operations teams will use it during real incidents.

Key configuration surfaces:

1. Provider toggle
   - enable/disable provider by corridor, payout method, amount range, or destination bank
   - require reason for disabling
   - show affected traffic before saving

2. Traffic split control
   - 100/0, 75/25, 50/50, 25/75, 0/100 presets
   - custom split with validation
   - staged rollout with time window
   - automatic rollback if error threshold is breached

3. Route priority editor
   - ranked fallback list
   - drag/reorder providers
   - show expected impact on cost, latency, and success rate

4. Scoring weights editor
   - sliders for reliability, speed, cost, FX, liquidity, and ops burden
   - corridor-specific templates
   - preview which historical transactions would have routed differently

5. Circuit breaker configuration
   - threshold builder
   - measurement window
   - auto-action: alert only, degrade, block new traffic, shift percentage
   - recovery rules

6. Policy simulator
   - test a sample transaction
   - show eligible routes
   - show rejected routes and why
   - show final route and score
   - compare current policy vs proposed policy

7. Approval and rollback
   - maker-checker for sensitive changes
   - scheduled activation
   - policy version history
   - one-click rollback to previous policy

Configuration should feel like moving sliders and switches in a cockpit, but with banking-grade guardrails.

### Compatibility and Performance Requirements

The interface must be dependable in bank environments:

- page loads under 3 seconds on a typical corporate network
- live dashboard remains usable with thousands of transactions per minute
- large tables use pagination or virtualization
- charts avoid heavy rendering that slows old machines
- exports available as CSV/XLSX/PDF where appropriate
- all critical actions work without relying only on drag-and-drop
- every real-time screen has a visible "last updated" timestamp
- offline/stale-data state is explicit
- role-based permissions are enforced in the UI and backend

### Design System Direction

Visual tone:

- premium fintech control room
- dark and light modes if affordable, but start with one polished theme
- color used for meaning, not decoration
- compact but readable tables
- crisp icons and tooltips
- strong empty, loading, error, and degraded states

Key components:

- health badges
- latency sparklines
- route score chips
- provider scorecards
- corridor matrix
- transaction timeline
- policy diff viewer
- traffic split control
- FX comparison table
- incident banner
- audit drawer

The UI should be impressive because it makes operators faster and calmer.

## MVP Recommendation

### Phase 1 - Lean Nigeria Inbound Reliability Switch

Focus: direct-to-account into Nigerian banks.

The goal is not maximum coverage. The goal is to prove that the switch can observe rail health, make better route decisions, and prevent avoidable transaction failures.

Integrations:

- Bank: start with one anchor bank, but design the model for multiple banks.
- Providers: choose 3-5 routes that represent different rail types:
  - one legacy IMTO, such as Western Union, MoneyGram, or Ria
  - one digital IMTO, such as Remitly, WorldRemit/Sendwave, or Small World
  - one B2B payout network, such as Thunes, TerraPay, or Onafriq
  - PAPSS where Africa-to-Africa use cases are in scope
  - bank-owned route only if the anchor bank already has one available
- Local rails: account validation/name enquiry, bank posting path, basic sanctions/compliance response, ledger/reconciliation files.

Dashboard:

- provider/corridor status board
- transaction lifecycle trace
- SLA and latency charts
- failed/stuck queue
- settlement/reconciliation workspace
- compliance exception queue

Routing:

- start with rules plus weighted scoring
- support route eligibility filtering
- support manual provider disablement
- support automatic circuit-breaker-based traffic shifting
- support full route-decision audit trail
- run new policies in shadow mode before activation

### MVP Build Scope

Build only the minimum pieces needed to route and monitor real transactions:

1. Transaction intake API
   - accepts transaction details from bank/channel
   - validates required fields
   - creates switch reference

2. Provider and route registry
   - supported corridors
   - payout methods
   - limits
   - fees/FX where available
   - provider status
   - bank policy rules

3. Routing decision service
   - eligibility filtering
   - simple weighted score
   - fallback list
   - route-decision audit log

4. Health monitor
   - provider API success/error rate
   - timeout rate
   - webhook or status update lag
   - time-to-credit
   - stuck transaction count

5. Circuit breaker
   - mark route degraded
   - pause new traffic
   - shift new traffic to next eligible route
   - allow operator override

6. Operations dashboard
   - live route health
   - transactions by state
   - failed/stuck queue
   - route selected and why
   - provider/corridor SLA view

7. Basic reconciliation
   - import provider/bank settlement files
   - match by reference, amount, currency, and beneficiary
   - show unmatched, delayed, duplicate, or mismatched items

### MVP UI Scope

The first UI should be polished, but ruthlessly focused:

1. Control room
   - global health
   - corridor/provider status
   - active incidents
   - stuck transactions
   - recommended action

2. Corridor detail
   - provider ranking
   - success rate
   - P50/P95 latency
   - cost/FX comparison
   - current traffic split

3. Transaction trace
   - lifecycle timeline
   - selected route and reason
   - current state
   - owner of next action

4. Route configuration
   - enable/disable provider
   - traffic split presets
   - fallback route order
   - preview impact
   - audit reason

5. FX and cost board
   - current provider rates
   - fees/spread
   - effective cost
   - stale-rate warnings

6. Latency drilldown
   - end-to-end latency
   - step-level waterfall
   - recent downtime events

Do not build every screen in the full information architecture for the pilot. Build the screens that help the bank see, decide, switch, and prove impact.

### MVP Explicitly Out of Scope

- cash pickup orchestration unless the anchor bank needs it immediately
- every bank in Nigeria
- every IMTO listed by CBN
- AI-based routing
- advanced fraud modeling
- provider commercial negotiation tooling
- automated settlement movement
- customer-facing dispute portal
- deep CRM integration
- fully automated compliance case management
- custom theming per bank
- complex drag-and-drop workflow builders
- advanced BI dashboards unrelated to routing reliability

### Delivery Order

1. Build monitoring and transaction trace before auto-switching.
2. Build manual route override before automatic circuit breakers.
3. Build rules-based routing before weighted optimization.
4. Build one real provider adapter end-to-end before adding many shallow adapters.
5. Build reconciliation exceptions before advanced settlement analytics.
6. Add cash pickup, wallets, and Africa-wide corridors only after account payout is stable.

### 90-Day Pilot Shape

Month 1:

- anchor bank discovery
- corridor/provider selection
- canonical transaction model
- provider registry
- transaction intake API
- first dashboard version

Month 2:

- first 2-3 provider/rail adapters
- transaction trace timeline
- route eligibility engine
- manual override
- provider health monitor

Month 3:

- weighted scoring
- circuit breaker automation
- fallback routing
- reconciliation exception view
- SLA scorecards
- pilot report showing saved failures, time saved, cost impact, and incident response improvement

### Phase 2 - Africa corridors

Add PAPSS and pan-African payout networks:

- PAPSS for Africa-to-Africa local currency payments.
- Onafriq for mobile money/wallet-heavy markets.
- KCB/Equity-style providers for East Africa.
- Ecobank Rapidtransfer, UBA AfriCash, AccessAfrica for bank-owned pan-African rails.

### Phase 3 - Intelligent routing

Use historical performance and real-time health to auto-select routes:

- cheapest route that meets SLA
- fastest route under amount/compliance constraints
- fallback route if provider/bank rail is degraded
- regulator-approved route only
- bank-preferred provider weighting

## Key Design Principle

Treat every transaction as a traceable workflow, not just a payment message. The switch should know:

- who initiated it
- which provider accepted it
- which bank received it
- why it is delayed
- whether customer value has been delivered
- whether settlement has matched
- who owns the next action

## Strategic Moat

The defensible asset is not the adapter code alone. It is the reliability intelligence layer built from cross-provider, cross-bank, cross-corridor performance data.

The more banks and providers connected, the better the switch becomes at answering:

- which provider is fastest for this exact corridor today
- which route is cheapest after fees, FX, and settlement cost
- which provider is likely to fail for this destination bank
- which corridor needs liquidity before SLA is breached
- which rail is showing early degradation before public downtime
- which provider disputes are repeating and need commercial escalation

This creates a network effect around operational truth.

## Sources

- CBN licensed IMTO list: https://www.cbn.gov.ng/PaymentsSystem/InternationalMoneyTransferOperators.html
- World Bank Remittance Prices Worldwide: https://datacatalog.worldbank.org/search/dataset/0037898/remittance-prices-worldwide
- Access Bank money transfer: https://www.accessbankplc.com/personal/money-transfer
- UBA remittance services: https://www.ubagroup.com/nigeria/remittance-services/
- FirstBank money transfer: https://www.firstbanknigeria.com/personal/money-transfer/
- GTBank Remitly and IMT services: https://www.gtbank.com/personal-banking/services/imt-services/remitly
- Zenith IMT: https://www.zenithbank.com/personal-banking/electronic-banking/imt/
- Ecobank remittance partners: https://www.ecobank.com/personal-banking/payments-transfers/remittance-partners
- Thunes and Ecobank partnership: https://www.thunes.com/news/thunes-and-ecobank-group-to-power-africas-instant-payments-for-the-next-billion-users/
- PAPSS: https://papss.com/
- PAPSS get connected: https://papss.com/get-connected/
- TerraPay network: https://www.terrapay.com/network/
- Onafriq network: https://onafriq.com/
- KCB Kenya IMT services: https://ke.kcbgroup.com/ways-of-banking/international-money-transfer-services?trk=test
- Equity Bank Kenya money transfer: https://equitygroupholdings.com/ke/diaspora-banking/money-transfer/
- Standard Bank MoneyGram: https://www.standardbank.co.za/southafrica/personal/products-and-services/bank-with-us/foreign-exchange/moneygram-transfers
- Absa Western Union: https://www.absa.co.za/personal/bank/international-banking/westernunion/
