<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  BadgeCheck,
  BarChart3,
  BellRing,
  CheckCircle2,
  CircleDollarSign,
  Clock3,
  Gauge,
  GitBranch,
  History,
  Network,
  ReceiptText,
  Search,
  ShieldCheck,
  SlidersHorizontal,
  TimerReset,
} from '@lucide/vue'
import ActionBar from './components/ActionBar.vue'
import CountryPair from './components/CountryPair.vue'
import DataFreshness from './components/DataFreshness.vue'
import DataTable from './components/DataTable.vue'
import EmptyState from './components/EmptyState.vue'
import HealthBadge from './components/HealthBadge.vue'
import KpiTile from './components/KpiTile.vue'
import PageHeader from './components/PageHeader.vue'
import Panel from './components/Panel.vue'
import ProviderMark from './components/ProviderMark.vue'
import RoutePath from './components/RoutePath.vue'
import RouteScoreChip from './components/RouteScoreChip.vue'
import StateBanner from './components/StateBanner.vue'
import TransactionTimeline from './components/TransactionTimeline.vue'
import TraceEmptyIllustration from './components/TraceEmptyIllustration.vue'
import UiButton from './components/UiButton.vue'
import { getDashboardMock } from './services/mockDashboard'
import { parseCorridor } from './services/identity'
import type { HealthState, ProviderToggle, ScreenId, ScoringWeight, TransactionRecord } from './types'

const dashboard = getDashboardMock()

const activeScreen = ref<ScreenId>('control')
const dateRange = ref(dashboard.dateRange.label)
const qaThresholdSeconds = ref(dashboard.qaPolicy.thresholdSeconds)
const warningThresholdSeconds = ref(dashboard.qaPolicy.warningSeconds)
const savedQaThresholdSeconds = ref(dashboard.qaPolicy.thresholdSeconds)
const savedWarningThresholdSeconds = ref(dashboard.qaPolicy.warningSeconds)

const transactionQuery = ref('')
const selectedTransactionReference = ref('')
const senderFilter = ref('All senders')
const destinationFilter = ref('All destinations')
const currencyFilter = ref('All currencies')
const timeFilter = ref('All timing')
const destinationTypeFilter = ref('All destination types')
const sortBy = ref('totalTimeDesc')
const fxBaseCurrency = ref('USD')
const fxTargetCurrency = ref('NGN')

const providerDraft = ref<ProviderToggle[]>(dashboard.routeConfig.providers.map((provider) => ({ ...provider })))
const scoringDraft = ref<ScoringWeight[]>(dashboard.routeConfig.weights.map((weight) => ({ ...weight })))
const selectedPreset = ref(dashboard.routeConfig.presets.find((preset) => preset.active)?.label ?? dashboard.routeConfig.presets[0]?.label ?? '')
const draftReason = ref(dashboard.routeConfig.workflow.change.reason)
const savedProviderDraft = ref(JSON.stringify(providerDraft.value))
const savedScoringDraft = ref(JSON.stringify(scoringDraft.value))
const savedPreset = ref(selectedPreset.value)
const savedDraftReason = ref(draftReason.value)

const auditQuery = ref('')
const selectedAuditTime = ref(dashboard.auditEvents[0]?.time ?? '')

const navigation = [
  { id: 'control' as ScreenId, label: 'Control Room', icon: Gauge, kicker: 'Live triage' },
  { id: 'transactions' as ScreenId, label: 'Transactions', icon: Search, kicker: 'Find and trace' },
  { id: 'corridors' as ScreenId, label: 'Routes', icon: Network, kicker: 'Health and fallbacks' },
  { id: 'policy' as ScreenId, label: 'Policy', icon: SlidersHorizontal, kicker: 'Draft defaults' },
  { id: 'incidents' as ScreenId, label: 'Incidents', icon: BellRing, kicker: 'Open work' },
  { id: 'fx' as ScreenId, label: 'Rates & costs', icon: CircleDollarSign, kicker: 'Economics' },
  { id: 'reconciliation' as ScreenId, label: 'Reconcile', icon: ReceiptText, kicker: 'Settlement breaks' },
  { id: 'providers' as ScreenId, label: 'Providers', icon: BarChart3, kicker: 'Partner health' },
  { id: 'audit' as ScreenId, label: 'Audit', icon: History, kicker: 'Decision log' },
]

const navigationGroups = [
  { label: 'Operate', items: navigation.slice(0, 4) },
  { label: 'Review', items: navigation.slice(4) },
]

const screenDescriptions: Record<ScreenId, string> = {
  control: 'Live route triage.',
  corridors: 'Corridors, routes, fallbacks.',
  transactions: 'Search and trace transfers.',
  incidents: 'Open route incidents.',
  policy: 'Draft routing defaults.',
  fx: 'Rates, cost, baseline.',
  reconciliation: 'Settlement breaks.',
  providers: 'Primary provider health.',
  audit: 'Decision log.',
}

const severityRank: Record<HealthState, number> = {
  blocked: 0,
  degraded: 1,
  watch: 2,
  stale: 3,
  recovery: 4,
  unknown: 5,
  healthy: 6,
}

const selectedScreen = computed(() => navigation.find((item) => item.id === activeScreen.value) ?? navigation[0])
const activeIncident = computed(() => dashboard.incidents[0] ?? null)
const sortedCorridors = computed(() => [...dashboard.corridors].sort((a, b) => severityRank[a.state] - severityRank[b.state]))
const sortedProviders = computed(() => [...dashboard.providerScores].sort((a, b) => a.rank - b.rank))
const weakestProvider = computed(() => sortedProviders.value[sortedProviders.value.length - 1])
const primaryIncidentCorridor = computed(() => activeIncident.value?.corridor ?? sortedCorridors.value[0]?.corridor ?? 'European Union -> Nigeria')
const primaryIncidentRoute = computed(() => sortedCorridors.value.find((corridor) => corridor.corridor === primaryIncidentCorridor.value) ?? sortedCorridors.value[0])
const simulationSample = computed(() => dashboard.policySimulationSamples[0])

const unique = (values: string[]) => [...new Set(values)]
const senderCountries = computed(() => unique(dashboard.transactions.map((transaction) => transaction.senderCountry)))
const destinationCountries = computed(() => unique(dashboard.transactions.map((transaction) => transaction.destinationCountry)))
const currencies = computed(() => unique(dashboard.transactions.flatMap((transaction) => [transaction.senderCurrency, transaction.destinationCurrency])))
const fxCurrencies = ['USD', 'EUR', 'GBP', 'KES', 'NGN']
const destinationTypes = computed(() => unique(dashboard.transactions.map((transaction) => transaction.destinationType)))
const qaThresholdLabel = computed(() => formatDuration(qaThresholdSeconds.value))
const warningThresholdLabel = computed(() => formatDuration(warningThresholdSeconds.value))
const thresholdIsValid = computed(() => warningThresholdSeconds.value >= 15 && qaThresholdSeconds.value >= warningThresholdSeconds.value)
const thresholdHasChanges = computed(
  () => qaThresholdSeconds.value !== savedQaThresholdSeconds.value || warningThresholdSeconds.value !== savedWarningThresholdSeconds.value,
)

