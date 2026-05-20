import type {
  ChangeHistoryItem,
  DowntimeEvent,
  FallbackRoute,
  HealthState,
  LatencyStep,
  PolicySimulationSample,
  ProviderToggle,
  ScoringWeight,
  TimelineStep,
  TrafficSplitPreset,
} from './types'

export const summary = {
  globalHealth: '97.2%',
  valueToday: '$18.4M',
  transactionsToday: '42,618',
  p95CreditTime: '42s',
  stuckTransactions: 18,
  activeIncidents: 2,
  lastUpdated: '14:32:18 UTC',
}

export const corridors = [
  {
    corridor: 'US -> Nigeria',
    payout: 'Bank account',
    state: 'healthy' as HealthState,
    selectedRoute: 'Thunes -> NIP',
    score: 94,
    p95: '31s',
    cost: '0.82%',
    split: '72 / 28',
    recommendation: 'Send more',
  },
  {
    corridor: 'UK -> Nigeria',
    payout: 'Bank account',
    state: 'watch' as HealthState,
    selectedRoute: 'Remitly -> NIP',
    score: 87,
    p95: '49s',
    cost: '0.91%',
    split: '60 / 40',
    recommendation: 'Watch webhook lag',
  },
  {
    corridor: 'EU -> Nigeria',
    payout: 'Bank account',
    state: 'degraded' as HealthState,
    selectedRoute: 'Ria -> NIP',
    score: 63,
    p95: '4m 18s',
    cost: '0.74%',
    split: '25 / 75',
    recommendation: 'Shift traffic',
  },
  {
    corridor: 'Kenya -> Nigeria',
    payout: 'Local account',
    state: 'recovery' as HealthState,
    selectedRoute: 'PAPSS',
    score: 79,
    p95: '58s',
    cost: '0.68%',
    split: '10 / 90',
    recommendation: 'Test recovery',
  },
]

export const providerScores = [
  {
    provider: 'Thunes',
    corridor: 'US -> Nigeria',
    successRate: '99.1%',
    p50: '14s',
    p95: '31s',
    p99: '58s',
    stuckRate: '0.04%',
    settlementExceptions: 2,
    state: 'healthy' as HealthState,
  },
  {
    provider: 'Remitly',
    corridor: 'UK -> Nigeria',
    successRate: '98.4%',
    p50: '18s',
    p95: '49s',
    p99: '1m 12s',
    stuckRate: '0.09%',
    settlementExceptions: 5,
    state: 'healthy' as HealthState,
  },
  {
    provider: 'Ria',
    corridor: 'EU -> Nigeria',
    successRate: '87.5%',
    p50: '46s',
    p95: '4m 18s',
    p99: '7m 44s',
    stuckRate: '1.18%',
    settlementExceptions: 19,
    state: 'degraded' as HealthState,
  },
  {
    provider: 'PAPSS',
    corridor: 'Kenya -> Nigeria',
    successRate: '96.8%',
    p50: '23s',
    p95: '58s',
    p99: '1m 36s',
    stuckRate: '0.21%',
    settlementExceptions: 4,
    state: 'recovery' as HealthState,
  },
]

export const timeline: TimelineStep[] = [
  { label: 'Received', owner: 'Bank channel', status: 'done', time: '14:29:11' },
  { label: 'Validated', owner: 'imsi-rails', status: 'done', time: '14:29:11' },
  { label: 'Route picked', owner: 'Route engine', status: 'done', time: '14:29:12' },
  { label: 'Provider accepted', owner: 'Thunes', status: 'done', time: '14:29:16' },
  { label: 'Bank rail posting', owner: 'NIP', status: 'current', time: '14:29:23' },
  { label: 'Settlement match', owner: 'Settlement', status: 'pending', time: 'pending' },
]

export const drilldownFilters = {
  provider: 'Ria',
  corridor: 'EU -> Nigeria',
  destinationBank: 'Access Bank',
  window: '15 min',
}

export const latencySummary = {
  endToEnd: '4m 18s',
  target: '90s',
  slowestStep: 'Provider accepted',
  affectedTransactions: 184,
}

export const latencySteps: LatencyStep[] = [
  { label: 'Bank submit', owner: 'Bank channel', durationMs: 420, targetMs: 800, state: 'healthy' },
  { label: 'Validation', owner: 'imsi-rails', durationMs: 260, targetMs: 500, state: 'healthy' },
  { label: 'FX lock', owner: 'Treasury rules', durationMs: 1_800, targetMs: 2_000, state: 'healthy' },
  { label: 'Provider accepted', owner: 'Ria', durationMs: 128_000, targetMs: 30_000, state: 'degraded' },
  { label: 'Webhook callback', owner: 'Ria', durationMs: 84_000, targetMs: 45_000, state: 'watch' },
  { label: 'Bank posting', owner: 'NIP', durationMs: 43_000, targetMs: 60_000, state: 'healthy' },
]

export const downtimeEvents: DowntimeEvent[] = [
  {
    time: '14:04',
    title: 'P95 target missed',
    actor: 'imsi-rails',
    state: 'watch',
    detail: 'EU -> Nigeria account payouts crossed 90s.',
  },
  {
    time: '14:13',
    title: 'Ria route degraded',
    actor: 'Ria adapter',
    state: 'degraded',
    detail: 'Timeouts reached 12.5% in the last 15 min.',
  },
  {
    time: '14:16',
    title: 'Traffic shift tested',
    actor: 'Ops analyst',
    state: 'healthy',
    detail: '25% to Thunes looked faster with acceptable cost.',
  },
  {
    time: '14:21',
    title: 'Recovery test started',
    actor: 'Circuit breaker',
    state: 'recovery',
    detail: 'Ria held at 10% while webhook lag is watched.',
  },
]

