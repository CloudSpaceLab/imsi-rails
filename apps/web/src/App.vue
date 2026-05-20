<script setup lang="ts">
import CorridorMatrix from './components/CorridorMatrix.vue'
import DataFreshness from './components/DataFreshness.vue'
import DowntimeTimeline from './components/DowntimeTimeline.vue'
import FxCostBoard from './components/FxCostBoard.vue'
import HealthBadge from './components/HealthBadge.vue'
import LatencyWaterfall from './components/LatencyWaterfall.vue'
import PolicySimulatorPanel from './components/PolicySimulatorPanel.vue'
import ProviderScorecards from './components/ProviderScorecards.vue'
import RouteConfigurationPanel from './components/RouteConfigurationPanel.vue'
import TransactionTimeline from './components/TransactionTimeline.vue'
import UiButton from './components/UiButton.vue'
import {
  changeHistory,
  corridors,
  downtimeEvents,
  drilldownFilters,
  fallbackRoutes,
  fxCostBoard,
  latencySteps,
  latencySummary,
  policySimulationSamples,
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
  <div class="app-shell" data-bank-theme="imsi">
    <aside class="sidebar" aria-label="Product navigation">
      <a class="brand" href="/" aria-label="imsi-rails">
        <span class="brand__mark" aria-hidden="true"></span>
        <span>imsi-rails</span>
      </a>
      <nav>
        <a class="is-active" href="#control-room">Live view</a>
        <a href="#corridors">Corridors</a>
        <a href="#transactions">Transactions</a>
        <a href="#routing-policy">Rules</a>
        <a href="#policy-simulator">Test</a>
        <a href="#fx-costs">FX/Cost</a>
        <a href="#audit">Audit</a>
      </nav>
    </aside>

    <main class="workspace" id="control-room">
      <header class="topbar">
        <div>
          <p class="eyebrow">Live routing</p>
          <h1>Operations</h1>
        </div>
        <DataFreshness :updated="summary.lastUpdated" />
      </header>

      <section class="summary-grid" aria-label="Operational summary">
        <article class="summary-card summary-card--primary">
          <span>Routes healthy</span>
          <strong>{{ summary.globalHealth }}</strong>
          <HealthBadge state="healthy" window="15 min" :updated="summary.lastUpdated" />
        </article>
        <article class="summary-card">
          <span>Value today</span>
          <strong>{{ summary.valueToday }}</strong>
          <small>{{ summary.transactionsToday }} transactions</small>
        </article>
        <article class="summary-card">
          <span>P95 credit time</span>
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
              <h2 id="corridors-title">Route board</h2>
              <p>15 min window</p>
            </div>
            <UiButton>{{ operationalActions.shiftTraffic }}</UiButton>
          </header>
          <CorridorMatrix :rows="corridors" />
        </section>

        <aside class="panel recommendation-panel" aria-labelledby="recommendation-title">
          <header class="panel__header">
            <h2 id="recommendation-title">Action needed</h2>
            <HealthBadge state="degraded" window="15 min" />
          </header>
          <div class="recommendation">
            <strong>Move 25% of EU -> Nigeria account payouts off Ria.</strong>
            <p>Ria missed the 90s P95 target for 15 min. Thunes is available and has fresh FX.</p>
            <div class="button-row">
              <UiButton>{{ operationalActions.previewPolicy }}</UiButton>
              <UiButton variant="secondary">{{ operationalActions.exportEvidence }}</UiButton>
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

        <PolicySimulatorPanel :samples="policySimulationSamples" />

        <FxCostBoard :board="fxCostBoard" />

        <LatencyWaterfall :filters="drilldownFilters" :summary="latencySummary" :steps="latencySteps" />

        <DowntimeTimeline :events="downtimeEvents" />

        <ProviderScorecards :providers="providerScores" />

        <section class="panel" id="transactions" aria-labelledby="trace-title">
          <header class="panel__header">
            <div>
              <h2 id="trace-title">Transaction trail</h2>
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
