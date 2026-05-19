<script setup lang="ts">
import HealthBadge from './HealthBadge.vue'
import RouteScoreChip from './RouteScoreChip.vue'
import type { HealthState } from '../types'

defineProps<{
  rows: Array<{
    corridor: string
    payout: string
    state: HealthState
    selectedRoute: string
    score: number
    p95: string
    cost: string
    split: string
    recommendation: string
  }>
}>()
</script>

<template>
  <div class="matrix" role="table" aria-label="Corridor health matrix">
    <div class="matrix__row matrix__row--head" role="row">
      <span role="columnheader">Corridor</span>
      <span role="columnheader">Route</span>
      <span role="columnheader">Score</span>
      <span role="columnheader">P95</span>
      <span role="columnheader">Cost</span>
      <span role="columnheader">Split</span>
      <span role="columnheader">Action</span>
    </div>
    <div v-for="row in rows" :key="row.corridor" class="matrix__row" role="row">
      <span role="cell" data-label="Corridor">
        <strong>{{ row.corridor }}</strong>
        <small>{{ row.payout }}</small>
      </span>
      <span role="cell" class="matrix__route" data-label="Route">
        <HealthBadge :state="row.state" window="15 min" />
        {{ row.selectedRoute }}
      </span>
      <span role="cell" data-label="Score"><RouteScoreChip :score="row.score" :reason="row.recommendation" /></span>
      <span role="cell" data-label="P95">{{ row.p95 }}</span>
      <span role="cell" data-label="Cost">{{ row.cost }}</span>
      <span role="cell" data-label="Split">{{ row.split }}</span>
      <span role="cell" data-label="Action">
        <button type="button" class="text-action">{{ row.recommendation }}</button>
      </span>
    </div>
  </div>
</template>
