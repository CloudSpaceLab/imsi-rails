import { getDashboardMock } from './mockDashboard'
import type { DashboardAnalytics, DashboardSummaryResponse, DashboardTimeseriesPoint, HealthState, UiScenario } from '../types'

const API_BASE = import.meta.env.VITE_IMSI_API_BASE ?? 'http://127.0.0.1:8080'
const useMock = import.meta.env.MODE === 'test' || import.meta.env.VITE_IMSI_DATA_MODE === 'mock'

export type DashboardQuery = {
  range?: string
  provider_id?: string
  corridor?: string
  payout_method?: string
  currency?: string
  analysis_lens?: string
  scenario?: UiScenario
}

export async function getDashboardSummary(query: DashboardQuery) {
  if (useMock) return mockSummary(query)
  const response = await fetch(`${API_BASE}/v1/dashboard/summary?${toSearchParams(query)}`, { credentials: 'include' })
  if (!response.ok) throw new Error('Unable to load dashboard summary')
  return (await response.json()) as DashboardSummaryResponse
}

export async function getDashboardTimeseries(query: DashboardQuery) {
  if (useMock) return mockTimeseries(query)
  const response = await fetch(`${API_BASE}/v1/dashboard/timeseries?${toSearchParams(query)}`, { credentials: 'include' })
  if (!response.ok) throw new Error('Unable to load dashboard timeseries')
  const payload = (await response.json()) as { points: DashboardTimeseriesPoint[] }
  return payload.points
}

export function connectDashboardLive(query: DashboardQuery, onSummary: (summary: DashboardSummaryResponse) => void) {
  if (useMock || typeof EventSource === 'undefined') return () => {}
  const source = new EventSource(`${API_BASE}/v1/dashboard/live?${toSearchParams(query)}`, { withCredentials: true })
  source.addEventListener('dashboard.summary', (event) => {
    onSummary(JSON.parse((event as MessageEvent).data) as DashboardSummaryResponse)
  })
  return () => source.close()
}

function toSearchParams(query: DashboardQuery) {
  const params = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => {
    if (value) params.set(key, value)
  })
  return params
}

function mockSummary(query: DashboardQuery = {}): DashboardSummaryResponse {
  const scenario = query.scenario ?? 'degraded-ria'
  const dashboard = getDashboardMock(scenario)
  const analytics = mockAnalyticsForScenario(scenario)
  const displayCurrency = query.currency ?? 'USD'
  const p95 = formatDuration(analytics.p95_seconds)
  const p99 = formatDuration(analytics.p99_seconds)
  const failedAndStalled = analytics.failed_count + analytics.stalled_count
  const trafficShifted = scenario === 'traffic-shift' ? '25%' : scenario === 'healthy' ? '0%' : scenario === 'pilot-report' ? '18%' : 'candidate 25%'
  const failuresAvoided = scenario === 'traffic-shift' ? '312' : scenario === 'pilot-report' ? '1,184' : scenario === 'healthy' ? '0' : '312 est.'
  const shiftState: HealthState = scenario === 'traffic-shift' || scenario === 'pilot-report' ? 'recovery' : scenario === 'healthy' ? 'healthy' : 'watch'
  return {
    context: {
      from: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      to: new Date().toISOString(),
      timezone: 'Africa/Lagos',
      currency: displayCurrency,
      analysis_lens: query.analysis_lens ?? 'reliability',
      provider_id: query.provider_id,
      corridor: query.corridor,
      payout_method: query.payout_method,
      scenario,
    },
    analytics,
    tiles: [
      { id: 'value-risk', label: 'Value at risk', value: convertUsdLabel(dashboard.summary.atRiskValue, displayCurrency), unit: '', state: dashboard.summary.stuckTransactions ? 'watch' : 'healthy', trend: `${dashboard.summary.stuckTransactions} stuck transfers`, drilldown: '/transactions?timing=Stalled+only' },
      { id: 'sla', label: 'SLA completion', value: analytics.sla_rate.toFixed(1), unit: '%', state: analytics.sla_rate >= 95 ? 'healthy' : 'watch', trend: `${analytics.sla_completed_count.toLocaleString()} completed on time`, drilldown: '/transactions?timing=Under+QA+policy' },
      { id: 'tail-latency', label: 'P95 / P99 credit', value: `${p95} / ${p99}`, unit: '', state: analytics.p95_seconds > 90 ? 'degraded' : 'healthy', trend: 'time-to-credit tail', drilldown: '/incidents?focus=latency' },
      { id: 'failed', label: 'Stuck or failed', value: failedAndStalled.toLocaleString(), unit: '', state: failedAndStalled ? 'degraded' : 'healthy', trend: 'needs operations review', drilldown: '/transactions?timing=Stalled+only' },
      { id: 'switching', label: 'Failures avoided', value: failuresAvoided, unit: '', state: shiftState, trend: `traffic shifted ${trafficShifted}`, drilldown: '/routes?focus=switching' },
    ],
    providers: dashboard.providerScores.map((provider) => ({
      provider_id: provider.provider.toLowerCase(),
      provider_name: provider.provider,
      corridor: provider.corridor,
      processed_count: Number(provider.trafficShare.replace('%', '')) * 100,
      processed_volume: 1000000,
      sla_completed_count: 900,
      sla_rate: Number(provider.successRate.replace('%', '')),
      p95_seconds: provider.p95.includes('m') ? 258 : Number(provider.p95.replace(/\D/g, '')),
      state: provider.state,
    })),
    generated_at: new Date().toISOString(),
  }
}

