<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { AlertTriangle, BadgeCheck, CheckCircle2, Clock3, FileCheck2, RefreshCw, ShieldAlert, ShieldCheck } from '@lucide/vue'
import ActionBar from '../components/ActionBar.vue'
import EmptyState from '../components/EmptyState.vue'
import HealthBadge from '../components/HealthBadge.vue'
import KpiTile from '../components/KpiTile.vue'
import Panel from '../components/Panel.vue'
import UiButton from '../components/UiButton.vue'
import { formatNgnShort, parseMoney, queueLabels, useRemediationQueue } from '../composables/useInboundTower'
import { useTowerRouting } from '../composables/useTowerRouting'
import type { CaseActionStep, DashboardMock, ExceptionDetailTab, RemediationCase } from '../types'

const props = defineProps<{
  dashboard: DashboardMock
  actorName: string
}>()

const exceptionTabs: Array<{ id: ExceptionDetailTab; label: string }> = [
  { id: 'summary', label: 'Case summary' },
  { id: 'evidence', label: 'Evidence gaps' },
  { id: 'requery', label: 'Backoff/requery' },
  { id: 'closure', label: 'Closure' },
  { id: 'audit', label: 'Audit' },
]

const { route, openPath, openException } = useTowerRouting()
const selectedReference = computed(() => (typeof route.params.reference === 'string' ? route.params.reference : ''))
const actorNameRef = computed(() => props.actorName)
const {
  queueFilter,
  exceptionEvidence,
  exceptionReason,
  exceptionActionMessage,
  exceptionActionError,
  exceptionActionPending,
  selectedRequeryMethod,
  queueTabs,
  closureReadinessLanes,
  filteredCases,
  selectedException,
  selectedExceptionEvidence,
  selectedExceptionInstruction,
  selectedExceptionAttempt,
  selectedExceptionRoutePenalty,
  selectedExceptionBackoff,
  selectedRequeryAttempts,
  automaticRequeryAttempts,
  manualRequeryAttempts,
  nextAutomaticAttempt,
  automationExhausted,
  exceptionCanSubmit,
  submitExceptionAction,
} = useRemediationQueue(props.dashboard, selectedReference, actorNameRef)

const selectedAction = ref<CaseActionStep['action']>('attach_evidence')
const exceptionTab = computed<ExceptionDetailTab>(() =>
  exceptionTabs.some((tab) => tab.id === route.query.tab) ? (route.query.tab as ExceptionDetailTab) : 'summary',
)
const valueAtRiskTotal = computed(() => formatNgnShort(props.dashboard.remediationCases.reduce((sum, item) => sum + parseMoney(item.valueAtRisk), 0)))
const selectedRequirements = computed(() =>
  selectedException.value ? props.dashboard.evidenceRequirements.filter((item) => item.instructionReference === selectedException.value?.instructionReference) : [],
)
const selectedCaseAudit = computed(() =>
  selectedException.value
    ? props.dashboard.auditEvents.filter((event) => `${event.object} ${event.reason}`.includes(selectedException.value?.instructionReference ?? ''))
    : [],
)
const allowedActionSteps = computed(() => {
  const item = selectedException.value
  if (!item) return []
  return props.dashboard.caseActionSteps.filter((step) => {
    if (!step.queues.includes(item.queue)) return false
    if (item.duplicateRisk === 'High') return step.action === 'attach_evidence'
    return step.duplicateRisk === 'Any' || step.duplicateRisk === item.duplicateRisk
  })
})
const selectedActionStep = computed(() => allowedActionSteps.value.find((step) => step.action === selectedAction.value) ?? allowedActionSteps.value[0] ?? null)
const automationStatusLabels = {
  cooldown: 'Cooldown',
  scheduled: 'Scheduled',
  running: 'Automation running',
  exhausted: 'Automation exhausted',
  manual_only: 'Manual only',
}
const makerCheckerStateLabels = {
  not_required: 'Not required',
  maker_required: 'Maker evidence required',
  checker_pending: 'Checker pending',
  approved: 'Approved',
  rejected: 'Rejected',
}
const requeryMethodOptions = computed(() => {
  const provider = selectedException.value?.provider ?? 'Provider'
  return Array.from(
    new Set([
      selectedException.value?.requeryMethod,
      `${provider} status API`,
      `${provider} callback replay`,
      'Rail session lookup',
      'Core ledger and suspense search',
    ].filter(Boolean) as string[]),
  )
})

