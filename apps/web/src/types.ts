export type HealthState = 'healthy' | 'watch' | 'degraded' | 'blocked' | 'recovery' | 'unknown' | 'stale'

export type TimelineStep = {
  label: string
  owner: string
  status: 'done' | 'current' | 'pending'
  time: string
}

export type LatencyStep = {
  label: string
  owner: string
  durationMs: number
  targetMs: number
  state: HealthState
}

export type DowntimeEvent = {
  time: string
  title: string
  actor: string
  state: HealthState
  detail: string
}

export type ProviderToggle = {
  provider: string
  route: string
  enabled: boolean
  state: HealthState
}

export type FallbackRoute = {
  rank: number
  provider: string
  route: string
  state: HealthState
}

export type TrafficSplitPreset = {
  label: string
  active: boolean
  split: string
}

export type ScoringWeight = {
  label: string
  value: number
}

export type ChangeHistoryItem = {
  time: string
  actor: string
  summary: string
}

export type PolicyRouteDecision = {
  provider: string
  route: string
  score: number
  p95: string
  cost: string
  state: HealthState
}

export type PolicyRejectedRoute = {
  provider: string
  route: string
  reason: string
}

export type ShadowReportMetric = {
  label: string
  value: string
  detail: string
}

export type ShadowReportRow = {
  bucket: string
  currentRoute: string
  proposedRoute: string
  result: string
}

export type PolicySimulationSample = {
  reference: string
  corridor: string
  origin: string
  destination: string
  amount: string
  payout: string
  current: PolicyRouteDecision
  proposed: PolicyRouteDecision
  rejectedRoutes: PolicyRejectedRoute[]
  reportMetrics: ShadowReportMetric[]
  reportRows: ShadowReportRow[]
}

