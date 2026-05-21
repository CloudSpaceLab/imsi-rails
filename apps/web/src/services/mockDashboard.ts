import { countryIdentities, providerIdentities } from './identity'
import type { DashboardMock, HealthState, UiScenario } from '../types'

const baseDashboard = (): DashboardMock => ({
  scenario: 'degraded-ria',
  viewState: 'ready',
  providerIdentities: Object.values(providerIdentities),
  countryIdentities: Object.values(countryIdentities),
  summary: {
    globalHealth: '84.6%',
    valueToday: '$18.4M',
    transactionsToday: '42,618',
    p95CreditTime: '4m 18s',
    stuckTransactions: 184,
    activeIncidents: 2,
    lastUpdated: '14:32:18 UTC',
    atRiskValue: '$2.7M',
    topRisk: 'Europe to Nigeria account payouts',
    safeAction: 'Review a 25% new-traffic shift to Thunes.',
    connection: {
      mode: 'polling',
      freshness: 'fresh',
      updatedAt: '14:32:18 UTC',
      nextPollIn: 'Sample operational data',
    },
    metrics: [
      { label: 'Routes healthy', value: '84.6%', detail: '15 min measured window', trend: '-12.6%', state: 'degraded' },
      { label: 'At-risk value', value: '$2.7M', detail: '184 unsettled transfers', trend: '+$680K', state: 'watch' },
      { label: 'P95 credit time', value: '4m 18s', detail: 'target 90s account payout', trend: '+2m 48s', state: 'degraded' },
      { label: 'Shift candidate', value: '25%', detail: 'new traffic only', trend: 'review needed', state: 'recovery' },
    ],
  },
  dateRange: {
    label: 'Today',
    start: '2026-05-20 00:00 WAT',
    end: '2026-05-20 14:32 WAT',
    timezone: 'Africa/Lagos',
  },
  qaPolicy: {
    name: 'Nigeria account payout QA',
    version: 'QA-NG-ACCT-v4',
    thresholdSeconds: 90,
    warningSeconds: 75,
    scope: 'Direct-to-account inbound Nigeria',
    completedWithinPolicy: '95.8%',
    breachRate: '4.2%',
    weekComparison: '+1.8% vs last week',
    monthComparison: '+3.4% vs 30-day avg',
    updatedAt: '2026-05-20 13:48 WAT',
  },
  operationalStats: [
    {
      label: 'Completed transactions',
      value: '40,918',
      detail: 'credited to destination in selected range',
      weekComparison: '+8.2% vs last week',
      monthComparison: '+14.6% vs 30-day avg',
      state: 'healthy',
    },
    {
      label: 'Stalled or delayed',
      value: '1,700',
      detail: 'over QA threshold or awaiting destination credit',
      weekComparison: '+312 vs last week',
      monthComparison: '-9.1% vs 30-day avg',
      state: 'degraded',
    },
    {
      label: 'Completed on time',
      value: '39,218',
      detail: 'under current 90s QA policy',
      weekComparison: '+1.8% vs last week',
      monthComparison: '+3.4% vs 30-day avg',
      state: 'healthy',
    },
    {
      label: 'Median total time',
      value: '38s',
      detail: 'sender initiation to destination credit',
      weekComparison: '-6s vs last week',
      monthComparison: '-11s vs 30-day avg',
      state: 'healthy',
    },
  ],
  visuals: {
    completionTrend: [
      { label: '07:00', value: 97.2 },
      { label: '08:00', value: 97.8 },
      { label: '09:00', value: 96.9 },
      { label: '10:00', value: 96.4 },
      { label: '11:00', value: 96.1 },
      { label: '12:00', value: 95.9 },
      { label: '13:00', value: 94.6 },
      { label: '14:00', value: 95.8 },
    ],
    volumeTrend: [
      { label: '07', value: 4120 },
      { label: '08', value: 4820 },
      { label: '09', value: 5310 },
      { label: '10', value: 5680 },
      { label: '11', value: 6220 },
      { label: '12', value: 5890 },
      { label: '13', value: 6460 },
      { label: '14', value: 4118 },
    ],
    latencyBands: [
      { label: 'P50', valueSeconds: 38, targetSeconds: 90, state: 'healthy' },
      { label: 'P95', valueSeconds: 258, targetSeconds: 90, state: 'degraded' },
      { label: 'P99', valueSeconds: 464, targetSeconds: 90, state: 'blocked' },
    ],
    exceptionBreakdown: [
      { label: 'Delayed', value: 1284, state: 'degraded' },
      { label: 'Stalled', value: 416, state: 'blocked' },
      { label: 'Recon', value: 31, state: 'watch' },
    ],
    hourHealth: [
      { label: '07', value: 97, state: 'healthy' },
      { label: '08', value: 98, state: 'healthy' },
      { label: '09', value: 97, state: 'healthy' },
      { label: '10', value: 96, state: 'healthy' },
      { label: '11', value: 96, state: 'watch' },
      { label: '12', value: 96, state: 'watch' },
      { label: '13', value: 95, state: 'degraded' },
      { label: '14', value: 96, state: 'watch' },
    ],
  },
  recommendation: {
    title: 'Move 25% of Europe to Nigeria account payouts off Ria.',
    trigger: 'Ria breached the 90s P95 target for 15 minutes and timeout rate reached 12.5%.',
    affectedTraffic: '184 transfers',
    affectedValue: '$2.7M',
    currentRoute: 'Ria -> NIP',
    suggestedRoute: 'Thunes -> NIP',
    nextAction: 'Review policy',
    evidence: 'Latency breach, timeout spike, fresh FX, healthy Thunes acceptance.',
    state: 'degraded',
  },
  corridors: [
    {
      corridor: 'EU -> Nigeria',
      payout: 'Bank account',
      state: 'degraded' as HealthState,
      selectedRoute: 'Ria -> NIP',
      score: 63,
      p95: '4m 18s',
      cost: '0.74%',
      split: '25 / 75',
      recommendation: 'Preview shift',
      risk: 'Provider acceptance lag',
      atRiskValue: '$2.7M',
      owner: 'Payments Ops',
      status: 'Review policy',
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
      recommendation: 'Hold split',
      risk: 'Recovery testing',
      atRiskValue: '$410K',
      owner: 'Route desk',
      status: 'Monitoring',
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
      recommendation: 'Watch lag',
      risk: 'Callback delay',
      atRiskValue: '$930K',
      owner: 'Provider ops',
      status: 'Watching lag',
    },
    {
      corridor: 'US -> Nigeria',
      payout: 'Bank account',
      state: 'healthy' as HealthState,
      selectedRoute: 'Thunes -> NIP',
      score: 94,
      p95: '31s',
      cost: '0.82%',
      split: '72 / 28',
      recommendation: 'Route more',
      risk: 'Inside policy',
      atRiskValue: '$0',
      owner: 'Route engine',
      status: 'Healthy route',
    },
  ],
  providerScores: [
    {
      rank: 1,
      provider: 'Thunes',
      corridor: 'United States -> Nigeria',
      successRate: '99.1%',
      p50: '14s',
      p95: '31s',
      p99: '58s',
      stuckRate: '0.04%',
      settlementExceptions: 2,
      state: 'healthy',
      supportSla: '12m',
      trafficShare: '47%',
    },
    {
      rank: 2,
      provider: 'Remitly',
      corridor: 'UK -> Nigeria',
      successRate: '98.4%',
      p50: '18s',
      p95: '49s',
      p99: '1m 12s',
      stuckRate: '0.09%',
      settlementExceptions: 5,
      state: 'watch',
      supportSla: '18m',
      trafficShare: '24%',
    },
    {
      rank: 3,
      provider: 'PAPSS',
      corridor: 'Kenya -> Nigeria',
      successRate: '96.8%',
      p50: '23s',
      p95: '58s',
      p99: '1m 36s',
      stuckRate: '0.21%',
      settlementExceptions: 4,
      state: 'recovery',
      supportSla: '22m',
      trafficShare: '9%',
    },
    {
      rank: 4,
      provider: 'Ria',
      corridor: 'EU -> Nigeria',
      successRate: '87.5%',
      p50: '46s',
      p95: '4m 18s',
      p99: '7m 44s',
      stuckRate: '1.18%',
      settlementExceptions: 19,
      state: 'degraded',
      supportSla: '41m',
      trafficShare: '20%',
    },
  ],
  transactions: [
    {
      reference: 'IMSI-txn_000000000001',
      providerReference: 'RIA-EU-7K2Q',
      bankReference: 'ACB-NIP-581209',
      senderCountry: 'Germany',
      destinationCountry: 'Nigeria',
      senderCurrency: 'EUR',
      destinationCurrency: 'NGN',
      destinationType: 'Local bank',
      provider: 'Ria',
      route: 'Ria -> NIP',
      amount: 'EUR 2,400',
      senderStartedAt: '14:29:11 UTC',
      destinationCreditedAt: '14:33:29 UTC',
      totalTime: '4m 18s',
      totalTimeSeconds: 258,
      qaLimitSeconds: 90,
      qaStatus: 'delayed',
      state: 'degraded',
      beneficiary: 'Access Bank / 012****549',
      currentOwner: 'Provider ops',
      blocker: 'Provider acceptance breached QA policy before destination credit.',
    },
    {
      reference: 'IMSI-txn_000000000014',
      providerReference: 'THN-US-91A2',
      bankReference: 'GTB-NIP-882104',
      senderCountry: 'United States',
      destinationCountry: 'Nigeria',
      senderCurrency: 'USD',
      destinationCurrency: 'NGN',
      destinationType: 'Local bank',
      provider: 'Thunes',
      route: 'Thunes -> NIP',
      amount: 'USD 950',
      senderStartedAt: '14:28:04 UTC',
      destinationCreditedAt: '14:28:37 UTC',
      totalTime: '33s',
      totalTimeSeconds: 33,
      qaLimitSeconds: 90,
      qaStatus: 'on_time',
      state: 'healthy',
      beneficiary: 'GTBank / 014****771',
      currentOwner: 'Route engine',
      blocker: 'None. Completed under QA policy.',
    },
    {
      reference: 'IMSI-txn_000000031822',
      providerReference: 'RMT-UK-55Q8',
      bankReference: 'UBA-NIP-102981',
      senderCountry: 'United Kingdom',
      destinationCountry: 'Nigeria',
      senderCurrency: 'GBP',
      destinationCurrency: 'NGN',
      destinationType: 'Local bank',
      provider: 'Remitly',
      route: 'Remitly -> NIP',
      amount: 'GBP 850',
      senderStartedAt: '14:21:40 UTC',
      destinationCreditedAt: 'pending',
      totalTime: '10m 52s',
      totalTimeSeconds: 652,
      qaLimitSeconds: 90,
      qaStatus: 'stalled',
      state: 'blocked',
      beneficiary: 'UBA / 208****314',
      currentOwner: 'Bank posting monitor',
      blocker: 'Destination credit has not been confirmed.',
    },
    {
      reference: 'IMSI-txn_000000034410',
      providerReference: 'PAPSS-KE-41B7',
      bankReference: 'FBN-NIP-991430',
      senderCountry: 'Kenya',
      destinationCountry: 'Nigeria',
      senderCurrency: 'KES',
      destinationCurrency: 'NGN',
      destinationType: 'International bank',
      provider: 'PAPSS',
      route: 'PAPSS',
      amount: 'KES 180,000',
      senderStartedAt: '14:18:12 UTC',
      destinationCreditedAt: '14:19:08 UTC',
      totalTime: '56s',
      totalTimeSeconds: 56,
      qaLimitSeconds: 90,
      qaStatus: 'on_time',
      state: 'recovery',
      beneficiary: 'FirstBank / 302****108',
      currentOwner: 'Route desk',
      blocker: 'Recovery check completed under QA policy.',
    },
  ],
  trace: {
    reference: 'IMSI-txn_000000000001',
    providerReference: 'RIA-EU-7K2Q',
    bankReference: 'ACB-NIP-581209',
    beneficiary: 'Access Bank / 012****549',
    corridor: 'EU -> Nigeria',
    amount: 'EUR 2,400',
    senderCountry: 'Germany',
    destinationCountry: 'Nigeria',
    senderCurrency: 'EUR',
    destinationCurrency: 'NGN',
    destinationType: 'Local bank',
    senderStartedAt: '14:29:11 UTC',
    destinationCreditedAt: '14:33:29 UTC',
    totalTime: '4m 18s',
    totalTimeSeconds: 258,
    qaLimitSeconds: 90,
    qaStatus: 'delayed',
    currentState: 'watch',
    currentOwner: 'NIP posting monitor',
    blocker: 'Provider accepted late; bank posting remains inside SLA.',
    safeAction: 'Do not reroute in flight. Keep monitoring settlement match.',
    policyVersion: 'v2026.05.20.14.05',
    fallbackRoutes: ['Thunes -> NIP', 'Remitly -> NIP', 'Manual review'],
    selectedRoute: {
      provider: 'Ria',
      route: 'Ria -> NIP',
      score: 63,
      p95: '4m 18s',
      cost: '0.74%',
      state: 'degraded',
      reason: 'Selected before circuit breaker opened.',
      confidence: 'Medium',
      policyVersion: 'v2026.05.20.14.05',
    },
    rejectedRoutes: [
      { provider: 'Thunes', route: 'Thunes -> NIP', reason: 'Fallback only under live split at decision time.' },
      { provider: 'Remitly', route: 'Remitly -> NIP', reason: 'EUR/NGN FX rate was stale.' },
      { provider: 'PAPSS', route: 'PAPSS', reason: 'EU origin unsupported for account payout.' },
    ],
    scoreInputs: [
      { label: 'Reliability', value: '87.5%', state: 'degraded' },
      { label: 'Speed', value: '4m 18s P95', state: 'degraded' },
      { label: 'Cost', value: '0.74%', state: 'healthy' },
      { label: 'FX age', value: '2 min', state: 'healthy' },
      { label: 'Policy', value: 'allowed at decision', state: 'watch' },
    ],
    timeline: [
      { label: 'Sender initiated', owner: 'Bank channel', status: 'done', time: '14:29:11', duration: '0s', source: 'Bank API', reference: 'IDEMP-9e2' },
      { label: 'Validated', owner: 'imsi-rails', status: 'done', time: '14:29:11', duration: '0.4s', source: 'Validation service', reference: 'VAL-810' },
      { label: 'Route picked', owner: 'Route engine', status: 'done', time: '14:29:12', duration: '1s', source: 'Policy v2026.05.20.14.05', reference: 'DEC-542' },
      { label: 'Provider accepted', owner: 'Ria', status: 'done', time: '14:31:20', duration: '2m 09s', source: 'Provider API', reference: 'RIA-EU-7K2Q', note: 'Late acceptance' },
      { label: 'Bank rail posting', owner: 'NIP', status: 'done', time: '14:31:23', duration: '2m 12s', source: 'Bank adapter', reference: 'ACB-NIP-581209' },
      { label: 'Destination credited', owner: 'Access Bank', status: 'done', time: '14:33:29', duration: '4m 18s', source: 'NIP confirmation', reference: 'ACB-NIP-581209' },
      { label: 'Settlement match', owner: 'Settlement', status: 'pending', time: 'pending', source: 'Reconciliation', reference: 'Batch pending' },
    ],
  },
  latency: {
    filters: {
      provider: 'Ria',
      corridor: 'EU -> Nigeria',
      destinationBank: 'Access Bank',
      window: '15 min',
    },
    summary: {
      endToEnd: '4m 18s',
      target: '90s',
      slowestStep: 'Provider accepted',
      affectedTransactions: 184,
    },
    steps: [
      { label: 'Bank submit', owner: 'Bank channel', durationMs: 420, targetMs: 800, state: 'healthy' },
      { label: 'Validation', owner: 'imsi-rails', durationMs: 260, targetMs: 500, state: 'healthy' },
      { label: 'FX lock', owner: 'Treasury rules', durationMs: 1_800, targetMs: 2_000, state: 'healthy' },
      { label: 'Provider accepted', owner: 'Ria', durationMs: 128_000, targetMs: 30_000, state: 'degraded' },
      { label: 'Webhook callback', owner: 'Ria', durationMs: 84_000, targetMs: 45_000, state: 'watch' },
      { label: 'Bank posting', owner: 'NIP', durationMs: 43_000, targetMs: 60_000, state: 'healthy' },
      { label: 'Settlement file', owner: 'Settlement Ops', durationMs: 62_000, targetMs: 120_000, state: 'healthy' },
    ],
  },
  downtimeEvents: [
    {
      time: '14:04',
      title: 'P95 target missed',
      actor: 'imsi-rails',
      state: 'watch',
      detail: 'Europe to Nigeria account payouts crossed 90s.',
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
      title: 'Traffic shift previewed',
      actor: 'Ops analyst',
      state: 'recovery',
      detail: '25% to Thunes looked faster with acceptable cost.',
    },
    {
      time: '14:21',
      title: 'Policy draft marked for review',
      actor: 'Policy workflow',
      state: 'watch',
      detail: 'Draft shift is visible in policy review.',
    },
  ],
  incidents: [
    {
      id: 'INC-2026-0520-014',
      title: 'Ria acceptance latency breach',
      severity: 'degraded',
      corridor: 'EU -> Nigeria',
      owner: 'Payments Ops',
      startedAt: '14:04 UTC',
      affectedTransactions: 184,
      affectedValue: '$2.7M',
      rootCause: 'Provider API timeout spike and webhook lag.',
      nextAction: 'Review 25% traffic shift to Thunes.',
      status: 'Review policy',
    },
    {
      id: 'INC-2026-0520-011',
      title: 'PAPSS recovery check',
      severity: 'recovery',
      corridor: 'Kenya -> Nigeria',
      owner: 'Route desk',
      startedAt: '13:20 UTC',
      affectedTransactions: 42,
      affectedValue: '$410K',
      rootCause: 'Recovery testing after settlement lag.',
      nextAction: 'Hold at 10% until P95 remains under 60s for 30 min.',
      status: 'Monitoring',
    },
  ],
  routeConfig: {
    providers: [
      { provider: 'Thunes', route: 'Europe to Nigeria / NIP', enabled: true, state: 'healthy' },
      { provider: 'Remitly', route: 'Europe to Nigeria / NIP', enabled: true, state: 'stale' },
      { provider: 'Ria', route: 'Europe to Nigeria / NIP', enabled: false, state: 'degraded' },
      { provider: 'PAPSS', route: 'Africa corridor only', enabled: true, state: 'recovery' },
    ],
    fallbackRoutes: [
      { rank: 1, provider: 'Thunes', route: 'NIP account payout', state: 'healthy' },
      { rank: 2, provider: 'Remitly', route: 'NIP account payout', state: 'stale' },
      { rank: 3, provider: 'PAPSS', route: 'Cross-border account payout', state: 'recovery' },
      { rank: 4, provider: 'Ria', route: 'NIP account payout', state: 'degraded' },
    ],
    presets: [
      { label: 'Incident shift', active: true, split: '75 / 25 / 0' },
      { label: 'Balanced', active: false, split: '50 / 30 / 20' },
      { label: 'Recovery', active: false, split: '80 / 10 / 10' },
    ],
    weights: [
      { label: 'Reliability', value: 42 },
      { label: 'Speed', value: 28 },
      { label: 'Cost', value: 16 },
      { label: 'FX', value: 14 },
    ],
    impact: {
      successRate: '+6.1%',
      p95: '-2m 47s',
      cost: '+0.08%',
    },
    history: [
      { time: '14:21', actor: 'Policy workflow', summary: 'Marked 25% shift to Thunes for review.' },
      { time: '14:16', actor: 'Ops analyst', summary: 'Previewed 25% shift from Ria to Thunes.' },
      { time: '13:48', actor: 'Treasury lead', summary: 'Raised FX freshness for Europe to Nigeria.' },
    ],
    workflow: {
      scope: [
        { label: 'Corridor', value: 'Europe to Nigeria' },
        { label: 'Payout', value: 'Bank account' },
        { label: 'Destination bank', value: 'Access Bank, UBA, GTBank' },
        { label: 'Amount band', value: 'EUR 0 - 5,000' },
      ],
      currentPolicy: [
        { label: 'Live split', value: 'Ria 75 / Thunes 25' },
        { label: 'Fallback', value: 'Ria, Thunes, Remitly' },
        { label: 'Breaker action', value: 'Alert only' },
        { label: 'FX max age', value: '20 min' },
      ],
      proposedPolicy: [
        { label: 'Live split', value: 'Thunes 75 / Ria 25', changed: true },
        { label: 'Fallback', value: 'Thunes, Remitly, Ria', changed: true },
        { label: 'Breaker action', value: 'Block new Ria traffic', changed: true },
        { label: 'FX max age', value: '10 min', changed: true },
      ],
      validation: [
        { label: 'Affected traffic', value: '184 active / 1,920 next hour', state: 'watch' },
        { label: 'Duplicate payout risk', value: 'New traffic only', state: 'healthy' },
        { label: 'Change mode', value: 'New traffic only', state: 'healthy' },
        { label: 'Current policy', value: 'Recorded', state: 'healthy' },
      ],
      change: {
        reason: 'Ria missed 90s P95 target for 15 minutes and Thunes is healthy with fresh FX.',
      },
    },
  },
  policySimulationSamples: [
    {
      reference: 'SIM-ROUTE-1042',
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
        state: 'degraded',
        reason: 'Cheapest route but currently degraded.',
        confidence: 'Low',
        policyVersion: 'v2026.05.20.14.05',
      },
      proposed: {
        provider: 'Thunes',
        route: 'Thunes -> NIP',
        score: 91,
        p95: '37s',
        cost: '0.82%',
        state: 'healthy',
        reason: 'Healthy route with fresh FX and lower intervention risk.',
        confidence: 'High',
        policyVersion: 'draft-v2026.05.20.14.21',
      },
      rejectedRoutes: [
        { provider: 'Ria', route: 'Ria -> NIP', reason: 'Circuit breaker blocks new traffic.' },
        { provider: 'Remitly', route: 'Remitly -> NIP', reason: 'EUR/NGN FX is stale.' },
        { provider: 'PAPSS', route: 'PAPSS', reason: 'EU account payouts are not supported.' },
      ],
      reportMetrics: [
        { label: 'Better route', value: '386', detail: 'of 500 historical transactions' },
        { label: 'P95 change', value: '-3m 41s', detail: 'vs live rules' },
        { label: 'Cost change', value: '+0.08%', detail: 'effective cost' },
      ],
      reportRows: [
        { bucket: 'Healthy payout', currentRoute: 'Ria -> NIP', proposedRoute: 'Thunes -> NIP', result: 'Faster route' },
        { bucket: 'Stale FX', currentRoute: 'Ria -> NIP', proposedRoute: 'Hold for refresh', result: 'Wait for FX' },
        { bucket: 'Manual review', currentRoute: 'Ria -> NIP', proposedRoute: 'No change', result: 'Same decision' },
      ],
    },
  ],
  fxCostBoard: {
    corridor: 'EU -> Nigeria',
    pair: 'EUR/NGN',
    window: '15 min',
    refreshedAt: '14:31 UTC',
    cheapestProvider: 'Ria',
    recommendedProvider: 'Thunes',
    rateAlert: 'Remitly rate is 22 min old. New EUR/NGN traffic is blocked until refresh.',
    decision: 'Ria is cheapest, but the route is degraded. Thunes costs +0.08% and is healthy, so it is the selected eligible route.',
    routes: [
      {
        provider: 'Thunes',
        route: 'Thunes -> NIP',
        pair: 'EUR/NGN',
        rate: '1,721.40',
        updatedAt: '14:31 UTC',
        state: 'healthy',
        fee: '0.32%',
        spread: '0.50%',
        effectiveCost: '0.82%',
        payoutTime: '37s',
        cheapest: false,
        recommended: true,
        note: 'Healthy route with fresh FX.',
      },
      {
        provider: 'Ria',
        route: 'Ria -> NIP',
        pair: 'EUR/NGN',
        rate: '1,724.10',
        updatedAt: '14:30 UTC',
        state: 'degraded',
        fee: '0.26%',
        spread: '0.48%',
        effectiveCost: '0.74%',
        payoutTime: '4m 18s',
        cheapest: true,
        recommended: false,
        note: 'Cheapest, but too slow right now.',
      },
      {
        provider: 'Remitly',
        route: 'Remitly -> NIP',
        pair: 'EUR/NGN',
        rate: '1,718.90',
        updatedAt: '14:09 UTC',
        state: 'stale',
        fee: '0.30%',
        spread: '0.56%',
        effectiveCost: '0.86%',
        payoutTime: '58s',
        cheapest: false,
        recommended: false,
        note: 'Rate is stale.',
      },
      {
        provider: 'PAPSS',
        route: 'PAPSS',
        pair: 'EUR/NGN',
        rate: '1,716.20',
        updatedAt: '14:28 UTC',
        state: 'watch',
        fee: '0.18%',
        spread: '0.65%',
        effectiveCost: '0.83%',
        payoutTime: '1m 12s',
        cheapest: false,
        recommended: false,
        note: 'Not eligible for this payout.',
      },
    ],
  },
  reconciliation: [
    { reference: 'REC-7781', provider: 'Ria', amount: 'EUR 2,400', age: '42m', reason: 'Late provider settlement file', owner: 'Settlement Ops', state: 'watch' },
    { reference: 'REC-7784', provider: 'Ria', amount: 'EUR 910', age: '31m', reason: 'Provider accepted, bank credit pending', owner: 'NIP monitor', state: 'degraded' },
    { reference: 'REC-7792', provider: 'Remitly', amount: 'GBP 850', age: '18m', reason: 'FX timestamp mismatch', owner: 'Treasury Ops', state: 'stale' },
  ],
  auditEvents: [
    { time: '14:21:44', actor: 'Ops analyst', action: 'Marked draft for review', object: 'Europe to Nigeria policy draft', reason: 'Ria P95 breach', state: 'watch' },
    { time: '14:16:08', actor: 'Ops analyst', action: 'Previewed traffic shift', object: 'Ria to Thunes 25%', reason: 'Latency breach', state: 'recovery' },
    { time: '14:13:22', actor: 'Circuit breaker', action: 'Marked degraded', object: 'Ria -> NIP', reason: 'Timeout rate 12.5%', state: 'degraded' },
    { time: '13:48:12', actor: 'Treasury lead', action: 'Changed FX freshness', object: 'EUR/NGN max age', reason: 'Stale-rate control', state: 'healthy' },
  ],
})

