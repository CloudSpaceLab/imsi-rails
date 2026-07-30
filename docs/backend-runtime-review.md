# Backend Runtime Review

Date: 2026-05-19

## Decision

Use Go for the backend core services.

Use Vue 3, Vite, and TypeScript for the UI and UI-adjacent work.

Do not use Rust as the default backend for the first implementation.

## Why This Changed

The original Rust recommendation optimized for raw resource efficiency. That was directionally reasonable, but too narrow.

The actual project realities are:

- provider and bank latency will dominate route scoring latency
- adapter work will involve messy APIs, SFTP, CSV, XML/SOAP, callback retries, and reconciliation files
- bank pilots need fast iteration and easy maintainability
- hiring and handoff risk matter as much as runtime performance
- the core scoring path is small enough that Go should easily meet the performance budget

## Practical Numbers

| Factor | Go | Rust | Node.js/TypeScript |
| --- | --- | --- | --- |
| Professional developer usage in Stack Overflow 2025 | 17.4% | 14.5% | JavaScript 68.8%, TypeScript 48.8% |
| Concurrency model | goroutines; a few KB initial stack per Go FAQ | async runtime choices and ownership discipline | event loop plus worker pool |
| Operational packaging | compiled binary | compiled binary | runtime + package dependency tree |
| Route scoring suitability | excellent | excellent | good if callbacks stay small |
| Adapter/integration delivery | strong | slower, more specialist | fastest for many teams |
| Core switching risk | low | low technically, higher delivery risk | event-loop blocking and dependency discipline |
| Recommended role | backend core | future specialty components only | UI/BFF/internal tooling |

## Key Insight

If a route decision takes:

- Rust: sub-millisecond to low milliseconds
- Go: sub-millisecond to low milliseconds
- Node.js: low milliseconds when implemented carefully

but provider/bank work takes:

- provider API call: hundreds of milliseconds to seconds
- webhooks/status callbacks: seconds to minutes
- SFTP/settlement/reconciliation: minutes to hours

then Rust's theoretical performance edge is not where the bank value is.

The bank value is in:

- correct routing policy
- safe failover boundaries
- provider normalization
- clear audit trail
- latency/downtime analysis
- easy operations UX
- fast integration delivery

Go is the better default for that combination.

## Node.js Position

Node.js/TypeScript remains very useful for:

- Vue/Vite frontend tooling
- admin/BFF services
- demo tooling
- report generation
- internal configuration utilities

It is not the preferred route-decision core because Node's own guidance stresses that servers stay fast when each client's work remains small, and that blocking the event loop can delay other clients.

## Rust Position

Rust remains valid for:

- future ultra-low-level adapters
- cryptographic or parsing-heavy components
- specialized high-throughput sidecars
- components where memory safety without GC is worth the team cost

Rust should not be the default implementation language until a measured bottleneck justifies it.

## Sources

- Go FAQ: https://go.dev/doc/faq
- Node.js event loop guidance: https://nodejs.org/learn/asynchronous-work/dont-block-the-event-loop
- Stack Overflow Developer Survey 2025: https://survey.stackoverflow.co/2025/technology

