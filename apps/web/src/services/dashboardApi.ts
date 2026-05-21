import { getDashboardMock } from './mockDashboard'
import type { DashboardAnalytics, DashboardSummaryResponse, DashboardTimeseriesPoint } from '../types'

const API_BASE = import.meta.env.VITE_IMSI_API_BASE ?? 'http://127.0.0.1:8080'
const useMock = import.meta.env.MODE === 'test' || import.meta.env.VITE_IMSI_DATA_MODE === 'mock'

export type DashboardQuery = {
  range?: string
  provider_id?: string
  corridor?: string
  payout_method?: string
  currency?: string
  analysis_lens?: string
}

export async function getDashboardSummary(query: DashboardQuery) {
  if (useMock) return mockSummary()
  const response = await fetch(`${API_BASE}/v1/dashboard/summary?${toSearchParams(query)}`, { credentials: 'include' })
  if (!response.ok) throw new Error('Unable to load dashboard summary')
  return (await response.json()) as DashboardSummaryResponse
}

export async function getDashboardTimeseries(query: DashboardQuery) {
  if (useMock) return mockTimeseries()
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

function mockSummary(): DashboardSummaryResponse {
  const dashboard = getDashboardMock()
  const analytics: DashboardAnalytics = {
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
  return {
    context: {
      from: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      to: new Date().toISOString(),
      timezone: 'Africa/Lagos',
      currency: 'USD',
      analysis_lens: 'reliability',
    },
    analytics,
    tiles: [
      { id: 'processed', label: 'Processed', value: '42,618', unit: 'txns', state: 'healthy', trend: 'selected range', drilldown: '/transactions' },
      { id: 'volume', label: 'Volume', value: '18.4M', unit: 'USD', state: 'healthy', trend: 'gross processed value', drilldown: '/transactions?metric=volume' },
      { id: 'sla', label: 'Completed in SLA', value: '95.8', unit: '%', state: 'healthy', trend: '39,218 completed on time', drilldown: '/transactions?timing=Under+QA+policy' },
      { id: 'p95', label: 'P95 credit time', value: dashboard.summary.p95CreditTime, unit: 'sec', state: 'degraded', trend: 'credited transactions', drilldown: '/incidents?focus=latency' },
      { id: 'failed', label: 'Failed/stalled', value: '1,700', unit: 'txns', state: 'degraded', trend: 'needs review', drilldown: '/transactions?timing=Stalled+only' },
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

function mockTimeseries(): DashboardTimeseriesPoint[] {
  return getDashboardMock().visuals.volumeTrend.map((point, index) => ({
    time: new Date(Date.now() - (8 - index) * 60 * 60 * 1000).toISOString(),
    processed_count: point.value,
    volume: point.value * 420,
    sla_rate: getDashboardMock().visuals.completionTrend[index]?.value ?? 95,
    p95_seconds: [42, 48, 51, 88, 96, 144, 258, 180][index] ?? 90,
    state: getDashboardMock().visuals.hourHealth[index]?.state ?? 'healthy',
  }))
}

