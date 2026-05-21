<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
  Download,
  FileCheck2,
  Gauge,
  GitBranch,
  History,
  Maximize2,
  Network,
  PauseCircle,
  PlayCircle,
  Plus,
  ReceiptText,
  Search,
  Send,
  ShieldCheck,
  SlidersHorizontal,
  TimerReset,
  Upload,
  X,
} from '@lucide/vue'
import ActionBar from './components/ActionBar.vue'
import CountryPair from './components/CountryPair.vue'
import DataFreshness from './components/DataFreshness.vue'
import DataTable from './components/DataTable.vue'
import EmptyState from './components/EmptyState.vue'
import HealthBadge from './components/HealthBadge.vue'
import KpiTile from './components/KpiTile.vue'
import LoginPanel from './components/LoginPanel.vue'
import PageHeader from './components/PageHeader.vue'
import Panel from './components/Panel.vue'
import ProviderMark from './components/ProviderMark.vue'
import RealtimeLineChart from './components/RealtimeLineChart.vue'
import RoutePath from './components/RoutePath.vue'
import RouteScoreChip from './components/RouteScoreChip.vue'
import StateBanner from './components/StateBanner.vue'
import TransactionTimeline from './components/TransactionTimeline.vue'
import TraceEmptyIllustration from './components/TraceEmptyIllustration.vue'
import UiButton from './components/UiButton.vue'
import { currentUser, hasPermission, loginLDAP, loginLocal, logout } from './services/authApi'
import { connectDashboardLive, getDashboardSummary, getDashboardTimeseries, type DashboardQuery } from './services/dashboardApi'
import { getDashboardMock } from './services/mockDashboard'
import { getCountryIdentity, parseCorridor } from './services/identity'
import { routeForScreen } from './router'
import type {
  DashboardSummaryResponse,
  DashboardTimeseriesPoint,
  HealthState,
  Permission,
  ProviderToggle,
  ScreenId,
  ScoringWeight,
  SessionUser,
  TransactionRecord,
  UiScenario,
} from './types'

type PolicyStatus = 'draft' | 'pending_approval' | 'active' | 'inactive'

type PolicyRule = {
  id: string
  name: string
  status: PolicyStatus
  origin: string
  destination: string
  payoutMethod: string
  provider: string
  fallback: string[]
  amountBand: string
  drafter: string
  approver?: string
  version: string
  updatedAt: string
}

const route = useRoute()
const router = useRouter()
const dashboard = reactive(getDashboardMock(normalizeDataState(route.query.scenario)))

const sessionUser = ref<SessionUser | null>(null)
const authReady = ref(false)
const authBusy = ref(false)
const authError = ref('')
const dashboardSummary = ref<DashboardSummaryResponse | null>(null)
const dashboardTimeseries = ref<DashboardTimeseriesPoint[]>([])
const currentActor = computed(() => sessionUser.value?.display_name ?? 'Operations Admin')
let disconnectLive: (() => void) | null = null

const activeScreen = computed<ScreenId>(() => (route.meta.screen as ScreenId | undefined) ?? 'control')
const dateRange = ref(String(route.query.range ?? 'today'))
const dashboardCurrency = ref(String(route.query.currency ?? 'USD'))
const selectedProviderId = ref(String(route.query.provider_id ?? 'all'))
const selectedCorridor = ref(String(route.query.corridor ?? 'all'))
const selectedPayoutMethod = ref(String(route.query.payout_method ?? 'all'))
const analysisLens = ref(String(route.query.analysis_lens ?? 'reliability'))
const dataState = ref<UiScenario>(normalizeDataState(route.query.scenario))
const qaThresholdSeconds = ref(dashboard.qaPolicy.thresholdSeconds)
const warningThresholdSeconds = ref(dashboard.qaPolicy.warningSeconds)
const savedQaThresholdSeconds = ref(dashboard.qaPolicy.thresholdSeconds)
const savedWarningThresholdSeconds = ref(dashboard.qaPolicy.warningSeconds)

const transactionQuery = ref('')
const selectedTransactionReference = ref('')
const senderFilter = ref('All senders')
const destinationFilter = ref('All destinations')
const currencyFilter = ref('All currencies')
const timeFilter = ref(String(route.query.timing ?? 'All timing'))
const destinationTypeFilter = ref('All destination types')
const sortBy = ref('totalTimeDesc')
const transactionPage = ref(1)
const transactionPageSize = ref(10)
const traceSheetOpen = ref(false)
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
const routePage = ref(1)
const routePageSize = ref(3)

const policyRules = ref<PolicyRule[]>([
  {
    id: 'POL-EU-NG-001',
    name: 'EU to Nigeria account payouts',
    status: 'active',
    origin: 'European Union',
    destination: 'Nigeria',
    payoutMethod: 'Bank account',
    provider: 'Thunes',
    fallback: ['Remitly', 'Ria'],
    amountBand: 'EUR 0 - 5,000',
    drafter: 'Treasury Lead',
    approver: 'Bank Admin',
    version: 'v2026.05.20.14.05',
    updatedAt: '14:05 UTC',
  },
  {
    id: 'POL-GB-NG-002',
    name: 'UK to Nigeria high-confidence fallback',
    status: 'pending_approval',
    origin: 'United Kingdom',
    destination: 'Nigeria',
    payoutMethod: 'Bank account',
    provider: 'Remitly',
    fallback: ['Thunes', 'Ria'],
    amountBand: 'GBP 0 - 3,000',
    drafter: 'Ops Lead',
    version: 'draft-v2026.05.20.14.21',
    updatedAt: '14:21 UTC',
  },
  {
    id: 'POL-KE-NG-003',
    name: 'Kenya to Nigeria PAPSS recovery',
    status: 'inactive',
    origin: 'Kenya',
    destination: 'Nigeria',
    payoutMethod: 'Local account',
    provider: 'PAPSS',
    fallback: ['Thunes'],
    amountBand: 'KES 0 - 500,000',
    drafter: 'Route Desk',
    approver: 'Ops Lead',
    version: 'v2026.05.20.13.20',
    updatedAt: '13:20 UTC',
  },
])
const selectedPolicyId = ref(policyRules.value[0]?.id ?? '')
const policyDraftForm = reactive({
  name: 'New corridor rule',
  origin: 'United States',
  destination: 'Nigeria',
  payoutMethod: 'Bank account',
  provider: 'Thunes',
  fallback: 'Remitly, Ria',
  amountBand: 'Any amount',
})
const bankWorkspaceName = 'Taj Bank'

const navigation = [
  { id: 'control' as ScreenId, label: 'Control Room', icon: Gauge, kicker: 'Volume and risk' },
  { id: 'transactions' as ScreenId, label: 'Transactions', icon: Search, kicker: 'Trace and reports' },
  { id: 'corridors' as ScreenId, label: 'Routes', icon: Network, kicker: 'Corridors and rails' },
  { id: 'policy' as ScreenId, label: 'Policy', icon: SlidersHorizontal, kicker: 'Rules and approvals' },
  { id: 'incidents' as ScreenId, label: 'Incidents', icon: BellRing, kicker: 'Root cause' },
  { id: 'fx' as ScreenId, label: 'Rates & costs', icon: CircleDollarSign, kicker: 'Route economics' },
  { id: 'reconciliation' as ScreenId, label: 'Reconcile', icon: ReceiptText, kicker: 'Settlement work' },
  { id: 'providers' as ScreenId, label: 'Providers', icon: BarChart3, kicker: 'SLA and exceptions' },
  { id: 'audit' as ScreenId, label: 'Audit', icon: History, kicker: 'Evidence' },
]

const navigationGroups = [
  { label: 'Operate', items: navigation.slice(0, 4) },
  { label: 'Review', items: navigation.slice(4) },
]