export const providerToggles: ProviderToggle[] = [
  { provider: 'Thunes', route: 'US -> Nigeria / NIP', enabled: true, state: 'healthy' as HealthState },
  { provider: 'Remitly', route: 'UK -> Nigeria / NIP', enabled: true, state: 'healthy' as HealthState },
  { provider: 'Ria', route: 'EU -> Nigeria / NIP', enabled: false, state: 'degraded' as HealthState },
  { provider: 'PAPSS', route: 'Kenya -> Nigeria / PAPSS', enabled: true, state: 'recovery' as HealthState },
]

export const fallbackRoutes: FallbackRoute[] = [
  { rank: 1, provider: 'Thunes', route: 'NIP account payout', state: 'healthy' as HealthState },
  { rank: 2, provider: 'Remitly', route: 'NIP account payout', state: 'healthy' as HealthState },
  { rank: 3, provider: 'PAPSS', route: 'Cross-border account payout', state: 'recovery' as HealthState },
  { rank: 4, provider: 'Ria', route: 'NIP account payout', state: 'degraded' as HealthState },
]

export const trafficSplitPresets: TrafficSplitPreset[] = [
  { label: 'Reliability', active: true, split: '70 / 20 / 10' },
  { label: 'Balanced', active: false, split: '50 / 30 / 20' },
  { label: 'Recovery', active: false, split: '80 / 10 / 10' },
]

export const scoringWeights: ScoringWeight[] = [
  { label: 'Reliability', value: 40 },
  { label: 'Speed', value: 25 },
  { label: 'Cost', value: 20 },
  { label: 'FX', value: 15 },
]

export const changeHistory: ChangeHistoryItem[] = [
  { time: '14:16', actor: 'Ops analyst', summary: 'Tested 25% shift from Ria to Thunes.' },
  { time: '13:48', actor: 'Treasury lead', summary: 'Raised FX freshness for EU -> Nigeria.' },
  { time: '12:22', actor: 'Switch operator', summary: 'Moved PAPSS into recovery fallback order.' },
]

export const routeConfigImpact = {
  successRate: '+6.1%',
  p95: '-2m 47s',
  cost: '+0.08%',
}

export const policySimulationSamples: PolicySimulationSample[] = [
  {
    reference: 'SIM-EU-NG-1042',
    corridor: 'EU -> Nigeria',
    origin: 'Germany',
    destination: 'Access Bank',
    amount: 'EUR 2,400',
    payout: 'Bank account',
    current: {
      provider: 'Ria',
      route: 'Ria -> NIP',
      score: 63,
      p95: '4m 18s',
      cost: '0.74%',
      state: 'degraded' as HealthState,
    },
    proposed: {
      provider: 'Thunes',
      route: 'Thunes -> NIP',
      score: 91,
      p95: '37s',
      cost: '0.82%',
      state: 'healthy' as HealthState,
    },
    rejectedRoutes: [
      { provider: 'Ria', route: 'Ria -> NIP', reason: 'Ria missed the active P95 target.' },
      { provider: 'Remitly', route: 'Remitly -> NIP', reason: 'EUR/NGN FX is stale.' },
      { provider: 'PAPSS', route: 'PAPSS', reason: 'EU account payouts are not supported.' },
    ],
    reportMetrics: [
      { label: 'Better route', value: '386', detail: 'of 500 transactions' },
      { label: 'P95 change', value: '-3m 41s', detail: 'vs live rules' },
      { label: 'Cost change', value: '+0.08%', detail: 'effective cost' },
    ],
    reportRows: [
      { bucket: 'Healthy payout', currentRoute: 'Ria -> NIP', proposedRoute: 'Thunes -> NIP', result: 'Faster route' },
      { bucket: 'Stale FX', currentRoute: 'Ria -> NIP', proposedRoute: 'Hold for refresh', result: 'Wait for FX' },
      { bucket: 'Manual review', currentRoute: 'Ria -> NIP', proposedRoute: 'No change', result: 'Same decision' },
    ],
  },
  {
    reference: 'SIM-UK-NG-2219',
    corridor: 'UK -> Nigeria',
    origin: 'United Kingdom',
    destination: 'GTBank',
    amount: 'GBP 850',
    payout: 'Bank account',
    current: {
      provider: 'Remitly',
      route: 'Remitly -> NIP',
      score: 87,
      p95: '49s',
      cost: '0.91%',
      state: 'watch' as HealthState,
    },
    proposed: {
      provider: 'Remitly',
      route: 'Remitly -> NIP',
      score: 89,
      p95: '46s',
      cost: '0.91%',
      state: 'healthy' as HealthState,
    },
    rejectedRoutes: [
      { provider: 'Ria', route: 'Ria -> NIP', reason: 'Ria is degraded for account payouts.' },
      { provider: 'PAPSS', route: 'PAPSS', reason: 'GBP origin is not supported.' },
      { provider: 'Thunes', route: 'Thunes -> NIP', reason: 'Live route is cheaper and still inside target.' },
    ],
    reportMetrics: [
      { label: 'Better route', value: '74', detail: 'of 500 transactions' },
      { label: 'P95 change', value: '-3s', detail: 'vs live rules' },
      { label: 'Cost change', value: '0.00%', detail: 'effective cost' },
    ],
    reportRows: [
      { bucket: 'Normal traffic', currentRoute: 'Remitly -> NIP', proposedRoute: 'Remitly -> NIP', result: 'No change' },
      { bucket: 'Webhook lag', currentRoute: 'Remitly -> NIP', proposedRoute: 'Thunes -> NIP', result: 'Use canary' },
      { bucket: 'FX refresh', currentRoute: 'Remitly -> NIP', proposedRoute: 'Remitly -> NIP', result: 'Same decision' },
    ],
  },
]

