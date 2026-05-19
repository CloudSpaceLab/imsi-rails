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
| Executive Management | reliable remittance revenue, customer trust, provider accountability |

### Sales Narrative

Lead with reliability, not technology:

1. Banks lose trust and revenue when international transfers fail or stall.
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

Defer these until the core switch is trusted:

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

This audit trail is essential for bank trust, provider disputes, and regulatory review.

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
