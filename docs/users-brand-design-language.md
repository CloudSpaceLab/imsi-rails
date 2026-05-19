# imsi-rails Users, Brand, and Design Language

## Purpose

This document defines the exact users, brand posture, copy system, visual language, and product UI rules for `imsi-rails`.

The goal is an internationally competitive bank-grade product: premium, calm, fast, precise, and trusted by people who operate critical cross-border money movement.

## Exact Users

`imsi-rails` is sold to banks, but the product must serve several different user groups with very different jobs.

### 1. Executive Buyer

Common titles:

- Group Head of Payments
- Head of Retail Banking
- Head of Digital Banking
- Head of Transaction Banking
- Chief Operating Officer
- Executive Director, Banking Operations

Primary question:

> Is this platform going to protect remittance revenue, reduce operational risk, and improve customer trust?

Needs:

- clear reliability improvement
- stronger provider accountability
- better remittance revenue visibility
- reduced customer complaints
- board-level summary of performance and risk
- confidence that the platform is not replacing the bank's provider relationships

Best surfaces:

- executive dashboard
- pilot success report
- provider scorecard summary
- incident impact report
- revenue/cost/reliability trendline

Copy tone:

- commercial
- confident
- outcome-led
- low jargon

### 2. Economic Buyer

Common titles:

- Head of Remittances
- Head of Diaspora Banking
- Head of International Money Transfer
- Product Owner, Cross-Border Payments

Primary question:

> Which providers should get more traffic, and how does that improve volume, cost, speed, and customer experience?

Needs:

- provider comparison by corridor
- FX and fee visibility
- success rate and time-to-credit
- route recommendation rationale
- evidence for provider negotiations
- controlled rollout of new providers

Best surfaces:

- corridor detail
- FX and cost board
- provider scorecards
- routing intelligence dashboard
- shadow-routing reports

Copy tone:

- comparative
- practical
- performance-oriented
- precise about tradeoffs

### 3. Daily Operator

Common titles:

- Payments Operations Analyst
- Remittance Operations Officer
- Settlement Operations Analyst
- Treasury Operations Analyst
- Customer Escalation Specialist

Primary question:

> What is broken, who owns it, and what should I do next?

Needs:

- live health state
- stuck transaction queue
- current owner of next action
- safe manual switches
- incident timeline
- provider escalation evidence
- reconciliation exceptions

Best surfaces:

- control room
- transaction trace
- incident detail
- route configuration
- reconciliation exceptions

Copy tone:

- direct
- action-oriented
- unambiguous
- calm during incidents

### 4. Technical Integrator

Common titles:

- Payments Integration Lead
- Solutions Architect
- Core Banking Engineer
- API Platform Engineer
- Infrastructure/DevOps Lead

Primary question:

> Can this integrate cleanly, run reliably, and stay observable inside bank infrastructure?

Needs:

- clear API contracts
- idempotency and retry behavior
- adapter isolation
- deployment architecture
- logs, metrics, traces
- security model
- performance budgets

Best surfaces:

- API docs
- integration dashboard
- adapter health view
- deployment guide
- observability and runbooks

Copy tone:

- exact
- contract-driven
- implementation-aware
- no marketing fluff

### 5. Risk, Compliance, and Audit User

Common titles:

- Compliance Manager
- AML Officer
- Operational Risk Lead
- Internal Audit
- Regulator Relations Lead

Primary question:

> Can we prove that only approved routes were used and that every sensitive decision was controlled?

Needs:

- approved provider/corridor controls
- route decision audit
- config change audit
- maker-checker approvals
- exception history
- exportable evidence
- policy versions

Best surfaces:

- audit log
- route decision detail
- policy diff
- compliance exceptions
- export center

Copy tone:

- formal
- evidence-based
- precise
- conservative

### 6. Provider Relationship Manager

Common titles:

- IMTO Partnerships Manager
- Strategic Partnerships Lead
- Vendor Manager
- Provider Operations Manager

Primary question:

> Which providers are performing, which are causing issues, and what evidence do we have?

Needs:

- provider SLA scorecard
- incident history
- support responsiveness
- reconciliation break rate
- volume shifted due to poor performance
- provider improvement reports

Best surfaces:

- provider profile
- provider scorecard
- incident exports
- monthly performance report

Copy tone:

- fair
- objective
- evidence-led
- non-accusatory

### 7. Bank Relationship / Sales User

This is an internal `imsi-rails` user, not the bank.

Primary question:

> Can we show the bank why this matters and prove value quickly?

Needs:

- demo environment
- landing page
- pilot report template
- sample degraded corridor story
- before/after routing impact
- executive summary deck inputs

Best surfaces:

- demo control room
- simulated incident mode
- pilot report
- sales narrative assets

