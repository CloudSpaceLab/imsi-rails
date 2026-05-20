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
  previewPolicy: 'Test change',
  rollbackPolicy: 'Roll back',
  exportEvidence: 'Export',
  openIncident: 'Open incident',
} as const

export const rejectionCopy = {
  stale_fx_rate: 'FX rate is stale.',
  circuit_breaker_open: 'Route is blocked.',
  unsupported_destination_bank: 'Destination bank is unsupported.',
  amount_outside_limits: 'Amount is outside route limits.',
  compliance_manual_review: 'Manual approval is required.',
} as const

