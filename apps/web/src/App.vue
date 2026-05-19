<script setup lang="ts">
import CorridorMatrix from './components/CorridorMatrix.vue'
import DataFreshness from './components/DataFreshness.vue'
import DowntimeTimeline from './components/DowntimeTimeline.vue'
import HealthBadge from './components/HealthBadge.vue'
import LatencyWaterfall from './components/LatencyWaterfall.vue'
import ProviderScorecards from './components/ProviderScorecards.vue'
import RouteConfigurationPanel from './components/RouteConfigurationPanel.vue'
import TransactionTimeline from './components/TransactionTimeline.vue'
import {
  changeHistory,
  corridors,
  downtimeEvents,
  drilldownFilters,
  fallbackRoutes,
  latencySteps,
  latencySummary,
  providerScores,
  providerToggles,
  routeConfigImpact,
  scoringWeights,
  summary,
  timeline,
  trafficSplitPresets,
} from './data'
import { operationalActions } from './copy'
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar" aria-label="Product navigation">
      <a class="brand" href="/" aria-label="imsi-rails">
        <span class="brand__mark" aria-hidden="true"></span>
        <span>imsi-rails</span>
      </a>
      <nav>
        <a class="is-active" href="#control-room">Control Room</a>
        <a href="#corridors">Corridors</a>
        <a href="#transactions">Transactions</a>
        <a href="#routing-policy">Routing Policy</a>
        <a href="#fx-costs">FX & Costs</a>
        <a href="#audit">Audit</a>
      </nav>
    </aside>

    <main class="workspace" id="control-room">
      <header class="topbar">
        <div>
          <p class="eyebrow">International Transfer Reliability</p>
          <h1>Control Room</h1>
        </div>
        <DataFreshness :updated="summary.lastUpdated" />
      </header>

      <section class="summary-grid" aria-label="Operational summary">
        <article class="summary-card summary-card--primary">
          <span>Global health</span>
          <strong>{{ summary.globalHealth }}</strong>
          <HealthBadge state="healthy" window="15 min" :updated="summary.lastUpdated" />
        </article>
        <article class="summary-card">
          <span>Value today</span>
          <strong>{{ summary.valueToday }}</strong>
          <small>{{ summary.transactionsToday }} transactions</small>
        </article>
        <article class="summary-card">
          <span>P95 time-to-credit</span>
          <strong>{{ summary.p95CreditTime }}</strong>
          <small>account payouts</small>
        </article>
        <article class="summary-card">
          <span>Stuck transactions</span>
          <strong>{{ summary.stuckTransactions }}</strong>
          <small>{{ summary.activeIncidents }} active incidents</small>
        </article>
      </section>

      <section class="content-grid">
        <section class="panel panel--wide" id="corridors" aria-labelledby="corridors-title">
          <header class="panel__header">
            <div>
              <h2 id="corridors-title">Corridor Matrix</h2>
              <p>15 min window</p>
            </div>
            <button type="button" class="action-button">{{ operationalActions.shiftTraffic }}</button>
          </header>
          <CorridorMatrix :rows="corridors" />
        </section>

        <aside class="panel recommendation-panel" aria-labelledby="recommendation-title">
          <header class="panel__header">
            <h2 id="recommendation-title">Recommended Action</h2>
            <HealthBadge state="degraded" window="15 min" />
          </header>
          <div class="recommendation">
            <strong>Shift 25% EU -> Nigeria account traffic away from Ria.</strong>
            <p>P95 time-to-credit breached policy for 15 minutes. Thunes is the next eligible route with higher reliability and fresh FX.</p>
            <div class="button-row">
              <button type="button" class="action-button">{{ operationalActions.previewPolicy }}</button>
              <button type="button" class="ghost-button">{{ operationalActions.exportEvidence }}</button>
            </div>
          </div>
        </aside>

        <RouteConfigurationPanel
          :providers="providerToggles"
          :fallback-routes="fallbackRoutes"
          :presets="trafficSplitPresets"
          :weights="scoringWeights"
          :impact="routeConfigImpact"
          :history="changeHistory"
        />

        <LatencyWaterfall :filters="drilldownFilters" :summary="latencySummary" :steps="latencySteps" />

        <DowntimeTimeline :events="downtimeEvents" />

        <ProviderScorecards :providers="providerScores" />

        <section class="panel" id="transactions" aria-labelledby="trace-title">
          <header class="panel__header">
            <div>
              <h2 id="trace-title">Transaction Trace</h2>
              <p>IMSI-txn_000000000001</p>
            </div>
            <HealthBadge state="watch" window="current" />
          </header>
          <TransactionTimeline :steps="timeline" />
        </section>
      </section>
    </main>
  </div>
</template>
