# imsi-rails

International money transfer reliability infrastructure for banks.

`imsi-rails` is a bank-facing switching and observability platform that lets banks onboard their IMTO providers into one control layer, monitor provider health, compare cost/speed/reliability, and route each transaction to the best eligible rail.

## Repo Contents

- [Market and platform brief](plan.md)
- [Research and architecture](docs/research-and-architecture.md)
- [Delivery masterplan](docs/masterplan.md)
- [Product requirements document](docs/prd.md)
- [Users, brand, and design language](docs/users-brand-design-language.md)
- [GitHub issue plan](docs/github-issues.md)
- [Static landing page](landing/index.html)

## Product Positioning

The platform is not a consumer remittance app or money exchange marketplace. It is the bank's IMTO control tower:

- one integration surface across IMTO providers and payout rails
- one live reliability dashboard
- one route decision engine
- one FX/cost comparison view
- one transaction trace and audit trail
- one safe configuration layer for switching traffic

## Landing Page

Open [landing/index.html](landing/index.html) directly in a browser.

The landing page is intentionally static and dependency-free so it remains fast, portable, and easy to publish.

## Web Control Room

The operational UI prototype lives in `apps/web`.

```bash
npm install
npm run web:dev
npm run web:build
```

Implementation notes:

- [Vue control room shell](docs/implementation-notes/vue-control-room-shell.md)
- [Latency and downtime drilldowns](docs/implementation-notes/latency-downtime-drilldowns.md)
- [Provider and corridor scorecards](docs/implementation-notes/provider-corridor-scorecards.md)
- [Route configuration UI](docs/implementation-notes/route-configuration-ui.md)
- [Policy simulator and shadow routing](docs/implementation-notes/policy-simulator-shadow-routing.md)
- [FX and cost board](docs/implementation-notes/fx-cost-board.md)