const policyHasChanges = computed(
  () =>
    JSON.stringify(providerDraft.value) !== savedProviderDraft.value ||
    JSON.stringify(scoringDraft.value) !== savedScoringDraft.value ||
    selectedPreset.value !== savedPreset.value ||
    draftReason.value.trim() !== savedDraftReason.value.trim(),
)

const routeStateCounts = computed(() =>
  dashboard.corridors.reduce(
    (counts, corridor) => {
      counts[corridor.state] = (counts[corridor.state] ?? 0) + 1
      return counts
    },
    {} as Record<HealthState, number>,
  ),
)

const filteredTransactions = computed(() => {
  const query = transactionQuery.value.trim().toLowerCase()
  return dashboard.transactions
    .filter((transaction) => {
      const stalled = isTransactionStalled(transaction)
      const withinPolicy = transactionWithinPolicy(transaction)
      const matchesQuery =
        !query ||
        transaction.reference.toLowerCase().includes(query) ||
        transaction.providerReference.toLowerCase().includes(query) ||
        transaction.bankReference.toLowerCase().includes(query) ||
        transaction.provider.toLowerCase().includes(query) ||
        transaction.beneficiary.toLowerCase().includes(query)
      const matchesSender = senderFilter.value === 'All senders' || transaction.senderCountry === senderFilter.value
      const matchesDestination = destinationFilter.value === 'All destinations' || transaction.destinationCountry === destinationFilter.value
      const matchesCurrency =
        currencyFilter.value === 'All currencies' ||
        transaction.senderCurrency === currencyFilter.value ||
        transaction.destinationCurrency === currencyFilter.value
      const matchesType = destinationTypeFilter.value === 'All destination types' || transaction.destinationType === destinationTypeFilter.value
      const matchesTiming =
        timeFilter.value === 'All timing' ||
        (timeFilter.value === 'Under QA policy' && withinPolicy) ||
        (timeFilter.value === 'Over QA policy' && !withinPolicy && !stalled) ||
        (timeFilter.value === 'Stalled only' && stalled)

      return matchesQuery && matchesSender && matchesDestination && matchesCurrency && matchesType && matchesTiming
    })
    .sort((a, b) => {
      if (sortBy.value === 'totalTimeAsc') return a.totalTimeSeconds - b.totalTimeSeconds
      if (sortBy.value === 'reference') return a.reference.localeCompare(b.reference)
      if (sortBy.value === 'sender') return a.senderCountry.localeCompare(b.senderCountry)
      if (sortBy.value === 'destination') return a.destinationCountry.localeCompare(b.destinationCountry)
      return b.totalTimeSeconds - a.totalTimeSeconds
    })
})

const transactionStatusCounts = computed(() =>
  dashboard.transactions.reduce(
    (counts, transaction) => {
      if (isTransactionStalled(transaction)) counts.stalled += 1
      else if (transactionWithinPolicy(transaction)) counts.onTime += 1
      else counts.overPolicy += 1
      return counts
    },
    { all: dashboard.transactions.length, onTime: 0, overPolicy: 0, stalled: 0 },
  ),
)

const hasTransactionFilters = computed(
  () =>
    Boolean(transactionQuery.value.trim()) ||
    senderFilter.value !== 'All senders' ||
    destinationFilter.value !== 'All destinations' ||
    currencyFilter.value !== 'All currencies' ||
    timeFilter.value !== 'All timing' ||
    destinationTypeFilter.value !== 'All destination types' ||
    sortBy.value !== 'totalTimeDesc',
)

const selectedTransaction = computed<TransactionRecord | null>(
  () => filteredTransactions.value.find((transaction) => transaction.reference === selectedTransactionReference.value) ?? null,
)

const selectedTimeline = computed(() => {
  const transaction = selectedTransaction.value
  if (!transaction) return []
  if (transaction.reference === dashboard.trace.reference) return dashboard.trace.timeline

  const isPending = transaction.destinationCreditedAt === 'pending'
  return [
    {
      label: 'Sender initiated',
      owner: 'Bank channel',
      status: 'done' as const,
      time: transaction.senderStartedAt,
      duration: '0s',
      source: 'Bank API',
      reference: transaction.bankReference,
    },
    {
      label: isPending ? 'Destination credit pending' : 'Destination credited',
      owner: isPending ? transaction.currentOwner : transaction.beneficiary,
      status: isPending ? ('current' as const) : ('done' as const),
      time: transaction.destinationCreditedAt,
      duration: transaction.totalTime,
      source: isPending ? transaction.route : 'Destination confirmation',
      reference: isPending ? transaction.providerReference : transaction.bankReference,
      note: transaction.blocker,
    },
  ]
})

const filteredAuditEvents = computed(() => {
  const query = auditQuery.value.trim().toLowerCase()
  if (!query) return dashboard.auditEvents
  return dashboard.auditEvents.filter((event) =>
    [event.time, event.actor, event.action, event.object, event.reason].some((value) => value.toLowerCase().includes(query)),
  )
})

const selectedAudit = computed(
  () => filteredAuditEvents.value.find((event) => event.time === selectedAuditTime.value) ?? filteredAuditEvents.value[0] ?? null,
)

const fxComparisonRoutes = computed(() =>
  dashboard.fxCostBoard.routes.map((route, index) => ({
    ...route,
    baselineRate: baselineRateFor(index),
    pair: `${fxBaseCurrency.value}/${fxTargetCurrency.value}`,
  })),
)

const activate = (screen: ScreenId) => {
  activeScreen.value = screen
}

const resetTransactionFilters = () => {
  transactionQuery.value = ''
  senderFilter.value = 'All senders'
  destinationFilter.value = 'All destinations'
  currencyFilter.value = 'All currencies'
  timeFilter.value = 'All timing'
  destinationTypeFilter.value = 'All destination types'
  sortBy.value = 'totalTimeDesc'
}

const saveQaThresholds = () => {
  if (!thresholdIsValid.value) return
  savedQaThresholdSeconds.value = qaThresholdSeconds.value
  savedWarningThresholdSeconds.value = warningThresholdSeconds.value
}