function mockTimeseries(query: DashboardQuery = {}): DashboardTimeseriesPoint[] {
  const dashboard = getDashboardMock(query.scenario ?? 'degraded-ria')
  return dashboard.visuals.volumeTrend.map((point, index) => ({
    time: new Date(Date.now() - (8 - index) * 60 * 60 * 1000).toISOString(),
    processed_count: point.value,
    volume: point.value * 420,
    sla_rate: dashboard.visuals.completionTrend[index]?.value ?? 95,
    p95_seconds: [42, 48, 51, 88, 96, 144, 258, 180][index] ?? 90,
    state: dashboard.visuals.hourHealth[index]?.state ?? 'healthy',
  }))
}

function mockAnalyticsForScenario(scenario: UiScenario): DashboardAnalytics {
  if (scenario === 'healthy') {
    return {
      processed_count: 38492,
      processed_volume: 16250000,
      completed_count: 38210,
      sla_completed_count: 37598,
      sla_rate: 98.4,
      failed_count: 112,
      stalled_count: 18,
      pending_count: 152,
      p50_seconds: 31,
      p95_seconds: 68,
      p99_seconds: 104,
    }
  }
  if (scenario === 'traffic-shift') {
    return {
      processed_count: 43104,
      processed_volume: 18980000,
      completed_count: 41736,
      sla_completed_count: 40271,
      sla_rate: 96.5,
      failed_count: 642,
      stalled_count: 226,
      pending_count: 500,
      p50_seconds: 36,
      p95_seconds: 132,
      p99_seconds: 246,
    }
  }
  if (scenario === 'pilot-report') {
    return {
      processed_count: 286420,
      processed_volume: 118700000,
      completed_count: 279840,
      sla_completed_count: 271725,
      sla_rate: 97.1,
      failed_count: 2706,
      stalled_count: 910,
      pending_count: 2964,
      p50_seconds: 34,
      p95_seconds: 121,
      p99_seconds: 231,
    }
  }
  return {
    processed_count: 42618,
    processed_volume: 18400000,
    completed_count: 40918,
    sla_completed_count: 39218,
    sla_rate: 95.8,
    failed_count: 1284,
    stalled_count: 416,
    pending_count: 0,
    p50_seconds: 38,
    p95_seconds: 258,
    p99_seconds: 464,
  }
}

function formatDuration(seconds: number) {
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return remainingSeconds ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`
}

function convertUsdLabel(label: string, currency: string) {
  const value = Number(label.replace(/[^\d.]/g, ''))
  if (!Number.isFinite(value)) return label
  const multiplier = label.toUpperCase().includes('M') ? 1_000_000 : label.toUpperCase().includes('K') ? 1_000 : 1
  const usd = value * multiplier
  const rates: Record<string, number> = {
    USD: 1,
    NGN: 1560,
    EUR: 0.92,
    GBP: 0.78,
    KES: 129,
  }
  const converted = usd * (rates[currency] ?? 1)
  if (converted >= 1_000_000_000) return `${currency} ${(converted / 1_000_000_000).toFixed(1)}B`
  if (converted >= 1_000_000) return `${currency} ${(converted / 1_000_000).toFixed(1)}M`
  if (converted >= 1_000) return `${currency} ${(converted / 1_000).toFixed(0)}K`
  return `${currency} ${converted.toFixed(0)}`
}

