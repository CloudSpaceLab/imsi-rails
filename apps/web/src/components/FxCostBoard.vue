<script setup lang="ts">
import HealthBadge from './HealthBadge.vue'
import UiButton from './UiButton.vue'
import type { FxCostBoard } from '../types'

defineProps<{
  board: FxCostBoard
}>()
</script>

<template>
  <section class="panel panel--full" id="fx-costs" aria-labelledby="fx-cost-title">
    <header class="panel__header">
      <div>
        <h2 id="fx-cost-title">FX and cost</h2>
        <p>{{ board.corridor }} / {{ board.pair }} / {{ board.window }}</p>
      </div>
      <UiButton variant="secondary">Refresh</UiButton>
    </header>

    <div class="fx-layout">
      <section class="fx-decision" aria-label="FX route decision">
        <div>
          <span>Use</span>
          <strong>{{ board.recommendedProvider }}</strong>
        </div>
        <div>
          <span>Cheapest</span>
          <strong>{{ board.cheapestProvider }}</strong>
        </div>
        <div>
          <span>Rates updated</span>
          <strong>{{ board.refreshedAt }}</strong>
        </div>
        <p>{{ board.decision }}</p>
      </section>

      <aside class="fx-alert" aria-label="Rate alert">
        <HealthBadge state="stale" />
        <strong>Stale rate</strong>
        <p>{{ board.rateAlert }}</p>
      </aside>

      <div class="fx-table" role="table" aria-label="FX and route cost">
        <div class="fx-table__row fx-table__row--head" role="row">
          <span role="columnheader">Provider</span>
          <span role="columnheader">Rate</span>
          <span role="columnheader">Updated</span>
          <span role="columnheader">Fee</span>
          <span role="columnheader">Spread</span>
          <span role="columnheader">Cost</span>
          <span role="columnheader">Payout</span>
          <span role="columnheader">Call</span>
        </div>

        <div v-for="route in board.routes" :key="route.provider" class="fx-table__row" role="row">
          <span role="cell" data-label="Provider">
            <strong>{{ route.provider }}</strong>
            <small>{{ route.route }}</small>
          </span>
          <span role="cell" data-label="Rate">
            <strong>{{ route.rate }}</strong>
            <small>{{ route.pair }}</small>
          </span>
          <span role="cell" data-label="Updated">{{ route.updatedAt }}</span>
          <span role="cell" data-label="Fee">{{ route.fee }}</span>
          <span role="cell" data-label="Spread">{{ route.spread }}</span>
          <span role="cell" data-label="Cost">
            <strong>{{ route.effectiveCost }}</strong>
          </span>
          <span role="cell" data-label="Payout">{{ route.payoutTime }}</span>
          <span role="cell" data-label="Call" class="fx-table__call">
            <HealthBadge :state="route.state" />
            <strong v-if="route.recommended">Use</strong>
            <strong v-else-if="route.cheapest">Cheapest</strong>
            <span v-else>{{ route.note }}</span>
          </span>
        </div>
      </div>
    </div>
  </section>
</template>
