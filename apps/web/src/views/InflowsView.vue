<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, FileCheck2, GitBranch, History, Inbox, Landmark, ListChecks, RefreshCw, Search, ShieldAlert } from '@lucide/vue'
import ActionBar from '../components/ActionBar.vue'
import DataTable from '../components/DataTable.vue'
import EmptyState from '../components/EmptyState.vue'
import HealthBadge from '../components/HealthBadge.vue'
import Panel from '../components/Panel.vue'
import TransactionTimeline from '../components/TransactionTimeline.vue'
import UiButton from '../components/UiButton.vue'
import {
  backoffForAttempt,
  confidenceState,
  finalLegStateLabels,
  outcomeConfidenceLabels,
  routePenaltyWidth,
  useInflows,
} from '../composables/useInboundTower'
import { useTowerRouting } from '../composables/useTowerRouting'
import type { DashboardMock, InflowDetailTab } from '../types'

const props = defineProps<{
  dashboard: DashboardMock
}>()

const inflowTabs: Array<{ id: InflowDetailTab; label: string; icon: unknown }> = [
  { id: 'summary', label: 'Summary', icon: Landmark },
  { id: 'route', label: 'Route decision', icon: GitBranch },
  { id: 'timeline', label: 'Timeline', icon: History },
  { id: 'evidence', label: 'Evidence', icon: FileCheck2 },
  { id: 'requery', label: 'Requery', icon: RefreshCw },
  { id: 'reconciliation', label: 'Reconciliation', icon: ListChecks },
  { id: 'audit', label: 'Audit', icon: ShieldAlert },
]

const { route, openPath, activate, openInflow, openRoute } = useTowerRouting()
const selectedReference = computed(() => (typeof route.params.reference === 'string' ? route.params.reference : ''))
const activeTab = computed<InflowDetailTab>(() =>
  inflowTabs.some((tab) => tab.id === route.query.tab) ? (route.query.tab as InflowDetailTab) : 'summary',
)
const {
  inflowQuery,
  inflowStateFilter,
  inflowRouteFilter,
  inflowPartnerFilter,
  inflowDestinationFilter,
  inflowValueFilter,
  inflowSlaAgeFilter,
  inflowOutcomeFilter,
  inflowStateOptions,
  inflowRouteOptions,
  inflowPartnerOptions,
  inflowDestinationOptions,
  inflowValueOptions,
  inflowSlaAgeOptions,
  inflowOutcomeOptions,
  proofLanes,
  filteredInflows,
  selectedInflow,
  selectedAttempt,
  selectedEvidence,
  selectedRequeries,
  selectedRemediationCase,
  selectedRoutePenalty,
  selectedTimeline,
} = useInflows(props.dashboard, selectedReference)

const isDetail = computed(() => Boolean(selectedReference.value && selectedInflow.value))
const selectedContract = computed(() => props.dashboard.inboundContracts.find((item) => item.id === selectedInflow.value?.contractId) ?? null)
const selectedAccount = computed(() => props.dashboard.standingAccounts.find((item) => item.id === selectedContract.value?.standingAccountId) ?? null)
const selectedRouteDecision = computed(() => props.dashboard.routeDecisions.find((item) => item.instructionReference === selectedInflow.value?.reference) ?? null)
const selectedRequirements = computed(() =>
  selectedInflow.value ? props.dashboard.evidenceRequirements.filter((item) => item.instructionReference === selectedInflow.value?.reference) : [],
)
const selectedReconciliation = computed(() =>
  selectedInflow.value ? props.dashboard.reconciliationMatches.find((item) => item.instructionReference === selectedInflow.value?.reference) ?? null : null,
)
const selectedAudit = computed(() =>
  selectedInflow.value
    ? props.dashboard.auditEvents.filter((event) => `${event.object} ${event.reason}`.includes(selectedInflow.value?.reference ?? ''))
    : [],
)
const nextRequery = computed(() => selectedRequeries.value.find((attempt) => attempt.completedAt === '-') ?? null)

function setTab(tab: InflowDetailTab) {
  if (!selectedInflow.value) return
  openPath(`/inflows/${encodeURIComponent(selectedInflow.value.reference)}`, { tab })
}

function openRouteContext() {
  if (selectedRoutePenalty.value) openRoute(selectedRoutePenalty.value)
  else activate('routes')
}
</script>

