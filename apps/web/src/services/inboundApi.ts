import type { ComplianceHold, CreditLeg, InboundSlaRow, RoutingContract } from '../types'

const API_BASE = import.meta.env.VITE_IMSI_API_BASE ?? 'http://127.0.0.1:8080'
const useMock = import.meta.env.MODE === 'test' || import.meta.env.VITE_IMSI_DATA_MODE === 'mock'

// inboundApiEnabled is true when the web app should fetch inbound data from the
// Go API. In mock/demo/test mode the app keeps using the synchronous fixtures.
export const inboundApiEnabled = !useMock

export type InboundSummary = {
  contractCount: number
  creditsInFlight: number
  reversals: number
  reconBreaks: number
  complianceHolds: number
}

export type CreditLegQuery = {
  state?: string
  contract_id?: string
  q?: string
}

export type CreditActionKind = 'attach_evidence' | 'resolve_recon' | 'manual_requery' | 'mark_completed_outside_platform' | 'approve_reversal' | 'close_case'

export type CreditActionPayload = {
  action: CreditActionKind
  note: string
  evidenceReference: string
  reasonCode: string
  requeryMethod?: string
  customerValueDelivered?: boolean
}

export type InboundAuditEvent = {
  time: string
  actor: string
  action: CreditActionKind | string
  reference: string
  object: string
  result: string
  note?: string
  evidenceReference?: string
  reasonCode?: string
  requeryMethod?: string
  state?: string
  makerCheckerStatus?: string
  customerValueDelivered: boolean
}

// applyCreditAction posts an operator intervention to the API. In mock/demo
// mode it returns null so the caller applies the transition client-side.
export async function applyCreditAction(reference: string, payload: CreditActionPayload): Promise<CreditLeg | null> {
  if (!inboundApiEnabled) return null
  const response = await fetch(`${API_BASE}/v1/inbound/credits/${encodeURIComponent(reference)}/actions`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      action: payload.action,
      note: payload.note,
      evidence_reference: payload.evidenceReference,
      reason_code: payload.reasonCode,
      requery_method: payload.requeryMethod,
      customer_value_delivered: payload.customerValueDelivered ?? false,
    }),
  })
  if (!response.ok) throw new Error(`Action ${payload.action} failed (${response.status})`)
  return camelize<CreditLeg>(await response.json())
}

export async function getInboundSummary(): Promise<InboundSummary> {
  const payload = await getJSON<Record<string, unknown>>('/v1/inbound/summary')
  return camelize<InboundSummary>(payload)
}

export async function getInboundContracts(): Promise<RoutingContract[]> {
  const payload = await getJSON<{ contracts: unknown[] }>('/v1/inbound/contracts')
  return (payload.contracts ?? []).map((item) => camelize<RoutingContract>(item))
}

export async function getInboundCredits(query: CreditLegQuery = {}): Promise<CreditLeg[]> {
  const payload = await getJSON<{ credit_legs: unknown[] }>(`/v1/inbound/credits?${toSearchParams(query)}`)
  return (payload.credit_legs ?? []).map((item) => camelize<CreditLeg>(item))
}

export async function getInboundSla(): Promise<InboundSlaRow[]> {
  const payload = await getJSON<{ rows: unknown[] }>('/v1/inbound/sla')
  return (payload.rows ?? []).map((item) => camelize<InboundSlaRow>(item))
}

export async function getInboundComplianceHolds(): Promise<ComplianceHold[]> {
  const payload = await getJSON<{ holds: unknown[] }>('/v1/inbound/compliance-holds')
  return (payload.holds ?? []).map((item) => camelize<ComplianceHold>(item))
}

export async function getInboundAudit(): Promise<InboundAuditEvent[]> {
  const payload = await getJSON<{ events: unknown[] }>('/v1/inbound/audit')
  return (payload.events ?? []).map((item) => camelize<InboundAuditEvent>(item))
}

async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, { credentials: 'include' })
  if (!response.ok) throw new Error(`Unable to load ${path}`)
  return (await response.json()) as T
}

function toSearchParams(query: CreditLegQuery) {
  const params = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => {
    if (value) params.set(key, value)
  })
  return params
}

// camelize recursively converts snake_case object keys (the Go API JSON
// convention) into the camelCase keys the web types use.
function camelize<T>(value: unknown): T {
  if (Array.isArray(value)) {
    return value.map((item) => camelize(item)) as unknown as T
  }
  if (value && typeof value === 'object') {
    const out: Record<string, unknown> = {}
    for (const [key, val] of Object.entries(value as Record<string, unknown>)) {
      const camelKey = key.replace(/_([a-z0-9])/g, (_, char: string) => char.toUpperCase())
      out[camelKey] = camelize(val)
    }
    return out as T
  }
  return value as T
}