const screenDescriptions: Record<ScreenId, string> = {
  control: 'Monitor route health, transfer volume, and open exceptions.',
  corridors: 'Compare corridor rails and open the right drilldown.',
  transactions: 'Search transfers, inspect status, and export reports.',
  incidents: 'Track degradation, root cause, and recovery owner.',
  policy: 'Review route rules and maker-checker status.',
  fx: 'Compare effective cost against route eligibility.',
  reconciliation: 'Clear settlement breaks by reason and owner.',
  providers: 'Review provider SLA, latency, volume, and exceptions.',
  audit: 'Inspect policy, routing, and operator events.',
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
const isPolicyCreateFlow = computed(() => activeScreen.value === 'policy' && route.path === '/policy/new')
const isRouteDetailFlow = computed(() => activeScreen.value === 'corridors' && typeof route.params.routeId === 'string')
const routeDetailId = computed(() => (typeof route.params.routeId === 'string' ? route.params.routeId : ''))
const currentPageTitle = computed(() => {
  if (isPolicyCreateFlow.value) return 'New policy'
  if (isRouteDetailFlow.value) return 'Route detail'
  return selectedScreen.value.label
})
const currentPageDescription = computed(() =>
  isPolicyCreateFlow.value
    ? 'Define corridor, payout method, primary rail, fallback order, and amount band.'
    : isRouteDetailFlow.value && selectedRouteDetail.value
      ? `${friendlyCorridorLabel(selectedRouteDetail.value.corridor.corridor)} / ${selectedRouteDetail.value.corridor.selectedRoute}`
      : screenDescriptions[selectedScreen.value.id],
)
const breadcrumbs = computed(() => {
  const crumbs: Array<{ label: string; path?: string }> = [{ label: bankWorkspaceName, path: '/' }]
  crumbs.push({ label: selectedScreen.value.label, path: routeForScreen(selectedScreen.value.id).path })
  if (isRouteDetailFlow.value) crumbs.push({ label: selectedRouteDetail.value ? friendlyCorridorLabel(selectedRouteDetail.value.corridor.corridor) : 'Route detail' })
  if (isPolicyCreateFlow.value) crumbs.push({ label: 'New policy' })
  return crumbs
})
const activeIncident = computed(() => dashboard.incidents[0] ?? null)
const sortedCorridors = computed(() => [...dashboard.corridors].sort((a, b) => severityRank[a.state] - severityRank[b.state]))
const sortedProviders = computed(() => [...dashboard.providerScores].sort((a, b) => a.rank - b.rank))
const routeDetailRows = computed(() =>
  sortedCorridors.value.map((corridor) => ({
    id: routeDetailIdFor(corridor),
    corridor,
    provider: routeProvider(corridor.selectedRoute),
    rail: routeRail(corridor.selectedRoute),
  })),
)
const selectedRouteDetail = computed(() => routeDetailRows.value.find((item) => item.id === routeDetailId.value) ?? null)
const selectedRouteProviderScore = computed(() => {
  const detail = selectedRouteDetail.value
  if (!detail) return null
  return sortedProviders.value.find((provider) => provider.provider.toLowerCase() === detail.provider.toLowerCase()) ?? null
})
const selectedRouteFx = computed(() => {
  const detail = selectedRouteDetail.value
  if (!detail) return null
  return dashboard.fxCostBoard.routes.find((route) => route.provider.toLowerCase() === detail.provider.toLowerCase()) ?? null
})
const selectedRouteIncidents = computed(() => {
  const detail = selectedRouteDetail.value
  if (!detail) return []
  const corridorKey = normalizeCorridorValue(detail.corridor.corridor)
  return dashboard.incidents.filter((incident) => normalizeCorridorValue(incident.corridor) === corridorKey)
})
const selectedRouteTransactions = computed(() => {
  const detail = selectedRouteDetail.value
  if (!detail) return []
  const provider = detail.provider.toLowerCase()
  const destination = getCountryIdentity(corridorParts(detail.corridor.corridor).destination).code
  const origin = getCountryIdentity(corridorParts(detail.corridor.corridor).origin).code
  return dashboard.transactions.filter((transaction) => {
    const transactionProvider = routeProvider(transaction.route).toLowerCase()
    const transactionDestination = getCountryIdentity(transaction.destinationCountry).code
    const transactionOrigin = getCountryIdentity(transaction.senderCountry).code
    const originMatches = origin === 'EU' ? ['DE', 'FR', 'ES', 'IT', 'NL', 'EU'].includes(transactionOrigin) : transactionOrigin === origin
    return transactionProvider === provider && transactionDestination === destination && originMatches
  })
})
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
const dashboardCurrencies = ['USD', 'NGN', 'EUR', 'GBP', 'KES']
const dashboardRanges = [
  { label: 'Today', value: 'today' },
  { label: '24 hours', value: '24h' },
  { label: '7 days', value: '7d' },
  { label: '30 days', value: '30d' },
]
const analysisLenses = [
  { label: 'Reliability', value: 'reliability' },
  { label: 'SLA', value: 'sla' },
  { label: 'Volume', value: 'volume' },
  { label: 'Cost', value: 'cost' },
]
const dataStateOptions: Array<{ label: string; value: UiScenario }> = [
  { label: 'Current pressure', value: 'degraded-ria' },
  { label: 'Normal operation', value: 'healthy' },
  { label: 'Shift in progress', value: 'traffic-shift' },
  { label: '30-day review', value: 'pilot-report' },
]
const providerOptions = computed(() => [
  { label: 'All IMTOs', value: 'all' },
  ...unique(dashboard.providerScores.map((provider) => provider.provider)).map((provider) => ({ label: provider, value: provider.toLowerCase() })),
])
const corridorOptions = computed(() => [
  { label: 'All corridors', value: 'all' },
  ...unique(dashboard.corridors.map((corridor) => corridor.corridor)).map((corridor) => ({
    label: friendlyCorridorLabel(corridor),
    value: normalizeCorridorValue(corridor),
  })),
])
const payoutOptions = [
  { label: 'All payouts', value: 'all' },
  { label: 'Bank account', value: 'bank_account' },
  { label: 'Cash pickup', value: 'cash_pickup' },
  { label: 'Wallet', value: 'wallet' },
]
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

const transactionPageCount = computed(() => Math.max(1, Math.ceil(filteredTransactions.value.length / transactionPageSize.value)))
const pagedTransactions = computed(() => {
  const start = (transactionPage.value - 1) * transactionPageSize.value
  return filteredTransactions.value.slice(start, start + transactionPageSize.value)
})
const transactionStartIndex = computed(() => (filteredTransactions.value.length ? (transactionPage.value - 1) * transactionPageSize.value + 1 : 0))
const transactionEndIndex = computed(() => Math.min(filteredTransactions.value.length, transactionPage.value * transactionPageSize.value))

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

const apiMetricTiles = computed(() => dashboardSummary.value?.tiles ?? [])
const providerComparisons = computed(() => dashboardSummary.value?.providers ?? [])
const volumeAnalytics = computed(() => {
  const analytics = dashboardSummary.value?.analytics
  const totalVolume = analytics?.processed_volume ?? dashboardTimeseries.value.reduce((sum, point) => sum + point.volume, 0)
  const processedCount = analytics?.processed_count ?? dashboardTimeseries.value.reduce((sum, point) => sum + point.processed_count, 0)
  const latestVolume = dashboardTimeseries.value[dashboardTimeseries.value.length - 1]?.volume ?? 0
  return {
    totalVolume,
    processedCount,
    latestVolume,
    totalLabel: formatCurrencyShort(totalVolume, dashboardCurrency.value),
    latestLabel: `${formatCurrencyShort(latestVolume, dashboardCurrency.value)} latest interval`,
    countLabel: `${Math.round(processedCount).toLocaleString()} transfers`,
  }
})
const providerVolumeRows = computed(() => {
  const rows = (providerComparisons.value.length
    ? providerComparisons.value.map((provider) => ({
        id: provider.provider_id,
        provider: provider.provider_name,
        corridor: provider.corridor,
        volume: provider.processed_volume,
        count: provider.processed_count,
        state: provider.state,
      }))
    : sortedProviders.value.map((provider) => {
        const share = parsePercent(provider.trafficShare)
        return {
          id: provider.provider.toLowerCase(),
          provider: provider.provider,
          corridor: provider.corridor,
          volume: volumeAnalytics.value.totalVolume * (share / 100),
          count: volumeAnalytics.value.processedCount * (share / 100),
          state: provider.state,
        }
      }))
    .sort((a, b) => b.volume - a.volume)
    .slice(0, 4)
  const maxVolume = Math.max(...rows.map((row) => row.volume), 1)
  return rows.map((row) => ({
    ...row,
    volumeLabel: formatCurrencyShort(row.volume, dashboardCurrency.value),
    countLabel: `${Math.round(row.count).toLocaleString()} transfers`,
    width: percentWidth((row.volume / maxVolume) * 100),
  }))
})
const routeVolumeRows = computed(() => {
  const fallbackShares = [34, 27, 21, 18]
  const rows = sortedCorridors.value
    .map((corridor, index) => {
      const provider = routeProvider(corridor.selectedRoute)
      const providerScore = sortedProviders.value.find((score) => score.provider.toLowerCase() === provider.toLowerCase())
      const providerComparison = providerComparisons.value.find((item) => item.provider_name.toLowerCase() === provider.toLowerCase())
      const share = providerScore ? parsePercent(providerScore.trafficShare) : fallbackShares[index] ?? 10
      const volume = providerComparison?.processed_volume ?? volumeAnalytics.value.totalVolume * (share / 100)
      return {
        id: `${corridor.corridor}-${corridor.selectedRoute}`,
        corridor: corridor.corridor,
        provider,
        route: corridor.selectedRoute,
        risk: corridor.risk,
        state: corridor.state,
        volume,
      }
    })
    .sort((a, b) => b.volume - a.volume)
    .slice(0, 4)
  const maxVolume = Math.max(...rows.map((row) => row.volume), 1)
  return rows.map((row) => ({
    ...row,
    volumeLabel: formatCurrencyShort(row.volume, dashboardCurrency.value),
    width: percentWidth((row.volume / maxVolume) * 100),
  }))
})
const operationsSnapshot = computed(() => {
  const route = primaryIncidentRoute.value
  const incident = activeIncident.value
  const settlementBreak = dashboard.reconciliation[0]
  return [
    {
      label: 'Bottleneck',
      value: incident?.rootCause ?? 'No active bottleneck',
      detail: incident?.owner ?? 'Routes inside policy',
      state: incident?.severity ?? ('healthy' as HealthState),
    },
    {
      label: 'Route to inspect',
      value: route?.selectedRoute ?? 'No route selected',
      detail: route ? `${friendlyCorridorLabel(route.corridor)} / ${route.risk}` : 'No route risk',
      state: route?.state ?? ('healthy' as HealthState),
    },
    {
      label: 'Latency owner',
      value: incident?.owner ?? 'No active owner',
      detail: incident ? `${incident.startedAt} / ${incident.status}` : 'No escalated latency work',
      state: incident?.severity ?? ('healthy' as HealthState),
    },
    {
      label: 'Settlement work',
      value: `${dashboard.reconciliation.length} open`,
      detail: settlementBreak ? `${settlementBreak.reason} / ${settlementBreak.owner}` : 'No settlement break',
      state: dashboard.reconciliation.length ? ('watch' as HealthState) : ('healthy' as HealthState),
    },
  ]
})
const routePageCount = computed(() => Math.max(1, Math.ceil(sortedCorridors.value.length / routePageSize.value)))
const pagedCorridors = computed(() => {
  const start = (routePage.value - 1) * routePageSize.value
  return sortedCorridors.value.slice(start, start + routePageSize.value)
})
const routeActionItems = computed(() =>
  sortedCorridors.value.map((corridor) => ({
    id: `${corridor.corridor}-${corridor.selectedRoute}`,
    corridor,
    action:
      corridor.state === 'healthy'
        ? 'Route more'
        : corridor.state === 'recovery'
          ? 'Hold recovery split'
          : corridor.state === 'watch'
            ? 'Watch callbacks'
            : 'Open policy review',
    })),
)
const pagedRouteActionItems = computed(() => {
  const start = (routePage.value - 1) * routePageSize.value
  return routeActionItems.value.slice(start, start + routePageSize.value)
})
const selectedPolicy = computed(() => policyRules.value.find((policy) => policy.id === selectedPolicyId.value) ?? policyRules.value[0] ?? null)
const policyStatusCounts = computed(() =>
  policyRules.value.reduce(
    (counts, policy) => {
      counts[policy.status] += 1
      return counts
    },
    { draft: 0, pending_approval: 0, active: 0, inactive: 0 } as Record<PolicyStatus, number>,
  ),
)
const selectedPolicyNeedsDifferentApprover = computed(() => Boolean(selectedPolicy.value && selectedPolicy.value.status === 'pending_approval' && selectedPolicy.value.drafter === currentActor.value))

const dashboardQuery = computed<DashboardQuery>(() => ({
  range: dateRange.value === 'Today' ? 'today' : dateRange.value,
  provider_id: selectedProviderId.value === 'all' ? undefined : selectedProviderId.value,
  corridor: selectedCorridor.value === 'all' ? undefined : selectedCorridor.value,
  payout_method: selectedPayoutMethod.value === 'all' ? undefined : selectedPayoutMethod.value,
  currency: dashboardCurrency.value,
  analysis_lens: analysisLens.value,
  scenario: dataState.value,
}))

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

const globalDashboardQuery = () => ({
  range: dateRange.value,
  currency: dashboardCurrency.value,
  provider_id: selectedProviderId.value,
  corridor: selectedCorridor.value,
  payout_method: selectedPayoutMethod.value,
  analysis_lens: analysisLens.value,
  scenario: dataState.value,
})

const pageQueryKeys: Record<ScreenId, string[]> = {
  control: [],
  corridors: ['focus'],
  transactions: ['timing', 'metric'],
  incidents: ['focus'],
  policy: ['policy_id'],
  fx: [],
  reconciliation: [],
  providers: [],
  audit: [],
}

const currentPageQuery = (screen: ScreenId) =>
  Object.fromEntries(pageQueryKeys[screen].filter((key) => typeof route.query[key] === 'string').map((key) => [key, String(route.query[key])]))

const activate = (screen: ScreenId, extraQuery: Record<string, string> = {}) => {
  void router.push({ path: routeForScreen(screen).path, query: { ...globalDashboardQuery(), ...extraQuery } })
}

const openPolicyCreateFlow = () => {
  void router.push({ path: '/policy/new', query: globalDashboardQuery() })
}

const openRouteDetail = (corridor: { corridor: string; selectedRoute: string }) => {
  void router.push({ path: `/routes/${routeDetailIdFor(corridor)}`, query: globalDashboardQuery() })
}

const syncDashboardQuery = () => {
  void router.replace({
    path: route.path,
    query: {
      ...globalDashboardQuery(),
      ...currentPageQuery(activeScreen.value),
    },
  })
}

const usingSampleData = import.meta.env.MODE === 'test' || import.meta.env.VITE_IMSI_DATA_MODE === 'mock'

const applyDashboardState = (scenario: UiScenario) => {
  const next = getDashboardMock(scenario)
  Object.assign(dashboard, next)
  selectedAuditTime.value = next.auditEvents[0]?.time ?? ''
  if (selectedTransactionReference.value && !next.transactions.some((transaction) => transaction.reference === selectedTransactionReference.value)) {
    selectedTransactionReference.value = ''
  }
}

const refreshDashboard = async () => {
  if (!sessionUser.value) return
  try {
    const [summary, timeseries] = await Promise.all([getDashboardSummary(dashboardQuery.value), getDashboardTimeseries(dashboardQuery.value)])
    dashboardSummary.value = summary
    dashboardTimeseries.value = timeseries
    dashboard.summary.lastUpdated = new Date(summary.generated_at).toLocaleTimeString('en-GB', { timeZone: 'UTC' }) + ' UTC'
    dashboard.summary.connection.mode = usingSampleData ? 'static' : 'api'
    dashboard.summary.connection.freshness = 'fresh'
    dashboard.summary.connection.updatedAt = summary.generated_at
    dashboard.summary.connection.nextPollIn = usingSampleData ? 'Static operational data' : 'API aggregation'
  } catch {
    dashboard.summary.connection.mode = sessionUser.value ? 'api' : dashboard.summary.connection.mode
    dashboard.summary.connection.freshness = 'stale'
    dashboard.summary.connection.nextPollIn = 'retry on filter change'
  }
}

const reconnectLive = () => {
  disconnectLive?.()
  disconnectLive = connectDashboardLive(dashboardQuery.value, (summary) => {
    dashboardSummary.value = summary
    dashboard.summary.lastUpdated = new Date(summary.generated_at).toLocaleTimeString('en-GB', { timeZone: 'UTC' }) + ' UTC'
    dashboard.summary.connection.mode = 'sse'
    dashboard.summary.connection.freshness = 'fresh'
    dashboard.summary.connection.updatedAt = summary.generated_at
    dashboard.summary.connection.nextPollIn = 'live stream'
  })
}

const handleLogin = async (payload: { mode: 'local' | 'ldap'; bankId: string; username: string; password: string }) => {
  authBusy.value = true
  authError.value = ''
  try {
    sessionUser.value =
      payload.mode === 'ldap'
        ? await loginLDAP(payload.bankId, payload.username, payload.password)
        : await loginLocal(payload.bankId, payload.username, payload.password)
    await refreshDashboard()
    reconnectLive()
  } catch (error) {
    authError.value = error instanceof Error ? error.message : 'Login failed'
  } finally {
    authBusy.value = false
  }
}

const handleLogout = async () => {
  await logout()
  sessionUser.value = null
}

const can = (permission: Permission) => hasPermission(sessionUser.value, permission)

const openDrilldown = (drilldown: string) => {
  const url = new URL(drilldown, window.location.origin)
  void router.push({
    path: url.pathname,
    query: {
      ...Object.fromEntries(url.searchParams.entries()),
      ...globalDashboardQuery(),
    },
  })
}

const selectTransaction = (transaction: TransactionRecord) => {
  selectedTransactionReference.value = transaction.reference
}

const openRouteTransaction = (transaction: TransactionRecord, corridor: string) => {
  selectTransaction(transaction)
  activate('transactions', { corridor: normalizeCorridorValue(corridor) })
}

const goToTransactionPage = (page: number) => {
  transactionPage.value = Math.min(transactionPageCount.value, Math.max(1, page))
}

const goToRoutePage = (page: number) => {
  routePage.value = Math.min(routePageCount.value, Math.max(1, page))
}

const openTraceSheet = () => {
  if (selectedTransaction.value) traceSheetOpen.value = true
}

const selectPolicy = (policy: PolicyRule) => {
  selectedPolicyId.value = policy.id
  void router.replace({ path: routeForScreen('policy').path, query: { ...globalDashboardQuery(), policy_id: policy.id } })
}

const csvEscape = (value: string | number) => {
  const text = String(value)
  return /[",\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
}

const downloadTransactionReport = () => {
  const headers = [
    'switch_reference',
    'provider_reference',
    'bank_reference',
    'provider',
    'route',
    'sender_country',
    'destination_country',
    'amount',
    'beneficiary',
    'status',
    'total_time_seconds',
    'qa_limit_seconds',
    'current_owner',
    'blocker',
  ]
  const rows = filteredTransactions.value.map((transaction) => [
    transaction.reference,
    transaction.providerReference,
    transaction.bankReference,
    transaction.provider,
    transaction.route,
    transaction.senderCountry,
    transaction.destinationCountry,
    transaction.amount,
    transaction.beneficiary,
    qaStatusLabel(transaction),
    transaction.totalTimeSeconds,
    transaction.qaLimitSeconds,
    transaction.currentOwner,
    transaction.blocker,
  ])
  const csv = [headers, ...rows].map((row) => row.map(csvEscape).join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `imsi-transactions-${dateRange.value}-${dashboardCurrency.value}.csv`
  link.click()
  URL.revokeObjectURL(url)
}

const createPolicyDraft = () => {
  const nextId = `POL-${String(policyRules.value.length + 1).padStart(3, '0')}`
  const draft: PolicyRule = {
    id: nextId,
    name: policyDraftForm.name.trim() || 'Untitled corridor rule',
    status: 'draft',
    origin: policyDraftForm.origin,
    destination: policyDraftForm.destination,
    payoutMethod: policyDraftForm.payoutMethod,
    provider: policyDraftForm.provider,
    fallback: policyDraftForm.fallback
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean),
    amountBand: policyDraftForm.amountBand,
    drafter: currentActor.value,
    version: `draft-${new Date().toISOString().slice(0, 16)}`,
    updatedAt: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }) + ' UTC',
  }
  policyRules.value = [draft, ...policyRules.value]
  selectedPolicyId.value = draft.id
  void router.push({ path: routeForScreen('policy').path, query: { ...globalDashboardQuery(), policy_id: draft.id } })
}