const resetQaThresholds = () => {
  qaThresholdSeconds.value = savedQaThresholdSeconds.value
  warningThresholdSeconds.value = savedWarningThresholdSeconds.value
}

const savePolicyDraft = () => {
  savedProviderDraft.value = JSON.stringify(providerDraft.value)
  savedScoringDraft.value = JSON.stringify(scoringDraft.value)
  savedPreset.value = selectedPreset.value
  savedDraftReason.value = draftReason.value
}

const resetPolicyDraft = () => {
  providerDraft.value = JSON.parse(savedProviderDraft.value) as ProviderToggle[]
  scoringDraft.value = JSON.parse(savedScoringDraft.value) as ScoringWeight[]
  selectedPreset.value = savedPreset.value
  draftReason.value = savedDraftReason.value
}

function formatDuration(seconds: number) {
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return remainingSeconds ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`
}

function isTransactionStalled(transaction: TransactionRecord) {
  return transaction.qaStatus === 'stalled' || transaction.destinationCreditedAt === 'pending'
}

function transactionWithinPolicy(transaction: TransactionRecord) {
  return !isTransactionStalled(transaction) && transaction.totalTimeSeconds <= qaThresholdSeconds.value
}

function qaStatusLabel(transaction: TransactionRecord) {
  if (isTransactionStalled(transaction)) return 'stalled'
  return transactionWithinPolicy(transaction) ? 'on time' : 'over policy'
}

function qaStateFor(transaction: TransactionRecord): HealthState {
  if (isTransactionStalled(transaction)) return 'blocked'
  if (transactionWithinPolicy(transaction)) return 'healthy'
  if (transaction.totalTimeSeconds <= qaThresholdSeconds.value * 1.5) return 'watch'
  return 'degraded'
}

function routeDecisionSummary(transaction: TransactionRecord) {
  if (transaction.destinationCreditedAt === 'pending') return `${transaction.provider} has not confirmed destination credit after ${transaction.totalTime}.`
  const policyResult = transaction.totalTimeSeconds <= qaThresholdSeconds.value ? 'inside' : 'outside'
  return `${transaction.provider} completed in ${transaction.totalTime}, ${policyResult} the ${formatDuration(qaThresholdSeconds.value)} QA policy.`
}

function stateLabel(state: HealthState) {
  const labels: Record<HealthState, string> = {
    healthy: 'Healthy',
    watch: 'Watch',
    degraded: 'Degraded',
    blocked: 'Blocked',
    recovery: 'Recovery',
    unknown: 'Unknown',
    stale: 'Stale',
  }
  return labels[state]
}

function corridorParts(corridor: string) {
  return parseCorridor(corridor)
}

function routeProvider(route: string) {
  return route.split('->')[0]?.trim() ?? route
}

function routeRail(route: string) {
  const parts = route.split('->').map((part) => part.trim())
  return parts.length > 1 ? parts.slice(1).join(' -> ') : ''
}

function percentWidth(value: string | number) {
  const parsed = typeof value === 'number' ? value : Number(String(value).replace(/[^\d.]/g, ''))
  return `${Math.max(3, Math.min(100, Number.isFinite(parsed) ? parsed : 0))}%`
}

function baselineRateFor(index: number) {
  const targetFactor: Record<string, number> = {
    NGN: 1,
    EUR: 0.00058,
    GBP: 0.00049,
    KES: 0.075,
    USD: 0.00064,
  }
  const baseFactor: Record<string, number> = {
    USD: 1,
    EUR: 0.92,
    GBP: 0.78,
    KES: 129,
    NGN: 1560,
  }
  const baseUsdRate = [1560.2, 1557.8, 1552.4, 1549.6][index] ?? 1550
  const adjusted = (baseUsdRate / (baseFactor[fxBaseCurrency.value] || 1)) * (targetFactor[fxTargetCurrency.value] || 1)
  return adjusted.toLocaleString('en-US', { maximumFractionDigits: adjusted > 100 ? 2 : 4 })
}
</script>

<template>
  <div class="app-shell" data-bank-theme="imsi">
    <aside class="sidebar" aria-label="Product navigation">
      <a class="brand" href="/" aria-label="imsi-rails">
        <span class="brand__mark" aria-hidden="true">
          <Activity :size="15" />
        </span>
        <span>
          <strong>imsi-rails</strong>
          <small>Route reliability switch</small>
        </span>
      </a>

      <nav class="primary-nav" aria-label="Primary">
        <section v-for="group in navigationGroups" :key="group.label" class="nav-group" :aria-label="group.label">
          <span class="nav-group__label">{{ group.label }}</span>
          <button
            v-for="item in group.items"
            :key="item.id"
            type="button"
            class="nav-item"
            :class="{ 'is-active': activeScreen === item.id }"
            @click="activate(item.id)"
          >
            <component :is="item.icon" :size="17" aria-hidden="true" />
            <span>
              <strong>{{ item.label }}</strong>
              <small>{{ item.kicker }}</small>
            </span>
          </button>
        </section>
      </nav>

      <section class="sidebar-status" aria-label="Connection status">
        <span>Data source</span>
        <strong>{{ dashboard.summary.connection.mode }}</strong>
        <small>{{ dashboard.summary.connection.nextPollIn }}</small>
      </section>
    </aside>

    <main class="workspace">
      <PageHeader
        eyebrow="Bank operations / Nigeria inbound"
        :title="selectedScreen.label"
        :description="screenDescriptions[selectedScreen.id]"
      >
        <template #actions>
          <DataFreshness
            :updated="dashboard.summary.lastUpdated"
            :mode="dashboard.summary.connection.mode"
            :stale="dashboard.summary.connection.freshness !== 'fresh'"
          />
        </template>
      </PageHeader>

      <StateBanner :scenario="dashboard.scenario" />

      <section v-if="activeScreen === 'control'" class="screen-stack">
        <section class="kpi-grid">
          <KpiTile label="Routes healthy" :value="dashboard.summary.globalHealth" :detail="dashboard.summary.topRisk" tone="degraded" :icon="Gauge" />
          <KpiTile label="At-risk value" :value="dashboard.summary.atRiskValue" :detail="`${dashboard.summary.stuckTransactions} transfers need review`" tone="watch" :icon="AlertTriangle" />
          <KpiTile label="P95 credit time" :value="dashboard.summary.p95CreditTime" :detail="`Target ${qaThresholdLabel}`" tone="degraded" :icon="Clock3" />
          <KpiTile label="Active incidents" :value="dashboard.summary.activeIncidents" detail="Open route work" tone="recovery" :icon="BellRing" />
        </section>

        <section class="dashboard-grid">
          <Panel title="Recommended next action" eyebrow="Triage" :accent="dashboard.recommendation.state" class="span-5">
            <div class="recommendation-card">
              <HealthBadge :state="dashboard.recommendation.state" window="15 min" />
              <h3>{{ dashboard.recommendation.title }}</h3>
              <p>{{ dashboard.recommendation.trigger }}</p>
              <dl class="metric-grid metric-grid--three">
                <div>
                  <dt>Traffic</dt>
                  <dd>{{ dashboard.recommendation.affectedTraffic }}</dd>
                </div>
                <div>
                  <dt>Value</dt>
                  <dd>{{ dashboard.recommendation.affectedValue }}</dd>
                </div>
                <div>
                  <dt>Evidence</dt>
                  <dd>{{ dashboard.recommendation.evidence }}</dd>
                </div>
              </dl>
              <RoutePath
                :origin="corridorParts(primaryIncidentCorridor).origin"
                :destination="corridorParts(primaryIncidentCorridor).destination"
                :provider="routeProvider(dashboard.recommendation.suggestedRoute)"
                :rail="routeRail(dashboard.recommendation.suggestedRoute)"
              />
              <ActionBar>
                <UiButton @click="activate('policy')">
                  Review policy
                  <ArrowRight :size="15" aria-hidden="true" />
                </UiButton>
                <UiButton variant="secondary" @click="activate('transactions')">Trace affected transfers</UiButton>
              </ActionBar>
            </div>
          </Panel>

          <Panel title="Top corridor risks" eyebrow="Routes" accent="degraded" class="span-7">
            <DataTable :empty="sortedCorridors.length === 0" empty-title="No corridor risk" empty-description="All monitored corridors are inside policy.">
              <table>
                <thead>
                  <tr>
                    <th>Corridor</th>
                    <th>Selected route</th>
                    <th>P95</th>
                    <th>Cost</th>
                    <th>Next team</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="corridor in sortedCorridors" :key="`${corridor.corridor}-${corridor.payout}`">
                    <td>
                      <CountryPair :origin="corridorParts(corridor.corridor).origin" :destination="corridorParts(corridor.corridor).destination" />
                      <small>{{ corridor.payout }}</small>
                    </td>
                    <td>
                      <ProviderMark :provider="routeProvider(corridor.selectedRoute)" />
                      <HealthBadge :state="corridor.state" :trigger="corridor.risk" />
                    </td>
                    <td><strong>{{ corridor.p95 }}</strong></td>
                    <td>{{ corridor.cost }}</td>
                    <td>
                      <strong>{{ corridor.owner }}</strong>
                      <small>{{ corridor.status }}</small>
                    </td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>

          <Panel title="Active incidents" eyebrow="Open work" :accent="activeIncident?.severity ?? 'healthy'" class="span-5">
            <EmptyState
              v-if="!activeIncident"
              title="No active incidents"
              description="No degraded routes currently require incident review."
              :icon="CheckCircle2"
              tone="success"
            />
            <div v-else class="incident-summary">
              <HealthBadge :state="activeIncident.severity" />
              <h3>{{ activeIncident.title }}</h3>
              <CountryPair :origin="corridorParts(activeIncident.corridor).origin" :destination="corridorParts(activeIncident.corridor).destination" />
              <dl class="metric-grid">
                <div>
                  <dt>Affected</dt>
                  <dd>{{ activeIncident.affectedTransactions }}</dd>
                </div>
                <div>
                  <dt>Value</dt>
                  <dd>{{ activeIncident.affectedValue }}</dd>
                </div>
                <div>
                  <dt>Next team</dt>
                  <dd>{{ activeIncident.owner }}</dd>
                </div>
              </dl>
              <p>{{ activeIncident.nextAction }}</p>
            </div>
          </Panel>

          <Panel title="Provider context" eyebrow="Summary" accent="healthy" class="span-7">
            <div class="provider-context">
              <article>
                <BarChart3 :size="18" aria-hidden="true" />
                <div>
                  <strong>Provider dashboard</strong>
                  <small>Scorecards, SLA, traffic, reconciliation.</small>
                </div>
              </article>
              <article>
                <AlertTriangle :size="18" aria-hidden="true" />
                <div>
                  <strong>{{ weakestProvider?.provider ?? 'None' }}</strong>
                  <small>{{ weakestProvider?.p95 ?? '-' }} P95 / {{ weakestProvider?.settlementExceptions ?? 0 }} breaks</small>
                </div>
              </article>
            </div>
            <ActionBar>
              <UiButton variant="secondary" @click="activate('providers')">Open Providers</UiButton>
            </ActionBar>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'corridors'" class="screen-stack">
        <section class="dashboard-grid">
          <Panel title="Corridor command center" eyebrow="Routes" accent="degraded" class="span-8">
            <DataTable :empty="sortedCorridors.length === 0" empty-title="No corridors configured" empty-description="Add corridor routes before monitoring can start.">
              <table>
                <thead>
                  <tr>
                    <th>Corridor</th>
                    <th>Current route</th>
                    <th>Score</th>
                    <th>Split</th>
                    <th>Recommendation</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="corridor in sortedCorridors" :key="`${corridor.corridor}-${corridor.selectedRoute}`">
                    <td>
                      <CountryPair :origin="corridorParts(corridor.corridor).origin" :destination="corridorParts(corridor.corridor).destination" />
                      <small>{{ corridor.payout }} / {{ corridor.atRiskValue }} at risk</small>
                    </td>
                    <td>
                      <ProviderMark :provider="routeProvider(corridor.selectedRoute)" />
                      <small>{{ routeRail(corridor.selectedRoute) || corridor.risk }}</small>
                    </td>
                    <td><RouteScoreChip :score="corridor.score" :reason="corridor.risk" :confidence="corridor.status" /></td>
                    <td>{{ corridor.split }}</td>
                    <td>
                      <HealthBadge :state="corridor.state" :trigger="corridor.recommendation" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>

          <Panel title="Route action" eyebrow="Current focus" :accent="primaryIncidentRoute?.state ?? 'unknown'" class="span-4">
            <div v-if="primaryIncidentRoute" class="route-focus">
              <RoutePath
                :origin="corridorParts(primaryIncidentRoute.corridor).origin"
                :destination="corridorParts(primaryIncidentRoute.corridor).destination"
                :provider="routeProvider(primaryIncidentRoute.selectedRoute)"
                :rail="routeRail(primaryIncidentRoute.selectedRoute)"
              />
              <dl class="metric-grid">
                <div>
                  <dt>P95</dt>
                  <dd>{{ primaryIncidentRoute.p95 }}</dd>
                </div>
                <div>
                  <dt>Cost</dt>
                  <dd>{{ primaryIncidentRoute.cost }}</dd>
                </div>
                <div>
                  <dt>Split</dt>
                  <dd>{{ primaryIncidentRoute.split }}</dd>
                </div>
              </dl>
              <p>{{ primaryIncidentRoute.risk }}</p>
              <ActionBar>
                <UiButton @click="activate('policy')">Adjust draft</UiButton>
                <UiButton variant="secondary" @click="activate('transactions')">Trace</UiButton>
              </ActionBar>
            </div>
          </Panel>

          <Panel title="Fallback order" eyebrow="Draft policy" accent="watch" class="span-6">
            <ol class="rank-list">
              <li v-for="route in dashboard.routeConfig.fallbackRoutes" :key="route.provider">
                <span class="rank">#{{ route.rank }}</span>
                <ProviderMark :provider="route.provider" show-category />
                <span>{{ route.route }}</span>
                <HealthBadge :state="route.state" />
              </li>
            </ol>
          </Panel>

          <Panel title="Route states" eyebrow="Fleet" accent="healthy" class="span-6">
            <div class="state-count-grid">
              <article v-for="state in ['healthy', 'watch', 'degraded', 'recovery']" :key="state">
                <HealthBadge :state="state as HealthState" />
                <strong>{{ routeStateCounts[state as HealthState] ?? 0 }}</strong>
              </article>
            </div>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'transactions'" class="screen-stack">
        <Panel title="Find, filter, trace" eyebrow="Transactions">
          <ActionBar>
            <label class="search-field">
              <Search :size="18" aria-hidden="true" />
              <input v-model="transactionQuery" type="search" placeholder="Reference, provider, bank, beneficiary" aria-label="Search transactions" />
            </label>
            <select v-model="timeFilter" aria-label="Timing filter">
              <option>All timing</option>
              <option>Under QA policy</option>
              <option>Over QA policy</option>
              <option>Stalled only</option>
            </select>
            <select v-model="senderFilter" aria-label="Sender filter">
              <option>All senders</option>
              <option v-for="sender in senderCountries" :key="sender">{{ sender }}</option>
            </select>
            <select v-model="destinationFilter" aria-label="Destination filter">
              <option>All destinations</option>
              <option v-for="destination in destinationCountries" :key="destination">{{ destination }}</option>
            </select>
            <select v-model="currencyFilter" aria-label="Currency filter">
              <option>All currencies</option>
              <option v-for="currency in currencies" :key="currency">{{ currency }}</option>
            </select>
            <select v-model="destinationTypeFilter" aria-label="Payout filter">
              <option>All destination types</option>
              <option v-for="type in destinationTypes" :key="type">{{ type }}</option>
            </select>
            <select v-model="sortBy" aria-label="Sort order">
              <option value="totalTimeDesc">Slowest first</option>
              <option value="totalTimeAsc">Fastest first</option>
              <option value="reference">Reference</option>
              <option value="sender">Sender</option>
              <option value="destination">Destination</option>
            </select>
            <UiButton variant="secondary" size="sm" :disabled="!hasTransactionFilters" @click="resetTransactionFilters">Reset</UiButton>
          </ActionBar>
          <div class="status-chips" aria-label="Transaction status counts">
            <button type="button" :class="{ 'is-active': timeFilter === 'All timing' }" @click="timeFilter = 'All timing'">All {{ transactionStatusCounts.all }}</button>
            <button type="button" :class="{ 'is-active': timeFilter === 'Under QA policy' }" @click="timeFilter = 'Under QA policy'">On time {{ transactionStatusCounts.onTime }}</button>
            <button type="button" :class="{ 'is-active': timeFilter === 'Over QA policy' }" @click="timeFilter = 'Over QA policy'">Over policy {{ transactionStatusCounts.overPolicy }}</button>
            <button type="button" :class="{ 'is-active': timeFilter === 'Stalled only' }" @click="timeFilter = 'Stalled only'">Stalled {{ transactionStatusCounts.stalled }}</button>
          </div>
        </Panel>

        <section class="transaction-layout">
          <Panel title="Results" :eyebrow="`${filteredTransactions.length} transfers`" class="transaction-list-panel">
            <DataTable
              :empty="filteredTransactions.length === 0"
              empty-title="No transactions match"
              empty-description="Clear filters or search a different reference."
            >
              <table>
                <thead>
                  <tr>
                    <th>Reference</th>
                    <th>Route</th>
                    <th>Amount</th>
                    <th>Elapsed</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="transaction in filteredTransactions"
                    :key="transaction.reference"
                    class="click-row"
                    :class="{ 'is-selected': selectedTransaction?.reference === transaction.reference }"
                    @click="selectedTransactionReference = transaction.reference"
                  >
                    <td>
                      <strong class="mono">{{ transaction.reference }}</strong>
                      <small>{{ transaction.providerReference }}</small>
                    </td>
                    <td>
                      <CountryPair :origin="transaction.senderCountry" :destination="transaction.destinationCountry" compact />
                      <small>{{ transaction.provider }} / {{ transaction.destinationType }}</small>
                    </td>
                    <td><strong>{{ transaction.amount }}</strong></td>
                    <td><strong>{{ transaction.totalTime }}</strong></td>
                    <td><HealthBadge :state="qaStateFor(transaction)" :trigger="qaStatusLabel(transaction)" /></td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>

          <Panel title="Trace detail" eyebrow="Selected transfer" class="transaction-detail-panel" :accent="selectedTransaction ? qaStateFor(selectedTransaction) : 'unknown'">
            <EmptyState
              v-if="!selectedTransaction"
              title="Select a transfer"
              description="Pick one row to trace timing, references, and route decision."
            >
              <TraceEmptyIllustration />
            </EmptyState>
            <div v-else class="transaction-detail">
              <div class="detail-heading">
                <div>
                  <h3 class="mono">{{ selectedTransaction.reference }}</h3>
                  <p>{{ selectedTransaction.amount }} / {{ selectedTransaction.beneficiary }}</p>
                </div>
                <HealthBadge :state="qaStateFor(selectedTransaction)" :trigger="qaStatusLabel(selectedTransaction)" />
              </div>
              <dl class="metric-grid metric-grid--four">
                <div>
                  <dt>Sender initiated</dt>
                  <dd>{{ selectedTransaction.senderStartedAt }}</dd>
                </div>
                <div>
                  <dt>Destination credited</dt>
                  <dd>{{ selectedTransaction.destinationCreditedAt }}</dd>
                </div>
                <div>
                  <dt>QA limit</dt>
                  <dd>{{ qaThresholdLabel }}</dd>
                </div>
                <div>
                  <dt>Next team</dt>
                  <dd>{{ selectedTransaction.currentOwner }}</dd>
                </div>
              </dl>
              <section class="detail-grid">
                <article>
                  <span class="eyebrow">References</span>
                  <dl class="reference-list">
                    <div>
                      <dt>Switch</dt>
                      <dd class="mono">{{ selectedTransaction.reference }}</dd>
                    </div>
                    <div>
                      <dt>Provider</dt>
                      <dd class="mono">{{ selectedTransaction.providerReference }}</dd>
                    </div>
                    <div>
                      <dt>Bank</dt>
                      <dd class="mono">{{ selectedTransaction.bankReference }}</dd>
                    </div>
                  </dl>
                </article>
                <article>
                  <span class="eyebrow">Route decision</span>
                  <RoutePath
                    :origin="selectedTransaction.senderCountry"
                    :destination="selectedTransaction.destinationCountry"
                    :provider="routeProvider(selectedTransaction.route)"
                    :rail="routeRail(selectedTransaction.route)"
                  />
                  <RouteScoreChip
                    v-if="selectedTransaction.reference === dashboard.trace.reference"
                    :score="dashboard.trace.selectedRoute.score"
                    :reason="dashboard.trace.selectedRoute.reason || dashboard.trace.selectedRoute.provider"
                    :confidence="dashboard.trace.selectedRoute.confidence"
                    :policy-version="dashboard.trace.policyVersion"
                  />
                  <p v-else>{{ routeDecisionSummary(selectedTransaction) }}</p>
                </article>
              </section>
              <aside class="state-note">
                <Clock3 :size="16" aria-hidden="true" />
                <span>{{ selectedTransaction.blocker }}</span>
                <strong>{{ selectedTransaction.totalTime }}</strong>
              </aside>
              <TransactionTimeline :steps="selectedTimeline" />
            </div>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'policy'" class="screen-stack">
        <section class="dashboard-grid">
          <Panel title="Active policy" eyebrow="Live defaults" accent="healthy" class="span-6">
            <dl class="definition-list">
              <div v-for="item in dashboard.routeConfig.workflow.currentPolicy" :key="item.label">
                <dt>{{ item.label }}</dt>
                <dd>{{ item.value }}</dd>
              </div>
            </dl>
          </Panel>

          <Panel title="Draft impact" eyebrow="Preview" accent="watch" class="span-6">
            <dl class="metric-grid metric-grid--three">
              <div>
                <dt>Success</dt>
                <dd>{{ dashboard.routeConfig.impact.successRate }}</dd>
              </div>
              <div>
                <dt>P95</dt>
                <dd>{{ dashboard.routeConfig.impact.p95 }}</dd>
              </div>
              <div>
                <dt>Cost</dt>
                <dd>{{ dashboard.routeConfig.impact.cost }}</dd>
              </div>
            </dl>
            <div class="validation-list">
              <span v-for="item in dashboard.routeConfig.workflow.validation" :key="item.label">
                <HealthBadge :state="item.state" />
                <strong>{{ item.value }}</strong>
                <small>{{ item.label }}</small>
              </span>
            </div>
          </Panel>

          <Panel title="Completion-time policy" eyebrow="QA thresholds" :accent="thresholdIsValid ? 'healthy' : 'degraded'" class="span-5">
            <template #actions>
              <UiButton variant="secondary" size="sm" :disabled="!thresholdHasChanges" @click="resetQaThresholds">Reset</UiButton>
              <UiButton size="sm" :disabled="!thresholdHasChanges || !thresholdIsValid" @click="saveQaThresholds">Save thresholds</UiButton>
            </template>
            <div class="form-grid">
              <label>
                <span>Healthy completion threshold</span>
                <input v-model.number="qaThresholdSeconds" type="number" min="15" step="15" />
                <small>Current healthy limit: {{ qaThresholdLabel }}.</small>
              </label>
              <label>
                <span>Warning threshold</span>
                <input v-model.number="warningThresholdSeconds" type="number" min="15" step="15" />
                <small>Warn before breach at {{ warningThresholdLabel }}.</small>
              </label>
            </div>
            <p v-if="!thresholdIsValid" class="form-error">Healthy threshold must be equal to or greater than the warning threshold.</p>
          </Panel>

          <Panel title="Draft route defaults" eyebrow="Editable" :accent="policyHasChanges ? 'watch' : 'healthy'" class="span-7">
            <template #actions>
              <span class="dashboard-chip">{{ policyHasChanges ? 'Unsaved draft' : 'Saved draft' }}</span>
              <UiButton variant="secondary" size="sm" :disabled="!policyHasChanges" @click="resetPolicyDraft">Reset</UiButton>
              <UiButton size="sm" :disabled="!policyHasChanges || !draftReason.trim()" @click="savePolicyDraft">Save draft</UiButton>
            </template>
            <section class="policy-edit-grid">
              <div>
                <span class="eyebrow">Provider defaults</span>
                <label v-for="provider in providerDraft" :key="provider.provider" class="toggle-row">
                  <input v-model="provider.enabled" type="checkbox" />
                  <ProviderMark :provider="provider.provider" />
                  <HealthBadge :state="provider.state" />
                </label>
              </div>
              <div>
                <span class="eyebrow">Traffic split</span>
                <div class="segmented-group">
                  <button
                    v-for="preset in dashboard.routeConfig.presets"
                    :key="preset.label"
                    type="button"
                    :class="{ 'is-selected': selectedPreset === preset.label }"
                    @click="selectedPreset = preset.label"
                  >
                    <strong>{{ preset.label }}</strong>
                    <small>{{ preset.split }}</small>
                  </button>
                </div>
              </div>
              <div class="span-2">
                <span class="eyebrow">Scoring weights</span>
                <label v-for="weight in scoringDraft" :key="weight.label" class="range-row">
                  <span>{{ weight.label }}</span>
                  <input v-model.number="weight.value" type="range" min="0" max="60" />
                  <strong>{{ weight.value }}</strong>
                </label>
              </div>
              <label class="span-2">
                <span>Audit reason</span>
                <textarea v-model="draftReason" rows="3" required />
              </label>
            </section>
          </Panel>

          <Panel title="Policy simulator" eyebrow="Shadow test" accent="healthy" class="span-7">
            <div v-if="simulationSample" class="simulation-grid">
              <section>
                <span class="eyebrow">Sample transaction</span>
                <h3 class="mono">{{ simulationSample.reference }}</h3>
                <dl class="metric-grid">
                  <div>
                    <dt>Corridor</dt>
                    <dd>
                      <CountryPair
                        :origin="corridorParts(simulationSample.corridor).origin"
                        :destination="corridorParts(simulationSample.corridor).destination"
                        compact
                      />
                    </dd>
                  </div>
                  <div>
                    <dt>Amount</dt>
                    <dd>{{ simulationSample.amount }}</dd>
                  </div>
                  <div>
                    <dt>Payout</dt>
                    <dd>{{ simulationSample.payout }}</dd>
                  </div>
                </dl>
              </section>
              <section>
                <span class="eyebrow">Route comparison</span>
                <div class="route-compare-row">
                  <ProviderMark :provider="simulationSample.current.provider" />
                  <strong>{{ simulationSample.current.score }}</strong>
                  <small>{{ simulationSample.current.p95 }} / {{ simulationSample.current.cost }}</small>
                </div>
                <div class="route-compare-row is-preferred">
                  <ProviderMark :provider="simulationSample.proposed.provider" />
                  <strong>{{ simulationSample.proposed.score }}</strong>
                  <small>{{ simulationSample.proposed.p95 }} / {{ simulationSample.proposed.cost }}</small>
                </div>
              </section>
            </div>
          </Panel>

          <Panel title="Policy history" eyebrow="Read-only" accent="recovery" class="span-5">
            <ol class="event-list">
              <li v-for="item in dashboard.routeConfig.history" :key="`${item.time}-${item.summary}`">
                <time>{{ item.time }}</time>
                <div>
                  <strong>{{ item.actor }}</strong>
                  <span>{{ item.summary }}</span>
                </div>
              </li>
            </ol>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'incidents'" class="screen-stack">
        <section class="dashboard-grid">
          <Panel title="Current incident" eyebrow="Open work" :accent="activeIncident?.severity ?? 'healthy'" class="span-5">
            <EmptyState v-if="!activeIncident" title="No active incidents" description="Routes are inside current operational thresholds." :icon="CheckCircle2" tone="success" />
            <div v-else class="incident-summary">
              <HealthBadge :state="activeIncident.severity" />
              <h3>{{ activeIncident.title }}</h3>
              <p>{{ activeIncident.rootCause }}</p>
              <dl class="metric-grid">
                <div>
                  <dt>Started</dt>
                  <dd>{{ activeIncident.startedAt }}</dd>
                </div>
                <div>
                  <dt>Affected</dt>
                  <dd>{{ activeIncident.affectedTransactions }}</dd>
                </div>
                <div>
                  <dt>Next team</dt>
                  <dd>{{ activeIncident.owner }}</dd>
                </div>
              </dl>
              <ActionBar>
                <UiButton @click="activate('policy')">Review policy</UiButton>
                <UiButton variant="secondary" @click="activate('transactions')">Trace transfers</UiButton>
              </ActionBar>
            </div>
          </Panel>

          <Panel title="Incident list" eyebrow="Active and recovery" accent="watch" class="span-7">
            <DataTable :empty="dashboard.incidents.length === 0" empty-title="No incidents" empty-description="No degraded routes require incident tracking.">
              <table>
                <thead>
                  <tr>
                    <th>Incident</th>
                    <th>Corridor</th>
                    <th>Affected</th>
                    <th>Next team</th>
                    <th>State</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="incident in dashboard.incidents" :key="incident.id">
                    <td>
                      <strong>{{ incident.title }}</strong>
                      <small class="mono">{{ incident.id }}</small>
                    </td>
                    <td>
                      <CountryPair :origin="corridorParts(incident.corridor).origin" :destination="corridorParts(incident.corridor).destination" compact />
                    </td>
                    <td>{{ incident.affectedTransactions }} / {{ incident.affectedValue }}</td>
                    <td>{{ incident.owner }}</td>
                    <td><HealthBadge :state="incident.severity" :trigger="incident.status" /></td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>

          <Panel title="Event timeline" eyebrow="Recent evidence" accent="degraded" class="span-7">
            <ol class="event-list event-list--timeline">
              <li v-for="event in dashboard.downtimeEvents" :key="`${event.time}-${event.title}`">
                <time>{{ event.time }}</time>
                <div>
                  <strong>{{ event.title }}</strong>
                  <span>{{ event.actor }}</span>
                  <p>{{ event.detail }}</p>
                </div>
                <HealthBadge :state="event.state" />
              </li>
            </ol>
          </Panel>

          <Panel title="Latency breakdown" eyebrow="Slowest route" accent="degraded" class="span-5">
            <div class="latency-list">
              <article v-for="step in dashboard.latency.steps" :key="step.label">
                <div>
                  <strong>{{ step.label }}</strong>
                  <small>{{ step.owner }}</small>
                </div>
                <span class="progress-track">
                  <i :class="`progress-fill--${step.state}`" :style="{ width: percentWidth((step.durationMs / Math.max(step.targetMs, 1)) * 40) }"></i>
                </span>
                <HealthBadge :state="step.state" :trigger="`${Math.round(step.durationMs / 1000)}s`" />
              </article>
            </div>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'fx'" class="screen-stack">
        <section class="dashboard-grid">
          <Panel title="Recommended route" eyebrow="Rates and costs" accent="healthy" class="span-4">
            <ProviderMark :provider="dashboard.fxCostBoard.recommendedProvider" show-category size="lg" />
            <p>{{ dashboard.fxCostBoard.decision }}</p>
            <div class="fx-baseline">
              <strong>{{ fxBaseCurrency }} baseline</strong>
              <small>Compare provider quotes against {{ fxBaseCurrency }}/{{ fxTargetCurrency }}.</small>
            </div>
            <div class="fx-control-grid">
              <label>
                <span>Base</span>
                <select v-model="fxBaseCurrency" aria-label="Base currency">
                  <option v-for="currency in fxCurrencies" :key="currency" :value="currency">{{ currency }}</option>
                </select>
              </label>
              <label>
                <span>Compare</span>
                <select v-model="fxTargetCurrency" aria-label="Comparison currency">
                  <option v-for="currency in fxCurrencies" :key="currency" :value="currency">{{ currency }}</option>
                </select>
              </label>
            </div>
            <dl class="metric-grid">
              <div>
                <dt>Pair</dt>
                <dd>{{ fxBaseCurrency }}/{{ fxTargetCurrency }}</dd>
              </div>
              <div>
                <dt>Window</dt>
                <dd>{{ dashboard.fxCostBoard.window }}</dd>
              </div>
              <div>
                <dt>Updated</dt>
                <dd>{{ dashboard.fxCostBoard.refreshedAt }}</dd>
              </div>
            </dl>
          </Panel>

          <Panel title="Provider comparison" :eyebrow="`${fxBaseCurrency} baseline`" accent="watch" class="span-8">
            <DataTable :empty="fxComparisonRoutes.length === 0" empty-title="No rates available" empty-description="Rates will appear when provider quotes refresh.">
              <table>
                <thead>
                  <tr>
                    <th>Provider</th>
                    <th>{{ fxBaseCurrency }} baseline</th>
                    <th>Fee / spread</th>
                    <th>Effective cost</th>
                    <th>Speed</th>
                    <th>Signal</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="route in fxComparisonRoutes" :key="route.provider">
                    <td>
                      <ProviderMark :provider="route.provider" />
                      <small>{{ route.route }}</small>
                    </td>
                    <td>
                      <strong>{{ route.baselineRate }}</strong>
                      <small>{{ route.pair }} / {{ route.updatedAt }}</small>
                    </td>
                    <td>{{ route.fee }} / {{ route.spread }}</td>
                    <td><strong>{{ route.effectiveCost }}</strong></td>
                    <td>{{ route.payoutTime }}</td>
                    <td>
                      <HealthBadge :state="route.state" :trigger="route.recommended ? 'recommended' : route.cheapest ? 'cheapest' : route.note" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'reconciliation'" class="screen-stack">
        <Panel title="Settlement breaks" eyebrow="Reconcile" accent="watch">
          <DataTable
            :empty="dashboard.reconciliation.length === 0"
            empty-title="No settlement breaks"
            empty-description="Provider files, bank postings, and settlement references are currently matched."
          >
            <table>
              <thead>
                <tr>
                  <th>Reference</th>
                  <th>Provider</th>
                  <th>Amount</th>
                  <th>Age</th>
                  <th>Reason</th>
                  <th>Next team</th>
                  <th>State</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in dashboard.reconciliation" :key="item.reference">
                  <td class="mono">{{ item.reference }}</td>
                  <td><ProviderMark :provider="item.provider" /></td>
                  <td><strong>{{ item.amount }}</strong></td>
                  <td>{{ item.age }}</td>
                  <td>{{ item.reason }}</td>
                  <td>{{ item.owner }}</td>
                  <td><HealthBadge :state="item.state" /></td>
                </tr>
              </tbody>
            </table>
          </DataTable>
        </Panel>
      </section>

      <section v-else-if="activeScreen === 'providers'" class="screen-stack">
        <section class="kpi-grid">
          <KpiTile label="Leader" :value="sortedProviders[0]?.provider ?? 'None'" :detail="`${sortedProviders[0]?.p95 ?? '-'} P95`" tone="healthy" :icon="CheckCircle2" />
          <KpiTile label="Weakest route" :value="weakestProvider?.provider ?? 'None'" :detail="`${weakestProvider?.settlementExceptions ?? 0} exceptions`" tone="degraded" :icon="AlertTriangle" />
          <KpiTile label="Total providers" :value="sortedProviders.length" detail="Connected in pilot snapshot" tone="brand" :icon="Network" />
          <KpiTile label="Traffic window" value="15m" detail="Success, P95, stuck rate" tone="recovery" :icon="TimerReset" />
        </section>
        <Panel title="Provider health dashboard" eyebrow="Primary view" accent="healthy">
          <DataTable :empty="sortedProviders.length === 0" empty-title="No providers" empty-description="Connect provider routes to start measuring performance.">
            <table>
              <thead>
                <tr>
                  <th>Rank</th>
                  <th>Provider</th>
                  <th>Corridor</th>
                  <th>Success</th>
                  <th>P50 / P95 / P99</th>
                  <th>Traffic</th>
                  <th>Reconciliation</th>
                  <th>State</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="provider in sortedProviders" :key="provider.provider">
                  <td class="rank">#{{ provider.rank }}</td>
                  <td><ProviderMark :provider="provider.provider" show-category /></td>
                  <td>
                    <CountryPair :origin="corridorParts(provider.corridor).origin" :destination="corridorParts(provider.corridor).destination" compact />
                  </td>
                  <td><strong>{{ provider.successRate }}</strong></td>
                  <td>{{ provider.p50 }} / {{ provider.p95 }} / {{ provider.p99 }}</td>
                  <td>{{ provider.trafficShare }}</td>
                  <td>{{ provider.settlementExceptions }} breaks</td>
                  <td><HealthBadge :state="provider.state" /></td>
                </tr>
              </tbody>
            </table>
          </DataTable>
        </Panel>
      </section>

      <section v-else-if="activeScreen === 'audit'" class="screen-stack">
        <Panel title="Decision log" eyebrow="Audit">
          <ActionBar>
            <label class="search-field">
              <Search :size="18" aria-hidden="true" />
              <input v-model="auditQuery" type="search" placeholder="Actor, object, reason, action" aria-label="Search audit events" />
            </label>
            <span class="dashboard-chip">Read-only log</span>
          </ActionBar>
        </Panel>

        <section class="audit-layout">
          <Panel title="Events" :eyebrow="`${filteredAuditEvents.length} records`" class="audit-list-panel">
            <DataTable :empty="filteredAuditEvents.length === 0" empty-title="No audit records" empty-description="No logged decisions match this search.">
              <table>
                <thead>
                  <tr>
                    <th>Time</th>
                    <th>Action</th>
                    <th>Actor</th>
                    <th>Object</th>
                    <th>State</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="event in filteredAuditEvents"
                    :key="`${event.time}-${event.action}`"
                    class="click-row"
                    :class="{ 'is-selected': selectedAudit?.time === event.time }"
                    @click="selectedAuditTime = event.time"
                  >
                    <td class="mono">{{ event.time }}</td>
                    <td><strong>{{ event.action }}</strong></td>
                    <td>{{ event.actor }}</td>
                    <td>{{ event.object }}</td>
                    <td><HealthBadge :state="event.state" /></td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>

          <Panel title="Log detail" eyebrow="Selected record" class="audit-detail-panel" :accent="selectedAudit?.state ?? 'unknown'">
            <EmptyState v-if="!selectedAudit" title="No record selected" description="Choose an audit event to inspect its reason and object." :icon="BadgeCheck" />
            <div v-else class="audit-detail">
              <BadgeCheck :size="22" aria-hidden="true" />
              <h3>{{ selectedAudit.action }}</h3>
              <dl class="definition-list">
                <div>
                  <dt>Time</dt>
                  <dd class="mono">{{ selectedAudit.time }}</dd>
                </div>
                <div>
                  <dt>Actor</dt>
                  <dd>{{ selectedAudit.actor }}</dd>
                </div>
                <div>
                  <dt>Object</dt>
                  <dd>{{ selectedAudit.object }}</dd>
                </div>
                <div>
                  <dt>Reason</dt>
                  <dd>{{ selectedAudit.reason }}</dd>
                </div>
              </dl>
              <aside class="state-note">
                <GitBranch :size="16" aria-hidden="true" />
                <span>Policy snapshot and route evidence are captured in the decision log.</span>
              </aside>
            </div>
          </Panel>
        </section>
      </section>
    </main>
  </div>
</template>
