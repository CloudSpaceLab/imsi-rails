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

