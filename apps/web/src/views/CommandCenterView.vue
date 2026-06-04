<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, ArrowRight, BadgeCheck, CheckCircle2, FileCheck2, Network, RefreshCw, ShieldAlert, WifiOff } from '@lucide/vue'
import ActionBar from '../components/ActionBar.vue'
import DashboardChart from '../components/DashboardChart.vue'
import HealthBadge from '../components/HealthBadge.vue'
import Panel from '../components/Panel.vue'
import ProviderMark from '../components/ProviderMark.vue'
import UiButton from '../components/UiButton.vue'
import { queueLabels, routePenaltyWidth, useCommandCenter } from '../composables/useInboundTower'
import { useTowerRouting } from '../composables/useTowerRouting'
import type { DashboardMock, HealthState, IncomingInstruction, RemediationCase, RoutePenalty } from '../types'

const props = defineProps<{
  dashboard: DashboardMock
}>()

const {
  openRemediationCases,
  routePenalties,
  nextSafestAction,
  valueAtRiskTotal,
} = useCommandCenter(props.dashboard)
const { activate, openPath, openInflow, openException, openRoute } = useTowerRouting()

const unavailableStates = new Set<HealthState>(['degraded', 'blocked', 'stale', 'unknown'])
const activeFailureStates = new Set(['failed_safe', 'failed_unsafe', 'reversal_pending'])

const localRoutes = computed(() => routePenalties.value.filter((routeRow) => routeRow.route !== 'Fidelity core'))
const workingLocalRoutes = computed(() => localRoutes.value.filter((routeRow) => !unavailableStates.has(routeRow.state)))
const bestLocalProvider = computed(() => [...localRoutes.value].sort((a, b) => a.penaltyScore - b.penaltyScore)[0] ?? null)
const mostPressuredProvider = computed(() => [...localRoutes.value].sort((a, b) => b.penaltyScore - a.penaltyScore)[0] ?? null)
const providerPerformanceRows = computed(() => [...localRoutes.value].sort((a, b) => a.penaltyScore - b.penaltyScore))

const failedOrReversedTransfers = computed(() =>
  props.dashboard.incomingInstructions.filter(
    (item) => activeFailureStates.has(item.currentState) || item.outcomeConfidence === 'safe_failed' || item.outcomeConfidence === 'unsafe_failed',
  ),
)
const reversalPendingCount = computed(() =>
  props.dashboard.incomingInstructions.filter((item) => item.currentState === 'reversal_pending').length,
)
const unsafeFailureCount = computed(() =>
  props.dashboard.incomingInstructions.filter((item) => item.currentState === 'failed_unsafe' || item.outcomeConfidence === 'unsafe_failed').length,
)
const failedTransferTone = computed<HealthState>(() => (failedOrReversedTransfers.value.length ? 'watch' : 'healthy'))
const failedTransferAnswer = computed(() => (failedOrReversedTransfers.value.length ? `Yes - ${failedOrReversedTransfers.value.length}` : 'No'))
const failedTransferSummary = computed(() => {
  if (!failedOrReversedTransfers.value.length) return 'No failed or reversed local transfer in this window.'
  return `${reversalPendingCount.value} reversal pending / ${unsafeFailureCount.value} unsafe / ${openRemediationCases.value.length} repair cases.`
})

const localProviderTone = computed<HealthState>(() => {
  if (!localRoutes.value.length) return 'unknown'
  if (workingLocalRoutes.value.length === localRoutes.value.length) return 'healthy'
  return workingLocalRoutes.value.length ? 'watch' : 'degraded'
})
const localProviderAnswer = computed(() => `${workingLocalRoutes.value.length}/${localRoutes.value.length}`)
const localProviderSummary = computed(() =>
  bestLocalProvider.value ? 'Current provider window' : 'No local provider telemetry is available.',
)
const providerQuickStats = computed(() =>
  providerPerformanceRows.value.slice(0, 4).map((routeRow) => ({
    label: routeRow.rail,
    p95: routeRow.p95CreditTime,
    timeout: routeRow.timeoutRate,
    openCases: routeRow.openCases,
    isBest: routeRow.route === bestLocalProvider.value?.route,
  })),
)

