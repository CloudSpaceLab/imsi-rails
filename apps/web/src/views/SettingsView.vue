<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { FileCheck2, GitBranch, KeyRound, Landmark, ListChecks, Network, Power, SlidersHorizontal, TimerReset, ToggleRight, UserCircle } from '@lucide/vue'
import DataTable from '../components/DataTable.vue'
import HealthBadge from '../components/HealthBadge.vue'
import Panel from '../components/Panel.vue'
import UiButton from '../components/UiButton.vue'
import { standingMeterWidth } from '../composables/useInboundTower'
import { useTowerRouting } from '../composables/useTowerRouting'
import { changePassword } from '../services/authApi'
import type { DashboardMock, FeatureSwitchKey, HealthState, RoutePenalty, SessionUser, SlaPolicy } from '../types'

export type SettingsTab = 'contracts' | 'sla' | 'integrations' | 'switches' | 'audit' | 'profile'

const props = defineProps<{
  dashboard: DashboardMock
  initialTab?: SettingsTab
  sessionUser?: SessionUser | null
}>()

const { route, openPath } = useTowerRouting()
const settingsTab = ref<SettingsTab>(normalizeSettingsTab(route.query.tab) ?? props.initialTab ?? 'contracts')
const settingTabs = [
  { id: 'contracts' as SettingsTab, label: 'Contracts', icon: Landmark },
  { id: 'sla' as SettingsTab, label: 'SLA and backoff', icon: TimerReset },
  { id: 'integrations' as SettingsTab, label: 'Integrations', icon: Network },
  { id: 'switches' as SettingsTab, label: 'Feature controls', icon: ToggleRight },
  { id: 'audit' as SettingsTab, label: 'Audit', icon: FileCheck2 },
  { id: 'profile' as SettingsTab, label: 'Profile', icon: UserCircle },
]
const selectedSlaId = ref(props.dashboard.slaPolicies[0]?.id ?? '')
const slaDraft = ref({
  duration: props.dashboard.slaPolicies[0]?.duration ?? '',
  cooldownWindow: props.dashboard.slaPolicies[0]?.cooldownWindow ?? '',
  lateSuccessWindow: props.dashboard.slaPolicies[0]?.lateSuccessWindow ?? '',
})
const slaMessage = ref('')
const passwordForm = ref({ current: '', next: '', confirm: '' })
const passwordBusy = ref(false)
const passwordMessage = ref('')
const passwordError = ref('')

const selectedSla = computed(() => props.dashboard.slaPolicies.find((policy) => policy.id === selectedSlaId.value) ?? props.dashboard.slaPolicies[0] ?? null)
const userRoles = computed(() => props.sessionUser?.roles.join(', ') || 'No role assigned')
const permissionSummary = computed(() => `${props.sessionUser?.permissions.length ?? 0} permissions`)
const providerControlRows = computed(() =>
  [...props.dashboard.routePenalties]
    .filter((routeRow) => routeRow.route !== 'Fidelity core')
    .sort((a, b) => a.penaltyScore - b.penaltyScore),
)

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
  return value === 'contracts' || value === 'sla' || value === 'integrations' || value === 'switches' || value === 'audit' || value === 'profile'
    ? value
    : null
}

function syncSlaDraft(policy: SlaPolicy | null) {
  slaDraft.value = {
    duration: policy?.duration ?? '',
    cooldownWindow: policy?.cooldownWindow ?? '',
    lateSuccessWindow: policy?.lateSuccessWindow ?? '',
  }
}

function appendAudit(action: string, object: string, reason: string, state: HealthState = 'recovery') {
  props.dashboard.auditEvents.unshift({
    time: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    actor: props.sessionUser?.display_name ?? 'Operations Admin',
    action,
    object,
    reason,
    state,
  })
}

function saveSlaPolicy() {
  const policy = selectedSla.value
  if (!policy) return
  policy.duration = slaDraft.value.duration.trim() || policy.duration
  policy.cooldownWindow = slaDraft.value.cooldownWindow.trim() || policy.cooldownWindow
  policy.lateSuccessWindow = slaDraft.value.lateSuccessWindow.trim() || policy.lateSuccessWindow
  policy.state = 'recovery'
  slaMessage.value = `${policy.name} updated. Audit event captured.`
  appendAudit('SLA policy updated', policy.id, `${policy.duration} SLA / ${policy.cooldownWindow} cooldown / ${policy.lateSuccessWindow} late-success window`)
}

async function submitPasswordChange() {
  passwordError.value = ''
  passwordMessage.value = ''
  if (!passwordForm.value.current || !passwordForm.value.next || !passwordForm.value.confirm) {
    passwordError.value = 'Current password, new password, and confirmation are required.'
    return
  }
  if (passwordForm.value.next.length < 12) {
    passwordError.value = 'New password must be at least 12 characters.'
    return
  }
  if (passwordForm.value.next !== passwordForm.value.confirm) {
    passwordError.value = 'New password and confirmation do not match.'
    return
  }
  passwordBusy.value = true
  try {
    await changePassword(passwordForm.value.current, passwordForm.value.next)
    passwordForm.value = { current: '', next: '', confirm: '' }
    passwordMessage.value = 'Password updated. Audit event captured.'
    appendAudit('Password changed', props.sessionUser?.username ?? 'current user', 'User changed own password from profile settings')
  } catch (error) {
    passwordError.value = error instanceof Error ? error.message : 'Password change failed'
  } finally {
    passwordBusy.value = false
  }
}