<template>
  <section class="screen-stack">
    <template v-if="!isDetail">
      <Panel title="Inflow search" eyebrow="Live instruction queue" accent="healthy">
        <ActionBar>
          <label class="search-field">
            <Search :size="18" aria-hidden="true" />
            <input v-model="inflowQuery" type="search" aria-label="Search inflows" placeholder="Switch ref, partner ref, bank ref, beneficiary, amount, batch" />
          </label>
          <select v-model="inflowStateFilter" aria-label="SLA state filter">
            <option v-for="option in inflowStateOptions" :key="option">{{ option }}</option>
          </select>
          <span class="dashboard-chip">{{ filteredInflows.length }} of {{ dashboard.incomingInstructions.length }}</span>
        </ActionBar>
        <div class="filter-row">
          <select v-model="inflowRouteFilter" aria-label="Route filter"><option v-for="option in inflowRouteOptions" :key="option">{{ option }}</option></select>
          <select v-model="inflowPartnerFilter" aria-label="Partner filter"><option v-for="option in inflowPartnerOptions" :key="option">{{ option }}</option></select>
          <select v-model="inflowDestinationFilter" aria-label="Destination bank filter"><option v-for="option in inflowDestinationOptions" :key="option">{{ option }}</option></select>
          <select v-model="inflowValueFilter" aria-label="Value band filter"><option v-for="option in inflowValueOptions" :key="option">{{ option }}</option></select>
          <select v-model="inflowSlaAgeFilter" aria-label="SLA age filter"><option v-for="option in inflowSlaAgeOptions" :key="option">{{ option }}</option></select>
          <select v-model="inflowOutcomeFilter" aria-label="Outcome filter"><option v-for="option in inflowOutcomeOptions" :key="option">{{ option }}</option></select>
        </div>
      </Panel>

      <Panel title="Repair queues" eyebrow="Grouped by next operations action" accent="watch">
        <div class="proof-lane-grid">
          <button v-for="lane in proofLanes" :key="lane.id" type="button" :class="`proof-lane proof-lane--${lane.state}`" @click="lane.items[0] && openInflow(lane.items[0].reference)">
            <span>{{ lane.label }}</span>
            <strong>{{ lane.items.length }}</strong>
            <small>{{ lane.items[0]?.reference ?? lane.empty }}</small>
          </button>
        </div>
      </Panel>

      <Panel title="Incoming transfers" eyebrow="Search by reference, beneficiary, bank, route, or amount" accent="watch">
        <DataTable :empty="filteredInflows.length === 0" empty-title="No active inflows" empty-description="No incoming instruction matches the selected monitoring filter.">
          <table class="compact-table">
            <thead>
              <tr>
                <th>Instruction</th>
                <th>Beneficiary</th>
                <th>Route</th>
                <th>SLA</th>
                <th>Outcome</th>
                <th>Next step</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="inflow in filteredInflows"
                :key="inflow.reference"
                class="click-row"
                role="button"
                tabindex="0"
                :aria-label="`Open inflow ${inflow.reference}`"
                @click="openInflow(inflow.reference)"
                @keydown.enter.prevent="openInflow(inflow.reference)"
                @keydown.space.prevent="openInflow(inflow.reference)"
              >
                <td data-label="Instruction"><strong class="mono">{{ inflow.reference }}</strong><small class="muted-line">{{ inflow.partnerReference }} / {{ inflow.settlementBatch }}</small></td>
                <td data-label="Beneficiary"><strong>{{ inflow.beneficiary }}</strong><small class="muted-line">{{ inflow.beneficiaryAccount }} / {{ inflow.destinationBank }}</small></td>
                <td data-label="Route">{{ inflow.route }}<small class="muted-line">{{ inflow.amount }}</small></td>
                <td data-label="SLA"><strong>{{ finalLegStateLabels[inflow.currentState] }}</strong><small class="muted-line">Deadline {{ inflow.slaDeadline }}</small></td>
                <td data-label="Outcome"><HealthBadge :state="confidenceState(inflow.outcomeConfidence)" :trigger="outcomeConfidenceLabels[inflow.outcomeConfidence]" /></td>
                <td data-label="Next step">
                  {{ inflow.staffAssignment?.staffName ?? inflow.owner }}
                  <small class="muted-line">{{ inflow.safeAction }}</small>
                </td>
              </tr>
            </tbody>
          </table>
        </DataTable>
      </Panel>
    </template>

    <template v-else-if="selectedInflow">
      <section class="detail-workspace">
        <Panel title="Inflow detail" eyebrow="Transaction trace" :accent="selectedInflow.state" class="detail-primary-panel">
          <div class="detail-heading">
            <div>
              <p class="section-kicker mono">{{ selectedInflow.reference }}</p>
              <h3>{{ selectedInflow.beneficiary }}</h3>
              <p>{{ selectedInflow.safeAction }}</p>
            </div>
            <HealthBadge :state="selectedInflow.state" :trigger="finalLegStateLabels[selectedInflow.currentState]" />
          </div>

          <div class="segmented-group detail-tabs">
            <button v-for="tab in inflowTabs" :key="tab.id" type="button" :class="{ 'is-selected': activeTab === tab.id }" @click="setTab(tab.id)">
              <component :is="tab.icon" :size="15" aria-hidden="true" />
              <strong>{{ tab.label }}</strong>
            </button>
          </div>

          <section v-if="activeTab === 'summary'" class="flow-tab-panel">
            <dl class="metric-grid metric-grid--three">
              <div><dt>Amount</dt><dd>{{ selectedInflow.amount }}</dd></div>
              <div><dt>Partner</dt><dd>{{ selectedInflow.origin }}</dd></div>
              <div><dt>Settlement batch</dt><dd>{{ selectedInflow.settlementBatch }}</dd></div>
              <div><dt>SLA state</dt><dd>{{ finalLegStateLabels[selectedInflow.currentState] }}</dd></div>
              <div><dt>Customer owner</dt><dd>{{ selectedInflow.staffAssignment?.staffName ?? selectedInflow.owner }}</dd></div>
              <div><dt>Assignment</dt><dd>{{ selectedInflow.staffAssignment?.assignmentRule ?? 'No customer owner required' }}</dd></div>
              <div><dt>Outcome</dt><dd>{{ outcomeConfidenceLabels[selectedInflow.outcomeConfidence] }}</dd></div>
            </dl>
            <aside v-if="selectedInflow.staffAssignment" class="owner-assignment-card">
              <strong>{{ selectedInflow.staffAssignment.staffName }}</strong>
              <span>{{ selectedInflow.staffAssignment.staffRole }} · {{ selectedInflow.staffAssignment.team }}</span>
            </aside>
            <div class="detail-status-rail">
              <article><strong>{{ selectedContract?.partner ?? selectedInflow.origin }}</strong><small>{{ selectedContract?.corridor ?? selectedInflow.destination }}</small></article>
              <article><strong>{{ selectedAccount?.availableLimit ?? '-' }}</strong><small>Standing account availability</small></article>
              <article><strong>{{ selectedInflow.destinationBank }}</strong><small>{{ selectedInflow.beneficiaryAccount }}</small></article>
            </div>
          </section>

          <section v-else-if="activeTab === 'route'" class="flow-tab-panel">
            <div class="selected-pressure-card" v-if="selectedRoutePenalty">
              <header>
                <div>
                  <span class="section-kicker">Final-leg route pressure</span>
                  <strong>{{ selectedRoutePenalty.route }}</strong>
                  <small>{{ selectedRoutePenalty.penaltyReason }}</small>
                </div>
                <HealthBadge :state="selectedRoutePenalty.state" :trigger="`${selectedRoutePenalty.penaltyScore} pressure`" />
              </header>
              <div class="route-penalty-meter" :aria-label="`${selectedRoutePenalty.penaltyScore} route pressure`">
                <span :style="{ width: routePenaltyWidth(selectedRoutePenalty.penaltyScore) }"></span>
              </div>
            </div>
            <dl class="metric-grid metric-grid--three">
              <div><dt>Policy version</dt><dd>{{ selectedRouteDecision?.policyVersion ?? 'Recorded in route audit' }}</dd></div>
              <div><dt>Selected route</dt><dd>{{ selectedRouteDecision?.selectedRoute ?? selectedInflow.route }}</dd></div>
              <div><dt>Decision time</dt><dd>{{ selectedRouteDecision?.decidedAt ?? selectedAttempt?.submittedAt ?? '-' }}</dd></div>
            </dl>
            <div class="score-input-grid">
              <article v-for="input in selectedRouteDecision?.scoreInputs ?? []" :key="input.label">
                <HealthBadge :state="input.state" />
                <strong>{{ input.label }}</strong>
                <small>{{ input.value }}</small>
              </article>
            </div>
            <div class="policy-stack">
              <article v-for="item in selectedRouteDecision?.rejectedRoutes ?? []" :key="`${item.provider}-${item.route}`">
                <span><strong>{{ item.route }}</strong><small>{{ item.provider }}</small></span>
                <p>{{ item.reason }}</p>
              </article>
            </div>
          </section>

          <section v-else-if="activeTab === 'timeline'" class="flow-tab-panel">
            <TransactionTimeline :steps="selectedTimeline" />
          </section>

          <section v-else-if="activeTab === 'evidence'" class="flow-tab-panel">
            <div class="requirement-grid">
              <article v-for="item in selectedRequirements" :key="item.id">
                <FileCheck2 :size="17" aria-hidden="true" />
                <span><strong>{{ item.label }}</strong><small>{{ item.capturedReference ?? 'Not captured' }} / {{ item.status }}</small></span>
                <HealthBadge :state="item.state" :trigger="item.required ? 'Required' : 'Optional'" />
              </article>
            </div>
            <div class="evidence-list">
              <article v-for="item in selectedEvidence" :key="`${item.type}-${item.reference}`">
                <FileCheck2 :size="18" aria-hidden="true" />
                <span><strong>{{ item.label }}</strong><small>{{ item.reference }} / {{ item.owner }}</small></span>
                <HealthBadge :state="item.state" :trigger="item.status" />
              </article>
            </div>
          </section>

          <section v-else-if="activeTab === 'requery'" class="flow-tab-panel">
            <section class="backoff-ladder">
              <header>
                <strong>Backoff ladder</strong>
                <small>{{ backoffForAttempt(dashboard, selectedAttempt)?.schedule.join(' / ') ?? 'No requery policy' }}</small>
              </header>
              <ol>
                <li v-for="attempt in selectedRequeries" :key="attempt.id">
                  <span>{{ attempt.attemptNumber }}</span>
                  <strong>{{ attempt.method }}</strong>
                  <small>{{ attempt.dueAt }} / {{ attempt.result }}</small>
                  <HealthBadge :state="attempt.state" />
                </li>
              </ol>
              <p v-if="selectedRequeries.length === 0">No requery scheduled.</p>
            </section>
            <aside class="state-note">
              <RefreshCw :size="16" aria-hidden="true" />
              <span>{{ nextRequery ? `Next automatic attempt ${nextRequery.dueAt}` : 'No automatic requery is pending.' }}</span>
            </aside>
          </section>

          <section v-else-if="activeTab === 'reconciliation'" class="flow-tab-panel">
            <dl class="metric-grid metric-grid--three">
              <div><dt>Match state</dt><dd>{{ selectedReconciliation?.matchState ?? 'No match record' }}</dd></div>
              <div><dt>Provider file</dt><dd>{{ selectedReconciliation?.providerFileReference ?? '-' }}</dd></div>
              <div><dt>Bank ledger</dt><dd>{{ selectedReconciliation?.bankLedgerReference ?? '-' }}</dd></div>
            </dl>
            <aside class="state-note" :class="`state-note--${selectedReconciliation?.state ?? 'unknown'}`">
              <ListChecks :size="16" aria-hidden="true" />
              <span>{{ selectedReconciliation?.mismatchReason ?? 'No break linked.' }}</span>
            </aside>
          </section>

          <section v-else class="flow-tab-panel">
            <div class="audit-preview">
              <article v-for="event in selectedAudit" :key="`${event.time}-${event.action}-${event.object}`">
                <span>{{ event.time }}</span>
                <strong>{{ event.action }}</strong>
                <small>{{ event.object }} / {{ event.reason }}</small>
              </article>
              <EmptyState v-if="selectedAudit.length === 0" title="No audit events" description="Actions recorded as the case progresses." :icon="ShieldAlert" />
            </div>
          </section>
        </Panel>

        <aside class="safe-action-panel">
          <UiButton size="sm" variant="secondary" @click="openPath('/inflows')">
            <ArrowLeft :size="15" aria-hidden="true" />
            Back to queue
          </UiButton>
          <article>
            <span>Safe action</span>
            <strong>{{ selectedInflow.safeAction }}</strong>
            <small>{{ selectedInflow.staffAssignment?.staffName ?? selectedInflow.owner }}</small>
          </article>
          <article v-if="selectedInflow.staffAssignment">
            <span>Customer owner</span>
            <strong>{{ selectedInflow.staffAssignment.staffName }}</strong>
            <small>{{ selectedInflow.staffAssignment.assignmentRule }} · {{ selectedInflow.staffAssignment.assignedAt }}</small>
          </article>
          <article>
            <span>Linked work</span>
            <strong>{{ selectedRemediationCase?.id ?? '—' }}</strong>
            <small>{{ selectedRemediationCase?.nextAction ?? 'No open case' }}</small>
          </article>
          <ActionBar>
            <UiButton v-if="selectedRemediationCase" @click="openPath(`/exceptions/${encodeURIComponent(selectedInflow.reference)}`, { tab: 'evidence' })">Open exception</UiButton>
            <UiButton variant="secondary" @click="openRouteContext">Route context</UiButton>
          </ActionBar>
        </aside>
      </section>
    </template>

    <EmptyState v-else title="No inflow selected" description="Select an incoming instruction to inspect final-leg evidence." :icon="Inbox" />
  </section>
</template>
