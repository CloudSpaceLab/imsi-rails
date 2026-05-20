<script setup lang="ts">
import { computed, ref } from 'vue'
import HealthBadge from './HealthBadge.vue'
import UiButton from './UiButton.vue'
import type { PolicyRouteDecision, PolicySimulationSample } from '../types'

const props = defineProps<{
  samples: PolicySimulationSample[]
}>()

const selectedIndex = ref(0)
const runCount = ref(1)
const shadowOnly = ref(true)

const sample = computed(() => props.samples[selectedIndex.value] ?? props.samples[0])
const runReference = computed(() => `test-${String(runCount.value).padStart(3, '0')}`)

const decisionFields = (decision: PolicyRouteDecision) => [
  { label: 'Score', value: String(decision.score) },
  { label: 'P95', value: decision.p95 },
  { label: 'Cost', value: decision.cost },
]

const runSimulation = () => {
  runCount.value += 1
}
</script>

<template>
  <section class="panel panel--full" id="policy-simulator" aria-labelledby="policy-simulator-title">
    <header class="panel__header">
      <div>
        <h2 id="policy-simulator-title">Test a route change</h2>
        <p>{{ runReference }} / {{ sample.reference }}</p>
      </div>
      <div class="button-row">
        <UiButton variant="secondary">Export</UiButton>
        <UiButton @click="runSimulation">Run</UiButton>
      </div>
    </header>

    <div class="simulator-layout">
      <section class="simulator-card simulator-card--sample" aria-labelledby="sample-title">
        <div class="simulator-card__top">
          <h3 id="sample-title">Test transaction</h3>
          <label class="simulator-switch">
            <input v-model="shadowOnly" type="checkbox" />
            <span>{{ shadowOnly ? 'Test only' : 'Can affect live' }}</span>
          </label>
        </div>

        <label class="simulator-field">
          <span>Transaction</span>
          <select v-model.number="selectedIndex" aria-label="Sample transaction">
            <option v-for="(item, index) in samples" :key="item.reference" :value="index">
              {{ item.reference }}
            </option>
          </select>
        </label>

        <dl class="sample-grid">
          <div>
            <dt>Corridor</dt>
            <dd>{{ sample.corridor }}</dd>
          </div>
          <div>
            <dt>Origin</dt>
            <dd>{{ sample.origin }}</dd>
          </div>
          <div>
            <dt>Destination</dt>
            <dd>{{ sample.destination }}</dd>
          </div>
          <div>
            <dt>Amount</dt>
            <dd>{{ sample.amount }}</dd>
          </div>
          <div>
            <dt>Payout</dt>
            <dd>{{ sample.payout }}</dd>
          </div>
          <div>
            <dt>Live traffic</dt>
            <dd>{{ shadowOnly ? 'No change' : 'Approval needed' }}</dd>
          </div>
        </dl>
      </section>

      <section class="simulator-card simulator-card--compare" aria-labelledby="compare-title">
        <h3 id="compare-title">Route result</h3>
        <div class="route-compare">
          <article>
            <div class="route-compare__head">
              <span>Live route</span>
              <HealthBadge :state="sample.current.state" />
            </div>
            <strong>{{ sample.current.route }}</strong>
            <small>{{ sample.current.provider }}</small>
            <dl>
              <div v-for="field in decisionFields(sample.current)" :key="`current-${field.label}`">
                <dt>{{ field.label }}</dt>
                <dd>{{ field.value }}</dd>
              </div>
            </dl>
          </article>

          <article class="route-compare__proposed">
            <div class="route-compare__head">
              <span>Test route</span>
              <HealthBadge :state="sample.proposed.state" />
            </div>
            <strong>{{ sample.proposed.route }}</strong>
            <small>{{ sample.proposed.provider }}</small>
            <dl>
              <div v-for="field in decisionFields(sample.proposed)" :key="`proposed-${field.label}`">
                <dt>{{ field.label }}</dt>
                <dd>{{ field.value }}</dd>
              </div>
            </dl>
          </article>
        </div>
      </section>

      <section class="simulator-card" aria-labelledby="reject-title">
        <h3 id="reject-title">Why other routes lost</h3>
        <ol class="rejection-list">
          <li v-for="route in sample.rejectedRoutes" :key="`${sample.reference}-${route.provider}`">
            <strong>{{ route.provider }}</strong>
            <span>{{ route.route }}</span>
            <p>{{ route.reason }}</p>
          </li>
        </ol>
      </section>

      <section class="simulator-card" aria-labelledby="report-title">
        <div class="simulator-card__top">
          <h3 id="report-title">Test results</h3>
          <span class="report-state">Ready to export</span>
        </div>

        <dl class="report-metrics">
          <div v-for="metric in sample.reportMetrics" :key="metric.label">
            <dt>{{ metric.label }}</dt>
            <dd>{{ metric.value }}</dd>
            <span>{{ metric.detail }}</span>
          </div>
        </dl>

        <div class="report-table" role="table" aria-label="Test route results">
          <div class="report-table__row report-table__row--head" role="row">
            <span role="columnheader">Case</span>
            <span role="columnheader">Live</span>
            <span role="columnheader">Test</span>
            <span role="columnheader">Outcome</span>
          </div>
          <div v-for="row in sample.reportRows" :key="`${sample.reference}-${row.bucket}`" class="report-table__row" role="row">
            <span role="cell" data-label="Case">{{ row.bucket }}</span>
            <span role="cell" data-label="Live">{{ row.currentRoute }}</span>
            <span role="cell" data-label="Test">{{ row.proposedRoute }}</span>
            <strong role="cell" data-label="Outcome">{{ row.result }}</strong>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>