function toggleProvider(routeRow: RoutePenalty) {
  const disabling = routeRow.state !== 'blocked'
  routeRow.state = disabling ? 'blocked' : 'watch'
  routeRow.penaltyReason = disabling ? 'Provider deactivated for new eligible traffic by administrator.' : 'Provider reactivated for controlled new traffic.'
  routeRow.trafficSplit = disabling ? '0% new eligible traffic' : 'Controlled recovery traffic only'
  appendAudit(
    disabling ? 'Provider deactivated' : 'Provider reactivated',
    routeRow.route,
    disabling ? 'New eligible traffic blocked from this provider in Settings' : 'Provider reopened for controlled new eligible traffic in Settings',
    disabling ? 'degraded' : 'recovery',
  )
}

watch(selectedSla, (policy) => syncSlaDraft(policy), { immediate: true })
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
      <Panel title="Partner contracts and standing accounts" eyebrow="Contracts and accounts" accent="healthy" class="span-8">
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
      <Panel title="Edit SLA policy" eyebrow="Administrator controlled" accent="recovery" class="span-6">
        <div class="form-grid">
          <label>
            <span>Policy</span>
            <select v-model="selectedSlaId" aria-label="SLA policy to edit">
              <option v-for="policy in dashboard.slaPolicies" :key="policy.id" :value="policy.id">{{ policy.name }}</option>
            </select>
          </label>
          <label>
            <span>Credit SLA</span>
            <input v-model="slaDraft.duration" aria-label="Credit SLA duration" placeholder="90s" />
          </label>
          <label>
            <span>Cooldown window</span>
            <input v-model="slaDraft.cooldownWindow" aria-label="Cooldown window" placeholder="15m" />
          </label>
          <label>
            <span>Late-success window</span>
            <input v-model="slaDraft.lateSuccessWindow" aria-label="Late success window" placeholder="30m" />
          </label>
          <aside v-if="slaMessage" class="state-note state-note--success">
            <FileCheck2 :size="16" aria-hidden="true" />
            <span>{{ slaMessage }}</span>
          </aside>
          <UiButton @click="saveSlaPolicy">Save SLA change</UiButton>
        </div>
      </Panel>
      <Panel title="Backoff policy" eyebrow="Route-specific requery" accent="healthy" class="span-12">
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
      <Panel title="Provider traffic controls" eyebrow="New eligible transfers only" accent="watch" class="span-12">
        <div class="provider-control-list">
          <article v-for="row in providerControlRows" :key="row.route">
            <span>
              <strong>{{ row.rail }}</strong>
              <small>{{ row.route }} / {{ row.trafficSplit }}</small>
            </span>
            <HealthBadge :state="row.state" :trigger="row.penaltyReason" />
            <UiButton :variant="row.state === 'blocked' ? 'secondary' : 'danger'" size="sm" @click="toggleProvider(row)">
              <Power :size="14" aria-hidden="true" />
              {{ row.state === 'blocked' ? 'Reactivate' : 'Deactivate' }}
            </UiButton>
          </article>
        </div>
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

    <section v-else-if="settingsTab === 'profile'" class="dashboard-grid">
      <Panel title="Current user profile" eyebrow="Signed-in operator" accent="healthy" class="span-5">
        <div class="profile-card">
          <UserCircle :size="28" aria-hidden="true" />
          <span>
            <strong>{{ sessionUser?.display_name ?? 'Unknown user' }}</strong>
            <small>{{ sessionUser?.email ?? 'No email on file' }}</small>
          </span>
          <dl class="definition-list">
            <div><dt>Bank ID</dt><dd>{{ sessionUser?.bank_id ?? '-' }}</dd></div>
            <div><dt>Username</dt><dd>{{ sessionUser?.username ?? '-' }}</dd></div>
            <div><dt>Roles</dt><dd>{{ userRoles }}</dd></div>
            <div><dt>Access</dt><dd>{{ permissionSummary }}</dd></div>
            <div><dt>Auth provider</dt><dd>{{ sessionUser?.auth_provider ?? '-' }}</dd></div>
          </dl>
        </div>
      </Panel>

      <Panel title="Change password" eyebrow="Account security" accent="watch" class="span-7">
        <form class="form-grid" @submit.prevent="submitPasswordChange">
          <label>
            <span>Current password</span>
            <input v-model="passwordForm.current" type="password" autocomplete="current-password" aria-label="Current password" />
          </label>
          <label>
            <span>New password</span>
            <input v-model="passwordForm.next" type="password" autocomplete="new-password" aria-label="New password" />
          </label>
          <label>
            <span>Confirm new password</span>
            <input v-model="passwordForm.confirm" type="password" autocomplete="new-password" aria-label="Confirm new password" />
          </label>
          <small>Minimum 12 characters. This action is recorded in the audit trail.</small>
          <p v-if="passwordError" class="form-error">{{ passwordError }}</p>
          <aside v-if="passwordMessage" class="state-note state-note--success">
            <KeyRound :size="16" aria-hidden="true" />
            <span>{{ passwordMessage }}</span>
          </aside>
          <UiButton :disabled="passwordBusy" @click="submitPasswordChange">
            <KeyRound :size="15" aria-hidden="true" />
            Update password
          </UiButton>
        </form>
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
