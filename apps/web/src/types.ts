export type HealthState = 'healthy' | 'watch' | 'degraded' | 'blocked' | 'recovery' | 'unknown' | 'stale'

export type ScreenId =
  | 'command'
  | 'inflows'
  | 'routes'
  | 'exceptions'
  | 'partners'
  | 'reports'
  | 'settings'

export type UiScenario =
  | 'healthy'
  | 'degraded'
  | 'degraded-ria'
  | 'traffic-shift'
  | 'pilot-report'
  | 'blocked'
  | 'stale-fx'
  | 'empty'
  | 'permission-denied'
  | 'loading'
  | 'api-failure'

export type DataViewState = 'ready' | 'loading' | 'empty' | 'stale' | 'error' | 'permission-denied'

export type Permission =
  | 'dashboard:read'
  | 'transactions:read'
  | 'transactions:trace'
  | 'providers:manage'
  | 'policy:draft'
  | 'policy:approve'
  | 'policy:activate'
  | 'incidents:manage'
  | 'reconciliation:manage'
  | 'fx:read'
  | 'audit:read'
  | 'audit:export'
  | 'users:manage'
  | 'identity:manage'
  | 'contracts:read'
  | 'credits:read'
  | 'compliance:manage'

export type Role =
  | 'platform_admin'
  | 'bank_admin'
  | 'ops_lead'
  | 'ops_analyst'
  | 'compliance_officer'
  | 'treasury_finance'
  | 'auditor'
  | 'viewer'
  | 'demo_viewer'

export type SessionUser = {
  id: string
  bank_id: string
  username: string
  email: string
  display_name: string
  roles: Role[]
  permissions: Permission[]
  auth_provider: string
}

export type DashboardContext = {
  from: string
  to: string
  timezone: string
  provider_id?: string
  corridor?: string
  payout_method?: string
  scenario?: UiScenario
  currency: string
  analysis_lens: string
}

export type DashboardAnalytics = {
  processed_count: number
  processed_volume: number
  completed_count: number
  sla_completed_count: number
  sla_rate: number
  failed_count: number
  stalled_count: number
  pending_count: number
  p50_seconds: number
  p95_seconds: number
  p99_seconds: number
}

export type MetricTile = {
  id: string
  label: string
  value: string
  unit: string
  state: HealthState
  trend: string
  drilldown: string
}

export type DashboardTimeseriesPoint = {
  time: string
  processed_count: number
  volume: number
  sla_rate: number
  p95_seconds: number
  state: HealthState
}

export type ProviderComparison = {
  provider_id: string
  provider_name: string
  corridor: string
  processed_count: number
  processed_volume: number
  sla_completed_count: number
  sla_rate: number
  p95_seconds: number
  state: HealthState
}

export type DashboardSummaryResponse = {
  context: DashboardContext
  analytics: DashboardAnalytics
  tiles: MetricTile[]
  providers: ProviderComparison[]
  generated_at: string
}

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
  mode: 'polling' | 'static' | 'api' | 'sse'
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

export type PartnerSlaHistory = {
  partner: string
  /** SLA compliance target as a percentage, e.g. 97 */
  target: number
  /** Daily SLA compliance percentage, oldest first; label is an ISO date */
  daily: ChartPoint[]
}

