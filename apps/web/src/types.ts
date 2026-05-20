export type HealthState = 'healthy' | 'watch' | 'degraded' | 'blocked' | 'recovery' | 'unknown' | 'stale'

export type ScreenId =
  | 'control'
  | 'corridors'
  | 'transactions'
  | 'incidents'
  | 'policy'
  | 'fx'
  | 'reconciliation'
  | 'providers'
  | 'audit'

export type UiScenario = 'healthy' | 'degraded' | 'blocked' | 'stale-fx' | 'empty' | 'permission-denied' | 'loading' | 'api-failure'

export type DataViewState = 'ready' | 'loading' | 'empty' | 'stale' | 'error' | 'permission-denied'

export type ProviderIdentity = {
  id: string
  name: string
  shortName: string
  mark: string
  category: 'Legacy IMTO' | 'Digital IMTO' | 'B2B payout network' | 'Pan-African rail' | 'Local payout rail' | 'Operational queue' | 'Provider'
  tone: HealthState
  color: string
}

export type CountryIdentity = {
  code: string
  name: string
  shortName: string
  flag: string
}

export type DataConnection = {
  mode: 'polling' | 'static'
  freshness: 'fresh' | 'stale' | 'estimated' | 'unavailable'
  updatedAt: string
  nextPollIn: string
}

export type SummaryMetric = {
  label: string
  value: string
  detail: string
  trend: string
  state: HealthState
}

export type DashboardStat = {
  label: string
  value: string
  detail: string
  weekComparison: string
  monthComparison: string
  state: HealthState
}

export type ChartPoint = {
  label: string
  value: number
}

export type LatencyBand = {
  label: string
  valueSeconds: number
  targetSeconds: number
  state: HealthState
}

export type DashboardBreakdown = {
  label: string
  value: number
  state: HealthState
}

export type DashboardVisuals = {
  completionTrend: ChartPoint[]
  volumeTrend: ChartPoint[]
  latencyBands: LatencyBand[]
  exceptionBreakdown: DashboardBreakdown[]
  hourHealth: Array<ChartPoint & { state: HealthState }>
}

export type QaPolicy = {
  name: string
  version: string
  thresholdSeconds: number
  warningSeconds: number
  scope: string
  completedWithinPolicy: string
  breachRate: string
  weekComparison: string
  monthComparison: string
  updatedAt: string
}

export type ControlSummary = {
  globalHealth: string
  valueToday: string
  transactionsToday: string
  p95CreditTime: string
  stuckTransactions: number
  activeIncidents: number
  lastUpdated: string
  atRiskValue: string
  topRisk: string
  safeAction: string
  metrics: SummaryMetric[]
  connection: DataConnection
}

export type Recommendation = {
  title: string
  trigger: string
  affectedTraffic: string
  affectedValue: string
  currentRoute: string
  suggestedRoute: string
  nextAction: string
  evidence: string
  state: HealthState
}

export type CorridorRow = {
  corridor: string
  payout: string
  state: HealthState
  selectedRoute: string
  score: number
  p95: string
  cost: string
  split: string
  recommendation: string
  risk: string
  atRiskValue: string
  owner: string
  status: string
}

export type ProviderScore = {
  provider: string
  corridor: string
  successRate: string
  p50: string
  p95: string
  p99: string
  stuckRate: string
  settlementExceptions: number
  state: HealthState
  rank: number
  supportSla: string
  trafficShare: string
}

export type TimelineStep = {
  label: string
  owner: string
  status: 'done' | 'current' | 'pending'
  time: string
  duration?: string
  source?: string
  reference?: string
  note?: string
}

export type TransactionRecord = {
  reference: string
  providerReference: string
  bankReference: string
  senderCountry: string
  destinationCountry: string
  senderCurrency: string
  destinationCurrency: string
  destinationType: 'Local bank' | 'International bank' | 'Wallet' | 'Cash pickup'
  provider: string
  route: string
  amount: string
  senderStartedAt: string
  destinationCreditedAt: string
  totalTime: string
  totalTimeSeconds: number
  qaLimitSeconds: number
  qaStatus: 'on_time' | 'delayed' | 'stalled'
  state: HealthState
  beneficiary: string
  currentOwner: string
  blocker: string
}