const failingPartnerSlas = computed(() => props.dashboard.inboundSla.filter((row) => row.state === 'degraded' || row.state === 'blocked'))
const partnerSlaTone = computed<HealthState>(() => (failingPartnerSlas.value.length ? 'degraded' : 'healthy'))
const partnerSlaAnswer = computed(() => (failingPartnerSlas.value.length ? `${failingPartnerSlas.value.length} failing` : 'None failing'))
const partnerSlaSummary = computed(() => {
  const topBreach = failingPartnerSlas.value[0]
  if (!topBreach) return 'All international banking partner SLAs are inside the current breach policy.'
  return `${topBreach.partner} is at ${topBreach.breachRate} breach rate; oldest breach ${topBreach.oldestBreach}.`
})

const switchingApiState = computed<HealthState>(() => {
  if (props.dashboard.viewState === 'error' || props.dashboard.summary.connection.freshness === 'unavailable') return 'degraded'
  if (props.dashboard.summary.connection.freshness === 'stale') return 'stale'
  if (props.dashboard.viewState === 'loading') return 'watch'
  return 'healthy'
})
const switchingApiAnswer = computed(() => {
  if (switchingApiState.value === 'healthy') return 'Live'
  if (switchingApiState.value === 'stale') return 'Stale status'
  if (switchingApiState.value === 'watch') return 'Loading'
  return 'Status issue'
})
const switchingApiSummary = computed(() => {
  if (switchingApiState.value === 'degraded') return 'Telemetry unavailable; keep settlement changes in manual review.'
  if (switchingApiState.value === 'stale') return 'Telemetry delayed; verify provider status before closing cases.'
  if (switchingApiState.value === 'watch') return 'Telemetry loading; wait for the next poll before shifting traffic.'
  return `Telemetry current / ${props.dashboard.summary.connection.nextPollIn} / ${props.dashboard.summary.lastUpdated}.`
})
const switchingPurpose = computed(() => {
  const rails = localRoutes.value.map((routeRow) => routeRow.rail).join(', ') || 'eligible local rails'
  return `Switches eligible local settlement traffic across ${rails}.`
})
const switchingAction = computed(() => {
  if (switchingApiState.value !== 'healthy') return 'Keep settlement traffic changes in manual review until readiness status is current.'
  if (mostPressuredProvider.value?.state === 'degraded' && bestLocalProvider.value) {
    return `Use ${bestLocalProvider.value.route} for eligible new traffic. Repair ${mostPressuredProvider.value.route} evidence separately.`
  }
  return nextSafestAction.value
})

const overallTone = computed<HealthState>(() => {
  if (switchingApiState.value === 'degraded' || failingPartnerSlas.value.length || unsafeFailureCount.value) return 'degraded'
  if (failedOrReversedTransfers.value.length || workingLocalRoutes.value.length !== localRoutes.value.length) return 'watch'
  return 'healthy'
})
const primarySignal = computed(() => {
  if (!openRemediationCases.value.length) return 'No failed-transfer or middleware repair case is open right now.'
  return `${openRemediationCases.value.length} repair case${openRemediationCases.value.length === 1 ? '' : 's'} need evidence before closure.`
})

const signalItems = computed(() => [
  {
    id: 'failed-transfers',
    label: 'Failed/reversed local transfers',
    value: failedTransferAnswer.value,
    detail: failedTransferSummary.value,
    tone: failedTransferTone.value,
    icon: failedOrReversedTransfers.value.length ? ShieldAlert : CheckCircle2,
    action: () => openPath('/exceptions', { queue: 'all', focus: 'failed' }),
  },
  {
    id: 'switching-api',
    label: 'Switching API status',
    value: switchingApiAnswer.value,
    detail: switchingApiSummary.value,
    tone: switchingApiState.value,
    icon: switchingApiState.value === 'healthy' ? Network : WifiOff,
    action: () => openPath('/settings', { tab: 'integrations', focus: 'api-health' }),
  },
  {
    id: 'local-providers',
    label: 'Local providers working',
    value: localProviderAnswer.value,
    detail: localProviderSummary.value,
    stats: providerQuickStats.value,
    tone: localProviderTone.value,
    icon: FileCheck2,
    action: () => openPath('/routes', { focus: 'provider-performance' }),
  },
  {
    id: 'partner-slas',
    label: 'International partner SLAs',
    value: partnerSlaAnswer.value,
    detail: partnerSlaSummary.value,
    tone: partnerSlaTone.value,
    icon: failingPartnerSlas.value.length ? AlertTriangle : BadgeCheck,
    action: () => openPath('/settings', { tab: 'sla', focus: 'breaches' }),
  },
])

