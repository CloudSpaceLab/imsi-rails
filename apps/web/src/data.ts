import type { HealthState, TimelineStep } from './types'

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
    recommendation: 'Route more',
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
    recommendation: 'Recovery test',
  },
]

export const providerScores = [
  { provider: 'Thunes', reliability: 98, speed: 96, cost: 88, fx: 91, state: 'healthy' as HealthState },
  { provider: 'Remitly', reliability: 96, speed: 90, cost: 84, fx: 89, state: 'healthy' as HealthState },
  { provider: 'Ria', reliability: 72, speed: 48, cost: 94, fx: 86, state: 'degraded' as HealthState },
  { provider: 'PAPSS', reliability: 87, speed: 82, cost: 91, fx: 88, state: 'recovery' as HealthState },
]

export const timeline: TimelineStep[] = [
  { label: 'Received', owner: 'Bank channel', status: 'done', time: '14:29:11' },
  { label: 'Validated', owner: 'imsi-rails', status: 'done', time: '14:29:11' },
  { label: 'Route selected', owner: 'Policy engine', status: 'done', time: '14:29:12' },
  { label: 'Provider accepted', owner: 'Thunes', status: 'done', time: '14:29:16' },
  { label: 'Bank rail posting', owner: 'NIP', status: 'current', time: '14:29:23' },
  { label: 'Settlement match', owner: 'Settlement', status: 'pending', time: 'pending' },
]

