# IMTO Orchestration and Transfer Rails Backend Architecture

Research date: 2026-05-21

## Purpose

This note turns the current `imsi-rails` product thesis into a backend architecture for a single bank-facing transfer rails API across IMTOs, payout networks, and African payment rails.

The central decision is this:

> Build a transaction orchestration and observability layer, not a thin HTTP proxy.

If the product is described commercially as a "single proxy" or "single integration surface," that must mean a robust orchestration layer: it routes, submits, monitors, reconciles, reports, and controls risk across rails. It must never mean a pass-through adapter that simply forwards requests to providers.

The provider APIs do not share the same lifecycle, finality boundary, error model, callback behavior, quote semantics, or reconciliation path. A useful orchestrator must normalize those differences while preserving enough provider-specific evidence to debug exceptions, prove latency ownership, and avoid duplicate payouts.

## Evidence From The Current Repo

The repo already has the right early foundation:

- `internal/core` models canonical transaction states, route registry, eligibility filtering, weighted scoring, route decisions, and default failover safety boundaries.
- `internal/intake` exposes an idempotent transaction intake service and records route decisions.
- `internal/health` ingests provider and route health samples, computes route snapshots, and evaluates circuit breaker state.
- `cmd/imsi-api/main.go` seeds the first provider set: Thunes, Remitly, Ria, and PAPSS.
- `plan.md` and `docs/masterplan.md` identify Tier 1 expansion providers: Western Union, MoneyGram, Ria, WorldRemit/Sendwave, Remitly, Transfast/Mastercard Cross-Border Services, PAPSS, TerraPay, Thunes, and Onafriq.

The repo intentionally does not yet implement real provider submission, provider callbacks, status polling, reconciliation, settlement matching, or persistent transaction event storage. Those are the pieces that turn route selection into a true transfer rail.

## Backend Review Findings

Reviewed against the Go backend on 2026-05-21.

| Area | Current implementation | Planning implication |
| --- | --- | --- |
| Transaction intake | `internal/intake/service.go` validates requests, enforces idempotency, selects a route, stores an in-memory record, and emits two in-memory lifecycle events. | The next layer must create durable transfers, attempts, provider operations, and outbox events after route selection. |
| Routing core | `internal/core/routing.go` supports static route capabilities, eligibility rejection, weighted scoring, and route decision audit shape. | The plan is accurate, but routing needs dynamic capability, quote, liquidity, prefund, FX, and settlement snapshots before real providers. |
| Lifecycle state | `internal/core/lifecycle.go` has canonical states and a conservative failover boundary. | The orchestrator must own state transitions and track attempt-level finality so adapters cannot mutate transfer state directly. |
| Sandbox adapter | `internal/core/adapter.go` simulates latency, timeout, rejection, failure, and duplicate callbacks. | It needs richer IMTO scenarios: quote expiry, unknown outcome, delayed/missing/out-of-order callbacks, settlement mismatch, and provider-instructed payout. |
| Health engine | `internal/health/service.go` accepts provider API, timeout, error, callback lag, and transaction outcome samples, then evaluates circuit breakers. | Health should be fed by provider operation events and rollups, not only direct sample ingestion. |
| Dashboard API | `internal/dashboard/service.go` aggregates seeded in-memory transaction metrics and emits short-lived SSE demo updates. | Production dashboard analytics must consume orchestration events/rollups and expose incident, exception, and reconciliation impact. |
| Auth/RBAC | `internal/auth` includes local login, LDAP/OIDC hooks, roles, and permissions including reconciliation, policy, audit, and incidents. | RBAC foundations exist, but orchestration actions still need permission checks, maker-checker, and audit events. |
| API contract | `api/openapi.yaml` covers `/v1/transactions`, health, auth/admin, and dashboard endpoints. | The public contract needs canonical `/v1/transfers`, provider callback ingress, timeline, manual actions, reconciliation, reports, and audit endpoints. |
| Persistence/eventing | `cmd/imsi-api/main.go` wires in-memory stores; NATS/MariaDB are documented but not implemented in the runtime. | MariaDB repositories, transactional outbox, and JetStream publication are required before a bank pilot. |

