<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Activity, Gauge, Inbox, LogOut, Network, Settings, ShieldAlert, ToggleRight } from '@lucide/vue'
import DataFreshness from './components/DataFreshness.vue'
import LoginPanel from './components/LoginPanel.vue'
import PageHeader from './components/PageHeader.vue'
import UiButton from './components/UiButton.vue'
import { normalizeDataState } from './composables/useInboundTower'
import { useTowerRouting } from './composables/useTowerRouting'
import { routeForScreen } from './router'
import { currentUser, loginLDAP, loginLocal, logout } from './services/authApi'
import { getDashboardMock } from './services/mockDashboard'
import CommandCenterView from './views/CommandCenterView.vue'
import ExceptionsView from './views/ExceptionsView.vue'
import InflowsView from './views/InflowsView.vue'
import RoutesView from './views/RoutesView.vue'
import SettingsView, { type SettingsTab } from './views/SettingsView.vue'
import type { ScreenId, SessionUser, UiScenario } from './types'

const { route, router, activate, openPath } = useTowerRouting()
const dashboard = reactive(getDashboardMock(normalizeDataState(route.query.scenario)))

const sessionUser = ref<SessionUser | null>(null)
const authReady = ref(false)
const authBusy = ref(false)
const authError = ref('')
const dataState = ref<UiScenario>(normalizeDataState(route.query.scenario))
const settingsTab = ref<SettingsTab>('contracts')

const navigation = [
  { id: 'command' as ScreenId, label: 'Command Center', icon: Gauge, kicker: 'Settlement desk' },
  { id: 'inflows' as ScreenId, label: 'Inflows', icon: Inbox, kicker: 'Transaction trace' },
  { id: 'routes' as ScreenId, label: 'Routes', icon: Network, kicker: 'Rail health' },
  { id: 'exceptions' as ScreenId, label: 'Exceptions', icon: ShieldAlert, kicker: 'Resolution queue' },
  { id: 'settings' as ScreenId, label: 'Settings', icon: Settings, kicker: 'Policy and controls' },
]

const screenDescriptions: Record<ScreenId, string> = {
  command: 'Failed transfers, provider performance, partner SLA breaches, and API telemetry.',
  inflows: 'Trace each inbound instruction from partner receipt through final-leg evidence and customer value proof.',
  routes: 'Compare provider and rail pressure before shifting new traffic or freezing a degraded endpoint.',
  exceptions: 'Control unresolved credits with evidence-first requery, manual completion, and reversal workflows.',
  settings: 'Manage contracts, standing accounts, SLA/backoff policy, feature rollout, and audit evidence.',
}

const activeScreen = computed<ScreenId>(() => (route.meta.screen as ScreenId | undefined) ?? 'command')
const activeNavItem = computed(() => navigation.find((item) => item.id === activeScreen.value) ?? navigation[0])
const pageTitle = computed(() => activeNavItem.value.label)
const pageDescription = computed(() => screenDescriptions[activeScreen.value])
const isDetailRoute = computed(() => Boolean(route.params.reference || route.params.routeId))
const actorName = computed(() => sessionUser.value?.display_name ?? 'Operations Admin')
const breadcrumbs = computed(() => {
  const crumbs: Array<{ label: string; path?: string }> = [
    { label: 'INSWITCH', path: '/' },
    { label: activeNavItem.value.label, path: routeForScreen(activeScreen.value).path },
  ]
  if (isDetailRoute.value) crumbs.push({ label: String(route.params.reference ?? route.params.routeId) })
  return crumbs
})

watch(
  () => route.query.scenario,
  (scenario) => {
    const nextState = normalizeDataState(scenario)
    if (nextState === dataState.value) return
    dataState.value = nextState
    Object.assign(dashboard, getDashboardMock(dataState.value))
  },
)

watch(
  () => route.query.tab,
  (tab) => {
    if (tab === 'contracts' || tab === 'sla' || tab === 'integrations' || tab === 'switches' || tab === 'audit' || tab === 'profile') settingsTab.value = tab
  },
  { immediate: true },
)

onMounted(async () => {
  try {
    sessionUser.value = await currentUser()
  } catch (error) {
    authError.value = error instanceof Error ? error.message : 'Unable to load session'
  } finally {
    authReady.value = true
  }
})