const submitSelectedPolicy = () => {
  if (!selectedPolicy.value || selectedPolicy.value.status !== 'draft' || !can('policy:draft')) return
  selectedPolicy.value.status = 'pending_approval'
  selectedPolicy.value.updatedAt = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }) + ' UTC'
}

const approveSelectedPolicy = () => {
  if (!selectedPolicy.value || selectedPolicy.value.status !== 'pending_approval' || selectedPolicyNeedsDifferentApprover.value || !can('policy:approve')) return
  selectedPolicy.value.approver = currentActor.value
  selectedPolicy.value.status = 'inactive'
  selectedPolicy.value.version = selectedPolicy.value.version.replace('draft-', 'approved-')
  selectedPolicy.value.updatedAt = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }) + ' UTC'
}

const activateSelectedPolicy = () => {
  if (!selectedPolicy.value || selectedPolicy.value.status === 'draft' || selectedPolicy.value.status === 'pending_approval' || !can('policy:activate')) return
  selectedPolicy.value.status = 'active'
  selectedPolicy.value.updatedAt = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }) + ' UTC'
}

const deactivateSelectedPolicy = () => {
  if (!selectedPolicy.value || selectedPolicy.value.status !== 'active' || !can('policy:activate')) return
  selectedPolicy.value.status = 'inactive'
  selectedPolicy.value.updatedAt = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }) + ' UTC'
}