This confirms the architecture should extend the existing core rather than replace it. The current code is a solid route-decision and monitoring foundation; the missing product layer is durable orchestration after a route is selected.

## Provider API Patterns We Must Normalize

The top providers fall into a few operational archetypes.

| Archetype | Providers | Typical API shape | Architecture implication |
| --- | --- | --- | --- |
| Payout network / aggregator | Thunes, TerraPay, Onafriq | discover capability, validate account/wallet, quote or price, create/confirm transaction, callback and status lookup | Model dynamic capabilities, quote freshness, asynchronous final status, balance/prefund, local partner failure reasons |
| Legacy IMTO / money transfer network | MoneyGram, Ria, Western Union | quote/update/commit or partner payout instruction, transaction status/sub-status, cash pickup/account/wallet modes, private partner workflows | Support both "we send to provider" and "provider instructs us to pay" directions, status subcodes, cash pickup finality, private adapter specs |
| Scheme or clearing rail | PAPSS, Mastercard Cross-Border Services/Transfast | payment instruction, status messages, balance/prefund, settlement reports/files, cancellation and return processes | Treat status, settlement, and clearing as separate timelines; build ISO/message-file adapters as first-class citizens |
| Bank-owned regional rail | AccessAfrica, UBA AfriCash, Ecobank Rapidtransfer, GTMT, First Global Transfer | often private APIs, batch files, internal bank posting, branch operations | Keep adapter runtime protocol-agnostic: REST, SOAP, SFTP, ISO 20022, CSV, and manual import |

Important provider observations from public docs:

- Thunes Money Transfer has a multi-step flow: payer discovery, optional credit-party information/verification, quote, transaction creation, confirmation, then callback or polling for status. Confirming can hold the source amount, and callback failures require status polling fallback.
- MoneyGram exposes both a send-side Transfer API flow of quote, update, and commit, and a payout-partner flow where MoneyGram calls the receiving institution for account validation and fund transfer, then the partner reports final status by webhook.
- MoneyGram webhooks are status/sub-status events. They should be acknowledged quickly, can be retried, and may omit PII, so the switch must correlate by identifiers rather than expecting a full transaction payload.
- TerraPay status handling is explicitly asynchronous: callback and view-transaction APIs expose final and pending states, and response codes indicate whether to retry, cancel, or continue status enquiry.
- PAPSS is not just a REST payout API. It is an instant cross-border payment system with instant payment, pre-funding, net settlement, and ISO 20022 status messages across participants and RTGS systems.
- Mastercard Cross-Border Services exposes quote, payment, retrieve payment, cancel, status-change, rate, balance, and reporting capabilities. Its retrieve-payment specification warns not to poll pending payments too aggressively.
- Onafriq publishes a developer portal for disbursements and remittance/bulk payment APIs, but public pages still imply partner-specific API selection and onboarding support.
- Ria publicly confirms API integration as a digital partnership model, but the operational details are partner-specific. Remitly is seeded in the repo, but public technical API details are not enough to design an adapter without partner documentation.

## Core Product Contract

The bank-facing "Inswitch Transfer Rails API" should not mirror provider APIs directly. It should expose stable transfer semantics:

1. Receive a transfer intent from a bank/channel.
2. Validate tenant, idempotency, beneficiary, compliance posture, and required fields.
3. Select an eligible route using cached policy, capability, FX/cost, liquidity, and health.
4. Create one or more provider attempts under a single canonical transfer.
5. Submit provider operations asynchronously when needed.
6. Normalize callbacks, status polling results, bank posting events, and reconciliation files into one lifecycle timeline.
7. Detect stuck, unknown, delayed, failed, duplicate, reversed, and mismatched states.
8. Explain every route decision, attempt, exception, and operator intervention.

Suggested external API groups:

| API group | Endpoints | Purpose |
| --- | --- | --- |
| Transfer intake | `POST /v1/transfers`, `GET /v1/transfers/{id}` | Accept transfer intent and return canonical transfer state |
| Quote and preview | `POST /v1/quotes`, `POST /v1/routes/preview` | Show eligible routes, prices, FX, and route rationale before commitment where needed |
| Lifecycle | `GET /v1/transfers/{id}/timeline`, `GET /v1/transfers/{id}/events` | Expose step-level trace and audit evidence |
| Actions | `POST /v1/transfers/{id}/cancel`, `POST /v1/transfers/{id}/retry`, `POST /v1/transfers/{id}/manual-actions` | Safe operational intervention |
| Provider callbacks | `POST /v1/provider-callbacks/{provider_id}` | Single secured ingress for webhooks and status notifications |
| Provider health | `POST /v1/health/samples`, `GET /v1/health/routes/{route_id}` | Existing health model, extended by adapter runtime |
| Capabilities | `GET /v1/providers/{id}/capabilities`, `GET /v1/routes` | What a bank can route to now, with freshness |
| Reconciliation | `POST /v1/reconciliation/files`, `GET /v1/reconciliation/exceptions` | Settlement and transaction matching |
| Reports | `POST /v1/reports`, `GET /v1/reports/{id}`, `GET /v1/reports/{id}/download` | Reconciliation, incident, provider, SLA, audit, and pilot evidence packs |
| Audit | `GET /v1/audit/events`, `GET /v1/route-decisions/{id}` | Decision and configuration proof |

The API response should be honest about finality. For example, `accepted` should mean "the switch accepted the transfer intent", not "beneficiary value has been delivered."

## Canonical Transfer Model

Use a three-level model:

1. `transfer`
   - The bank/customer intent.
   - Stable `switch_reference`, `bank_id`, sender, receiver, amount, payout method, compliance posture, selected policy version.

2. `transfer_attempt`
   - A selected route attempt against one provider/rail.
   - Contains route decision snapshot, provider id, route id, attempt number, safety boundary, finality state, and fallback eligibility.

3. `provider_operation`
   - Each external call or file/message exchange.
   - Examples: quote, beneficiary validation, create transaction, confirm transaction, status query, cancel, reversal request, bank posting, settlement file import.

This avoids a common trap: pretending every provider has one request and one response. Thunes and MoneyGram may create a quote resource before a transaction exists. PAPSS may emit multiple ISO status messages. A provider callback can arrive before a synchronous response is processed. A single transfer needs a timeline, not one status field.

## Provider Adapter Contract

Each provider adapter should be isolated behind a contract like this:

```text
ProviderAdapter
  Capabilities(ctx, scope) -> CapabilitySnapshot
  Quote(ctx, transfer, route) -> QuoteResult
  ValidateBeneficiary(ctx, transfer, route) -> ValidationResult
  CreateTransfer(ctx, transfer, attempt, quote) -> ProviderOperationResult
  ConfirmTransfer(ctx, transfer, attempt) -> ProviderOperationResult
  QueryStatus(ctx, providerReference | externalID) -> ProviderStatusResult
  Cancel(ctx, attempt) -> ProviderOperationResult
  ParseCallback(rawRequest) -> ProviderEvent
  ParseReconciliationFile(file) -> ReconciliationBatch
```

Every adapter must also declare:

- supported corridors, payout methods, currencies, amount limits, destination banks/wallets, required fields, and capability freshness
- provider idempotency behavior and which fields form a unique external reference
- safe retry rules for each operation
- safe failover boundary
- status mapping from provider states into canonical states
- polling policy and callback expectations
- rate limits and backoff strategy
- authentication, signing, encryption, and IP allowlist requirements
- PII classification for request, response, callback, and logs
- settlement and reconciliation file formats

## Backend Architecture

### 1. Bank Edge API

Responsibilities:

- tenant auth, RBAC, idempotency, request validation
- input canonicalization: country/currency codes, payout method, destination bank identifiers, amount precision
- no heavy analytics or provider-specific logic
- write transfer intent and route decision atomically

### 2. Transfer Orchestrator

Responsibilities:

- owns the canonical transfer state machine
- creates provider attempts
- advances states from provider events, bank posting events, compliance events, and reconciliation events
- prevents illegal transitions and duplicate value delivery
- emits lifecycle events and timeline spans

This is the heart of the system. Provider adapters should never directly mutate final transfer state. They emit facts; the orchestrator decides what those facts mean.

