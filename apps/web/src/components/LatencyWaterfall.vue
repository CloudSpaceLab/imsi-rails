<script setup lang="ts">
import HealthBadge from './HealthBadge.vue'
import type { HealthState, LatencyStep } from '../types'

defineProps<{
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
}>()

const widthFor = (step: LatencyStep) => `${Math.min(100, Math.max(8, (step.durationMs / Math.max(step.targetMs, 1)) * 42))}%`
const seconds = (durationMs: number) => {
  if (durationMs >= 60_000) {
    return `${Math.round(durationMs / 1_000)}s`
  }
  return `${(durationMs / 1_000).toFixed(1)}s`
}

const badgeState = (state: HealthState) => state
</script>

<template>
  <section class="panel panel--wide" id="latency" aria-labelledby="latency-title">
    <header class="panel__header">
      <div>
        <h2 id="latency-title">Latency Waterfall</h2>
        <p>{{ filters.provider }} / {{ filters.corridor }} / {{ filters.destinationBank }} / {{ filters.window }}</p>
      </div>
      <HealthBadge state="degraded" window="15 min" />
    </header>

    <div class="filter-bar" aria-label="Latency filters">
      <label>
        <span>Provider</span>
        <select :value="filters.provider" aria-label="Provider">
          <option>{{ filters.provider }}</option>
          <option>Thunes</option>
          <option>Remitly</option>
        </select>
      </label>
      <label>
        <span>Corridor</span>
        <select :value="filters.corridor" aria-label="Corridor">
          <option>{{ filters.corridor }}</option>
          <option>US -> Nigeria</option>
          <option>UK -> Nigeria</option>
        </select>
      </label>
      <label>
        <span>Bank</span>
        <select :value="filters.destinationBank" aria-label="Destination bank">
          <option>{{ filters.destinationBank }}</option>
          <option>GTBank</option>
          <option>UBA</option>
        </select>
      </label>
      <label>
        <span>Window</span>
        <select :value="filters.window" aria-label="Time window">
          <option>{{ filters.window }}</option>
          <option>1 hour</option>
          <option>24 hours</option>
        </select>
      </label>
    </div>

    <div class="latency-summary" aria-label="Latency summary">
      <div>
        <span>End-to-end</span>
        <strong>{{ summary.endToEnd }}</strong>
      </div>
      <div>
        <span>Target</span>
        <strong>{{ summary.target }}</strong>
      </div>
      <div>
        <span>Slowest step</span>
        <strong>{{ summary.slowestStep }}</strong>
      </div>
      <div>
        <span>Affected</span>
        <strong>{{ summary.affectedTransactions }}</strong>
      </div>
    </div>

    <ol class="waterfall" aria-label="Step latency waterfall">
      <li v-for="step in steps" :key="step.label" class="waterfall__row">
        <div>
          <strong>{{ step.label }}</strong>
          <span>{{ step.owner }}</span>
        </div>
        <div class="waterfall__track" :aria-label="`${step.label} latency ${seconds(step.durationMs)}`">
          <span class="waterfall__bar" :class="`waterfall__bar--${badgeState(step.state)}`" :style="{ width: widthFor(step) }"></span>
        </div>
        <strong>{{ seconds(step.durationMs) }}</strong>
        <small>target {{ seconds(step.targetMs) }}</small>
      </li>
    </ol>
  </section>
</template>