export const getDashboardMock = (scenario: UiScenario = 'degraded'): DashboardMock => {
  const dashboard = baseDashboard()
  dashboard.scenario = scenario === 'degraded' ? 'degraded-ria' : scenario

  if (scenario === 'healthy') {
    dashboard.summary.globalHealth = '98.7%'
    dashboard.summary.p95CreditTime = '1m 8s'
    dashboard.summary.stuckTransactions = 12
    dashboard.summary.activeIncidents = 0
    dashboard.summary.atRiskValue = '$180K'
    dashboard.summary.topRisk = 'No corridor outside policy'
    dashboard.summary.safeAction = 'Keep current routing policy active.'
    dashboard.summary.metrics = [
      { label: 'Routes healthy', value: '98.7%', detail: '15 min measured window', trend: '+1.4%', state: 'healthy' },
      { label: 'At-risk value', value: '$180K', detail: '12 transfers pending review', trend: '-$2.5M', state: 'healthy' },
      { label: 'P95 credit time', value: '1m 8s', detail: 'target 90s account payout', trend: '-3m 10s', state: 'healthy' },
      { label: 'Shift candidate', value: '0%', detail: 'no shift needed', trend: 'inside policy', state: 'healthy' },
    ]
    dashboard.recommendation = {
      ...dashboard.recommendation,
      title: 'No route action needed.',
      trigger: 'All monitored routes are inside the current operating policy.',
      affectedTraffic: '0 transfers',
      affectedValue: '$0',
      currentRoute: 'Ria -> NIP',
      suggestedRoute: 'Keep current allocation',
      nextAction: 'Keep monitoring',
      evidence: 'Success and latency are inside SLA.',
      state: 'healthy',
    }
    dashboard.incidents = []
    dashboard.reconciliation = dashboard.reconciliation.slice(0, 1).map((item) => ({
      ...item,
      age: '8m',
      reason: 'Provider file awaiting scheduled refresh',
      state: 'healthy',
    }))
    dashboard.corridors = dashboard.corridors.map((corridor) => ({
      ...corridor,
      state: corridor.state === 'healthy' ? 'healthy' : 'watch',
      score: Math.max(corridor.score, 91),
      risk: 'Inside policy',
      recommendation: 'Monitor',
      status: 'Healthy',
    }))
    dashboard.providerScores = dashboard.providerScores.map((provider) => ({
      ...provider,
      successRate: provider.provider === 'Ria' ? '97.2%' : provider.successRate,
      p95: provider.provider === 'Ria' ? '1m 6s' : provider.p95,
      state: provider.state === 'degraded' ? 'watch' : provider.state,
      settlementExceptions: provider.provider === 'Ria' ? 2 : provider.settlementExceptions,
    }))
    dashboard.visuals.hourHealth = dashboard.visuals.hourHealth.map((point) => ({ ...point, state: 'healthy' }))
  }

  if (scenario === 'traffic-shift') {
    dashboard.summary.globalHealth = '91.4%'
    dashboard.summary.p95CreditTime = '2m 12s'
    dashboard.summary.stuckTransactions = 71
    dashboard.summary.activeIncidents = 1
    dashboard.summary.atRiskValue = '$940K'
    dashboard.summary.topRisk = 'Europe to Nigeria is recovering after a traffic shift'
    dashboard.summary.safeAction = 'Review a staged increase from 25% to 50% on Thunes.'
    dashboard.recommendation = {
      ...dashboard.recommendation,
      title: 'Traffic shift is reducing failed credits.',
      trigger: 'Ria callback lag is down, but destination credit tail latency is still above target.',
      affectedTraffic: '25% shifted',
      affectedValue: '$940K monitored',
      currentRoute: 'Ria -> NIP',
      suggestedRoute: 'Thunes -> NIP',
      nextAction: 'Hold 25% split for one more window, then test 50%.',
      evidence: 'Projected 312 failures avoided in the current window.',
      state: 'recovery',
    }
    dashboard.corridors = dashboard.corridors.map((corridor) =>
      corridor.corridor.includes('European Union') || corridor.corridor.startsWith('EU')
        ? { ...corridor, state: 'recovery', selectedRoute: 'Thunes -> NIP', score: 86, split: '75% Ria / 25% Thunes', recommendation: 'Hold recovery split', risk: 'Ria callbacks improving', status: 'Recovery testing', owner: 'Route Ops' }
        : corridor,
    )
    dashboard.auditEvents = [
      { time: '14:28:14', actor: 'Ops lead', action: 'Held recovery split', object: 'Europe to Nigeria 25%', reason: 'P95 improving after shift', state: 'recovery' },
      ...dashboard.auditEvents,
    ]
  }

  if (scenario === 'pilot-report') {
    dashboard.summary.globalHealth = '96.8%'
    dashboard.summary.valueToday = '$118.7M'
    dashboard.summary.transactionsToday = '286,420'
    dashboard.summary.p95CreditTime = '2m 1s'
    dashboard.summary.stuckTransactions = 910
    dashboard.summary.activeIncidents = 1
    dashboard.summary.atRiskValue = '$1.2M'
    dashboard.summary.topRisk = 'Ria latency improved after staged shifts'
    dashboard.summary.safeAction = 'Export the operating evidence for bank review.'
    dashboard.dateRange.label = '30-day review'
    dashboard.recommendation = {
      ...dashboard.recommendation,
      title: 'Operating evidence is ready for review.',
      trigger: 'Routing changes reduced avoidable failures and improved SLA completion.',
      affectedTraffic: '286,420 transfers',
      affectedValue: '$118.7M processed',
      currentRoute: 'Mixed provider allocation',
      suggestedRoute: 'Policy v2026.05.20.14.30',
      nextAction: 'Review provider scorecards and export the decision log.',
      evidence: '1,184 failures avoided across the review window.',
      state: 'healthy',
    }
    dashboard.incidents = dashboard.incidents.slice(0, 1).map((incident) => ({
      ...incident,
      title: 'Ria callback incident closed with recovery monitoring',
      status: 'Resolved - monitoring',
      severity: 'recovery',
    }))
  }

  if (scenario === 'loading') {
    dashboard.viewState = 'loading'
  }

  if (scenario === 'api-failure') {
    dashboard.viewState = 'error'
    dashboard.summary.connection.freshness = 'unavailable'
    dashboard.summary.connection.mode = 'static'
    dashboard.summary.connection.nextPollIn = 'last verified state'
  }

  if (scenario === 'permission-denied') {
    dashboard.viewState = 'permission-denied'
    dashboard.auditEvents = []
  }

  if (scenario === 'stale-fx') {
    dashboard.viewState = 'stale'
    dashboard.summary.connection.freshness = 'stale'
    dashboard.fxCostBoard.routes = dashboard.fxCostBoard.routes.map((route) =>
      route.provider === 'Remitly' ? { ...route, state: 'stale' as HealthState, note: 'Rate is stale.' } : route,
    )
  }

  if (scenario === 'empty') {
    dashboard.viewState = 'empty'
    dashboard.incidents = []
    dashboard.reconciliation = []
    dashboard.transactions = []
    dashboard.auditEvents = []
    dashboard.summary.stuckTransactions = 0
    dashboard.summary.activeIncidents = 0
    dashboard.recommendation = {
      ...dashboard.recommendation,
      title: 'No route action needed.',
      trigger: 'All monitored routes are inside policy.',
      affectedTraffic: '0 transfers',
      affectedValue: '$0',
      nextAction: 'Keep monitoring',
      state: 'healthy',
    }
  }

  return dashboard
}