### 3. Routing and Policy Service

Responsibilities:

- route eligibility and scoring
- policy versioning
- route decision audit
- fallback route list
- shadow routing
- traffic split and circuit breaker enforcement

The existing `internal/core` model is a good start. Extend it with capability snapshots, quote freshness, liquidity snapshots, settlement risk, and provider-specific finality metadata.

### 4. Adapter Runtime

Responsibilities:

- executes provider operations outside the hot path
- enforces per-provider timeouts, retries, backoff, rate limits, and circuit breaker state
- stores raw request/response envelopes with redaction
- emits provider latency samples and normalized provider events
- supports REST, SOAP, SFTP, CSV, ISO 20022, batch files, and manual import

Adapters should be worker-owned, not API-thread-owned. The intake API can return quickly while the worker drives the provider attempt.

### 5. Webhook Inbox and Status Poller

Responsibilities:

- verify signatures, auth headers, source IPs, timestamps, replay windows, and schemas
- persist raw callbacks before acknowledging
- deduplicate by provider event id, provider reference, subscription id, timestamp, and payload hash
- acknowledge quickly, then process asynchronously
- handle duplicate, delayed, out-of-order, and missing callbacks
- query provider status when callback state is missing, stale, or ambiguous

This needs its own inbox table and processing queue. Webhook handling cannot be "parse and update transaction inline" because providers retry and deliver out of order.

### 6. Event Bus and Outbox

Use the outbox pattern from MariaDB to NATS JetStream:

- database transaction writes transfer state and outbox event
- publisher relays committed events to NATS
- dashboard, health, alerting, reconciliation, and analytics consume events

Core subjects:

- `transfer.created`
- `route.decision.recorded`
- `provider.operation.started`
- `provider.operation.completed`
- `provider.callback.received`
- `transfer.state.changed`
- `transfer.exception.created`
- `health.route.changed`
- `circuit_breaker.changed`
- `reconciliation.item.matched`
- `audit.event.recorded`

### 7. Observability and Timeline Engine

Measure both technical latency and business latency.

Technical spans:

- intake validation
- route decision
- quote call
- beneficiary validation call
- provider create call
- provider confirm call
- status query
- callback processing
- bank name enquiry
- bank posting
- reconciliation import

Business milestones:

- transfer received
- route selected
- provider accepted
- payout submitted
- value delivered
- customer/bank notified
- settlement received
- reconciled

Store all timeline events with:

- `transfer_id`
- `attempt_id`
- `provider_operation_id`
- `trace_id`
- `source_system`
- `canonical_state`
- `provider_state`
- `occurred_at`
- `observed_at`
- `recorded_at`
- `owner`
- `latency_ms`
- references and redacted raw payload link

The dashboard should be able to answer: "Where did the time go?" without needing provider logs.

### 8. Exception Engine

Exceptions should be first-class records, not just failed statuses.

Core exception categories:

| Category | Example trigger | Owner |
| --- | --- | --- |
| `unknown_outcome` | create/confirm timed out after provider may have accepted | switch/provider |
| `callback_missing` | no callback within expected SLA | provider |
| `callback_duplicate` | duplicate provider event | switch |
| `callback_out_of_order` | terminal status followed by older pending status | switch/provider |
| `provider_rejected` | validation, compliance, inactive payer, unsupported bank | provider/bank |
| `beneficiary_validation_failed` | invalid account, wallet not found, name mismatch | bank/customer/provider |
| `compliance_hold` | sanctions, KYC, purpose, RFI/manual review | compliance |
| `liquidity_or_prefund_low` | balance below required threshold | treasury/provider |
| `bank_posting_failed` | NIP/core banking failure after provider acceptance | bank |
| `duplicate_payout_risk` | retry/failover requested after finality boundary | switch/operator |
| `stuck_pending` | PENDING beyond route SLA | provider/switch |
| `reversal_pending` | failed after debit or value delivered incorrectly | provider/settlement |
| `reconciliation_break` | provider says paid, bank ledger missing or amount mismatch | finance/ops |
| `fx_or_fee_mismatch` | settlement amount/rate differs from quote | finance/provider |

