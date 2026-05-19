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

Use Vue 3 with Vite and TypeScript for the product application, uPlot for dense time-series charts, and TanStack Table for large operational tables.

The landing page remains dependency-free static HTML/CSS/JS.

## Rationale

Vue 3 is the better practical choice for the first bank-facing product:

- larger hiring and maintenance pool than Svelte in most markets
- mature enterprise UI ecosystem without forcing a heavy component library
- strong TypeScript support through the official Vue tooling
- Vite-native development speed and build tooling
- official Vue Router and Pinia options for app routing and state
- easier stakeholder acceptance than a smaller framework

Svelte remains elegant and lightweight, but the gain is not large enough to justify the ecosystem and hiring tradeoff for this product. The product's UI performance should come from disciplined data loading, rollups, virtualization, and charting choices, not from framework novelty.

uPlot is small and fast for time-series visualization. TanStack Table is headless, which keeps the design system in our control while supporting virtualization and complex table behavior.

## Consequences

- The product app must keep initial JavaScript small and avoid unnecessary component libraries.
- Large chart views must use rollups/downsampling before rendering.
- Large tables must use pagination or virtualization.
- The landing page must stay static unless a clear publishing need justifies a build pipeline.
- Vue component primitives should be built from the brand/design-language guide before adopting any broad component suite.

## Sources

- Vue TypeScript guide: https://vuejs.org/guide/typescript/overview
- Vue FAQ: https://vuejs.org/about/faq.html
- Vite guide: https://vite.dev/guide/
- Vue Router: https://router.vuejs.org/
- Pinia: https://pinia.vuejs.org/