const resetTransactionFilters = () => {
  transactionQuery.value = ''
  senderFilter.value = 'All senders'
  destinationFilter.value = 'All destinations'
  currencyFilter.value = 'All currencies'
  timeFilter.value = 'All timing'
  destinationTypeFilter.value = 'All destination types'
  sortBy.value = 'totalTimeDesc'
  transactionPage.value = 1
}

watch(filteredTransactions, () => {
  transactionPage.value = 1
})

watch(sortedCorridors, () => {
  routePage.value = 1
})

watch([dateRange, dashboardCurrency, selectedProviderId, selectedCorridor, selectedPayoutMethod, analysisLens], () => {
  syncDashboardQuery()
  void refreshDashboard()
  reconnectLive()
})

watch(dataState, () => {
  applyDashboardState(dataState.value)
  syncDashboardQuery()
  void refreshDashboard()
  reconnectLive()
})

watch(
  () => route.query,
  (query) => {
    const nextRange = typeof query.range === 'string' ? query.range : 'today'
    const nextCurrency = typeof query.currency === 'string' ? query.currency : 'USD'
    const nextProvider = typeof query.provider_id === 'string' ? query.provider_id : 'all'
    const nextCorridor = typeof query.corridor === 'string' ? query.corridor : 'all'
    const nextPayout = typeof query.payout_method === 'string' ? query.payout_method : 'all'
    const nextLens = typeof query.analysis_lens === 'string' ? query.analysis_lens : 'reliability'
    const nextDataState = normalizeDataState(query.scenario)
    const nextPolicyId = typeof query.policy_id === 'string' ? query.policy_id : ''

    if (dateRange.value !== nextRange) dateRange.value = nextRange
    if (dashboardCurrency.value !== nextCurrency) dashboardCurrency.value = nextCurrency
    if (selectedProviderId.value !== nextProvider) selectedProviderId.value = nextProvider
    if (selectedCorridor.value !== nextCorridor) selectedCorridor.value = nextCorridor
    if (selectedPayoutMethod.value !== nextPayout) selectedPayoutMethod.value = nextPayout
    if (analysisLens.value !== nextLens) analysisLens.value = nextLens
    if (dataState.value !== nextDataState) dataState.value = nextDataState
    if (nextPolicyId && policyRules.value.some((policy) => policy.id === nextPolicyId) && selectedPolicyId.value !== nextPolicyId) {
      selectedPolicyId.value = nextPolicyId
    }
  },
)

watch(
  () => route.query.timing,
  (timing) => {
    if (typeof timing === 'string') timeFilter.value = timing
  },
)

onMounted(async () => {
  try {
    sessionUser.value = await currentUser()
    if (sessionUser.value) {
      await refreshDashboard()
      reconnectLive()
    }
  } catch {
    sessionUser.value = null
  } finally {
    authReady.value = true
  }
})

onBeforeUnmount(() => disconnectLive?.())

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

function formatCurrencyShort(value: number, currency: string) {
  const abs = Math.abs(value)
  const digits = abs >= 1_000_000 ? 1 : 0
  if (abs >= 1_000_000_000) return `${currency} ${(value / 1_000_000_000).toFixed(1)}B`
  if (abs >= 1_000_000) return `${currency} ${(value / 1_000_000).toFixed(digits)}M`
  if (abs >= 1_000) return `${currency} ${(value / 1_000).toFixed(0)}K`
  return `${currency} ${value.toFixed(0)}`
}

function parsePercent(value: string) {
  const parsed = Number(value.replace(/[^\d.]/g, ''))
  return Number.isFinite(parsed) ? parsed : 0
}

function normalizeDataState(value: unknown): UiScenario {
  if (value === 'healthy') return 'healthy'
  if (value === 'traffic-shift') return 'traffic-shift'
  if (value === 'pilot-report') return 'pilot-report'
  if (value === 'degraded' || value === 'degraded-ria') return 'degraded-ria'
  if (value === 'stale-fx') return 'stale-fx'
  if (value === 'empty') return 'empty'
  if (value === 'permission-denied') return 'permission-denied'
  if (value === 'loading') return 'loading'
  if (value === 'api-failure') return 'api-failure'
  return 'degraded-ria'
}

function normalizeCorridorValue(corridor: string) {
  return corridor
    .replace(/\s+to\s+/i, ' -> ')
    .split('->')
    .map((part) => {
      const value = part.trim().toLowerCase()
      if (value === 'united states') return 'US'
      if (value === 'united kingdom' || value === 'uk') return 'GB'
      if (value === 'nigeria') return 'NG'
      if (value === 'kenya') return 'KE'
      if (value === 'eu' || value === 'europe' || value === 'european union') return 'EU'
      return part.trim().slice(0, 2).toUpperCase()
    })
    .join(' -> ')
}