async function handleLogin(payload: { mode: 'local' | 'ldap'; bankId: string; username: string; password: string }) {
  authBusy.value = true
  authError.value = ''
  try {
    sessionUser.value =
      payload.mode === 'ldap'
        ? await loginLDAP(payload.bankId, payload.username, payload.password)
        : await loginLocal(payload.bankId, payload.username, payload.password)
  } catch (error) {
    authError.value = error instanceof Error ? error.message : 'Unable to sign in'
  } finally {
    authBusy.value = false
  }
}

async function signOut() {
  await logout()
  sessionUser.value = null
}

function setDataState(value: UiScenario) {
  dataState.value = value
  void router.push({ path: route.path, query: { ...route.query, scenario: value } })
}

function handleDataStateChange(event: Event) {
  setDataState((event.target as HTMLSelectElement).value as UiScenario)
}

function openSettingsTab(tab: SettingsTab) {
  settingsTab.value = tab
  openPath('/settings', { tab })
}
</script>

<template>
  <div v-if="!authReady" class="login-shell">
    <section class="login-card">
      <Activity :size="24" aria-hidden="true" />
      <p>Loading inbound settlement tower...</p>
    </section>
  </div>

  <LoginPanel v-else-if="!sessionUser" :error="authError" :busy="authBusy" @login="handleLogin" />

  <div v-else class="app-shell">
    <aside class="sidebar" aria-label="Product navigation">
      <RouterLink class="brand" to="/" @click="activate('command')">
        <span class="brand__mark" aria-hidden="true">
          <Activity :size="18" />
        </span>
        <span>
          <strong>INSWITCH</strong>
          <small>Fidelity settlement operations</small>
        </span>
      </RouterLink>

      <nav class="primary-nav">
        <section class="nav-group" aria-label="Inbound settlement">
          <p class="nav-group__label">Operate</p>
          <button
            v-for="item in navigation"
            :key="item.id"
            type="button"
            class="nav-item"
            :class="{ 'is-active': activeScreen === item.id }"
            @click="activate(item.id)"
          >
            <component :is="item.icon" :size="18" aria-hidden="true" />
            <span>
              <strong>{{ item.label }}</strong>
              <small>{{ item.kicker }}</small>
            </span>
          </button>
        </section>
      </nav>

      <section class="sidebar-status">
        <span>Monitoring feed</span>
        <label class="data-state-control">
          <select :value="dataState" aria-label="Monitoring feed" @change="handleDataStateChange">
            <option value="degraded-ria">Incident active</option>
            <option value="healthy">All routes nominal</option>
            <option value="traffic-shift">Recovery shift</option>
            <option value="empty">No active records</option>
            <option value="api-failure">Connector outage</option>
          </select>
        </label>
        <small>{{ sessionUser.display_name }} / ops console</small>
        <button type="button" class="sidebar-link" @click="openSettingsTab('profile')">
          Profile and password
        </button>
        <button type="button" class="sidebar-link" @click="signOut">
          <LogOut :size="13" aria-hidden="true" />
          Sign out
        </button>
      </section>
    </aside>

    <main class="workspace">
      <PageHeader :title="pageTitle" :description="pageDescription" :breadcrumbs="breadcrumbs" eyebrow="Settlement operations">
        <template #actions>
          <DataFreshness :updated="dashboard.summary.lastUpdated" :mode="dashboard.summary.connection.nextPollIn" :stale="dashboard.summary.connection.freshness !== 'fresh'" />
          <UiButton v-if="activeScreen === 'exceptions' && dashboard.remediationCases[0]" size="sm" @click="activate('exceptions')">
            <ShieldAlert :size="15" aria-hidden="true" />
            Work top case
          </UiButton>
          <UiButton v-if="activeScreen === 'settings'" size="sm" variant="secondary" @click="openSettingsTab('switches')">
            <ToggleRight :size="15" aria-hidden="true" />
            Feature switches
          </UiButton>
        </template>
      </PageHeader>

      <CommandCenterView v-if="activeScreen === 'command'" :dashboard="dashboard" :actor-name="actorName" />
      <InflowsView v-else-if="activeScreen === 'inflows'" :dashboard="dashboard" />
      <RoutesView v-else-if="activeScreen === 'routes'" :dashboard="dashboard" @settings-tab="openSettingsTab" />
      <ExceptionsView v-else-if="activeScreen === 'exceptions'" :dashboard="dashboard" :actor-name="actorName" />
      <SettingsView v-else :dashboard="dashboard" :initial-tab="settingsTab" :session-user="sessionUser" />
    </main>
  </div>
</template>