watch(
  () => route.query.queue,
  (queue) => {
    if (typeof queue === 'string' && (queue === 'all' || queue in queueLabels)) queueFilter.value = queue as 'all' | RemediationCase['queue']
  },
  { immediate: true },
)

watch(
  allowedActionSteps,
  (steps) => {
    selectedAction.value = steps[0]?.action ?? 'attach_evidence'
  },
  { immediate: true },
)

function setQueue(queue: 'all' | RemediationCase['queue']) {
  queueFilter.value = queue
  openPath('/exceptions', { queue })
}

function setExceptionTab(tab: ExceptionDetailTab) {
  if (!selectedException.value) return
  openPath(`/exceptions/${encodeURIComponent(selectedException.value.instructionReference)}`, { tab, queue: queueFilter.value })
}

function selectCase(reference: string) {
  openPath(`/exceptions/${encodeURIComponent(reference)}`, { queue: queueFilter.value, tab: exceptionTab.value })
}

function submitSelectedAction() {
  const action = selectedActionStep.value?.action
  if (!action) return
  if (action === 'manual_requery') void submitExceptionAction('requery')
  else if (action === 'mark_completed_outside_platform') void submitExceptionAction('manual_complete')
  else if (action === 'approve_reversal') void submitExceptionAction('approve_reversal')
  else void submitExceptionAction('attach_evidence')
}
</script>

