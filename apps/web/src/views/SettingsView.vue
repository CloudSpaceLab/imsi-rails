<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { FileCheck2, GitBranch, Landmark, ListChecks, Network, SlidersHorizontal, TimerReset, ToggleRight } from '@lucide/vue'
import DataTable from '../components/DataTable.vue'
import HealthBadge from '../components/HealthBadge.vue'
import Panel from '../components/Panel.vue'
import { standingMeterWidth } from '../composables/useInboundTower'
import { useTowerRouting } from '../composables/useTowerRouting'
import type { DashboardMock, FeatureSwitchKey } from '../types'

export type SettingsTab = 'contracts' | 'sla' | 'integrations' | 'switches' | 'audit'

const props = defineProps<{
  dashboard: DashboardMock
  initialTab?: SettingsTab
}>()

const { route, openPath } = useTowerRouting()
const settingsTab = ref<SettingsTab>(normalizeSettingsTab(route.query.tab) ?? props.initialTab ?? 'contracts')
const settingTabs = [
  { id: 'contracts' as SettingsTab, label: 'Contracts', icon: Landmark },
  { id: 'sla' as SettingsTab, label: 'SLA and backoff', icon: TimerReset },
  { id: 'integrations' as SettingsTab, label: 'Integrations', icon: Network },
  { id: 'switches' as SettingsTab, label: 'Feature controls', icon: ToggleRight },
  { id: 'audit' as SettingsTab, label: 'Audit', icon: FileCheck2 },
]

watch(
  () => props.initialTab,
  (tab) => {
    if (tab) settingsTab.value = tab
  },
)

watch(
  () => route.query.tab,
  (tab) => {
    const normalized = normalizeSettingsTab(tab)
    if (normalized) settingsTab.value = normalized
  },
)

function featureEnabled(key: FeatureSwitchKey) {
  return props.dashboard.featureSwitches.find((feature) => feature.key === key)?.enabled ?? false
}

function toggleFeature(key: FeatureSwitchKey) {
  const feature = props.dashboard.featureSwitches.find((item) => item.key === key)
  if (feature) feature.enabled = !feature.enabled
}

const hiddenFeatureCopy = computed(() => ({
  advancedFxCosts: featureEnabled('advancedFxCosts') ? 'Available to administrators' : 'Disabled for operators',
  policySimulator: featureEnabled('policySimulator') ? 'Shadow replay enabled' : 'Disabled for operators',
  providerCommercialScorecards: featureEnabled('providerCommercialScorecards') ? 'Commercial review enabled' : 'Disabled for operators',
}))

function setSettingsTab(tab: SettingsTab) {
  settingsTab.value = tab
  openPath('/settings', { tab })
}

function normalizeSettingsTab(value: unknown): SettingsTab | null {
  return value === 'contracts' || value === 'sla' || value === 'integrations' || value === 'switches' || value === 'audit' ? value : null
}
</script>