Each exception needs severity, value at risk, SLA clock, current owner, next recommended action, and linked evidence.

### 9. Reconciliation Engine

Reconciliation should be automatic by default. The system should remove routine manual reconciliation work, but it should not claim that manual intervention will never be needed. The safer bank-grade promise is:

> Manual reconciliation becomes exception review, not a daily matching process.

The reconciliation engine should ingest provider files, bank posting files, settlement reports, callbacks, ledger references, and transfer events. It should automatically match by switch reference, provider reference, bank posting reference, amount, currency, beneficiary account/wallet, settlement batch, and value date.

Recommended reconciliation states:

| State | Meaning | Manual action |
| --- | --- | --- |
| `auto_matching` | Matching is still running inside expected SLA | None |
| `auto_matched` | Provider, bank, and settlement evidence agree | None |
| `pending_external_evidence` | Waiting for callback, provider report, bank file, or settlement file | None unless SLA is breached |
| `exception_review_required` | Available evidence conflicts or is insufficient for safe resolution | Required |
| `manually_resolved` | Operator resolved with reason, evidence, and audit trail | Completed |
| `written_off_or_adjusted` | Finance approved an adjustment/write-off path | Completed with maker-checker |

Pending auto-reconciliations should remain out of the manual work queue until they cross an SLA, conflict with another evidence source, or create value-at-risk. Operators should review the exception queue, not re-run routine matching by hand.

Manual review should be reserved for cases such as:

- provider says paid, but the bank ledger has no matching credit
- provider and bank agree on reference but disagree on amount, currency, FX rate, fee, or value date
- duplicate provider callback or duplicate bank posting risk
- missing, late, malformed, or partial settlement file
- timeout or ambiguous provider status creates an `unknown_outcome`
- reversal, refund, dispute, cash pickup, or compliance hold needs human confirmation
- settlement report does not match the provider-paid population

Every manual resolution must record actor, role, timestamp, reason code, supporting evidence, approval status, financial impact, and whether customer value was delivered.

### 10. Reporting and Evidence Packs

Reports should be generated from immutable events, reconciliation records, incident timelines, and audit logs. They should not depend on screenshots or operator memory.

Core report types:

| Report | Trigger | Contents | Audience |
| --- | --- | --- | --- |
| Reconciliation exception report | Daily, weekly, monthly, or on demand | unmatched items, duplicates, amount/FX mismatches, aging, value at risk, owner, resolution status | Finance, operations, provider managers |
| Settlement report | Settlement cycle close or on demand | expected settlement, received settlement, prefund movements, batch totals, suspense items, settlement aging | Treasury, finance |
| Incident report | Incident open/close or postmortem | timeline, affected corridors/routes, transactions impacted, customer value at risk, circuit breaker actions, operator actions, root cause, prevention steps | Operations, executives, provider managers |
| Provider SLA report | Weekly/monthly/provider review | success rate, P50/P95/P99 latency, callback lag, error rate, stuck rate, reconciliation breaks, support response time | Head of remittances, provider relationship manager |
| Route decision audit report | Regulator/audit request or policy review | selected routes, rejected routes, policy version, scoring inputs, overrides, maker-checker approvals | Risk, compliance, audit |
| Exception resolution report | Shift handover or management review | open exceptions, SLA breaches, manual resolutions, unresolved value at risk, next owner | Operations leads |
| Pilot impact report | Pilot close or steering committee | failures avoided, traffic shifted, time-to-credit improvement, cost/FX impact, incident response improvement | Executive sponsors |

Report jobs should support:

- filters by bank, corridor, provider, payout method, currency, date range, settlement batch, exception type, and incident id
- CSV/XLSX exports for operations and finance
- PDF evidence packs for executives, audit, and provider escalation
- scheduled delivery with role-based access controls
- reproducible report versions tied to source event ranges and generation time
- redaction policies for PII and commercially sensitive provider terms
- report run audit: requester, parameters, generated files, access history, and retention policy

Incident and reconciliation reports should be linked. For example, if a provider incident caused late settlement or missing callbacks, the incident report should show the downstream reconciliation impact, and the reconciliation exception should link back to the incident timeline.

