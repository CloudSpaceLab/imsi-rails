# imsi-rails Reality Register

This register captures practical constraints that must shape implementation. It exists to keep the masterplan honest.

## 1. Banks Move Slowly

Reality:

- procurement, security review, legal review, risk review, and integration approval can take longer than engineering
- bank teams may be fragmented across remittances, IT, operations, compliance, treasury, and vendor management

Implication:

- first pilot must feel like operational relief, not a transformation program
- support sandbox mode, file samples, and read-only monitoring before full switching
- create clear onboarding artifacts and evidence packs

## 2. Provider Integrations Will Be Uneven

Reality:

- some IMTOs expose modern APIs
- some rely on SOAP, SFTP, CSV, email reports, or portals
- status codes, references, time zones, and reversal semantics differ widely

Implication:

- adapter framework is core product infrastructure
- canonical transaction state is non-negotiable
- every adapter needs idempotency and normalization tests

## 3. Auto-Switching Can Create Real Money Risk

Reality:

- unsafe failover can cause duplicate payouts or unresolved settlement disputes
- providers differ on when a transaction is final, cancellable, reversible, or merely accepted

Implication:

- failover must be conservative by default
- each provider needs explicit safe/unsafe states
- auto-switching should start with new-transaction routing before in-flight failover

## 4. Active Health Probes May Not Be Allowed

Reality:

- banks or providers may restrict active probes in production
- some route health must be inferred from passive transaction outcomes and callbacks

Implication:

- health engine must support active and passive signals
- every health score must expose measurement source and freshness
- absence of signal is not the same as healthy

## 5. Data Residency and Tenant Boundaries Matter

Reality:

- banks may require country-specific hosting, data segregation, audit export, or private deployment
- multi-bank reliability intelligence can be commercially powerful but sensitive

Implication:

- design for strict tenant boundaries from the start
- aggregate cross-bank insights only where contracts allow
- never assume SaaS deployment is acceptable to every bank

## 6. Neutrality Must Be Credible

Reality:

- banks will not trust routing if providers can secretly influence route selection
- providers may challenge scorecards that shift volume away from them

Implication:

- bank policy owns routing
- route decisions must be explainable
- provider scorecards must use observable, timestamped evidence
- avoid hidden provider-paid routing incentives

## 7. FX and Cost Data Can Be Stale

Reality:

- provider FX rates may be delayed, batch-updated, corridor-specific, or valid only for certain amounts
- cheapest route may be risky if the rate is stale or settlement quality is weak

Implication:

- FX freshness must affect eligibility/scoring
- UI must explain when the cheapest route is not selected
- cost comparison needs timestamps and validity windows

## 8. Reconciliation Is Part of Reliability

Reality:

- a transfer is not truly done for bank operations until settlement and reconciliation are clean
- provider success does not always mean bank records match

Implication:

- route score should eventually include reconciliation break rate
- basic reconciliation exceptions belong in the pilot
- settlement aging must be visible early

## 9. Bank Users Need Different Views

Reality:

- operators need actions
- executives need proof
- compliance needs evidence
- integrators need contracts and logs

Implication:

- avoid one generic dashboard
- default to operational clarity
- provide rollups, exports, and audit drawers for non-operator users

## 10. Performance Is A Sales Feature

Reality:

- bank environments can involve older machines, constrained networks, locked-down browsers, and strict infrastructure review

Implication:

- landing page stays static
- product app must keep bundle size low
- tables must virtualize
- charts must use rollups
- routing core must stay small and benchmarked