const repairQueue = computed(() => [...openRemediationCases.value].sort((a, b) => casePriority(a) - casePriority(b)))
const activeWorkItems = computed(() => repairQueue.value.slice(0, 3))
const providerRows = computed(() =>
  providerPerformanceRows.value
    .map((routeRow) => ({
      route: routeRow,
      action: providerAction(routeRow),
      isBest: routeRow.route === bestLocalProvider.value?.route,
      isPressured: routeRow.route === mostPressuredProvider.value?.route,
    })),
)
const partnerSlaRows = computed(() =>
  [...props.dashboard.inboundSla].sort((a, b) => stateOrder(a.state) - stateOrder(b.state) || b.agingBreaches - a.agingBreaches),
)
const switchingChecks = computed(() => [
  {
    label: 'Status',
    value: switchingApiAnswer.value,
    detail: props.dashboard.summary.connection.mode === 'static' ? 'Static fallback mode' : props.dashboard.summary.connection.nextPollIn,
    state: switchingApiState.value,
  },
  {
    label: 'Local rail coverage',
    value: `${workingLocalRoutes.value.length}/${localRoutes.value.length}`,
    detail: localRoutes.value.map((routeRow) => routeRow.rail).join(', ') || 'No local rails configured',
    state: localProviderTone.value,
  },
  {
    label: 'Reliability move',
    value: bestLocalProvider.value ? 'Shift eligible' : 'Manual review',
    detail: switchingAction.value,
    state: mostPressuredProvider.value?.state === 'degraded' ? 'watch' : switchingApiState.value,
  },
] satisfies Array<{ label: string; value: string; detail: string; state: HealthState }>)

const slaTrendLabels = computed(() => props.dashboard.visuals.completionTrend.map((point) => point.label))
const slaTrendDatasets = computed(() => [
  {
    label: 'SLA completion',
    values: props.dashboard.visuals.completionTrend.map((point) => point.value),
    color: '#0a66ff',
    fill: true,
  },
])
const providerChartLabels = computed(() => providerPerformanceRows.value.map((row) => row.rail))
const providerChartDatasets = computed(() => [
  {
    label: 'Timeout rate',
    values: providerPerformanceRows.value.map((row) => toPercent(row.timeoutRate)),
    color: '#b54708',
  },
  {
    label: 'Open cases',
    values: providerPerformanceRows.value.map((row) => row.openCases),
    color: '#0a66ff',
  },
])
const exceptionChartLabels = computed(() => ['Failed/reversed', 'Cooldown', 'Requerying', 'Manual queue'])
const exceptionChartDatasets = computed(() => [
  {
    label: 'Cases',
    values: [
      failedOrReversedTransfers.value.length,
      props.dashboard.incomingInstructions.filter((item) => item.currentState === 'cooldown' || item.currentState === 'sla_breached').length,
      props.dashboard.incomingInstructions.filter((item) => item.currentState === 'requerying').length,
      openRemediationCases.value.length,
    ],
  },
])

function stateOrder(state: HealthState) {
  const order: Record<HealthState, number> = {
    blocked: 0,
    degraded: 1,
    watch: 2,
    stale: 3,
    recovery: 4,
    unknown: 5,
    healthy: 6,
  }
  return order[state]
}

function casePriority(item: RemediationCase) {
  if (item.queue === 'exhausted') return 0
  if (item.queue === 'reversal') return 1
  if (item.queue === 'requerying') return 2
  return 3
}

function providerAction(routeRow: RoutePenalty) {
  if (routeRow.state === 'degraded' || routeRow.state === 'blocked') return 'Contain new eligible traffic'
  if (routeRow.route === bestLocalProvider.value?.route) return 'Best performer now'
  if (routeRow.state === 'watch') return 'Working, monitor closely'
  return 'Working'
}

function toPercent(value: string) {
  return Number.parseFloat(value.replace('%', '')) || 0
}

function queueLabel(item: RemediationCase) {
  return queueLabels[item.queue]
}

function workStatus(item: RemediationCase) {
  if (item.makerCheckerState === 'checker_pending') return 'Checker pending'
  if (item.queue === 'requerying' || item.automationStatus === 'running') return 'Requery running'
  if (item.queue === 'reversal') return 'Reversal ready'
  if (item.state === 'degraded') return 'Needs evidence'
  return 'Route degraded'
}