export type TransactionTrace = {
  reference: string
  providerReference: string
  bankReference: string
  beneficiary: string
  corridor: string
  amount: string
  senderCountry: string
  destinationCountry: string
  senderCurrency: string
  destinationCurrency: string
  destinationType: TransactionRecord['destinationType']
  senderStartedAt: string
  destinationCreditedAt: string
  totalTime: string
  totalTimeSeconds: number
  qaLimitSeconds: number
  qaStatus: TransactionRecord['qaStatus']
  currentState: HealthState
  currentOwner: string
  blocker: string
  safeAction: string
  selectedRoute: PolicyRouteDecision
  rejectedRoutes: PolicyRejectedRoute[]
  policyVersion: string
  fallbackRoutes: string[]
  scoreInputs: Array<{ label: string; value: string; state: HealthState }>
  timeline: TimelineStep[]
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

export type Incident = {
  id: string
  title: string
  severity: HealthState
  corridor: string
  owner: string
  startedAt: string
  affectedTransactions: number
  affectedValue: string
  rootCause: string
  nextAction: string
  status: string
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
  reason?: string
  confidence?: string
  policyVersion?: string
}

export type PolicyRejectedRoute = {
  provider: string
  route: string
  reason: string
}

export type PolicyWorkflow = {
  scope: Array<{ label: string; value: string }>
  currentPolicy: Array<{ label: string; value: string }>
  proposedPolicy: Array<{ label: string; value: string; changed?: boolean }>
  validation: Array<{ label: string; value: string; state: HealthState }>
  change: {
    reason: string
  }
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

export type FxCostRoute = {
  provider: string
  route: string
  pair: string
  rate: string
  updatedAt: string
  state: HealthState
  fee: string
  spread: string
  effectiveCost: string
  payoutTime: string
  cheapest: boolean
  recommended: boolean
  note: string
}

export type FxCostBoard = {
  corridor: string
  pair: string
  window: string
  refreshedAt: string
  cheapestProvider: string
  recommendedProvider: string
  rateAlert: string
  decision: string
  routes: FxCostRoute[]
}

export type ReconciliationItem = {
  reference: string
  provider: string
  amount: string
  age: string
  reason: string
  owner: string
  state: HealthState
}

export type AuditEvent = {
  time: string
  actor: string
  action: string
  object: string
  reason: string
  state: HealthState
}

export type DashboardMock = {
  scenario: UiScenario
  viewState: DataViewState
  providerIdentities: ProviderIdentity[]
  countryIdentities: CountryIdentity[]
  summary: ControlSummary
  dateRange: {
    label: string
    start: string
    end: string
    timezone: string
  }
  qaPolicy: QaPolicy
  operationalStats: DashboardStat[]
  visuals: DashboardVisuals
  recommendation: Recommendation
  corridors: CorridorRow[]
  providerScores: ProviderScore[]
  transactions: TransactionRecord[]
  trace: TransactionTrace
  latency: {
    filters: {
      provider: string
      corridor: string
      destinationBank: string
      window: string
    }
    summary: {
      endToEnd: string
      target: string
      slowestStep: string
      affectedTransactions: number
    }
    steps: LatencyStep[]
  }
  downtimeEvents: DowntimeEvent[]
  incidents: Incident[]
  routeConfig: {
    providers: ProviderToggle[]
    fallbackRoutes: FallbackRoute[]
    presets: TrafficSplitPreset[]
    weights: ScoringWeight[]
    impact: {
      successRate: string
      p95: string
      cost: string
    }
    history: ChangeHistoryItem[]
    workflow: PolicyWorkflow
  }
  policySimulationSamples: PolicySimulationSample[]
  fxCostBoard: FxCostBoard
  reconciliation: ReconciliationItem[]
  auditEvents: AuditEvent[]
}
