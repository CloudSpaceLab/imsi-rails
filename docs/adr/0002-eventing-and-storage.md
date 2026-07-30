# ADR 0002: Eventing and Storage

Status: accepted

Date: 2026-05-19

## Context

The platform needs durable transaction history, live dashboard updates, auditability, and deep reliability analysis. These needs have different performance profiles.

The system must avoid putting analytics or dashboard queries in the route-decision hot path.

## Decision

Use:

- PostgreSQL as the system of record
- NATS Core for low-latency live event fanout
- NATS JetStream for durable lifecycle event replay
- PostgreSQL rollups for pilot analytics
- ClickHouse only when the bank pilot volume justifies high-cardinality analytical drilldowns

## Rationale

PostgreSQL provides bank-friendly relational durability, constraints, indexes, and auditability. NATS provides a lightweight messaging layer for live state changes and durable event replay through JetStream. ClickHouse is excellent for high-volume analytics, but adding it too early increases operational burden.

## Consequences

- The MVP can run with PostgreSQL and NATS only.
- Analytics pipelines must be consumers of durable events, not dependencies of route decisions.
- Route decisions must persist their own inputs, scores, rejection reasons, and policy version.
- ClickHouse adoption requires a measured need: event volume, query latency, or retention pressure that PostgreSQL rollups cannot serve efficiently.