function openFirstFailure() {
  const item: IncomingInstruction | undefined = failedOrReversedTransfers.value[0]
  if (item) openInflow(item.reference)
  else activate('inflows')
}
</script>

<template>
  <section class="screen-stack command-workbench">
    <section class="ops-command-bar" :class="`ops-command-bar--${overallTone}`">
      <div class="ops-command-bar__copy">
        <p class="eyebrow">Primary dashboard</p>
        <h2>Settlement reliability</h2>
        <p>{{ primarySignal }} Failed transfers, local provider performance, partner SLA breaches, and API telemetry stay on one operating screen.</p>
      </div>
      <div class="ops-command-bar__meta">
        <HealthBadge :state="overallTone" :trigger="valueAtRiskTotal" />
        <strong>{{ valueAtRiskTotal }} at risk</strong>
        <small>{{ nextSafestAction }}</small>
      </div>
    </section>

    <section class="ops-signal-strip" aria-label="Primary operating answers">
      <button
        v-for="item in signalItems"
        :key="item.id"
        type="button"
        class="ops-signal"
        :class="`ops-signal--${item.tone}`"
        @click="item.action"
      >
        <span class="ops-signal__icon">
          <component :is="item.icon" :size="17" aria-hidden="true" />
        </span>
        <span>
          <small>{{ item.label }}</small>
          <strong>{{ item.value }}</strong>
        </span>
        <div v-if="item.stats?.length" class="ops-signal__stats" aria-label="Local provider quick stats">
          <span v-for="stat in item.stats" :key="stat.label" :class="{ 'is-best': stat.isBest }">
            <small>{{ stat.label }}</small>
            <strong>{{ stat.p95 }}</strong>
            <em>{{ stat.timeout }} timeout / {{ stat.openCases }} open</em>
          </span>
        </div>
        <p v-else>{{ item.detail }}</p>
      </button>
    </section>

    <section class="active-work-strip" aria-label="Active repair work">
      <button v-for="item in activeWorkItems" :key="item.id" type="button" @click="openException(item.instructionReference)">
        <span>
          <small>{{ workStatus(item) }}</small>
          <strong>{{ item.instructionReference }}</strong>
        </span>
        <span>
          <strong>{{ item.owner }}</strong>
          <small>{{ item.route }} / {{ item.age }}</small>
        </span>
        <span>
          <strong>{{ item.valueAtRisk }}</strong>
          <small>{{ item.nextAction }}</small>
        </span>
        <HealthBadge :state="item.state" />
      </button>
    </section>

    <section class="ops-chart-grid" aria-label="Operational charts">
      <Panel title="SLA completion trend" eyebrow="8-hour window" :accent="partnerSlaTone">
        <DashboardChart
          kind="line"
          :labels="slaTrendLabels"
          :datasets="slaTrendDatasets"
          unit="%"
          summary="SLA completion rate by hour for the selected settlement window."
          @drilldown="openPath('/settings', { tab: 'sla', focus: 'breaches' })"
        />
      </Panel>
      <Panel title="Local provider performance" eyebrow="Current window" :accent="localProviderTone">
        <DashboardChart
          kind="bar"
          :labels="providerChartLabels"
          :datasets="providerChartDatasets"
          summary="Timeout rate and open repair cases by local settlement provider."
          @drilldown="openPath('/routes', { focus: 'provider-performance' })"
        />
      </Panel>
      <Panel title="Exception mix" eyebrow="Current queue" :accent="failedTransferTone">
        <DashboardChart
          kind="doughnut"
          :labels="exceptionChartLabels"
          :datasets="exceptionChartDatasets"
          summary="Failed, cooldown, requerying, and manual queue cases in the current window."
          @drilldown="openPath('/exceptions', { queue: 'all', focus: 'failed' })"
        />
      </Panel>
    </section>

    <section class="ops-workbench-grid">
      <Panel title="Transactions and middleware calls to fix" eyebrow="Evidence-first worklist" :accent="openRemediationCases[0]?.state ?? 'healthy'">
        <div class="repair-worklist">
          <button
            v-for="item in repairQueue"
            :key="item.id"
            type="button"
            class="repair-ticket"
            @click="openException(item.instructionReference)"
          >
            <span class="repair-ticket__reference">
              <strong>{{ item.instructionReference }}</strong>
              <small>{{ queueLabel(item) }} / {{ item.route }}</small>
            </span>
            <span class="repair-ticket__issue">
              <strong>{{ item.title }}</strong>
              <small>{{ item.nextAction }}</small>
            </span>
            <span class="repair-ticket__value">
              <strong>{{ item.valueAtRisk }}</strong>
              <small>{{ item.age }} open</small>
            </span>
            <HealthBadge :state="item.state" />
          </button>
          <article v-if="repairQueue.length === 0" class="quiet-empty">
            <CheckCircle2 :size="18" aria-hidden="true" />
            <span>
              <strong>No open repair queue</strong>
              <small>No failed transfer, reversal, requery, or ledger evidence task is waiting.</small>
            </span>
          </article>
        </div>
        <ActionBar>
          <UiButton @click="activate('exceptions')">
            <ArrowRight :size="15" aria-hidden="true" />
            Work exceptions
          </UiButton>
          <UiButton variant="secondary" @click="openFirstFailure">Trace failed transfer</UiButton>
        </ActionBar>
      </Panel>

      <Panel title="Switching API telemetry" eyebrow="Settlement reliability layer" :accent="switchingApiState">
        <div class="switching-brief">
          <Network :size="18" aria-hidden="true" />
          <span>
            <strong>Local settlement switching layer</strong>
            <small>{{ switchingPurpose }}</small>
          </span>
        </div>
        <div class="switching-check-list">
          <article v-for="check in switchingChecks" :key="check.label">
            <span>
              <strong>{{ check.label }}</strong>
              <small>{{ check.detail }}</small>
            </span>
            <strong>{{ check.value }}</strong>
            <HealthBadge :state="check.state" />
          </article>
        </div>
        <aside class="state-note" :class="`state-note--${switchingApiState}`">
          <RefreshCw :size="16" aria-hidden="true" />
          <span>{{ switchingApiSummary }}</span>
        </aside>
      </Panel>
    </section>

    <section class="ops-workbench-grid ops-workbench-grid--balanced">
      <Panel title="Local provider performance" eyebrow="Working count and best performer" :accent="localProviderTone">
        <div class="provider-operating-list">
          <button
            v-for="row in providerRows"
            :key="row.route.route"
            type="button"
            class="provider-operating-row"
            :class="{ 'is-best': row.isBest, 'is-pressured': row.isPressured }"
            @click="openRoute(row.route)"
          >
            <ProviderMark :provider="row.route.route" show-name />
            <span>
              <strong>{{ row.action }}</strong>
              <small>{{ row.route.penaltyReason }}</small>
            </span>
            <dl>
              <div><dt>P95</dt><dd>{{ row.route.p95CreditTime }}</dd></div>
              <div><dt>Open</dt><dd>{{ row.route.openCases }}</dd></div>
              <div><dt>Timeout</dt><dd>{{ row.route.timeoutRate }}</dd></div>
            </dl>
            <HealthBadge :state="row.route.state" :trigger="`${row.route.penaltyScore} pressure`" />
            <div class="route-penalty-meter" :aria-label="`${row.route.penaltyScore} route pressure`">
              <span :style="{ width: routePenaltyWidth(row.route.penaltyScore) }"></span>
            </div>
          </button>
        </div>
      </Panel>

      <Panel title="International banking partner SLAs" eyebrow="Breach status by contract" :accent="partnerSlaTone">
        <div class="partner-sla-list">
          <button v-for="row in partnerSlaRows" :key="row.contractId" type="button" @click="openPath('/settings', { tab: 'sla', focus: row.state === 'degraded' || row.state === 'blocked' ? 'breaches' : 'contracts' })">
            <span>
              <strong>{{ row.partner }}</strong>
              <small>{{ row.corridor }} / SLA {{ row.creditSla }}</small>
            </span>
            <dl>
              <div><dt>Breach</dt><dd>{{ row.breachRate }}</dd></div>
              <div><dt>Aging</dt><dd>{{ row.agingBreaches }}</dd></div>
              <div><dt>Oldest</dt><dd>{{ row.oldestBreach }}</dd></div>
            </dl>
            <small>{{ row.recommendedAction }}</small>
            <HealthBadge :state="row.state" />
          </button>
        </div>
      </Panel>
    </section>
  </section>
</template>