## Hot User Flows and Risk Points

### Flow 1: Direct-to-account inbound transfer

Risk points:

- route picked with stale FX or stale capability data
- account validation passes but posting fails later
- provider accepts transfer but callback is delayed
- provider status says completed but bank credit is missing
- timeout on submit creates unknown outcome
- failover after provider acceptance causes duplicate credit risk

Architecture response:

- separate provider acceptance, bank posting, beneficiary credit, and reconciliation states
- query status before retrying unknown submissions
- block auto-failover after provider finality boundary
- reconcile provider paid status against bank ledger/posting reference

### Flow 2: Quote-lock and confirm provider

Risk points:

- quotation expires before confirm
- quote and transfer use different external references
- provider holds balance after confirmation
- confirm response fails but provider has accepted
- bank tells customer a rate that cannot be honored

Architecture response:

- model quote as its own provider operation with expiry time and rate source
- persist quote id, quote external id, FX rate, fee, and exact amount precision
- confirm with provider idempotency key
- on confirm timeout, move to `unknown_outcome` and status-query before retry

### Flow 3: Provider-instructed payout partner flow

Risk points:

- provider calls bank/switch for account validation before transfer exists locally
- validation and fund transfer calls arrive separately
- status webhook must report back to provider
- provider may retry instructions if acknowledgement is slow
- PII may be absent from later status webhooks

Architecture response:

- support inbound provider instructions as transfer creation source
- map provider instruction id to switch transfer id immediately
- ack provider within strict SLA after durable inbox write
- process crediting asynchronously and report final status through provider adapter

### Flow 4: Cash pickup

Risk points:

- "available for pickup" is not the same as "paid"
- branch/agent liquidity and operating hours affect success
- customer pickup can happen long after send
- cancellation/reversal rules vary by provider and pickup status

Architecture response:

- use separate canonical states: `available_for_pickup`, `paid_cash`, `expired`, `cancel_requested`, `cancelled`
- route by agent coverage and cash availability where data exists
- do not apply direct-to-account SLA assumptions to cash pickup

### Flow 5: PAPSS or clearing-style rail

Risk points:

- payment and settlement are separate but related lifecycles
- direct vs indirect participant changes liquidity and responsibility
- RTGS, central bank, and settlement timing affect resolution
- ISO 20022 status messages need schema validation and message correlation

Architecture response:

- model payment status and settlement status separately
- keep pre-funding and net settlement timelines visible
- parse ISO messages into normalized events while storing raw XML/JSON
- route decisions must account for prefund balance and settlement window

## Safe Failover Model

The default current core rule is right: failover is only safe before the transaction may have been accepted by a provider or downstream bank.

Recommended attempt safety states:

| Safety state | Meaning | Auto-failover |
| --- | --- | --- |
| `not_submitted` | No external provider call was made | Safe |
| `submitted_no_ack` | Request sent, no conclusive response | Unsafe until status query proves not accepted |
| `rejected_pre_finality` | Provider rejected before value movement | Safe if idempotency and amount/rate still valid |
| `accepted_value_pending` | Provider accepted or balance held | Unsafe |
| `value_delivered` | Beneficiary credited or cash paid | Never |
| `reversal_or_refund_pending` | Compensation needed | Never |

Auto-switching should move new traffic away from degraded routes. It should not blindly move in-flight traffic unless the adapter says the exact operation is pre-finality and idempotently safe.

## Data Model Additions

Add persistent tables around the existing domain:

```text
transfers
transfer_attempts
provider_operations
provider_references
route_decisions
transfer_events
timeline_spans
webhook_inbox
status_poll_jobs
exceptions
provider_capability_snapshots
provider_balance_snapshots
fx_quotes
reconciliation_files
reconciliation_items
report_definitions
report_runs
report_exports
audit_events
outbox_events
```

Key uniqueness constraints:

- `(bank_id, idempotency_key)` on transfer intake
- `(provider_id, provider_external_id)` on provider operations
- `(provider_id, provider_reference)` on provider references
- `(provider_id, callback_event_id)` when the provider sends event ids
- payload hash fallback for providers without event ids
- one active unresolved `unknown_outcome` exception per transfer attempt

