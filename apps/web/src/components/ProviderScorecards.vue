<script setup lang="ts">
import HealthBadge from './HealthBadge.vue'
import type { HealthState } from '../types'

defineProps<{
  providers: Array<{
    provider: string
    reliability: number
    speed: number
    cost: number
    fx: number
    state: HealthState
  }>
}>()
</script>

<template>
  <section class="panel" aria-labelledby="provider-scorecards-title">
    <header class="panel__header">
      <h2 id="provider-scorecards-title">Provider Scorecards</h2>
      <span class="window-label">15 min</span>
    </header>
    <div class="provider-grid">
      <article v-for="provider in providers" :key="provider.provider" class="provider-card">
        <div class="provider-card__top">
          <strong>{{ provider.provider }}</strong>
          <HealthBadge :state="provider.state" />
        </div>
        <dl>
          <div>
            <dt>Reliability</dt>
            <dd>{{ provider.reliability }}</dd>
          </div>
          <div>
            <dt>Speed</dt>
            <dd>{{ provider.speed }}</dd>
          </div>
          <div>
            <dt>Cost</dt>
            <dd>{{ provider.cost }}</dd>
          </div>
          <div>
            <dt>FX</dt>
            <dd>{{ provider.fx }}</dd>
          </div>
        </dl>
      </article>
    </div>
  </section>
</template>

