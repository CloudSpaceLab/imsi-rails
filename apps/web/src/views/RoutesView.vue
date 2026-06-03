<script setup lang="ts">
import { computed } from 'vue'
import DataTable from '../components/DataTable.vue'
import EmptyState from '../components/EmptyState.vue'
import HealthBadge from '../components/HealthBadge.vue'
import Panel from '../components/Panel.vue'
import ProviderMark from '../components/ProviderMark.vue'
import UiButton from '../components/UiButton.vue'
import ActionBar from '../components/ActionBar.vue'
import { routePenaltyWidth, useRouteHealth } from '../composables/useInboundTower'
import { useTowerRouting } from '../composables/useTowerRouting'
import type { DashboardMock, HealthState, RouteDetailTab } from '../types'

const props = defineProps<{
  dashboard: DashboardMock
}>()

const emit = defineEmits<{
  settingsTab: [tab: 'sla']
}>()

const routeTabs: Array<{ id: RouteDetailTab; label: string }> = [
  { id: 'overview', label: 'Overview' },
  { id: 'traffic', label: 'Traffic' },
  { id: 'cases', label: 'Cases' },
  { id: 'history', label: 'History' },
  { id: 'policy', label: 'Policy' },
]

const { route, openPath, activate, openRoute, openInflow, openException } = useTowerRouting()
const selectedRouteId = computed(() => (typeof route.params.routeId === 'string' ? route.params.routeId : ''))
const routeTab = computed<RouteDetailTab>(() => (routeTabs.some((tab) => tab.id === route.query.tab) ? (route.query.tab as RouteDetailTab) : 'overview'))
const { routePenalties, selectedRoutePenalty } = useRouteHealth(props.dashboard, selectedRouteId)
const providerRankings = computed(() => [...props.dashboard.providerScores].sort((a, b) => a.rank - b.rank))
const bestProvider = computed(() => providerRankings.value[0] ?? null)
const worstProvider = computed(() => providerRankings.value[providerRankings.value.length - 1] ?? null)
const providerBoardAccent = computed(() => worstProvider.value?.state ?? 'healthy')
const externalRoutes = computed(() => routePenalties.value.filter((routeRow) => routeRow.route !== 'Fidelity core'))
const lowestPenaltyExternalRoute = computed(
  () => [...externalRoutes.value].sort((a, b) => a.penaltyScore - b.penaltyScore)[0] ?? routePenalties.value[routePenalties.value.length - 1] ?? null,
)
const routeTrafficControls = computed(() =>
  routePenalties.value
    .map((routeRow) => {
      const isCore = routeRow.route === 'Fidelity core'
      const isLowestExternal = routeRow.route === lowestPenaltyExternalRoute.value?.route
      const isHighPressure = routeRow.penaltyScore >= 30 || routeRow.state === 'degraded' || routeRow.state === 'blocked'
      const stance = isCore
        ? 'Maintain direct credits'
        : isHighPressure
          ? 'Contain new traffic'
          : isLowestExternal && routeRow.state === 'healthy'
            ? 'Increase eligible traffic'
            : isLowestExternal
              ? 'Recovery test only'
              : routeRow.state === 'watch'
                ? 'Limit allocation'
                : 'Maintain allocation'
      const lane = isCore || (isLowestExternal && routeRow.state === 'healthy') ? 'leader' : isHighPressure ? 'risk' : routeRow.state === 'watch' ? 'watch' : 'steady'
      const accent: HealthState = lane === 'leader' ? 'healthy' : lane === 'risk' ? 'degraded' : lane === 'watch' ? 'watch' : routeRow.state
      const reason = isCore
        ? `${routeRow.p95CreditTime} P95 and ${routeRow.timeoutRate} timeout on Fidelity beneficiaries.`
        : isHighPressure
          ? `${routeRow.penaltyReason} Keep fallback on ${routeRow.fallbackOrder}.`
          : isLowestExternal
            ? `${routeRow.penaltyReason} Use controlled traffic while proof improves.`
            : `${routeRow.trafficSplit}; watch ${routeRow.timeoutRate} timeout and ${routeRow.lateSuccessRate} late success.`

      return { route: routeRow, stance, lane, accent, reason }
    })
    .sort((a, b) => {
      const priority = { leader: 0, risk: 1, watch: 2, steady: 3 }
      return priority[a.lane as keyof typeof priority] - priority[b.lane as keyof typeof priority] || b.route.penaltyScore - a.route.penaltyScore
    }),
)
const selectedRouteControl = computed(() => routeTrafficControls.value.find((item) => item.route.route === selectedRoutePenalty.value?.route) ?? null)
const linkedInflows = computed(() =>
  selectedRoutePenalty.value ? props.dashboard.incomingInstructions.filter((item) => item.route === selectedRoutePenalty.value?.route) : [],
)
const linkedCases = computed(() =>
  selectedRoutePenalty.value ? props.dashboard.remediationCases.filter((item) => item.route === selectedRoutePenalty.value?.route) : [],
)
const selectedRouteWindows = computed(() =>
  selectedRoutePenalty.value ? props.dashboard.routeHealthWindows.filter((item) => item.route === selectedRoutePenalty.value?.route) : [],
)

