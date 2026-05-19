export const healthStates = {
  healthy: 'Healthy',
  watch: 'Watch',
  degraded: 'Degraded',
  blocked: 'Blocked',
  recovery: 'Recovery',
  unknown: 'Unknown',
  stale: 'Stale data',
} as const

export const operationalActions = {
  shiftTraffic: 'Shift traffic',
  pauseNewTraffic: 'Pause new traffic',
  previewPolicy: 'Preview policy',
  rollbackPolicy: 'Roll back policy',
  exportEvidence: 'Export evidence',
  openIncident: 'Open incident',
} as const

export const rejectionCopy = {
  stale_fx_rate: 'Rejected because the FX rate is stale.',
  circuit_breaker_open: 'Rejected because the route is blocked by circuit breaker.',
  unsupported_destination_bank: 'Rejected because the destination bank is unsupported.',
  amount_outside_limits: 'Rejected because the amount is outside route limits.',
  compliance_manual_review: 'Manual approval required by compliance policy.',
} as const