Copy tone:

- crisp
- premium
- credible
- not exaggerated

## User Priority

MVP user priority:

1. Daily Operator
2. Head of Remittances / Economic Buyer
3. Technical Integrator
4. Risk and Compliance
5. Executive Buyer

Reasoning:

- Daily operators prove the product works.
- Remittance/product owners sponsor expansion.
- Technical integrators decide whether implementation is painful.
- Risk/compliance can block adoption if auditability is weak.
- Executives need proof, not every workflow.

## Brand Strategy

### Brand Idea

`imsi-rails` is the reliability layer beneath international money movement.

It turns fragmented IMTO relationships into measurable, configurable, and self-correcting bank infrastructure.

### Brand Promise

Every connected IMTO rail becomes observable, comparable, and safely switchable.

### Positioning Statement

For banks that operate international money transfer programs, `imsi-rails` is an IMTO reliability control tower that monitors every connected provider, scores every eligible route, and switches traffic to the best-performing rail according to bank policy.

Unlike remittance apps or provider portals, `imsi-rails` gives banks one operating truth across all providers, corridors, FX/cost data, failures, and route decisions.

### Personality

- precise
- calm
- premium
- decisive
- technical without being cold
- operational without feeling old
- neutral and fair to providers
- serious about money movement

### Brand Tension

The product should feel like a high-end aviation control room for money movement:

- alive, but not noisy
- intelligent, but explainable
- powerful, but safe
- beautiful, but never decorative at the expense of clarity

## Naming System

Preferred product phrase:

- IMTO reliability infrastructure for banks

Secondary phrases:

- international money transfer reliability layer
- bank-grade IMTO control tower
- routing and switching intelligence for IMTO rails
- provider performance control plane
- cross-border transfer reliability dashboard

Avoid:

- remittance app
- money exchange app
- crypto rail
- public bidding marketplace
- generic payment gateway
- all-in-one financial super app

## Copy System

### Voice

Clear, confident, operational, and evidence-led.

Write like the product is used by serious people during real incidents. Avoid hype. Avoid vague fintech language.

### Copy Rules

- Lead with the operational outcome.
- Prefer verbs that imply control: monitor, route, switch, trace, prove, compare, recover.
- Explain tradeoffs directly.
- Use "eligible route" instead of "best route" when compliance or policy constraints matter.
- Use "recommended" when the platform advises, and "selected" when the engine decides.
- Always distinguish live data from historical data.
- Never imply the system can guarantee external provider uptime.
- Never hide uncertainty; mark stale, delayed, estimated, or unavailable data clearly.

### Example Headlines

Good:

- One control tower for every IMTO provider.
- Route every transfer through the best eligible rail.
- Detect degraded corridors before customers complain.
- See why traffic moved, who owns the issue, and what happens next.
- Let provider performance decide traffic allocation.

Avoid:

- The future of remittance.
- Send money faster.
- Banking made simple.
- One app for all your money needs.
- AI-powered global payments revolution.

### In-Product Copy Patterns

Status:

- Healthy
- Watch
- Degraded
- Blocked
- Recovery testing
- Stale data

Actions:

- Shift traffic
- Pause new traffic
- Preview policy
- Run shadow test
- Roll back policy
- Export evidence
- Open incident
- Assign owner

Explanations:

- Selected because it met SLA at the lowest effective cost.
- Rejected because the FX rate is stale.
- Rejected because the route is blocked by circuit breaker.
- Recommended shift because P95 time-to-credit breached policy for 15 minutes.
- Manual approval required because the change affects high-value traffic.

## Visual Design Language

### Visual North Star

Premium operational intelligence.

The UI should look internationally competitive against high-end fintech, cloud infrastructure, observability, and treasury platforms, while staying simpler and more readable than generic BI dashboards.

### Design Mood

- dark operational canvas
- luminous signal colors
- compact but breathable layouts
- crisp typography
- purposeful motion
- sharp charts
- quiet confidence

### Color System

Color must carry meaning.

Core palette:

- Ink: primary dark background
- Panel: elevated surfaces
- Paper: primary text on dark
- Mist: secondary text
- Mint: healthy/success/primary action
- Amber: warning/watch
- Coral: degraded/critical
- Cyan: live data/latency/flow
- Violet: policy/simulation/shadow mode

Rules:

- Do not overuse gradients.
- Do not use color only; pair color with labels/icons.
- Red/coral is only for real risk.
- Mint is for positive state and primary action, not decoration everywhere.
- Violet marks simulation, policy previews, and shadow routing.
- Cyan marks live flow, latency, and route movement.

### Typography

Principles:

- Use a modern sans-serif with strong numerals.
- Data-heavy UI needs tabular numbers.
- Keep headings compact inside operational panels.
- Do not use giant marketing type inside the product app.
- Use consistent case:
  - Title Case for main sections
  - sentence case for explanations and form labels
  - uppercase only for short operational labels

### Layout

Principles:

- Dense, not cramped.
- Top-level screens should be scannable in 5 seconds.
- Drilldowns should reveal depth progressively.
- High-risk actions should stay visually close to their impact preview.
- Every real-time screen needs a visible freshness indicator.

Preferred structures:

- control-room summary row
- corridor matrix
- provider scorecard grid
- transaction timeline
- side drawer for audit/context
- split view for policy current vs proposed
- right-side recommendation panel

Avoid:

- nested cards inside cards
- decorative panels that do not carry data
- hero-style marketing composition inside the product app
- dashboards that require horizontal scrolling for core tasks

### Motion

Motion should indicate state change, not entertain.

Use motion for:

- live value updates
- route movement
- incident state transitions
- policy preview differences
- loading skeletons

Avoid:

- continuous animations in dense dashboards
- motion that distracts during incidents
- animations longer than 250 ms for operational actions

## Component Language

### Health Badge

Purpose:

- communicate route/provider state instantly

States:

- Healthy
- Watch
- Degraded
- Blocked
- Recovery
- Unknown

Required metadata:

- state
- measurement window
- last updated
- trigger reason

### Route Score Chip

Purpose:

- show why a route is recommended or selected

Contents:

- total score
- primary reason
- confidence
- policy version

Example:

- 92 / Selected / lowest cost within SLA
- 81 / Eligible / higher P95 latency
- 0 / Rejected / circuit breaker open

### Corridor Matrix

Purpose:

- let operators scan origin-destination risk quickly

Cells show:

- health state
- transaction volume
- P95 latency
- degraded provider count
- recommended action

### Transaction Timeline

Purpose:

- show the lifecycle of one transaction from intake to settlement

Rules:

- every step has timestamp, status, owner, source system, and reference
- current blocker is visually dominant
- audit details open in a drawer

### FX Comparison Table

Purpose:

- compare provider economics without becoming a trading terminal

Columns:

- provider
- rate
- rate age
- fee
- spread
- effective cost
- speed
- reliability
- recommendation

Rules:

- stale rates are visibly blocked or downgraded
- cheapest route must show if it was not selected and why

### Traffic Split Control

Purpose:

- safely move volume between routes

Rules:

- show current split and proposed split
- show affected corridor/volume
- require reason
- preview expected impact
- maker-checker for sensitive changes
- rollback is visible after activation

### Policy Diff Viewer

Purpose:

- help approvers understand what changes before approving

Must show:

- current policy
- proposed policy
- changed fields
- expected historical impact
- risk level
- rollback target

## Product Screen Personality

### Control Room

Feeling:

- calm command center

Must answer:

- what is happening now?
- what is at risk?
- where should we act?

### Corridor Detail

Feeling:

- performance lab

Must answer:

- which provider is best for this route right now?
- what changed?
- what should traffic do?

### Transaction Trace

Feeling:

- forensic timeline

Must answer:

- where is this transfer?
- who has it?
- why is it delayed?
- can we safely intervene?

### Route Configuration

Feeling:

- protected cockpit

Must answer:

- what will this change affect?
- is it safe?
- who approved it?
- how do we roll back?

### FX and Cost Board

Feeling:

- economics lens

Must answer:

- which provider is cheapest?
- which provider is worth the cost?
- which rates are stale?

### Latency Drilldown

Feeling:

- diagnostic console

Must answer:

- where exactly is time being lost?
- is this systemic or isolated?
- did switching help?

## Accessibility and International Readiness

Requirements:

- high color contrast
- no information conveyed by color alone
- keyboard navigable critical workflows
- screen-reader labels for critical status and controls
- date/time formats configurable by bank
- currency formatting by corridor
- UTC plus bank-local time display
- clear distinction between country, corridor, currency, rail, provider, and bank
- exportable evidence for audit and provider conversations

## Design Quality Gates

A screen is not done until it passes these checks:

- the primary user can answer the screen's main question in under 5 seconds
- live/stale data state is obvious
- all critical numbers have units and time windows
- every recommended action explains why
- every destructive or traffic-changing action has preview and rollback path
- keyboard use works for critical controls
- mobile/tablet executive read-only view does not break
- performance budget is respected

## Brand Decision

The design language should make `imsi-rails` feel like infrastructure with taste.

The product should not look like:

- a generic admin dashboard
- a consumer money transfer app
- a crypto trading terminal
- a decorative startup landing page
- a spreadsheet with charts

It should look and sound like the bank's trusted international transfer control layer.