export type DashboardVisuals = {
  completionTrend: ChartPoint[]
  partnerSlaHistory: PartnerSlaHistory[]
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

export type InflowDetailTab = 'summary' | 'route' | 'timeline' | 'evidence' | 'requery' | 'reconciliation' | 'audit'

export type ExceptionDetailTab = 'summary' | 'evidence' | 'requery' | 'closure' | 'audit'

export type RouteDetailTab = 'overview' | 'traffic' | 'cases' | 'history' | 'policy'

export type EvidenceRequirement = {
  id: string
  instructionReference: string
  slaPolicyId: string
  sourceTable: 'outcome_evidence'
  label: string
  evidenceType: OutcomeEvidence['type']
  required: boolean
  capturedReference?: string
  status: string
  state: HealthState
}

export type CaseActionStep = {
  action: 'attach_evidence' | 'manual_requery' | 'mark_completed_outside_platform' | 'approve_reversal'
  label: string
  queues: RemediationCase['queue'][]
  duplicateRisk: 'Any' | 'Low' | 'Medium' | 'High'
  requiresEvidence: boolean
  safetyChecklist: string[]
  resultingState: string
}

export type LinkedWorkItem = {
  id: string
  label: string
  reference: string
  path: string
  detail: string
  owner: string
  state: HealthState
}

export type IntegrationHealth = {
  id: string
  sourceTable: 'api_health_windows'
  serviceName: string
  endpoint: string
  provider: string
  successRate: string
  timeoutRate: string
  p95Latency: string
  polling: string
  callbackStatus: string
  lastSuccessAt: string
  nextAction: string
  state: HealthState
}

export type RouteDecisionRecord = {
  instructionReference: string
  sourceTable: 'route_decisions'
  policyVersion: string
  selectedRoute: string
  selectedRail: string
  score: number
  scoreInputs: Array<{ label: string; value: string; state: HealthState }>
  rejectedRoutes: PolicyRejectedRoute[]
  decisionReason: string
  decidedAt: string
}

export type ReconciliationMatch = {
  instructionReference: string
  sourceTable: 'reconciliation_matches'
  settlementBatch: string
  providerFileReference: string
  bankLedgerReference: string
  matchState: string
  mismatchReason: string
  resolvedAt: string
  state: HealthState
}

export type RouteHealthWindow = {
  route: string
  sourceTable: 'route_health_windows'
  rail: string
  window: string
  submittedCount: number
  creditedCount: number
  failedCount: number
  timeoutCount: number
  openCases: number
  p95CreditTime: string
  lateSuccessCount: number
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

export type ContractEndpoint = {
  name: string
  kind: 'MFB' | 'Fintech wallet' | 'Local bank' | 'Agent network' | 'Payment gateway'
  rail: string
  reach: string
  reliability: string
  p95Credit: string
  openCredits: number
  callbackLag: string
  settlementLag: string
  owner: string
  state: HealthState
}

export type RoutingContract = {
  id: string
  partner: string
  partnerType: 'Foreign bank' | 'IMTO' | 'Fintech'
  originCountry: string
  destinationCountry: string
  corridor: string
  payoutMethods: string
  endpointCount: number
  monthlyVolume: string
  deliveryReliability: string
  creditSla: string
  slaBreachRate: string
  lastSettlement: string
  availableLimit: string
  settlementModel: string
  prefundStatus: string
  cutoffWindow: string
  openExceptions: number
  oldestOpenCredit: string
  nextReview: string
  state: HealthState
  status: string
  owner: string
  note: string
  endpoints: ContractEndpoint[]
}

export type CreditLegState =
  | 'received'
  | 'crediting'
  | 'credited'
  | 'cooldown'
  | 'requerying'
  | 'manual_remediation'
  | 'completed_outside_platform'
  | 'reversal_pending'
  | 'reversed'
  | 'recon_break'

export type EvidenceItem = {
  label: string
  reference: string
  status: string
  owner: string
  state: HealthState
}

export type CreditLegRequeryAttempt = {
  id: string
  dueAt: string
  completedAt: string
  method: string
  result: string
  trigger: 'automatic' | 'manual'
  state: HealthState
}

export type CreditLeg = {
  reference: string
  contractId: string
  partner: string
  partnerReference: string
  originCountry: string
  destinationCountry: string
  beneficiary: string
  beneficiaryAccount: string
  endpoint: string
  endpointKind: ContractEndpoint['kind']
  amount: string
  valueAtRisk: string
  receivedAt: string
  creditedAt: string
  elapsed: string
  slaDueAt: string
  slaState: HealthState
  creditState: CreditLegState
  state: HealthState
  reversal: boolean
  reconState: string
  evidenceStatus: string
  railReference: string
  bankPostingReference: string
  settlementBatch: string
  callbackLag: string
  lastCallbackAt: string
  postingAttempts: number
  makerCheckerStatus: string
  nextAction: string
  owner: string
  blocker: string
  riskFlags: string[]
  evidence: EvidenceItem[]
  requeryAttempts?: CreditLegRequeryAttempt[]
  timeline: TimelineStep[]
}

export type InboundSlaRow = {
  contractId: string
  partner: string
  corridor: string
  creditSla: string
  p95Credit: string
  breachRate: string
  agingBreaches: number
  valueAtRisk: string
  oldestBreach: string
  breachReason: string
  owner: string
  recommendedAction: string
  state: HealthState
  trend: string
}

export type ImtoPartnerStatus = 'active' | 'onboarding' | 'suspended' | 'draft'
export type ImtoRiskRating = 'low' | 'medium' | 'high'
export type ImtoIntegrationMode = 'REST' | 'SOAP' | 'SFTP'
export type ImtoApprovalState = 'approved' | 'checker_pending' | 'maker_draft' | 'rejected'

export type ImtoDueDiligenceItem = {
  item: string
  status: 'complete' | 'pending' | 'overdue'
}

export type ImtoPartner = {
  id: string
  name: string
  country: string
  corridor: string
  status: ImtoPartnerStatus
  riskRating: ImtoRiskRating
  integrationMode: ImtoIntegrationMode
  iso20022Ready: boolean
  creditSla: string
  settlementCurrency: string
  prefundingModel: 'prefunded' | 'credit-line' | 'hybrid'
  contractEnd: string
  onboardingStage: string
  approvalState: ImtoApprovalState
  dueDiligence: ImtoDueDiligenceItem[]
  state: HealthState
}

export type ComplianceHold = {
  reference: string
  partner: string
  type: 'Sanctions' | 'KYC' | 'RFI' | 'Purpose'
  beneficiary: string
  amount: string
  age: string
  dueAt: string
  decisionSla: string
  requiredEvidence: string
  partnerRfiReference: string
  owner: string
  valueAtRisk: string
  state: HealthState
  note: string
}

export type FeatureSwitchKey =
  | 'externalDebitRailOps'
  | 'walletAndCashPickup'
  | 'advancedFxCosts'
  | 'providerCommercialScorecards'
  | 'policySimulator'
  | 'multiBankPortfolio'
  | 'advancedComplianceCases'
  | 'settlementFileImports'

export type FeatureSwitch = {
  key: FeatureSwitchKey
  label: string
  enabled: boolean
  defaultEnabled: boolean
  reason: string
}

export type CreditFacility = {
  limit: string
  utilized: string
  available: string
  rules: string
  state: HealthState
}

export type StandingAccount = {
  id: string
  partner: string
  currency: string
  prefundBalance: string
  availableLimit: string
  projectedExhaustion: string
  creditFacility?: CreditFacility
  state: HealthState
}

export type InboundContract = {
  id: string
  partner: string
  corridor: string
  standingAccountId: string
  slaPolicyId: string
  permittedRails: string[]
  permittedDestinationBanks: string[]
  settlementModel: string
  owner: string
  state: HealthState
}

export type FinalLegState =
  | 'received'
  | 'validated'
  | 'routing'
  | 'submitted'
  | 'accepted'
  | 'credit_pending'
  | 'credited'
  | 'sla_breached'
  | 'cooldown'
  | 'requerying'
  | 'outcome_uncertain'
  | 'manual_remediation'
  | 'completed_outside_platform'
  | 'reversal_pending'
  | 'failed_safe'
  | 'failed_unsafe'
  | 'reconciled'

export type OutcomeConfidence =
  | 'proven_credited'
  | 'probably_pending'
  | 'uncertain'
  | 'safe_failed'
  | 'unsafe_failed'

export type StaffAssignment = {
  id: string
  sourceTable: 'staff_case_assignments'
  instructionReference: string
  staffName: string
  staffRole: string
  team: string
  assignedAt: string
  assignmentRule: string
  responsibility: string
  state: HealthState
}

export type IncomingInstruction = {
  reference: string
  partnerReference: string
  contractId: string
  origin: string
  destination: string
  amount: string
  beneficiary: string
  beneficiaryAccount: string
  destinationBank: string
  settlementBatch: string
  receivedAt: string
  slaDeadline: string
  currentState: FinalLegState
  outcomeConfidence: OutcomeConfidence
  route: string
  owner: string
  staffAssignment?: StaffAssignment
  safeAction: string
  valueAtRisk: string
  state: HealthState
}

export type FinalLegAttempt = {
  id: string
  instructionReference: string
  route: string
  rail: string
  destinationBank: string
  submittedAt: string
  acceptedAt: string
  creditedAt: string
  p95CreditTime: string
  timeoutRate: string
  lateSuccessRisk: string
  state: FinalLegState
  health: HealthState
  decisionReason: string
}

export type SlaPolicy = {
  id: string
  name: string
  duration: string
  cooldownWindow: string
  lateSuccessWindow: string
  evidenceRequired: string[]
  state: HealthState
}

export type BackoffPolicy = {
  id: string
  route: string
  slaPolicyId: string
  maxRequeryAttempts: number
  schedule: string[]
  rerouteAfterAcceptance: boolean
  state: HealthState
}

export type RequeryAttempt = {
  id: string
  instructionReference: string
  attemptNumber: number
  dueAt: string
  completedAt: string
  method: string
  result: string
  trigger: 'automatic' | 'manual'
  state: HealthState
}

export type ProviderEscalation = {
  id: string
  sourceTable: 'provider_escalations'
  instructionReference: string
  provider: string
  channel: 'email' | 'provider_portal' | 'api_ticket'
  recipient: string
  triggerAfter: string
  status: 'not_due' | 'ready_to_send' | 'sent' | 'acknowledged'
  lastSentAt: string
  nextEscalationAt: string
  templateSubject: string
  requiredEvidence: string[]
  state: HealthState
}

export type OutcomeEvidence = {
  instructionReference: string
  type: 'partner_callback' | 'rail_session' | 'ledger_posting' | 'settlement_batch' | 'operator_note' | 'audit_event'
  label: string
  reference: string
  status: string
  owner: string
  state: HealthState
}

export type RemediationCase = {
  id: string
  instructionReference: string
  apiReference?: string
  queue: 'cooldown' | 'requerying' | 'exhausted' | 'recon_break' | 'completed_outside_platform' | 'reversal'
  route: string
  provider: string
  requeryMethod: string
  automaticAttempts: number
  maxAutomaticAttempts: number
  nextRequeryAt: string
  automationStatus: 'cooldown' | 'scheduled' | 'running' | 'exhausted' | 'manual_only'
  title: string
  theory: string
  valueAtRisk: string
  age: string
  duplicateRisk: string
  evidenceGap: string
  nextAction: string
  owner: string
  staffAssignment?: StaffAssignment
  statusRetrievalPlan: string[]
  escalationAfter: string
  duplicatePaymentGuard: string
  providerEscalation?: ProviderEscalation
  makerChecker: string
  makerCheckerState: 'not_required' | 'maker_required' | 'checker_pending' | 'approved' | 'rejected'
  safeClosure: string
  state: HealthState
}

export type RoutePenalty = {
  route: string
  rail: string
  destinationBank: string
  p95CreditTime: string
  timeoutRate: string
  lateSuccessRate: string
  openCases: number
  penaltyScore: number
  penaltyReason: string
  trafficSplit: string
  fallbackOrder: string
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
  routingContracts: RoutingContract[]
  creditLegs: CreditLeg[]
  inboundSla: InboundSlaRow[]
  imtoPartners: ImtoPartner[]
  complianceHolds: ComplianceHold[]
  featureSwitches: FeatureSwitch[]
  inboundContracts: InboundContract[]
  standingAccounts: StandingAccount[]
  incomingInstructions: IncomingInstruction[]
  finalLegAttempts: FinalLegAttempt[]
  slaPolicies: SlaPolicy[]
  backoffPolicies: BackoffPolicy[]
  requeryAttempts: RequeryAttempt[]
  providerEscalations: ProviderEscalation[]
  outcomeEvidence: OutcomeEvidence[]
  remediationCases: RemediationCase[]
  routePenalties: RoutePenalty[]
  routeDecisions: RouteDecisionRecord[]
  evidenceRequirements: EvidenceRequirement[]
  reconciliationMatches: ReconciliationMatch[]
  routeHealthWindows: RouteHealthWindow[]
  integrationHealth: IntegrationHealth[]
  caseActionSteps: CaseActionStep[]
  staffAssignments: StaffAssignment[]
}
