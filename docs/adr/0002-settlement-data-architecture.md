# ADR 0002: Settlement Data Architecture

Date: 2026-06-03

Status: Proposed

## Context

The primary dashboard must not be a demo layer with hand-picked metrics. Bank operators need charts and queues that map directly to the working model of inbound settlement: partner instruction, prefunded account, local final-leg attempt, provider/API call, evidence capture, SLA outcome, remediation, reconciliation, and audit.

## Decision

Use MariaDB as the operational system of record, an outbox/event stream for live processing, and dashboard rollups built from immutable operational events.

The dashboard must be fed by rollups, not by UI-only fixture math.

The initial MariaDB DDL draft is in `db/settlement_core.sql`.

## Core Tables

### Contracts and Liquidity

`partner_contracts`

- contract id, partner id, corridor, settlement model, active SLA policy, permitted rails, destination bank scope, owner, status

`standing_accounts`

- partner contract id, currency, prefund balance, available limit, credit facility, projected exhaustion, last ledger refresh

`sla_policies`

- duration, cooldown window, late-success observation window, required evidence, requery schedule, escalation rules

### Transfer Workflow

`incoming_instructions`

- switch reference, partner reference, contract id, beneficiary, amount, destination bank, received timestamp, SLA deadline, current state, outcome confidence, value at risk

`route_decisions`

- instruction reference, eligible routes, rejected routes, selected route, score inputs, policy version, decision reason, actor/system origin

`final_leg_attempts`

- instruction reference, route, rail/provider, destination bank, submitted/accepted/credited timestamps, provider status, rail status, final-leg state

`provider_calls`

- attempt id, endpoint, request id, idempotency key, response code, timeout flag, latency ms, raw status, normalized status

### Evidence and Repair

`outcome_evidence`

- instruction reference, evidence type, reference, status, owner, captured timestamp, source hash/object reference

`requery_attempts`

- instruction reference, attempt number, trigger, method, due time, completed time, result, next action

`remediation_cases`

- instruction reference, queue, theory of outcome, duplicate-risk level, evidence gap, owner, maker-checker state, safe closure path

`reconciliation_matches`

- instruction reference, settlement batch, provider file row, bank ledger row, match state, mismatch reason, resolved timestamp

`audit_events`

- actor, action, object type, object id, reason, previous state, new state, policy version, timestamp

## Event Model

Every state-changing write records a `domain_events` row inside the same database transaction and publishes through `outbox_events`.

Important event names:

- `instruction.received`
- `route.decision_recorded`
- `final_leg.submitted`
- `final_leg.accepted`
- `final_leg.credited`
- `sla.breached`
- `requery.scheduled`
- `requery.completed`
- `outcome.evidence_attached`
- `remediation.case_opened`
- `remediation.case_closed`
- `reconciliation.break_detected`
- `reconciliation.matched`
- `route.health_window_closed`
- `circuit_breaker.changed`

## Dashboard Rollups

Rollups are computed outside the routing hot path.

`route_health_windows`

- route, rail, destination bank, window start/end, submitted count, accepted count, credited count, failed count, timeout count, open cases, P50/P95/P99 credit time, late-success count, state

`partner_sla_windows`

- partner contract, corridor, SLA target, total instructions, credited in SLA, breached count, oldest open breach, value at risk, state

`exception_windows`

- queue, route, provider, count, value at risk, oldest age, duplicate-risk count, maker-checker pending count

`api_health_windows`

- service name, endpoint, success rate, timeout rate, P95 latency, stale telemetry flag, last successful poll

The primary dashboard should read from these four rollups:

- SLA completion trend: `partner_sla_windows`
- Local provider performance chart: `route_health_windows`
- Exception mix chart: `exception_windows`
- API telemetry status: `api_health_windows`

## Query and Storage Notes

- Partition high-volume tables by day or month: `incoming_instructions`, `final_leg_attempts`, `provider_calls`, `domain_events`, and `audit_events`.
- Keep raw provider payloads outside the hot tables; store object references and hashes in MariaDB.
- Use MariaDB composite indexes for open work: open remediation cases, unresolved reconciliation breaks, uncredited accepted attempts, and overdue SLA deadlines.
- Use immutable audit rows; never update an audit event in place.
- Use idempotency keys for intake, provider submission, callback ingestion, and manual repair actions.

## Consequences

- The UI can show premium charts without inventing numbers in the browser.
- Operators see the same truth across dashboard, transaction trace, exceptions, and audit.
- Requery, late-success, and duplicate-credit risk become first-class database concepts.
- Future AI/routing optimization can learn from structured route decisions and operational outcomes instead of screenshots or loosely named metrics.
