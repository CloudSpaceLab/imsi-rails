<script setup lang="ts">
import HealthBadge from './HealthBadge.vue'
import type { HealthState } from '../types'

defineProps<{
  providers: Array<{
    provider: string
    corridor: string
    successRate: string
    p50: string
    p95: string
    p99: string
    stuckRate: string
    settlementExceptions: number
    state: HealthState
  }>
}>()
</script>

<template>
  <section class="panel" aria-labelledby="provider-scorecards-title">
    <header class="panel__header">
      <div>
        <h2 id="provider-scorecards-title">Provider performance</h2>
        <p>15 min provider view</p>
      </div>
      <label class="window-select">
        <span>Window</span>
        <select value="15 min" aria-label="Scorecard time window">
          <option>15 min</option>
          <option>1 hour</option>
          <option>24 hours</option>
        </select>
      </label>
    </header>
    <div class="provider-grid">
      <article v-for="provider in providers" :key="provider.provider" class="provider-card">
        <div class="provider-card__top">
          <div>
            <strong>{{ provider.provider }}</strong>
            <span class="provider-card__corridor">{{ provider.corridor }}</span>
          </div>
          <HealthBadge :state="provider.state" />
        </div>
        <dl>
          <div>
            <dt>Success</dt>
            <dd>{{ provider.successRate }}</dd>
          </div>
          <div>
            <dt>P50</dt>
            <dd>{{ provider.p50 }}</dd>
          </div>
          <div>
            <dt>P95</dt>
            <dd>{{ provider.p95 }}</dd>
          </div>
          <div>
            <dt>P99</dt>
            <dd>{{ provider.p99 }}</dd>
          </div>
          <div>
            <dt>Stuck</dt>
            <dd>{{ provider.stuckRate }}</dd>
          </div>
          <div>
            <dt>Exceptions</dt>
            <dd>{{ provider.settlementExceptions }}</dd>
          </div>
        </dl>
      </article>
    </div>
  </section>
</template>

