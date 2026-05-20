# UI/UX Research Notes

Research date: 2026-05-20

## Dashboard Role

The Overview is a monitoring surface, not a transaction workbench. It should show the current state, the goal, what changed, and what action is safest. Detail tools such as row sorting belong inside detail screens like Transactions.

Supporting sources:

- Microsoft Power BI dashboard guidance says dashboards are for monitoring current state at a glance; detail belongs in drill-in reports, and clutter should be removed from the dashboard.
- Fluent 2 layout guidance emphasizes spacing, proximity, hierarchy, and using space to show what matters most.
- Fluent 2 accessibility guidance emphasizes predictable structure, scannable headings, contrast, responsive reflow, and concise meaningful text.
- Apple HIG emphasizes clear hierarchy where controls support content rather than dominate it.
- Bach et al., Dashboard Design Patterns, frames dashboard work as a set of tradeoffs around screen space, interaction, and information shown.

## Product Implications

- Overview controls should be global context controls: date range, corridor, payout type, QA policy scope.
- Overview should not expose transaction table controls like sort order.
- Sort, search, and row-level filtering belong on Transactions.
- Status color should be sparse and semantic: healthy, watch, degraded, blocked, stale, recovery.
- The first viewport should answer three questions:
  - Are transfers completing within policy?
  - Where is the operating risk?
  - What action is safe and audit-ready?

## Applied Corrections

- Removed transaction sorting from Overview.
- Added dashboard-level scope filters for corridor and payout type.
- Kept transaction sorting inside the Transactions screen only.
- Reframed the Overview hierarchy around QA completion health, risk, and next action.

## Sources

- Microsoft Power BI dashboard design tips: https://learn.microsoft.com/en-us/power-bi/create-reports/service-dashboards-design-tips
- Fluent 2 layout: https://fluent2.microsoft.design/layout
- Fluent 2 accessibility: https://fluent2.microsoft.design/accessibility
- Apple Human Interface Guidelines: https://developer.apple.com/design/human-interface-guidelines/
- Dashboard Design Patterns: https://arxiv.org/abs/2205.00757
