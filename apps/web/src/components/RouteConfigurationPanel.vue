<script setup lang="ts">
import HealthBadge from './HealthBadge.vue'
import UiButton from './UiButton.vue'
import type { ChangeHistoryItem, FallbackRoute, ProviderToggle, ScoringWeight, TrafficSplitPreset } from '../types'

defineProps<{
  providers: ProviderToggle[]
  fallbackRoutes: FallbackRoute[]
  presets: TrafficSplitPreset[]
  weights: ScoringWeight[]
  impact: {
    successRate: string
    p95: string
    cost: string
  }
  history: ChangeHistoryItem[]
}>()
</script>

<template>
  <section class="panel panel--full" id="routing-policy" aria-labelledby="routing-policy-title">
    <header class="panel__header">
      <div>
        <h2 id="routing-policy-title">Routing rules</h2>
        <p>EU -> Nigeria / bank account</p>
      </div>
      <UiButton>Check</UiButton>
    </header>

    <div class="config-grid">
      <section class="config-block" aria-labelledby="provider-toggles-title">
        <h3 id="provider-toggles-title">Provider status</h3>
        <div class="toggle-list">
          <label v-for="provider in providers" :key="provider.provider" class="toggle-row">
            <input type="checkbox" :checked="provider.enabled" />
            <span>
              <strong>{{ provider.provider }}</strong>
              <small>{{ provider.route }}</small>
            </span>
            <HealthBadge :state="provider.state" />
          </label>
        </div>
      </section>

      <section class="config-block" aria-labelledby="fallback-title">
        <h3 id="fallback-title">Fallback order</h3>
        <ol class="fallback-list">
          <li v-for="route in fallbackRoutes" :key="route.provider">
            <span class="rank">{{ route.rank }}</span>
            <div>
              <strong>{{ route.provider }}</strong>
              <small>{{ route.route }}</small>
            </div>
            <HealthBadge :state="route.state" />
            <div class="rank-actions">
              <UiButton variant="icon" size="sm" :aria-label="`Move ${route.provider} up`" :title="`Move ${route.provider} up`">^</UiButton>
              <UiButton variant="icon" size="sm" :aria-label="`Move ${route.provider} down`" :title="`Move ${route.provider} down`">v</UiButton>
            </div>
          </li>
        </ol>
      </section>

      <section class="config-block" aria-labelledby="split-title">
        <h3 id="split-title">Traffic split</h3>
        <div class="preset-row" role="group" aria-label="Traffic split presets">
          <UiButton v-for="preset in presets" :key="preset.label" variant="choice" :selected="preset.active">
            <strong>{{ preset.label }}</strong>
            <span>{{ preset.split }}</span>
          </UiButton>
        </div>
      </section>

      <section class="config-block" aria-labelledby="weights-title">
        <h3 id="weights-title">Route weights</h3>
        <div class="weight-list">
          <label v-for="weight in weights" :key="weight.label">
            <span>{{ weight.label }}</span>
            <input type="range" min="0" max="60" :value="weight.value" />
            <strong>{{ weight.value }}</strong>
          </label>
        </div>
      </section>

      <section class="config-block" aria-labelledby="impact-title">
        <h3 id="impact-title">Impact check</h3>
        <dl class="impact-grid">
          <div>
            <dt>Success</dt>
            <dd>{{ impact.successRate }}</dd>
          </div>
          <div>
            <dt>P95</dt>
            <dd>{{ impact.p95 }}</dd>
          </div>
          <div>
            <dt>Cost</dt>
            <dd>{{ impact.cost }}</dd>
          </div>
        </dl>
        <label class="reason-field">
          <span>Reason</span>
          <textarea required rows="3">Ria missed the 90s P95 target for 15 minutes.</textarea>
        </label>
      </section>

      <section class="config-block" aria-labelledby="history-title">
        <h3 id="history-title">Recent changes</h3>
        <ol class="history-list">
          <li v-for="item in history" :key="`${item.time}-${item.summary}`">
            <time>{{ item.time }}</time>
            <div>
              <strong>{{ item.actor }}</strong>
              <span>{{ item.summary }}</span>
            </div>
          </li>
        </ol>
      </section>
    </div>
  </section>
</template>