<template>
  <section class="screen-stack">
    <section class="kpi-grid">
      <KpiTile label="Manual queue" :value="dashboard.remediationCases.length" detail="Open remediation cases" tone="degraded" :icon="ShieldAlert" />
      <KpiTile label="Value at risk" :value="valueAtRiskTotal" detail="Requires evidence before closure" tone="watch" :icon="AlertTriangle" />
      <KpiTile label="Oldest unresolved" :value="dashboard.remediationCases[2]?.age ?? '0m'" detail="Sorted by risk and age" tone="watch" :icon="Clock3" />
      <KpiTile label="Duplicate-risk watch" :value="dashboard.remediationCases.filter((item) => item.duplicateRisk !== 'Low').length" detail="Evidence attachment only until risk clears" tone="degraded" :icon="ShieldCheck" />
    </section>

    <Panel title="Exception queues" eyebrow="Cooldown, requery, exhausted, recon, reversals" accent="watch">
      <div class="segmented-group exception-tabs">
        <button v-for="tab in queueTabs" :key="tab.id" type="button" :class="{ 'is-selected': queueFilter === tab.id }" @click="setQueue(tab.id)">
          <strong>{{ tab.label }}</strong>
          <small>{{ tab.count }} open</small>
        </button>
      </div>
    </Panel>

    <Panel title="Closure readiness" eyebrow="Evidence, automation, checker, reversal" accent="degraded">
      <div class="closure-readiness-grid">
        <button
          v-for="lane in closureReadinessLanes"
          :key="lane.id"
          type="button"
          :class="`closure-readiness-card closure-readiness-card--${lane.state}`"
          :disabled="lane.items.length === 0"
          @click="lane.items[0] && openException(lane.items[0].instructionReference)"
        >
          <span>{{ lane.label }}</span>
          <strong>{{ lane.items.length }}</strong>
          <small>{{ lane.items[0]?.instructionReference ?? 'No open case' }}</small>
        </button>
      </div>
    </Panel>

    <section class="exception-workspace">
      <Panel title="Remediation queue" eyebrow="Evidence-first closure" accent="degraded">
        <div class="remediation-queue">
          <button
            v-for="item in filteredCases"
            :key="item.id"
            type="button"
            :class="{ 'is-selected': selectedException?.id === item.id }"
            @click="selectCase(item.instructionReference)"
          >
            <span>
              <strong>{{ item.title }}</strong>
              <small>{{ item.instructionReference }} / {{ queueLabels[item.queue] }}</small>
            </span>
            <span>
              <strong>{{ item.valueAtRisk }}</strong>
              <small>{{ item.staffAssignment?.staffName ?? item.owner }} / {{ item.age }} / duplicate risk {{ item.duplicateRisk }}</small>
            </span>
            <HealthBadge :state="item.state" />
          </button>
          <EmptyState v-if="filteredCases.length === 0" title="No exceptions in this queue" description="There is nothing to remediate for the selected queue." :icon="CheckCircle2" tone="success" />
        </div>
      </Panel>

      <Panel title="Case detail" eyebrow="Safe action and evidence" :accent="selectedException?.state ?? 'unknown'" class="case-panel">
        <template v-if="selectedException">
          <div class="detail-heading">
            <div>
              <p class="section-kicker">{{ selectedException.id }}</p>
              <h3>{{ selectedException.title }}</h3>
              <p>{{ selectedException.theory }}</p>
            </div>
            <HealthBadge :state="selectedException.state" :trigger="queueLabels[selectedException.queue]" />
          </div>

          <div class="segmented-group detail-tabs">
            <button v-for="tab in exceptionTabs" :key="tab.id" type="button" :class="{ 'is-selected': exceptionTab === tab.id }" @click="setExceptionTab(tab.id)">
              <strong>{{ tab.label }}</strong>
            </button>
          </div>

          <aside v-if="selectedException.staffAssignment" class="owner-assignment-card">
            <strong>Customer owner: {{ selectedException.staffAssignment.staffName }}</strong>
            <span>{{ selectedException.staffAssignment.staffRole }} / {{ selectedException.staffAssignment.team }}</span>
            <span>{{ selectedException.staffAssignment.assignmentRule }}</span>
            <small>{{ selectedException.staffAssignment.responsibility }}</small>
          </aside>

          <section v-if="exceptionTab === 'summary'" class="flow-tab-panel">
            <dl class="metric-grid">
              <div><dt>Value</dt><dd>{{ selectedException.valueAtRisk }}</dd></div>
              <div><dt>Age</dt><dd>{{ selectedException.age }}</dd></div>
              <div><dt>Duplicate risk</dt><dd>{{ selectedException.duplicateRisk }}</dd></div>
              <div><dt>Customer owner</dt><dd>{{ selectedException.staffAssignment?.staffName ?? selectedException.owner }}</dd></div>
              <div><dt>Assignment</dt><dd>{{ selectedException.staffAssignment?.assignmentRule ?? 'Manual queue assignment' }}</dd></div>
            </dl>
            <aside class="state-note">
              <ShieldAlert :size="16" aria-hidden="true" />
              <span>{{ selectedException.nextAction }}</span>
            </aside>
            <div class="route-context-row">
              <article><strong>{{ selectedExceptionInstruction?.destinationBank ?? selectedException.provider }}</strong><small>{{ selectedExceptionAttempt?.id ?? selectedException.instructionReference }}</small></article>
              <article><strong>{{ selectedExceptionRoutePenalty?.penaltyScore ?? 'N/A' }}</strong><small>Route penalty</small></article>
              <article><strong>{{ selectedExceptionBackoff?.schedule.join(' / ') ?? 'No policy' }}</strong><small>{{ selectedExceptionBackoff?.maxRequeryAttempts ?? selectedException.maxAutomaticAttempts }} automatic max</small></article>
            </div>
          </section>

          <section v-else-if="exceptionTab === 'evidence'" class="flow-tab-panel">
            <div class="requirement-grid">
              <article v-for="item in selectedRequirements" :key="item.id">
                <FileCheck2 :size="17" aria-hidden="true" />
                <span><strong>{{ item.label }}</strong><small>{{ item.capturedReference ?? 'Not captured' }} / {{ item.status }}</small></span>
                <HealthBadge :state="item.state" :trigger="item.required ? 'Required' : 'Optional'" />
              </article>
            </div>
            <div class="evidence-list">
              <article v-for="item in selectedExceptionEvidence" :key="`${item.label}-${item.reference}`">
                <FileCheck2 :size="18" aria-hidden="true" />
                <span><strong>{{ item.label }}</strong><small>{{ item.reference }} / {{ item.status }}</small></span>
                <HealthBadge :state="item.state" />
              </article>
            </div>
          </section>

          <section v-else-if="exceptionTab === 'requery'" class="flow-tab-panel">
            <dl class="metric-grid">
              <div><dt>Automatic attempts</dt><dd>{{ selectedException.automaticAttempts }} / {{ selectedException.maxAutomaticAttempts }}</dd></div>
              <div><dt>Next requery</dt><dd>{{ selectedException.nextRequeryAt }}</dd></div>
              <div><dt>Automation</dt><dd>{{ automationStatusLabels[selectedException.automationStatus] }}</dd></div>
              <div><dt>Pending attempt</dt><dd>{{ nextAutomaticAttempt?.dueAt ?? (automationExhausted ? 'Manual only' : 'No pending attempt') }}</dd></div>
            </dl>
            <div class="attempt-list">
              <article v-for="attempt in selectedRequeryAttempts" :key="attempt.id">
                <span class="attempt-index">#{{ attempt.attemptNumber }}</span>
                <span><strong>{{ attempt.method }}</strong><small>{{ attempt.dueAt }} -> {{ attempt.completedAt }} / {{ attempt.result }}</small></span>
                <HealthBadge :state="attempt.state" :trigger="attempt.trigger" />
              </article>
            </div>
            <small>{{ automaticRequeryAttempts.length }} automatic / {{ manualRequeryAttempts.length }} manual</small>
          </section>

          <section v-else-if="exceptionTab === 'closure'" class="flow-tab-panel">
            <div class="action-stepper">
              <article class="action-step"><span>1</span><strong>Choose action</strong><small>{{ selectedActionStep?.resultingState ?? 'No action available' }}</small></article>
              <article class="action-step"><span>2</span><strong>Attach evidence</strong><small>Reference required</small></article>
              <article class="action-step"><span>3</span><strong>Record reason</strong><small>Operator note required</small></article>
              <article class="action-step"><span>4</span><strong>Review checklist</strong><small>Submit audited action</small></article>
            </div>
            <div class="action-choice-grid">
              <button v-for="step in allowedActionSteps" :key="step.action" type="button" :class="{ 'is-selected': selectedAction === step.action }" @click="selectedAction = step.action">
                <strong>{{ step.label }}</strong>
                <small>{{ step.resultingState }}</small>
              </button>
            </div>
            <label>
              <span>Requery method</span>
              <select v-model="selectedRequeryMethod" aria-label="Requery method" :disabled="selectedAction !== 'manual_requery'">
                <option v-for="method in requeryMethodOptions" :key="method" :value="method">{{ method }}</option>
              </select>
            </label>
            <label>
              <span>Evidence reference</span>
              <input v-model="exceptionEvidence" aria-label="Evidence reference" placeholder="NIP session, ledger posting, callback, or case ID" />
            </label>
            <label>
              <span>Reason and operator note</span>
              <textarea v-model="exceptionReason" aria-label="Resolution reason" rows="4" placeholder="Record what was checked and why the action is safe."></textarea>
            </label>
            <div class="checklist-stack">
              <article v-for="item in selectedActionStep?.safetyChecklist ?? []" :key="item">
                <BadgeCheck :size="15" aria-hidden="true" />
                <span>{{ item }}</span>
              </article>
            </div>
            <dl class="metric-grid maker-checker-grid">
              <div><dt>State</dt><dd>{{ makerCheckerStateLabels[selectedException.makerCheckerState] }}</dd></div>
              <div><dt>Evidence gap</dt><dd>{{ selectedException.evidenceGap }}</dd></div>
              <div><dt>Safe closure</dt><dd>{{ selectedException.safeClosure }}</dd></div>
            </dl>
            <p v-if="!exceptionCanSubmit" class="form-error">Evidence and reason are required before manual completion, requery, reversal approval, or evidence attachment.</p>
            <p v-if="exceptionActionMessage" class="state-note state-note--success">
              <BadgeCheck :size="16" aria-hidden="true" />
              <span>{{ exceptionActionMessage }} Audit event captured.</span>
            </p>
            <p v-if="exceptionActionError" class="form-error">{{ exceptionActionError }}</p>
            <ActionBar>
              <UiButton :disabled="!selectedActionStep || !exceptionCanSubmit || exceptionActionPending" @click="submitSelectedAction">
                <RefreshCw :size="15" aria-hidden="true" />
                Submit action
              </UiButton>
            </ActionBar>
          </section>

          <section v-else class="flow-tab-panel">
            <div class="audit-preview">
              <article v-for="event in selectedCaseAudit" :key="`${event.time}-${event.action}-${event.object}`">
                <span>{{ event.time }}</span>
                <strong>{{ event.action }}</strong>
                <small>{{ event.object }} / {{ event.reason }}</small>
              </article>
              <EmptyState v-if="selectedCaseAudit.length === 0" title="No audit event for this case" description="Submitted actions will appear in this immutable trail." :icon="ShieldAlert" />
            </div>
          </section>
        </template>
        <EmptyState v-else title="No case selected" description="Choose a remediation case to inspect evidence and safe actions." :icon="ShieldAlert" />
      </Panel>
    </section>
  </section>
</template>