function routeDetailIdFor(corridor: { corridor: string; selectedRoute: string }) {
  return `${normalizeCorridorValue(corridor.corridor)}-${corridor.selectedRoute}`
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
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

function metricIcon(id: string) {
  if (id === 'tail-latency') return Clock3
  if (id === 'failed' || id === 'value-risk') return AlertTriangle
  if (id === 'switching') return ShieldCheck
  if (id === 'sla') return CheckCircle2
  return Gauge
}

function metricTileDetail(tile: { id: string; trend: string }) {
  return tile.trend
}

function policyStatusLabel(status: PolicyStatus) {
  const labels: Record<PolicyStatus, string> = {
    draft: 'Draft',
    pending_approval: 'Pending approval',
    active: 'Active',
    inactive: 'Inactive',
  }
  return labels[status]
}

function corridorParts(corridor: string) {
  return parseCorridor(corridor)
}

function friendlyCorridorLabel(corridor: string) {
  const parts = parseCorridor(corridor)
  return `${getCountryIdentity(parts.origin).name} -> ${getCountryIdentity(parts.destination).name}`
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
  <LoginPanel v-if="authReady && !sessionUser" :busy="authBusy" :error="authError" @login="handleLogin" />
  <main v-else-if="!authReady" class="login-shell">
    <section class="login-card">
      <strong>Loading secure session</strong>
      <p>Checking bank access and role permissions.</p>
    </section>
  </main>
  <div v-else class="app-shell" data-bank-theme="imsi">
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
        <label class="data-state-control">
          <span>Data state</span>
          <select v-model="dataState" aria-label="Data state">
            <option v-for="option in dataStateOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
          </select>
        </label>
        <strong>{{ sessionUser?.display_name }}</strong>
        <small>{{ sessionUser?.roles.join(', ') }}</small>
        <button type="button" class="sidebar-link" @click="handleLogout">Sign out</button>
      </section>
    </aside>

    <main class="workspace">
      <PageHeader
        :title="currentPageTitle"
        :description="currentPageDescription"
        :breadcrumbs="breadcrumbs"
      >
        <template #actions>
          <UiButton v-if="activeScreen === 'policy' && !isPolicyCreateFlow" size="sm" @click="openPolicyCreateFlow">
            <Plus :size="15" aria-hidden="true" />
            New policy
          </UiButton>
          <UiButton v-if="isPolicyCreateFlow" size="sm" variant="secondary" @click="activate('policy')">
            Back to policies
          </UiButton>
          <DataFreshness
            :updated="dashboard.summary.lastUpdated"
            :mode="dashboard.summary.connection.mode"
            :stale="dashboard.summary.connection.freshness !== 'fresh'"
          />
        </template>
      </PageHeader>

      <StateBanner :scenario="dashboard.scenario" />

      <section class="control-bar" aria-label="Dashboard context controls">
        <label>
          <span>Date range</span>
          <select v-model="dateRange" aria-label="Dashboard date range">
            <option v-for="range in dashboardRanges" :key="range.value" :value="range.value">{{ range.label }}</option>
          </select>
        </label>
        <label>
          <span>IMTO</span>
          <select v-model="selectedProviderId" aria-label="Dashboard provider">
            <option v-for="provider in providerOptions" :key="provider.value" :value="provider.value">{{ provider.label }}</option>
          </select>
        </label>
        <label>
          <span>Corridor</span>
          <select v-model="selectedCorridor" aria-label="Dashboard corridor">
            <option v-for="corridor in corridorOptions" :key="corridor.value" :value="corridor.value">{{ corridor.label }}</option>
          </select>
        </label>
        <label>
          <span>Payout</span>
          <select v-model="selectedPayoutMethod" aria-label="Dashboard payout method">
            <option v-for="payout in payoutOptions" :key="payout.value" :value="payout.value">{{ payout.label }}</option>
          </select>
        </label>
        <label>
          <span>Currency</span>
          <select v-model="dashboardCurrency" aria-label="Dashboard display currency">
            <option v-for="currency in dashboardCurrencies" :key="currency">{{ currency }}</option>
          </select>
        </label>
        <label>
          <span>Analysis</span>
          <select v-model="analysisLens" aria-label="Dashboard analysis lens">
            <option v-for="lens in analysisLenses" :key="lens.value" :value="lens.value">{{ lens.label }}</option>
          </select>
        </label>
      </section>

      <section v-if="activeScreen === 'control'" class="screen-stack">
        <section class="kpi-grid kpi-grid--five">
          <KpiTile
            v-for="tile in apiMetricTiles"
            :key="tile.id"
            :label="tile.label"
            :value="`${tile.value}${tile.unit === '%' ? '%' : ''}`"
            :detail="metricTileDetail(tile)"
            :tone="tile.state"
            :icon="metricIcon(tile.id)"
            clickable
            @click="openDrilldown(tile.drilldown)"
          />
          <template v-if="apiMetricTiles.length === 0">
            <KpiTile label="Routes healthy" :value="dashboard.summary.globalHealth" :detail="dashboard.summary.topRisk" tone="degraded" :icon="Gauge" clickable @click="activate('corridors')" />
            <KpiTile label="At-risk value" :value="dashboard.summary.atRiskValue" :detail="`${dashboard.summary.stuckTransactions} transfers need review`" tone="watch" :icon="AlertTriangle" clickable @click="activate('transactions')" />
            <KpiTile label="P95 credit time" :value="dashboard.summary.p95CreditTime" :detail="`Target ${qaThresholdLabel}`" tone="degraded" :icon="Clock3" clickable @click="activate('incidents')" />
            <KpiTile label="Active incidents" :value="dashboard.summary.activeIncidents" detail="Open route work" tone="recovery" :icon="BellRing" clickable @click="activate('incidents')" />
          </template>
        </section>

        <section class="dashboard-grid">
          <Panel title="Transfer volume overview" :eyebrow="`Selected range / ${dashboardCurrency}`" accent="healthy" class="span-12">
            <div class="volume-command-grid">
              <article class="volume-widget volume-widget--trend">
                <header>
                  <span>Total volume for all transfers</span>
                  <strong>{{ volumeAnalytics.totalLabel }}</strong>
                  <small>{{ volumeAnalytics.countLabel }} / {{ volumeAnalytics.latestLabel }}</small>
                </header>
                <RealtimeLineChart
                  :points="dashboardTimeseries"
                  metric="volume"
                  :label="`Total transfer volume trend in ${dashboardCurrency}`"
                  color="#56d6ff"
                  :height="236"
                />
              </article>

              <article class="volume-widget">
                <header>
                  <span>Volume per top providers</span>
                  <strong>{{ providerVolumeRows[0]?.provider ?? 'No provider' }}</strong>
                  <small>{{ providerVolumeRows[0]?.volumeLabel ?? 'No processed volume' }}</small>
                </header>
                <div class="volume-rank-list">
                  <button
                    v-for="provider in providerVolumeRows"
                    :key="provider.id"
                    type="button"
                    class="volume-rank-row"
                    @click="activate('providers', { provider_id: provider.id })"
                  >
                    <span>
                      <ProviderMark :provider="provider.provider" />
                      <small>{{ provider.corridor }}</small>
                    </span>
                    <strong>{{ provider.volumeLabel }}</strong>
                    <em :style="{ width: provider.width }"></em>
                  </button>
                </div>
              </article>

              <article class="volume-widget">
                <header>
                  <span>Volume per top routes</span>
                  <strong>{{ routeVolumeRows[0]?.route ?? 'No route' }}</strong>
                  <small>{{ routeVolumeRows[0]?.volumeLabel ?? 'No processed volume' }}</small>
                </header>
                <div class="volume-rank-list">
                  <button
                    v-for="routeRow in routeVolumeRows"
                    :key="routeRow.id"
                    type="button"
                    class="volume-rank-row"
                    @click="activate('corridors', { corridor: normalizeCorridorValue(routeRow.corridor), focus: 'volume' })"
                  >
                    <span>
                      <CountryPair :origin="corridorParts(routeRow.corridor).origin" :destination="corridorParts(routeRow.corridor).destination" compact />
                      <small>{{ routeRow.provider }} / {{ routeRow.risk }}</small>
                    </span>
                    <strong>{{ routeRow.volumeLabel }}</strong>
                    <em :style="{ width: routeRow.width }"></em>
                  </button>
                </div>
              </article>
            </div>
          </Panel>

          <Panel title="Risk and ownership" :eyebrow="`${analysisLens} / ${dashboardCurrency}`" accent="healthy" class="span-12">
            <div class="operations-picture">
              <article v-for="item in operationsSnapshot" :key="item.label">
                <HealthBadge :state="item.state" />
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
                <small>{{ item.detail }}</small>
              </article>
            </div>
          </Panel>

          <Panel title="Corridors needing attention" eyebrow="Routes" accent="degraded" class="span-7">
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

          <Panel title="Active service incidents" eyebrow="Open work" :accent="activeIncident?.severity ?? 'healthy'" class="span-5">
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

          <Panel title="Control evidence" eyebrow="Controls" accent="healthy" class="span-7">
            <div class="provider-context">
              <article>
                <ShieldCheck :size="18" aria-hidden="true" />
                <div>
                  <strong>Decision trail</strong>
                  <small>Route, policy version, rejected options, and scoring inputs are retained.</small>
                </div>
              </article>
              <article>
                <History :size="18" aria-hidden="true" />
                <div>
                  <strong>Change control</strong>
                  <small>Maker-checker approval separates policy drafting from activation.</small>
                </div>
              </article>
            </div>
            <ActionBar>
              <UiButton variant="secondary" @click="activate('audit')">Open Audit</UiButton>
              <UiButton variant="secondary" @click="activate('policy')">Open Policy</UiButton>
            </ActionBar>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'corridors' && isRouteDetailFlow" class="screen-stack">
        <section v-if="selectedRouteDetail" class="dashboard-grid">
          <Panel title="Route summary" eyebrow="Route detail" :accent="selectedRouteDetail.corridor.state" class="span-5">
            <div class="selected-policy">
              <RoutePath
                :origin="corridorParts(selectedRouteDetail.corridor.corridor).origin"
                :destination="corridorParts(selectedRouteDetail.corridor.corridor).destination"
                :provider="selectedRouteDetail.provider"
                :rail="selectedRouteDetail.rail"
              />
              <dl class="metric-grid">
                <div>
                  <dt>Score</dt>
                  <dd>{{ selectedRouteDetail.corridor.score }}</dd>
                </div>
                <div>
                  <dt>P95</dt>
                  <dd>{{ selectedRouteDetail.corridor.p95 }}</dd>
                </div>
                <div>
                  <dt>Cost</dt>
                  <dd>{{ selectedRouteDetail.corridor.cost }}</dd>
                </div>
                <div>
                  <dt>Traffic split</dt>
                  <dd>{{ selectedRouteDetail.corridor.split }}</dd>
                </div>
              </dl>
              <aside class="state-note">
                <AlertTriangle :size="16" aria-hidden="true" />
                <span>{{ selectedRouteDetail.corridor.risk }}</span>
                <strong>{{ selectedRouteDetail.corridor.owner }}</strong>
              </aside>
              <ActionBar>
                <UiButton variant="secondary" @click="activate('transactions', { corridor: normalizeCorridorValue(selectedRouteDetail.corridor.corridor) })">View transactions</UiButton>
                <UiButton variant="secondary" @click="activate('providers', { provider_id: selectedRouteDetail.provider.toLowerCase() })">Provider</UiButton>
                <UiButton @click="activate('policy', { corridor: normalizeCorridorValue(selectedRouteDetail.corridor.corridor) })">Policy</UiButton>
              </ActionBar>
            </div>
          </Panel>

          <Panel title="Recent route transactions" :eyebrow="`${selectedRouteTransactions.length} matching transfers`" accent="healthy" class="span-7">
            <DataTable
              :empty="selectedRouteTransactions.length === 0"
              empty-title="No matching transactions"
              empty-description="No recent transfer records match this route in the current data window."
            >
              <table>
                <thead>
                  <tr>
                    <th>Reference</th>
                    <th>Amount</th>
                    <th>Elapsed</th>
                    <th>Owner</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="transaction in selectedRouteTransactions"
                    :key="transaction.reference"
                    class="click-row"
                    @click="openRouteTransaction(transaction, selectedRouteDetail.corridor.corridor)"
                  >
                    <td>
                      <strong class="mono">{{ transaction.reference }}</strong>
                      <small>{{ transaction.providerReference }}</small>
                    </td>
                    <td>{{ transaction.amount }}</td>
                    <td>{{ transaction.totalTime }}</td>
                    <td>{{ transaction.currentOwner }}</td>
                    <td><HealthBadge :state="qaStateFor(transaction)" :trigger="qaStatusLabel(transaction)" /></td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>

          <Panel title="Provider performance" eyebrow="Route provider" :accent="selectedRouteProviderScore?.state ?? 'unknown'" class="span-6">
            <div v-if="selectedRouteProviderScore" class="analytics-strip analytics-strip--compact">
              <article>
                <span>Success</span>
                <strong>{{ selectedRouteProviderScore.successRate }}</strong>
              </article>
              <article>
                <span>P50 / P95 / P99</span>
                <strong>{{ selectedRouteProviderScore.p50 }} / {{ selectedRouteProviderScore.p95 }} / {{ selectedRouteProviderScore.p99 }}</strong>
              </article>
              <article>
                <span>Stuck rate</span>
                <strong>{{ selectedRouteProviderScore.stuckRate }}</strong>
              </article>
              <article>
                <span>Exceptions</span>
                <strong>{{ selectedRouteProviderScore.settlementExceptions }}</strong>
              </article>
            </div>
            <EmptyState v-else title="No provider score" description="Provider scorecard data has not been published for this route." />
          </Panel>

          <Panel title="Incidents and owner" eyebrow="Operational work" :accent="selectedRouteIncidents[0]?.severity ?? selectedRouteDetail.corridor.state" class="span-6">
            <div class="event-list">
              <article v-if="selectedRouteIncidents.length === 0" class="state-note">
                <CheckCircle2 :size="16" aria-hidden="true" />
                <span>No active incident is linked to this route.</span>
              </article>
              <article v-for="incident in selectedRouteIncidents" :key="incident.id" class="state-note">
                <BellRing :size="16" aria-hidden="true" />
                <span>{{ incident.title }} / {{ incident.rootCause }}</span>
                <strong>{{ incident.owner }}</strong>
              </article>
            </div>
          </Panel>

          <Panel title="Policy and fallback" eyebrow="Routing rules" accent="watch" class="span-6">
            <ol class="rank-list">
              <li>
                <span class="rank">#1</span>
                <ProviderMark :provider="selectedRouteDetail.provider" show-category />
                <span>{{ selectedRouteDetail.corridor.selectedRoute }}</span>
                <HealthBadge :state="selectedRouteDetail.corridor.state" :trigger="selectedRouteDetail.corridor.status" />
              </li>
              <li v-for="fallback in dashboard.routeConfig.fallbackRoutes" :key="fallback.provider">
                <span class="rank">#{{ fallback.rank + 1 }}</span>
                <ProviderMark :provider="fallback.provider" show-category />
                <span>{{ fallback.route }}</span>
                <HealthBadge :state="fallback.state" />
              </li>
            </ol>
          </Panel>

          <Panel title="Cost and quote state" eyebrow="Rates and costs" :accent="selectedRouteFx?.state ?? 'unknown'" class="span-6">
            <div v-if="selectedRouteFx" class="definition-list">
              <div>
                <dt>Effective cost</dt>
                <dd>{{ selectedRouteFx.effectiveCost }}</dd>
              </div>
              <div>
                <dt>Fee / spread</dt>
                <dd>{{ selectedRouteFx.fee }} / {{ selectedRouteFx.spread }}</dd>
              </div>
              <div>
                <dt>Payout time</dt>
                <dd>{{ selectedRouteFx.payoutTime }}</dd>
              </div>
              <div>
                <dt>Quote age</dt>
                <dd>{{ selectedRouteFx.updatedAt }}</dd>
              </div>
            </div>
            <EmptyState v-else title="No quote data" description="No current rate or fee record is available for this route." />
          </Panel>
        </section>
        <Panel v-else title="Route not found" eyebrow="Routes" accent="watch">
          <EmptyState title="Route not found" description="Choose a route from the route list." :icon="Network">
            <UiButton @click="activate('corridors')">Back to routes</UiButton>
          </EmptyState>
        </Panel>
      </section>

      <section v-else-if="activeScreen === 'corridors'" class="screen-stack">
        <section class="dashboard-grid">
          <Panel title="Corridor route list" eyebrow="Routes" accent="degraded" class="span-12">
            <DataTable :empty="sortedCorridors.length === 0" empty-title="No corridors configured" empty-description="Add corridor routes before monitoring can start.">
              <table>
                <thead>
                  <tr>
                    <th>Corridor</th>
                    <th>Selected route</th>
                    <th>Score</th>
                    <th>Traffic split</th>
                    <th>Next action</th>
                    <th>Open</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="corridor in pagedCorridors" :key="`${corridor.corridor}-${corridor.selectedRoute}`">
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
                    <td>
                      <ActionBar compact>
                        <UiButton size="sm" @click="openRouteDetail(corridor)">Open</UiButton>
                        <UiButton size="sm" variant="secondary" @click="activate('transactions', { corridor: normalizeCorridorValue(corridor.corridor) })">Trace</UiButton>
                        <UiButton size="sm" variant="secondary" @click="activate('providers', { provider_id: routeProvider(corridor.selectedRoute).toLowerCase() })">Provider</UiButton>
                        <UiButton size="sm" @click="activate('policy', { corridor: normalizeCorridorValue(corridor.corridor) })">Policy</UiButton>
                      </ActionBar>
                    </td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
            <div class="pagination-bar">
              <span>{{ pagedCorridors.length }} of {{ sortedCorridors.length }} corridors</span>
              <div>
                <UiButton size="sm" variant="secondary" :disabled="routePage === 1" @click="goToRoutePage(routePage - 1)">Previous</UiButton>
                <strong>Page {{ routePage }} / {{ routePageCount }}</strong>
                <UiButton size="sm" variant="secondary" :disabled="routePage === routePageCount" @click="goToRoutePage(routePage + 1)">Next</UiButton>
              </div>
            </div>
          </Panel>

          <Panel title="Route action queue" eyebrow="Worklist" accent="watch" class="span-7">
            <div class="action-queue">
              <article v-for="item in pagedRouteActionItems" :key="item.id">
                <div class="route-action-main">
                  <CountryPair :origin="corridorParts(item.corridor.corridor).origin" :destination="corridorParts(item.corridor.corridor).destination" compact />
                  <div>
                    <strong>{{ item.action }}</strong>
                    <small>{{ item.corridor.risk }} / {{ item.corridor.owner }}</small>
                  </div>
                </div>
                <HealthBadge :state="item.corridor.state" />
                <ActionBar compact>
                  <UiButton size="sm" variant="secondary" @click="openRouteDetail(item.corridor)">Open route</UiButton>
                  <UiButton size="sm" variant="secondary" @click="activate('transactions', { corridor: normalizeCorridorValue(item.corridor.corridor) })">Trace</UiButton>
                  <UiButton size="sm" @click="activate('policy', { corridor: normalizeCorridorValue(item.corridor.corridor) })">Open policy</UiButton>
                </ActionBar>
              </article>
            </div>
            <div class="pagination-bar">
              <span>{{ pagedRouteActionItems.length }} of {{ routeActionItems.length }} actions</span>
              <div>
                <UiButton size="sm" variant="secondary" :disabled="routePage === 1" @click="goToRoutePage(routePage - 1)">Previous</UiButton>
                <strong>Page {{ routePage }} / {{ routePageCount }}</strong>
                <UiButton size="sm" variant="secondary" :disabled="routePage === routePageCount" @click="goToRoutePage(routePage + 1)">Next</UiButton>
              </div>
            </div>
          </Panel>

          <Panel title="Fallback order" eyebrow="Draft policy" accent="watch" class="span-5">
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
        <Panel title="Transfer search and reports" eyebrow="Transactions">
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
            <template #actions>
              <UiButton size="sm" variant="secondary" @click="downloadTransactionReport">
                <Download :size="15" aria-hidden="true" />
                Export CSV
              </UiButton>
            </template>
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
                    v-for="transaction in pagedTransactions"
                    :key="transaction.reference"
                    class="click-row"
                    :class="{ 'is-selected': selectedTransaction?.reference === transaction.reference }"
                    @click="selectTransaction(transaction)"
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
            <div class="pagination-bar">
              <span>{{ transactionStartIndex }}-{{ transactionEndIndex }} of {{ filteredTransactions.length }}</span>
              <label>
                Rows
                <select v-model.number="transactionPageSize" aria-label="Rows per page">
                  <option :value="10">10</option>
                  <option :value="25">25</option>
                  <option :value="50">50</option>
                </select>
              </label>
              <div>
                <UiButton size="sm" variant="secondary" :disabled="transactionPage === 1" @click="goToTransactionPage(transactionPage - 1)">Previous</UiButton>
                <strong>Page {{ transactionPage }} / {{ transactionPageCount }}</strong>
                <UiButton size="sm" variant="secondary" :disabled="transactionPage === transactionPageCount" @click="goToTransactionPage(transactionPage + 1)">Next</UiButton>
              </div>
            </div>
          </Panel>

          <Panel title="Transfer detail" eyebrow="Current selection" class="transaction-detail-panel" :accent="selectedTransaction ? qaStateFor(selectedTransaction) : 'unknown'">
            <EmptyState
              v-if="!selectedTransaction"
              title="Select a transfer"
              description="Pick one row to see the key status and open the full trace."
            >
              <TraceEmptyIllustration />
            </EmptyState>
            <div v-else class="transaction-detail">
              <div class="detail-heading">
                <div>
                  <h3 class="mono">{{ selectedTransaction.reference }}</h3>
                  <p>{{ selectedTransaction.amount }} / {{ selectedTransaction.beneficiary }}</p>
                  <p class="compact-route">
                    {{ selectedTransaction.senderCountry }} -> {{ selectedTransaction.destinationCountry }}
                    <span>{{ routeProvider(selectedTransaction.route) }}{{ routeRail(selectedTransaction.route) ? ` / ${routeRail(selectedTransaction.route)}` : '' }}</span>
                  </p>
                </div>
                <HealthBadge :state="qaStateFor(selectedTransaction)" :trigger="qaStatusLabel(selectedTransaction)" />
              </div>
              <dl class="compact-trace-facts">
                <div>
                  <dt>Elapsed</dt>
                  <dd>{{ selectedTransaction.totalTime }}</dd>
                </div>
                <div>
                  <dt>Owner</dt>
                  <dd>{{ selectedTransaction.currentOwner }}</dd>
                </div>
              </dl>
              <aside class="compact-state-note">
                <Clock3 :size="16" aria-hidden="true" />
                <span>{{ selectedTransaction.blocker }}</span>
              </aside>
              <ActionBar>
                <UiButton size="sm" @click="openTraceSheet">
                  <Maximize2 :size="15" aria-hidden="true" />
                  Open full trace
                </UiButton>
              </ActionBar>
            </div>
          </Panel>
        </section>

        <section v-if="traceSheetOpen && selectedTransaction" class="side-sheet-backdrop" role="dialog" aria-modal="true" aria-label="Full transaction trace">
          <aside class="side-sheet">
            <header>
              <div>
                <span class="eyebrow">Full trace</span>
                <h2 class="mono">{{ selectedTransaction.reference }}</h2>
              </div>
              <UiButton variant="icon" aria-label="Close full trace" @click="traceSheetOpen = false">
                <X :size="18" aria-hidden="true" />
              </UiButton>
            </header>
            <div class="sheet-section">
              <RoutePath
                :origin="selectedTransaction.senderCountry"
                :destination="selectedTransaction.destinationCountry"
                :provider="routeProvider(selectedTransaction.route)"
                :rail="routeRail(selectedTransaction.route)"
              />
              <dl class="metric-grid metric-grid--four">
                <div>
                  <dt>Amount</dt>
                  <dd>{{ selectedTransaction.amount }}</dd>
                </div>
                <div>
                  <dt>Elapsed</dt>
                  <dd>{{ selectedTransaction.totalTime }}</dd>
                </div>
                <div>
                  <dt>Owner</dt>
                  <dd>{{ selectedTransaction.currentOwner }}</dd>
                </div>
                <div>
                  <dt>Status</dt>
                  <dd>{{ qaStatusLabel(selectedTransaction) }}</dd>
                </div>
              </dl>
            </div>
            <TransactionTimeline :steps="selectedTimeline" />
          </aside>
        </section>
      </section>

      <section v-else-if="activeScreen === 'policy' && isPolicyCreateFlow" class="screen-stack">
        <section class="dashboard-grid">
          <Panel title="Policy scope" eyebrow="New policy" accent="recovery" class="span-8">
            <template #actions>
              <UiButton size="sm" :disabled="!can('policy:draft')" @click="createPolicyDraft">
                <Plus :size="15" aria-hidden="true" />
                Save draft
              </UiButton>
            </template>
            <div class="policy-rule-form">
              <label>
                <span>Name</span>
                <input v-model="policyDraftForm.name" type="text" />
              </label>
              <label>
                <span>Origin country/region</span>
                <select v-model="policyDraftForm.origin">
                  <option>European Union</option>
                  <option>United Kingdom</option>
                  <option>United States</option>
                  <option>Kenya</option>
                </select>
              </label>
              <label>
                <span>Destination</span>
                <select v-model="policyDraftForm.destination">
                  <option>Nigeria</option>
                </select>
              </label>
              <label>
                <span>Payout method</span>
                <select v-model="policyDraftForm.payoutMethod">
                  <option>Bank account</option>
                  <option>Local account</option>
                  <option>Wallet</option>
                  <option>Cash pickup</option>
                </select>
              </label>
              <label>
                <span>Primary rail</span>
                <select v-model="policyDraftForm.provider">
                  <option>Thunes</option>
                  <option>Remitly</option>
                  <option>Ria</option>
                  <option>PAPSS</option>
                </select>
              </label>
              <label>
                <span>Fallback order</span>
                <input v-model="policyDraftForm.fallback" type="text" />
              </label>
              <label class="span-2">
                <span>Amount band</span>
                <input v-model="policyDraftForm.amountBand" type="text" />
              </label>
            </div>
          </Panel>

          <Panel title="Approval path" eyebrow="Maker-checker" accent="healthy" class="span-4">
            <div class="selected-policy">
              <dl class="definition-list">
                <div>
                  <dt>Drafter</dt>
                  <dd>{{ currentActor }}</dd>
                </div>
                <div>
                  <dt>Status after save</dt>
                  <dd>Draft</dd>
                </div>
                <div>
                  <dt>Approval required</dt>
                  <dd>Yes</dd>
                </div>
                <div>
                  <dt>Activation</dt>
                  <dd>After approval</dd>
                </div>
              </dl>
              <aside class="state-note">
                <ShieldCheck :size="16" aria-hidden="true" />
                <span>Drafts cannot be activated by the same user who creates them.</span>
              </aside>
              <ActionBar>
                <UiButton variant="secondary" @click="activate('policy')">Cancel</UiButton>
                <UiButton :disabled="!can('policy:draft')" @click="createPolicyDraft">Save draft</UiButton>
              </ActionBar>
            </div>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'policy'" class="screen-stack">
        <section class="dashboard-grid">
          <Panel title="Policy inventory" eyebrow="Maker-checker" accent="healthy" class="span-7">
            <template #actions>
              <span class="dashboard-chip">{{ policyStatusCounts.active }} active</span>
              <span class="dashboard-chip">{{ policyStatusCounts.pending_approval }} pending approval</span>
              <UiButton size="sm" @click="openPolicyCreateFlow">
                <Plus :size="15" aria-hidden="true" />
                New
              </UiButton>
            </template>
            <div class="policy-list">
              <button
                v-for="policy in policyRules"
                :key="policy.id"
                type="button"
                :class="{ 'is-selected': selectedPolicy?.id === policy.id }"
                @click="selectPolicy(policy)"
              >
                <span>
                  <strong>{{ policy.name }}</strong>
                  <small>{{ policy.id }} / {{ policy.version }}</small>
                </span>
                <CountryPair :origin="policy.origin" :destination="policy.destination" compact />
                <ProviderMark :provider="policy.provider" />
                <HealthBadge :state="policy.status === 'active' ? 'healthy' : policy.status === 'pending_approval' ? 'watch' : policy.status === 'inactive' ? 'stale' : 'recovery'" :trigger="policyStatusLabel(policy.status)" />
              </button>
            </div>
          </Panel>

          <Panel title="Selected policy" eyebrow="Scope and approval" :accent="selectedPolicy?.status === 'active' ? 'healthy' : 'watch'" class="span-5">
            <div v-if="selectedPolicy" class="selected-policy">
              <CountryPair :origin="selectedPolicy.origin" :destination="selectedPolicy.destination" />
              <dl class="definition-list">
                <div>
                  <dt>Payout</dt>
                  <dd>{{ selectedPolicy.payoutMethod }}</dd>
                </div>
                <div>
                  <dt>Amount band</dt>
                  <dd>{{ selectedPolicy.amountBand }}</dd>
                </div>
                <div>
                  <dt>Primary rail</dt>
                  <dd><ProviderMark :provider="selectedPolicy.provider" /></dd>
                </div>
                <div>
                  <dt>Fallbacks</dt>
                  <dd>{{ selectedPolicy.fallback.join(', ') }}</dd>
                </div>
                <div>
                  <dt>Drafter</dt>
                  <dd>{{ selectedPolicy.drafter }}</dd>
                </div>
                <div>
                  <dt>Approver</dt>
                  <dd>{{ selectedPolicy.approver ?? 'Required' }}</dd>
                </div>
              </dl>
              <p v-if="selectedPolicyNeedsDifferentApprover" class="form-error">Maker-checker requires another user to approve this draft.</p>
              <ActionBar>
                <UiButton size="sm" :disabled="selectedPolicy.status !== 'draft' || !can('policy:draft')" @click="submitSelectedPolicy">
                  <Send :size="15" aria-hidden="true" />
                  Submit
                </UiButton>
                <UiButton size="sm" variant="secondary" :disabled="selectedPolicy.status !== 'pending_approval' || selectedPolicyNeedsDifferentApprover || !can('policy:approve')" @click="approveSelectedPolicy">
                  <FileCheck2 :size="15" aria-hidden="true" />
                  Approve
                </UiButton>
                <UiButton size="sm" variant="secondary" :disabled="selectedPolicy.status === 'draft' || selectedPolicy.status === 'pending_approval' || selectedPolicy.status === 'active' || !can('policy:activate')" @click="activateSelectedPolicy">
                  <PlayCircle :size="15" aria-hidden="true" />
                  Activate
                </UiButton>
                <UiButton size="sm" variant="secondary" :disabled="selectedPolicy.status !== 'active' || !can('policy:activate')" @click="deactivateSelectedPolicy">
                  <PauseCircle :size="15" aria-hidden="true" />
                  Deactivate
                </UiButton>
              </ActionBar>
            </div>
          </Panel>

          <Panel title="Completion-time policy" eyebrow="QA thresholds" :accent="thresholdIsValid ? 'healthy' : 'degraded'" class="span-5">
            <template #actions>
              <UiButton variant="secondary" size="sm" :disabled="!thresholdHasChanges" @click="resetQaThresholds">Reset</UiButton>
              <UiButton size="sm" :disabled="!thresholdHasChanges || !thresholdIsValid || !can('policy:draft')" @click="saveQaThresholds">Save thresholds</UiButton>
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

          <Panel title="Policy impact check" eyebrow="Shadow test" accent="healthy" class="span-7">
            <div v-if="simulationSample" class="simulation-grid">
              <section>
                <span class="eyebrow">Replay transaction</span>
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

          <Panel title="Policy change history" eyebrow="Change log" accent="recovery" class="span-5">
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

          <Panel title="Latency root cause" eyebrow="Ownership" accent="degraded" class="span-5">
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
          <Panel title="Eligible route vs cheapest quote" eyebrow="Rates and costs" accent="healthy" class="span-5">
            <div class="fx-decision-grid">
              <article>
                <span>Selected eligible route</span>
                <ProviderMark :provider="dashboard.fxCostBoard.recommendedProvider" show-category size="lg" />
              </article>
              <article>
                <span>Cheapest quoted route</span>
                <ProviderMark :provider="dashboard.fxCostBoard.cheapestProvider" show-category />
              </article>
            </div>
            <aside class="state-note">
              <CircleDollarSign :size="16" aria-hidden="true" />
              <span>{{ dashboard.fxCostBoard.decision }}</span>
              <strong>{{ dashboard.fxCostBoard.window }}</strong>
            </aside>
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

          <Panel title="Cost, speed, and eligibility" :eyebrow="`${fxBaseCurrency} baseline`" accent="watch" class="span-7">
            <DataTable :empty="fxComparisonRoutes.length === 0" empty-title="No rates available" empty-description="Rates will appear when provider quotes refresh.">
              <table class="fx-cost-table">
                <thead>
                  <tr>
                    <th>Provider</th>
                    <th>Rate / cost</th>
                    <th>Speed</th>
                    <th>Why it matters</th>
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
                      <small>Fee/spread {{ route.fee }} / {{ route.spread }} - effective {{ route.effectiveCost }}</small>
                    </td>
                    <td>{{ route.payoutTime }}</td>
                    <td>
                      <HealthBadge :state="route.state" :trigger="route.recommended ? 'selected' : route.cheapest ? 'cheapest' : route.note" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>

          <Panel title="Route quote checks" eyebrow="Eligibility and freshness" accent="recovery" class="span-12">
            <div class="fx-rule-grid">
              <article v-for="route in fxComparisonRoutes" :key="`${route.provider}-rule`">
                <ProviderMark :provider="route.provider" />
                <HealthBadge :state="route.state" :trigger="route.note" />
                <dl>
                  <div>
                    <dt>Rate age</dt>
                    <dd>{{ route.updatedAt }}</dd>
                  </div>
                  <div>
                    <dt>Effective cost</dt>
                    <dd>{{ route.effectiveCost }}</dd>
                  </div>
                  <div>
                    <dt>Payout time</dt>
                    <dd>{{ route.payoutTime }}</dd>
                  </div>
                </dl>
              </article>
            </div>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'reconciliation'" class="screen-stack">
        <section class="dashboard-grid">
          <Panel title="Settlement breaks to clear" eyebrow="Reconciliation" accent="watch" class="span-8">
            <template #actions>
              <UiButton size="sm" variant="secondary">
                <Upload :size="15" aria-hidden="true" />
                Import file
              </UiButton>
              <UiButton size="sm" variant="secondary">
                <Download :size="15" aria-hidden="true" />
                Export breaks
              </UiButton>
            </template>
            <DataTable
              :empty="dashboard.reconciliation.length === 0"
              empty-title="No settlement breaks"
              empty-description="Provider files, bank postings, and settlement references match for this view."
            >
              <table>
                <thead>
                  <tr>
                    <th>Reference</th>
                    <th>Provider</th>
                    <th>What happened</th>
                    <th>Amount</th>
                    <th>Age</th>
                    <th>Owner</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in dashboard.reconciliation" :key="item.reference">
                    <td class="mono">{{ item.reference }}</td>
                    <td><ProviderMark :provider="item.provider" /></td>
                    <td>
                      <HealthBadge :state="item.state" :trigger="item.reason" />
                    </td>
                    <td><strong>{{ item.amount }}</strong></td>
                    <td>{{ item.age }}</td>
                    <td>{{ item.owner }}</td>
                    <td>
                      <ActionBar compact>
                        <UiButton size="sm" variant="secondary" @click="activate('transactions')">Trace</UiButton>
                        <UiButton size="sm">Resolve</UiButton>
                      </ActionBar>
                    </td>
                  </tr>
                </tbody>
              </table>
            </DataTable>
          </Panel>

          <Panel title="Break categories" eyebrow="Exception type" accent="recovery" class="span-4">
            <div class="recon-lanes">
              <article>
                <FileCheck2 :size="18" aria-hidden="true" />
                <strong>Provider file late</strong>
                <small>{{ dashboard.reconciliation.filter((item) => item.reason.includes('settlement file')).length }} open</small>
              </article>
              <article>
                <Clock3 :size="18" aria-hidden="true" />
                <strong>Credit pending</strong>
                <small>{{ dashboard.reconciliation.filter((item) => item.reason.includes('credit pending')).length }} open</small>
              </article>
              <article>
                <CircleDollarSign :size="18" aria-hidden="true" />
                <strong>FX mismatch</strong>
                <small>{{ dashboard.reconciliation.filter((item) => item.reason.includes('FX')).length }} open</small>
              </article>
            </div>
          </Panel>

          <Panel title="How to close a break" eyebrow="Operator steps" accent="healthy" class="span-12">
            <div class="resolution-grid">
              <article>
                <BadgeCheck :size="18" aria-hidden="true" />
                <strong>Confirm references</strong>
                <small>Compare switch, provider, bank, and settlement batch IDs.</small>
              </article>
              <article>
                <CircleDollarSign :size="18" aria-hidden="true" />
                <strong>Validate amount and FX</strong>
                <small>Check amount, currency, rate timestamp, fee, and spread.</small>
              </article>
              <article>
                <Clock3 :size="18" aria-hidden="true" />
                <strong>Confirm beneficiary outcome</strong>
                <small>Record credit, reversal, or the team blocking closure.</small>
              </article>
              <article>
                <History :size="18" aria-hidden="true" />
                <strong>Close with evidence</strong>
                <small>Save resolution note, owner, and supporting reference.</small>
              </article>
            </div>
          </Panel>
        </section>
      </section>

      <section v-else-if="activeScreen === 'providers'" class="screen-stack">
        <section class="kpi-grid">
          <KpiTile label="Best SLA" :value="sortedProviders[0]?.provider ?? 'None'" :detail="`${sortedProviders[0]?.p95 ?? '-'} P95`" tone="healthy" :icon="CheckCircle2" />
          <KpiTile label="Most exceptions" :value="weakestProvider?.provider ?? 'None'" :detail="`${weakestProvider?.settlementExceptions ?? 0} exceptions`" tone="degraded" :icon="AlertTriangle" />
          <KpiTile label="Total providers" :value="sortedProviders.length" detail="Connected provider routes" tone="brand" :icon="Network" />
          <KpiTile label="Measurement window" value="15m" detail="Success, P95, stuck rate" tone="recovery" :icon="TimerReset" />
        </section>
        <Panel title="Provider action queue" eyebrow="SLA and exceptions" :accent="dashboard.recommendation.state">
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
        <Panel v-if="providerComparisons.length" title="Provider volume and SLA" :eyebrow="`${dateRange} / ${dashboardCurrency}`" accent="healthy">
          <DataTable :empty="providerComparisons.length === 0" empty-title="No provider analytics" empty-description="No provider records match the current dashboard context.">
            <table>
              <thead>
                <tr>
                  <th>Provider</th>
                  <th>Corridor</th>
                  <th>Processed</th>
                  <th>Volume</th>
                  <th>Completed in SLA</th>
                  <th>P95</th>
                  <th>State</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="provider in providerComparisons" :key="provider.provider_id">
                  <td><ProviderMark :provider="provider.provider_name" show-category /></td>
                  <td>{{ provider.corridor }}</td>
                  <td>{{ provider.processed_count.toLocaleString() }}</td>
                  <td>{{ provider.processed_volume.toLocaleString(undefined, { maximumFractionDigits: 0 }) }} {{ dashboardCurrency }}</td>
                  <td>{{ provider.sla_completed_count.toLocaleString() }} / {{ provider.sla_rate.toFixed(1) }}%</td>
                  <td>{{ formatDuration(provider.p95_seconds) }}</td>
                  <td><HealthBadge :state="provider.state" /></td>
                </tr>
              </tbody>
            </table>
          </DataTable>
        </Panel>
        <Panel title="Provider scorecard" eyebrow="SLA, latency, exceptions" accent="healthy">
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
        <Panel title="Audit trail" eyebrow="Evidence">
          <ActionBar>
            <label class="search-field">
              <Search :size="18" aria-hidden="true" />
              <input v-model="auditQuery" type="search" placeholder="Actor, object, reason, action" aria-label="Search audit events" />
            </label>
            <span class="dashboard-chip">Immutable audit trail</span>
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

          <Panel title="Event detail" eyebrow="Selected event" class="audit-detail-panel" :accent="selectedAudit?.state ?? 'unknown'">
            <EmptyState v-if="!selectedAudit" title="Select an event" description="Choose an audit event to inspect its reason and object." :icon="BadgeCheck" />
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
                <span>Stored with policy version, route inputs, actor, timestamp, and review reason.</span>
              </aside>
            </div>
          </Panel>
        </section>
      </section>
    </main>
  </div>
</template>
