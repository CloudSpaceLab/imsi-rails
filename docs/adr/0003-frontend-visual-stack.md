# ADR 0003: Frontend Visual Stack

Status: accepted

Date: 2026-05-19

## Context

`imsi-rails` needs a premium control-room UI that remains fast on conservative bank machines and corporate networks.

The interface must support:

- live provider/corridor health
- dense operational tables
- latency and downtime drilldowns
- FX and cost comparison
- safe route configuration
- accessible, keyboard-friendly workflows

## Decision

Use SvelteKit for the product application, uPlot for dense time-series charts, and TanStack Table for large operational tables.

The landing page remains dependency-free static HTML/CSS/JS.

## Rationale

Svelte compiles components into lean JavaScript and is a good fit for performance-sensitive operations software. uPlot is small and fast for time-series visualization. TanStack Table is headless, which keeps the design system in our control while supporting virtualization and complex table behavior.

## Consequences

- The product app must keep initial JavaScript small and avoid unnecessary component libraries.
- Large chart views must use rollups/downsampling before rendering.
- Large tables must use pagination or virtualization.
- The landing page must stay static unless a clear publishing need justifies a build pipeline.

