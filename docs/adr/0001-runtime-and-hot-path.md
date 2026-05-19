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

Use Go for the backend routing/runtime services.

The routing hot path must not depend on:

- analytical queries
- reconciliation matching
- provider health probes
- provider callbacks
- dashboard rollup generation
- external provider status checks that can be served by cached state

## Rationale

Go is the best delivery/runtime tradeoff for `imsi-rails`.

The deciding realities:

- The route decision target is under 20 ms p95 excluding external provider/bank calls. A simple eligibility/scoring path should run well under that target in Go.
- The real latency budget will be dominated by provider APIs, bank posting paths, webhooks, SFTP, settlement files, and reconciliation delays, often measured in hundreds of milliseconds to minutes.
- Go gives low operational overhead: single compiled binaries, straightforward deployment, built-in concurrency, strong standard library, mature HTTP tooling, and good support for Postgres, NATS, SFTP, XML/SOAP, and observability.
- Go's goroutines are cheap enough for adapter-heavy workloads. The official Go FAQ describes goroutines as having little overhead beyond a stack of a few kilobytes and says hundreds of thousands can be practical in one address space.
- Hiring and maintenance risk are lower than Rust for bank integration work.

Rust remains excellent for performance-sensitive systems, but here the likely bottleneck is integration delivery and operational correctness rather than raw compute. Node.js/TypeScript remains attractive for UI/BFF and internal tools, but the core switching service should avoid event-loop blocking risks and runtime dependency sprawl.

## Consequences

- The first implementation should use Go modules and a small core domain package.
- The API should be built with Go's standard `net/http` stack or a lightweight router only if needed.
- The first implementation must include benchmarks for route eligibility/scoring.
- Services should expose OpenTelemetry traces and metrics from the start.
- Adapters must be isolated so slow providers cannot stall the routing core.
- All hot-path decisions must be explainable from stored route-decision inputs.

## Sources

- Go FAQ on goroutines: https://go.dev/doc/faq
- Node.js guidance on not blocking the event loop: https://nodejs.org/learn/asynchronous-work/dont-block-the-event-loop
- Stack Overflow Developer Survey 2025 language usage: https://survey.stackoverflow.co/2025/technology