<template>
  <section class="screen-stack">
    <Panel title="Settings sections" eyebrow="Operational configuration" accent="healthy">
      <div class="segmented-group settings-tabs">
        <button v-for="tab in settingTabs" :key="tab.id" type="button" :class="{ 'is-selected': settingsTab === tab.id }" @click="setSettingsTab(tab.id)">
          <component :is="tab.icon" :size="16" aria-hidden="true" />
          <strong>{{ tab.label }}</strong>
        </button>
      </div>
    </Panel>

    <section v-if="settingsTab === 'contracts'" class="dashboard-grid">
      <Panel title="Partner contracts and standing accounts" eyebrow="Govern inbound settlement" accent="healthy" class="span-8">
        <DataTable :empty="dashboard.inboundContracts.length === 0" empty-title="No contracts" empty-description="No inbound contracts are configured.">
          <table class="compact-table">
            <thead>
              <tr>
                <th>Contract</th>
                <th>Corridor</th>
                <th>Rails</th>
                <th>Settlement</th>
                <th>Owner</th>
                <th>State</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="contract in dashboard.inboundContracts" :key="contract.id">
                <td data-label="Contract"><strong>{{ contract.partner }}</strong><small class="muted-line mono">{{ contract.id }}</small></td>
                <td data-label="Corridor">{{ contract.corridor }}</td>
                <td data-label="Rails">{{ contract.permittedRails.join(', ') }}</td>
                <td data-label="Settlement">{{ contract.settlementModel }}</td>
                <td data-label="Owner">{{ contract.owner }}</td>
                <td data-label="State"><HealthBadge :state="contract.state" /></td>
              </tr>
            </tbody>
          </table>
        </DataTable>
      </Panel>
      <Panel title="Standing account meters" eyebrow="Prefund and credit line" accent="watch" class="span-4">
        <div class="account-stack">
          <article v-for="account in dashboard.standingAccounts" :key="account.id">
            <strong>{{ account.partner }}</strong>
            <small>{{ account.prefundBalance }} prefund / {{ account.availableLimit }} available</small>
            <div class="progress-track">
              <i :class="`progress-fill--${account.state}`" :style="{ width: standingMeterWidth(account) }"></i>
            </div>
            <small>{{ account.projectedExhaustion }}</small>
          </article>
        </div>
      </Panel>
    </section>

    <section v-else-if="settingsTab === 'sla'" class="dashboard-grid">
      <Panel title="SLA definitions" eyebrow="Credit proof policy" accent="watch" class="span-6">
        <div class="policy-stack">
          <article v-for="policy in dashboard.slaPolicies" :key="policy.id">
            <HealthBadge :state="policy.state" />
            <span>
              <strong>{{ policy.name }}</strong>
              <small>{{ policy.duration }} SLA / {{ policy.cooldownWindow }} cooldown / {{ policy.lateSuccessWindow }} late-success window</small>
            </span>
            <p>{{ policy.evidenceRequired.join(', ') }}</p>
          </article>
        </div>
      </Panel>
      <Panel title="Backoff policy" eyebrow="Route-specific requery" accent="healthy" class="span-6">
        <div class="policy-stack">
          <article v-for="policy in dashboard.backoffPolicies" :key="policy.id">
            <HealthBadge :state="policy.state" />
            <span>
              <strong>{{ policy.route }}</strong>
              <small>{{ policy.maxRequeryAttempts }} attempts / reroute after acceptance: {{ policy.rerouteAfterAcceptance ? 'Allowed' : 'Frozen' }}</small>
            </span>
            <p>{{ policy.schedule.join(' -> ') }}</p>
          </article>
        </div>
      </Panel>
    </section>

    <section v-else-if="settingsTab === 'integrations'" class="dashboard-grid">
      <Panel title="Switching API and provider endpoints" eyebrow="API health windows" accent="healthy" class="span-12">
        <DataTable :empty="dashboard.integrationHealth.length === 0" empty-title="No integration health" empty-description="No API or provider endpoint telemetry is available.">
          <table class="compact-table">
            <thead>
              <tr>
                <th>Service</th>
                <th>Endpoint</th>
                <th>Success</th>
                <th>Timeout</th>
                <th>P95</th>
                <th>Callback</th>
                <th>Action</th>
                <th>State</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in dashboard.integrationHealth" :key="item.id">
                <td data-label="Service"><strong>{{ item.serviceName }}</strong><small class="muted-line">{{ item.provider }} / {{ item.polling }}</small></td>
                <td data-label="Endpoint" class="mono">{{ item.endpoint }}</td>
                <td data-label="Success">{{ item.successRate }}</td>
                <td data-label="Timeout">{{ item.timeoutRate }}</td>
                <td data-label="P95">{{ item.p95Latency }}</td>
                <td data-label="Callback">{{ item.callbackStatus }}<small class="muted-line">{{ item.lastSuccessAt }}</small></td>
                <td data-label="Action">{{ item.nextAction }}</td>
                <td data-label="State"><HealthBadge :state="item.state" /></td>
              </tr>
            </tbody>
          </table>
        </DataTable>
      </Panel>
    </section>

    <section v-else-if="settingsTab === 'switches'" class="dashboard-grid">
      <Panel title="Feature controls" eyebrow="Control rollout scope" accent="healthy" class="span-7">
        <div class="switch-list">
          <label v-for="feature in dashboard.featureSwitches" :key="feature.key" class="toggle-row">
            <input type="checkbox" :checked="feature.enabled" @change="toggleFeature(feature.key)" />
            <span>
              <strong>{{ feature.label }}</strong>
              <small>{{ feature.reason }}</small>
            </span>
            <HealthBadge :state="feature.enabled ? 'recovery' : 'stale'" :trigger="feature.enabled ? 'Enabled' : 'Off by default'" />
          </label>
        </div>
      </Panel>
      <Panel title="Advanced modules" eyebrow="Not in operator navigation" accent="stale" class="span-5">
        <div class="feature-preview">
          <article>
            <SlidersHorizontal :size="18" aria-hidden="true" />
            <strong>Advanced FX costs</strong>
            <small>{{ hiddenFeatureCopy.advancedFxCosts }}</small>
          </article>
          <article>
            <GitBranch :size="18" aria-hidden="true" />
            <strong>Policy simulator</strong>
            <small>{{ hiddenFeatureCopy.policySimulator }}</small>
          </article>
          <article>
            <ListChecks :size="18" aria-hidden="true" />
            <strong>Provider scorecards</strong>
            <small>{{ hiddenFeatureCopy.providerCommercialScorecards }}</small>
          </article>
        </div>
      </Panel>
    </section>

    <section v-else class="dashboard-grid">
      <Panel title="Audit trail and exports" eyebrow="Evidence retention" accent="healthy" class="span-12">
        <DataTable :empty="dashboard.auditEvents.length === 0" empty-title="No audit events" empty-description="No audit evidence is available for this window.">
          <table class="compact-table">
            <thead>
              <tr>
                <th>Time</th>
                <th>Actor</th>
                <th>Action</th>
                <th>Object</th>
                <th>Reason</th>
                <th>State</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="event in dashboard.auditEvents" :key="`${event.time}-${event.object}`">
                <td data-label="Time" class="mono">{{ event.time }}</td>
                <td data-label="Actor">{{ event.actor }}</td>
                <td data-label="Action">{{ event.action }}</td>
                <td data-label="Object">{{ event.object }}</td>
                <td data-label="Reason">{{ event.reason }}</td>
                <td data-label="State"><HealthBadge :state="event.state" /></td>
              </tr>
            </tbody>
          </table>
        </DataTable>
      </Panel>
    </section>
  </section>
</template>