## What To Build First

1. Persistent transfer/orchestration schema with attempts and provider operations.
2. Outbox and webhook inbox, even before the first real provider.
3. Provider adapter interface with a richer sandbox adapter that simulates:
   - quote expiry
   - timeout after acceptance
   - delayed callback
   - duplicate callback
   - out-of-order callback
   - settlement mismatch
4. Thunes-style adapter first because it exercises discovery, quote, create, confirm, callback, polling, and balance hold behavior.
5. MoneyGram payout-partner simulation next because it exercises provider-instructed payout and provider status reporting.
6. PAPSS-style message adapter simulation because it forces separate payment, prefund, and settlement timelines.
7. Reconciliation state machine that keeps pending auto-matches out of the manual queue until evidence conflicts or SLA is breached.
8. Exception engine and timeline UI before automatic in-flight failover.
9. Report generation for reconciliation exceptions, incidents, provider SLA, settlement, and route-decision audit.

## Architecture Decisions

1. The public API should expose canonical transfer semantics, not provider-shaped endpoints.
2. Provider adapters emit immutable facts; the orchestrator owns state transitions.
3. Webhooks are written to an inbox before processing and acknowledged quickly.
4. Every provider operation gets a span, a timeout policy, a retry policy, and an idempotency policy.
5. Unknown outcome is a first-class exception state.
6. Auto-switching applies primarily to new transfers. In-flight failover is opt-in per provider operation and safety boundary.
7. Provider status polling is a fallback, not a replacement for webhooks, and must obey provider-specific polling limits.
8. Reconciliation is part of reliability, not a finance afterthought.
9. Routine reconciliation should be automatic; manual work should be exception review with evidence, reason codes, and approvals.
10. Reports must be generated from immutable events, reconciliation state, incident timelines, and audit logs.
11. Raw provider payloads should be stored with redaction and retention controls because they are essential for dispute evidence.
12. Analytics and dashboard rollups must consume events outside the routing hot path.

## Open Provider Diligence Checklist

For every real provider, collect the following before implementation:

- full API specification, including sandbox and production base URLs
- authentication, signing, encryption, mTLS, VPN, and IP allowlist requirements
- idempotency support and recommended retry behavior
- transaction states, sub-states, and finality boundaries
- quote expiry, rate lock, fee, amount precision, and rounding rules
- callback schema, retry rules, delivery ordering guarantees, and event ids
- status polling limits and expected polling backoff
- cancellation, refund, reversal, and recall process
- account/wallet validation behavior and name match rules
- compliance/RFI/manual review behavior
- liquidity, prefunding, balance, and settlement reports
- reconciliation files and field-level matching rules
- incident, reconciliation, settlement, provider SLA, and audit reporting requirements
- provider support escalation path and SLA

## Source Notes

- Thunes Money Transfer API v2: https://docs.thunes.com/money-transfer/v2/
- MoneyGram Transfer API: https://developer.moneygram.com/moneygram-developer/docs/transfer-api
- MoneyGram Payout Partners: https://developer.moneygram.com/moneygram-developer/docs/payout-partners
- MoneyGram Status API: https://developer.moneygram.com/moneygram-developer/docs/status-api-overview
- MoneyGram Webhooks: https://developer.moneygram.com/moneygram-developer/docs/webhooks-introduction
- TerraPay API Suite: https://developers.terrapay.com/docs/
- PAPSS How It Works: https://papss.com/how-it-works/
- PAPSS Get Connected: https://papss.com/get-connected/
- Mastercard Cross-Border Services product guide for banks: https://static.developer.mastercard.com/content/cross-border-services/uploads/Cross-Border%20Services%20Product%20Guide_Bank.pdf
- Mastercard Retrieve Payment specifications: https://static.developer.mastercard.com/content/cross-border-services/uploads/Retrieve_Payment_Specifications.pdf
- Onafriq developer portal: https://developers.onafriq.com/
- Onafriq disbursements: https://onafriq.com/services/disbursements
- Ria digital partnerships: https://www.riamoneytransfer.com/en-us/become-a-digital-partner-how/
