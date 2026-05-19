# ADR 0001: Runtime and Hot Path

Status: accepted

Date: 2026-05-19

## Context

`imsi-rails` must make route decisions for bank international money transfer traffic with low CPU, low memory, predictable latency, and clear auditability.

The routing path should stay small:

- validate transaction intake
- read cached provider/corridor capability
- read cached bank policy
- read cached health/circuit-breaker state
- read fresh-enough FX/cost snapshot
- calculate eligible routes
- score routes
- persist transaction and route decision
- emit lifecycle events

## Decision

Use Rust with Tokio and Axum for the backend routing/runtime services.

The routing hot path must not depend on:

- analytical queries
- reconciliation matching
- provider health probes
- provider callbacks
- dashboard rollup generation
- external provider status checks that can be served by cached state

## Rationale

Rust gives predictable resource usage without a garbage collector, strong type safety, and a good fit for small long-running services. Tokio provides async concurrency for network-heavy adapter and API workloads. Axum sits on Tokio, Hyper, and Tower, which gives composable middleware for timeouts, tracing, auth, and rate limits.

Go remains an acceptable fallback for future peripheral services where team hiring speed is more important than memory predictability. The routing core should remain Rust unless a later benchmark or hiring reality forces a change.

## Consequences

- The first implementation must include benchmarks for route eligibility/scoring.
- Services should expose OpenTelemetry traces and metrics from the start.
- Adapters must be isolated so slow providers cannot stall the routing core.
- All hot-path decisions must be explainable from stored route-decision inputs.