function openPolicy() {
  emit('settingsTab', 'sla')
  openPath('/settings', { tab: 'sla', focus: 'route-policy' })
}

function setRouteTab(tab: RouteDetailTab) {
  if (!selectedRoutePenalty.value) return
  openPath(`/routes/${selectedRouteId.value || selectedRoutePenalty.value.route.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '')}`, { tab })
}
</script>

<template>
  <section class="screen-stack">
    <section class="dashboard-grid">
      <Panel title="Payment provider scorecards" eyebrow="Best/worst by credit proof telemetry" :accent="providerBoardAccent" class="span-8">
        <div class="provider-ranking-list provider-ranking-list--routes" aria-label="Payment provider ranking">
          <button v-for="provider in providerRankings" :key="provider.provider" type="button" @click="activate('inflows')">
            <span class="provider-rank-badge">#{{ provider.rank }}</span>
            <ProviderMark :provider="provider.provider" show-name show-category />
            <dl>
              <div><dt>Success</dt><dd>{{ provider.successRate }}</dd></div>
              <div><dt>P95</dt><dd>{{ provider.p95 }}</dd></div>
              <div><dt>Stuck</dt><dd>{{ provider.stuckRate }}</dd></div>
              <div><dt>Cases</dt><dd>{{ provider.settlementExceptions }}</dd></div>
            </dl>
            <HealthBadge :state="provider.state" :trigger="`${provider.trafficShare} traffic`" />
          </button>
        </div>
      </Panel>

      <Panel title="Traffic instruction" eyebrow="Provider routing stance" :accent="worstProvider?.state ?? 'healthy'" class="span-4">
        <div class="next-action-card">
          <HealthBadge :state="bestProvider?.state ?? 'healthy'" :trigger="bestProvider ? `${bestProvider.provider} leads` : 'No leader' " />
          <h3>Prefer {{ bestProvider?.provider ?? 'healthy providers' }} for new payouts</h3>
          <p>
            {{ bestProvider?.provider ?? 'The leading provider' }} is currently strongest on success and P95 credit proof.
            Keep {{ worstProvider?.provider ?? 'degraded providers' }} under watch until stuck rate and settlement exceptions recover.
          </p>
          <div class="decision-callouts">
            <article>
              <span>Best route signal</span>
              <strong>{{ bestProvider?.successRate ?? '-' }} success</strong>
              <small>{{ bestProvider?.p95 ?? '-' }} P95 / {{ bestProvider?.trafficShare ?? '0%' }} traffic</small>
            </article>
            <article class="is-risk">
              <span>Provider to contain</span>
              <strong>{{ worstProvider?.provider ?? 'None' }}</strong>
              <small>{{ worstProvider?.stuckRate ?? '0%' }} stuck / {{ worstProvider?.settlementExceptions ?? 0 }} exceptions</small>
            </article>
          </div>
        </div>
      </Panel>
    </section>

    <Panel title="Route traffic controls" eyebrow="Allocation stance by final-leg rail" accent="watch">
      <div class="route-control-grid">
        <button
          v-for="control in routeTrafficControls"
          :key="control.route.route"
          type="button"
          :class="`route-control-card route-control-card--${control.lane}`"
          @click="openRoute(control.route)"
        >
          <span>
            <small>Traffic action</small>
            <strong>{{ control.stance }}</strong>
          </span>
          <div>
            <strong>{{ control.route.route }}</strong>
            <small>{{ control.route.rail }} / {{ control.route.destinationBank }}</small>
          </div>
          <dl>
            <div><dt>Split</dt><dd>{{ control.route.trafficSplit }}</dd></div>
            <div><dt>Timeout</dt><dd>{{ control.route.timeoutRate }}</dd></div>
            <div><dt>Fallback</dt><dd>{{ control.route.fallbackOrder }}</dd></div>
          </dl>
          <p>{{ control.reason }}</p>
          <HealthBadge :state="control.accent" :trigger="`${control.route.penaltyScore} penalty`" />
        </button>
      </div>
    </Panel>

    <Panel title="Route detail" eyebrow="Rail drilldown and linked work" :accent="selectedRoutePenalty?.state ?? 'unknown'">
      <template v-if="selectedRoutePenalty">
        <div class="detail-heading">
          <div>
            <p class="section-kicker">{{ selectedRoutePenalty.rail }}</p>
            <h3>{{ selectedRoutePenalty.route }}</h3>
            <p>{{ selectedRoutePenalty.penaltyReason }}</p>
          </div>
          <HealthBadge :state="selectedRoutePenalty.state" :trigger="`${selectedRoutePenalty.penaltyScore} route pressure`" />
        </div>

        <div class="segmented-group detail-tabs">
          <button v-for="tab in routeTabs" :key="tab.id" type="button" :class="{ 'is-selected': routeTab === tab.id }" @click="setRouteTab(tab.id)">
            <strong>{{ tab.label }}</strong>
          </button>
        </div>

        <section v-if="routeTab === 'overview'" class="flow-tab-panel">
          <dl class="metric-grid metric-grid--four">
            <div><dt>P95</dt><dd>{{ selectedRoutePenalty.p95CreditTime }}</dd></div>
            <div><dt>Timeout</dt><dd>{{ selectedRoutePenalty.timeoutRate }}</dd></div>
            <div><dt>Late success</dt><dd>{{ selectedRoutePenalty.lateSuccessRate }}</dd></div>
            <div><dt>Open cases</dt><dd>{{ selectedRoutePenalty.openCases }}</dd></div>
          </dl>
          <div class="route-penalty-meter">
            <span :style="{ width: routePenaltyWidth(selectedRoutePenalty.penaltyScore) }"></span>
          </div>
        </section>

        <section v-else-if="routeTab === 'traffic'" class="flow-tab-panel">
          <div class="route-context-row">
            <article><strong>{{ selectedRouteControl?.stance ?? 'Maintain allocation' }}</strong><small>Suggested stance</small></article>
            <article><strong>{{ selectedRoutePenalty.trafficSplit }}</strong><small>Current split</small></article>
            <article><strong>{{ selectedRoutePenalty.fallbackOrder }}</strong><small>Fallback order</small></article>
          </div>
          <div class="policy-stack">
            <article v-for="preset in dashboard.routeConfig.presets" :key="preset.label">
              <HealthBadge :state="preset.active ? 'recovery' : 'stale'" :trigger="preset.active ? 'Active' : 'Preset'" />
              <span><strong>{{ preset.label }}</strong><small>{{ preset.split }}</small></span>
            </article>
          </div>
        </section>

        <section v-else-if="routeTab === 'cases'" class="flow-tab-panel">
          <div class="linked-work-grid">
            <button v-for="item in linkedInflows" :key="item.reference" type="button" @click="openInflow(item.reference)">
              <strong>{{ item.reference }}</strong>
              <small>{{ item.beneficiary }} / {{ item.currentState }}</small>
              <HealthBadge :state="item.state" />
            </button>
            <button v-for="item in linkedCases" :key="item.id" type="button" @click="openException(item.instructionReference)">
              <strong>{{ item.id }}</strong>
              <small>{{ item.title }} / {{ item.age }}</small>
              <HealthBadge :state="item.state" />
            </button>
          </div>
          <EmptyState v-if="linkedInflows.length === 0 && linkedCases.length === 0" title="No linked work" description="No inflow or exception is currently attached to this route." />
        </section>

        <section v-else-if="routeTab === 'history'" class="flow-tab-panel">
          <div class="window-history-list">
            <article v-for="window in selectedRouteWindows" :key="`${window.route}-${window.window}`">
              <span><strong>{{ window.window }}</strong><small>{{ window.submittedCount }} submitted / {{ window.creditedCount }} credited</small></span>
              <dl>
                <div><dt>P95</dt><dd>{{ window.p95CreditTime }}</dd></div>
                <div><dt>Timeouts</dt><dd>{{ window.timeoutCount }}</dd></div>
                <div><dt>Open</dt><dd>{{ window.openCases }}</dd></div>
              </dl>
              <HealthBadge :state="window.state" />
            </article>
          </div>
          <div class="audit-preview">
            <article v-for="event in dashboard.downtimeEvents" :key="`${event.time}-${event.title}`">
              <span>{{ event.time }}</span>
              <strong>{{ event.title }}</strong>
              <small>{{ event.detail }}</small>
            </article>
          </div>
        </section>

        <section v-else class="flow-tab-panel">
          <div class="policy-stack">
            <article v-for="item in dashboard.routeConfig.workflow.currentPolicy" :key="`current-${item.label}`">
              <span><strong>{{ item.label }}</strong><small>{{ item.value }}</small></span>
            </article>
            <article v-for="item in dashboard.routeConfig.workflow.validation" :key="`validation-${item.label}`">
              <HealthBadge :state="item.state" />
              <span><strong>{{ item.label }}</strong><small>{{ item.value }}</small></span>
            </article>
          </div>
          <ActionBar>
            <UiButton size="sm" @click="openPolicy">Open SLA policy</UiButton>
            <UiButton size="sm" variant="secondary" @click="activate('exceptions')">Linked exceptions</UiButton>
          </ActionBar>
        </section>
      </template>
    </Panel>

    <section class="dashboard-grid">
      <Panel title="Final-leg rail pressure" eyebrow="Eligible rails for new traffic" accent="watch" class="span-8">
        <div class="route-health-strip route-health-strip--stacked">
          <button
            v-for="routeRow in routePenalties"
            :key="routeRow.route"
            type="button"
            :class="{ 'is-selected': selectedRoutePenalty?.route === routeRow.route }"
            @click="openRoute(routeRow)"
          >
            <span>
              <strong>{{ routeRow.route }}</strong>
              <small>{{ routeRow.penaltyReason }}</small>
            </span>
            <div class="route-penalty-meter">
              <span :style="{ width: routePenaltyWidth(routeRow.penaltyScore) }"></span>
            </div>
            <HealthBadge :state="routeRow.state" :trigger="`${routeRow.penaltyScore} penalty`" />
          </button>
        </div>
      </Panel>

      <Panel title="Highest pressure rail" eyebrow="Penalty explanation" :accent="selectedRoutePenalty?.state ?? 'unknown'" class="span-4">
        <template v-if="selectedRoutePenalty">
          <div class="selected-route-card">
            <h3>{{ selectedRoutePenalty.route }}</h3>
            <p>{{ selectedRoutePenalty.penaltyReason }}</p>
            <dl class="metric-grid">
              <div><dt>P95</dt><dd>{{ selectedRoutePenalty.p95CreditTime }}</dd></div>
              <div><dt>Timeout</dt><dd>{{ selectedRoutePenalty.timeoutRate }}</dd></div>
              <div><dt>Late success</dt><dd>{{ selectedRoutePenalty.lateSuccessRate }}</dd></div>
              <div><dt>Open cases</dt><dd>{{ selectedRoutePenalty.openCases }}</dd></div>
            </dl>
            <ActionBar>
              <UiButton size="sm" variant="secondary" @click="activate('exceptions')">Linked cases</UiButton>
              <UiButton size="sm" @click="openPolicy">Policy</UiButton>
            </ActionBar>
          </div>
        </template>
      </Panel>
    </section>

    <Panel title="Route matrix" eyebrow="Destination bank, traffic split, and fallback order" accent="healthy">
      <DataTable :empty="routePenalties.length === 0" empty-title="No routes" empty-description="No route health is available for this window.">
        <table class="compact-table">
          <thead>
            <tr>
              <th>Route</th>
              <th>Destination</th>
              <th>Traffic split</th>
              <th>Fallback</th>
              <th>Penalty</th>
              <th>State</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="routeRow in routePenalties"
              :key="routeRow.route"
              class="click-row"
              role="button"
              tabindex="0"
              :aria-label="`Open route ${routeRow.route}`"
              @click="openRoute(routeRow)"
              @keydown.enter.prevent="openRoute(routeRow)"
              @keydown.space.prevent="openRoute(routeRow)"
            >
              <td data-label="Route"><strong>{{ routeRow.route }}</strong><small class="muted-line">{{ routeRow.rail }}</small></td>
              <td data-label="Destination">{{ routeRow.destinationBank }}</td>
              <td data-label="Traffic split">{{ routeRow.trafficSplit }}</td>
              <td data-label="Fallback">{{ routeRow.fallbackOrder }}</td>
              <td data-label="Penalty">
                <strong>{{ routeRow.penaltyScore }}</strong>
                <small class="muted-line">{{ routeRow.penaltyReason }}</small>
              </td>
              <td data-label="State"><HealthBadge :state="routeRow.state" /></td>
            </tr>
          </tbody>
        </table>
      </DataTable>
    </Panel>
  </section>
</template>
