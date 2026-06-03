import { computed, ref, watch, type Ref } from 'vue'
import type {
  BackoffPolicy,
  DashboardMock,
  FinalLegAttempt,
  FinalLegState,
  HealthState,
  IncomingInstruction,
  OutcomeConfidence,
  RemediationCase,
  RoutePenalty,
  TimelineStep,
  UiScenario,
} from '../types'
import { applyCreditAction, type CreditActionKind, type CreditActionPayload } from '../services/inboundApi'

export const stateRank: Record<HealthState, number> = {
  blocked: 0,
  degraded: 1,
  watch: 2,
  stale: 3,
  recovery: 4,
  unknown: 5,
  healthy: 6,
}

export const finalLegStateLabels: Record<FinalLegState, string> = {
  received: 'Received',
  validated: 'Validated',
  routing: 'Routing',
  submitted: 'Submitted',
  accepted: 'Accepted',
  credit_pending: 'Credit pending',
  credited: 'Credited',
  sla_breached: 'SLA breached',
  cooldown: 'Cooldown',
  requerying: 'Requerying',
  outcome_uncertain: 'Outcome uncertain',
  manual_remediation: 'Manual remediation',
  completed_outside_platform: 'Completed outside platform',
  reversal_pending: 'Reversal pending',
  failed_safe: 'Failed safe',
  failed_unsafe: 'Failed unsafe',
  reconciled: 'Reconciled',
}

export const outcomeConfidenceLabels: Record<OutcomeConfidence, string> = {
  proven_credited: 'Proven credited',
  probably_pending: 'Probably pending',
  uncertain: 'Outcome uncertain',
  safe_failed: 'Safe failed',
  unsafe_failed: 'Unsafe failed',
}

export const queueLabels: Record<RemediationCase['queue'], string> = {
  cooldown: 'Cooldown',
  requerying: 'Requerying',
  exhausted: 'Automation exhausted',
  recon_break: 'Recon breaks',
  completed_outside_platform: 'Completed outside platform',
  reversal: 'Reversals',
}

export function useCommandCenter(dashboard: DashboardMock) {
  const mainStandingAccount = computed(() => dashboard.standingAccounts[0] ?? null)
  const openRemediationCases = computed(() => dashboard.remediationCases)
  const providerRankings = computed(() => [...dashboard.providerScores].sort((a, b) => a.rank - b.rank))
  const bestProvider = computed(() => providerRankings.value[0] ?? null)
  const worstProvider = computed(() => providerRankings.value[providerRankings.value.length - 1] ?? null)
  const providerTrafficLanes = computed(() =>
    providerRankings.value.map((provider) => {
      const isLeader = provider.rank === 1
      const isWorst = provider.provider === worstProvider.value?.provider
      const shouldContain = ['blocked', 'degraded'].includes(provider.state) || isWorst
      const stance = isLeader ? 'Route more' : shouldContain ? 'Contain new traffic' : provider.state === 'watch' ? 'Watch before shift' : 'Maintain share'
      const reason = isLeader
        ? `${provider.successRate} success and ${provider.p95} P95 credit proof lead this window.`
        : shouldContain
          ? `${provider.stuckRate} stuck, ${provider.settlementExceptions} exceptions, ${provider.supportSla} support SLA.`
          : `${provider.trafficShare} live share with ${provider.p95} P95 and ${provider.stuckRate} stuck.`

      return {
        provider,
        stance,
        reason,
        lane: isLeader ? 'leader' : shouldContain ? 'risk' : 'steady',
      }
    }),
  )
  const routePenalties = computed(() => [...dashboard.routePenalties].sort((a, b) => b.penaltyScore - a.penaltyScore))
  const externalRoutePenalties = computed(() => routePenalties.value.filter((route) => route.route !== 'Fidelity core'))
  const worstFinalLegRoute = computed(() => externalRoutePenalties.value[0] ?? routePenalties.value[0] ?? null)
  const bestFinalLegRoute = computed(
    () => [...externalRoutePenalties.value].sort((a, b) => a.penaltyScore - b.penaltyScore)[0] ?? routePenalties.value[routePenalties.value.length - 1] ?? null,
  )
  const highestRiskInstructions = computed(() =>
    [...dashboard.incomingInstructions].sort((a, b) => stateRank[a.state] - stateRank[b.state]).slice(0, 4),
  )
  const nextSafestAction = computed(() => openRemediationCases.value[0]?.nextAction ?? dashboard.summary.safeAction)
  const valueAtRiskTotal = computed(() => formatNgnShort(openRemediationCases.value.reduce((sum, item) => sum + parseMoney(item.valueAtRisk), 0)))
  const unknownOutcomeCount = computed(() =>
    dashboard.incomingInstructions.filter((item) => item.outcomeConfidence === 'uncertain' || item.outcomeConfidence === 'unsafe_failed').length,
  )
  const creditedWithinSlaCount = computed(() =>
    dashboard.incomingInstructions.filter((item) => item.currentState === 'credited' || item.currentState === 'reconciled').length,
  )
  const openBreachCount = computed(() =>
    dashboard.incomingInstructions.filter((item) =>
      ['sla_breached', 'cooldown', 'requerying', 'outcome_uncertain', 'manual_remediation'].includes(item.currentState),
    ).length,
  )
  const slaLanes = computed(() => [
    {
      id: 'within',
      label: 'Within SLA',
      state: 'healthy' as HealthState,
      items: dashboard.incomingInstructions.filter((item) =>
        ['received', 'validated', 'routing', 'submitted', 'accepted', 'credit_pending', 'credited', 'reconciled'].includes(item.currentState),
      ),
    },
    {
      id: 'cooldown',
      label: 'Cooldown',
      state: 'watch' as HealthState,
      items: dashboard.incomingInstructions.filter((item) => item.currentState === 'cooldown' || item.currentState === 'sla_breached'),
    },
    {
      id: 'requerying',
      label: 'Requerying',
      state: 'watch' as HealthState,
      items: dashboard.incomingInstructions.filter((item) => item.currentState === 'requerying'),
    },
    {
      id: 'manual',
      label: 'Manual remediation',
      state: 'degraded' as HealthState,
      items: dashboard.incomingInstructions.filter((item) =>
        ['manual_remediation', 'outcome_uncertain', 'failed_unsafe'].includes(item.currentState),
      ),
    },
    {
      id: 'reconciled',
      label: 'Reconciled',
      state: 'healthy' as HealthState,
      items: dashboard.incomingInstructions.filter((item) => item.currentState === 'reconciled'),
    },
  ])

  return {
    mainStandingAccount,
    openRemediationCases,
    providerRankings,
    providerTrafficLanes,
    bestProvider,
    worstProvider,
    routePenalties,
    bestFinalLegRoute,
    worstFinalLegRoute,
    highestRiskInstructions,
    nextSafestAction,
    valueAtRiskTotal,
    unknownOutcomeCount,
    creditedWithinSlaCount,
    openBreachCount,
    slaLanes,
  }
}

export function useInflows(dashboard: DashboardMock, selectedReference: Ref<string>) {
  const inflowQuery = ref('')
  const inflowStateFilter = ref('All states')
  const inflowRouteFilter = ref('All routes')
  const inflowPartnerFilter = ref('All partners')
  const inflowDestinationFilter = ref('All destination banks')
  const inflowValueFilter = ref('All values')
  const inflowSlaAgeFilter = ref('All SLA ages')
  const inflowOutcomeFilter = ref('All outcomes')
  const inflowStateOptions = computed(() => [
    'All states',
    ...Array.from(new Set(dashboard.incomingInstructions.map((item) => finalLegStateLabels[item.currentState]))),
  ])
  const inflowRouteOptions = computed(() => ['All routes', ...Array.from(new Set(dashboard.incomingInstructions.map((item) => item.route)))])
  const inflowPartnerOptions = computed(() => ['All partners', ...Array.from(new Set(dashboard.incomingInstructions.map((item) => item.origin)))])
  const inflowDestinationOptions = computed(() => ['All destination banks', ...Array.from(new Set(dashboard.incomingInstructions.map((item) => item.destinationBank)))])
  const inflowValueOptions = ['All values', 'Under NGN 1M', 'NGN 1M - 3M', 'Over NGN 3M']
  const inflowSlaAgeOptions = ['All SLA ages', 'Inside SLA', 'SLA risk', 'Manual age']
  const inflowOutcomeOptions = computed(() => ['All outcomes', ...Array.from(new Set(dashboard.incomingInstructions.map((item) => outcomeConfidenceLabels[item.outcomeConfidence])))])
  const proofLanes = computed(() => [
    {
      id: 'proved',
      label: 'Proof captured',
      state: 'healthy' as HealthState,
      items: dashboard.incomingInstructions.filter((item) => ['credited', 'reconciled'].includes(item.currentState)),
      empty: 'No proven credits',
    },
    {
      id: 'cooldown',
      label: 'Cooldown watch',
      state: 'watch' as HealthState,
      items: dashboard.incomingInstructions.filter((item) => ['cooldown', 'sla_breached'].includes(item.currentState)),
      empty: 'No cooldown items',
    },
    {
      id: 'requery',
      label: 'Requery active',
      state: 'watch' as HealthState,
      items: dashboard.incomingInstructions.filter((item) => item.currentState === 'requerying'),
      empty: 'No requerying credits',
    },
    {
      id: 'manual-proof',
      label: 'Manual proof',
      state: 'degraded' as HealthState,
      items: dashboard.incomingInstructions.filter((item) =>
        ['manual_remediation', 'outcome_uncertain', 'failed_unsafe'].includes(item.currentState),
      ),
      empty: 'No manual proof queue',
    },
    {
      id: 'reversal',
      label: 'Reversal ready',
      state: 'recovery' as HealthState,
      items: dashboard.incomingInstructions.filter((item) =>
        ['reversal_pending', 'failed_safe', 'completed_outside_platform'].includes(item.currentState),
      ),
      empty: 'No reversals pending',
    },
  ])
  const filteredInflows = computed(() => {
    const query = inflowQuery.value.trim().toLowerCase()
    return dashboard.incomingInstructions.filter((item) => {
      if (inflowStateFilter.value !== 'All states' && finalLegStateLabels[item.currentState] !== inflowStateFilter.value) return false
      if (inflowRouteFilter.value !== 'All routes' && item.route !== inflowRouteFilter.value) return false
      if (inflowPartnerFilter.value !== 'All partners' && item.origin !== inflowPartnerFilter.value) return false
      if (inflowDestinationFilter.value !== 'All destination banks' && item.destinationBank !== inflowDestinationFilter.value) return false
      if (inflowOutcomeFilter.value !== 'All outcomes' && outcomeConfidenceLabels[item.outcomeConfidence] !== inflowOutcomeFilter.value) return false
      if (!matchesValueBand(item.valueAtRisk || item.amount, inflowValueFilter.value)) return false
      if (!matchesSlaAge(item.currentState, inflowSlaAgeFilter.value)) return false
      if (!query) return true
      return [
        item.reference,
        item.partnerReference,
        item.contractId,
        item.beneficiary,
        item.beneficiaryAccount,
        item.destinationBank,
        item.amount,
        item.settlementBatch,
        item.route,
      ].some((value) => value.toLowerCase().includes(query))
    })
  })
  const selectedInflow = computed(
    () =>
      dashboard.incomingInstructions.find((item) => item.reference.toLowerCase() === selectedReference.value.toLowerCase()) ??
      filteredInflows.value[0] ??
      dashboard.incomingInstructions[0] ??
      null,
  )
  const selectedAttempt = computed(() =>
    selectedInflow.value ? dashboard.finalLegAttempts.find((attempt) => attempt.instructionReference === selectedInflow.value?.reference) ?? null : null,
  )
  const selectedEvidence = computed(() =>
    selectedInflow.value ? dashboard.outcomeEvidence.filter((item) => item.instructionReference === selectedInflow.value?.reference) : [],
  )
  const selectedRequeries = computed(() =>
    selectedInflow.value ? dashboard.requeryAttempts.filter((attempt) => attempt.instructionReference === selectedInflow.value?.reference) : [],
  )
  const selectedRemediationCase = computed(() =>
    selectedInflow.value ? dashboard.remediationCases.find((item) => item.instructionReference === selectedInflow.value?.reference) ?? null : null,
  )
  const selectedRoutePenalty = computed(() =>
    selectedAttempt.value ? dashboard.routePenalties.find((penalty) => penalty.route === selectedAttempt.value?.route) ?? null : null,
  )
  const selectedTimeline = computed<TimelineStep[]>(() => buildTimeline(dashboard, selectedInflow.value, selectedAttempt.value))

  return {
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
  }
}

function matchesValueBand(value: string, band: string) {
  if (band === 'All values') return true
  const amount = parseMoney(value)
  if (band === 'Under NGN 1M') return amount < 1_000_000
  if (band === 'NGN 1M - 3M') return amount >= 1_000_000 && amount <= 3_000_000
  return amount > 3_000_000
}

function matchesSlaAge(state: FinalLegState, filter: string) {
  if (filter === 'All SLA ages') return true
  if (filter === 'Inside SLA') return ['received', 'validated', 'routing', 'submitted', 'accepted', 'credit_pending', 'credited', 'reconciled'].includes(state)
  if (filter === 'SLA risk') return ['sla_breached', 'cooldown', 'requerying'].includes(state)
  return ['manual_remediation', 'outcome_uncertain', 'failed_unsafe', 'reversal_pending'].includes(state)
}

export function useRouteHealth(dashboard: DashboardMock, selectedRouteId: Ref<string>) {
  const routePenalties = computed(() => [...dashboard.routePenalties].sort((a, b) => b.penaltyScore - a.penaltyScore))
  const selectedRoutePenalty = computed(
    () => routePenalties.value.find((penalty) => routeSlug(penalty) === selectedRouteId.value) ?? routePenalties.value[0] ?? null,
  )

  return {
    routePenalties,
    selectedRoutePenalty,
  }
}

export function useRemediationQueue(dashboard: DashboardMock, selectedReference: Ref<string>, actorName: Ref<string>) {
  const queueFilter = ref<'all' | RemediationCase['queue']>('all')
  const exceptionEvidence = ref('')
  const exceptionReason = ref('')
  const exceptionActionMessage = ref('')
  const exceptionActionError = ref('')
  const exceptionActionPending = ref(false)
  const selectedRequeryMethod = ref('')
  const queueTabs = computed(() => [
    { id: 'all' as const, label: 'All queues', count: dashboard.remediationCases.length },
    ...Object.entries(queueLabels).map(([id, label]) => ({
      id: id as RemediationCase['queue'],
      label,
      count: dashboard.remediationCases.filter((item) => item.queue === id).length,
    })),
  ])
  const closureReadinessLanes = computed(() => [
    {
      id: 'evidence_blocked',
      label: 'Evidence blocked',
      state: 'degraded' as HealthState,
      items: dashboard.remediationCases.filter(
        (item) => item.makerCheckerState === 'maker_required' || (item.duplicateRisk !== 'Low' && item.queue !== 'requerying'),
      ),
    },
    {
      id: 'requery_active',
      label: 'Requery active',
      state: 'watch' as HealthState,
      items: dashboard.remediationCases.filter((item) => item.queue === 'requerying' || item.automationStatus === 'running'),
    },
    {
      id: 'checker_review',
      label: 'Checker review',
      state: 'recovery' as HealthState,
      items: dashboard.remediationCases.filter(
        (item) => item.makerCheckerState === 'checker_pending' || item.queue === 'completed_outside_platform',
      ),
    },
    {
      id: 'reversal_ready',
      label: 'Reversal ready',
      state: 'watch' as HealthState,
      items: dashboard.remediationCases.filter((item) => item.queue === 'reversal'),
    },
  ])
  const filteredCases = computed(() => {
    const cases =
      queueFilter.value === 'all' ? dashboard.remediationCases : dashboard.remediationCases.filter((item) => item.queue === queueFilter.value)
    return [...cases].sort((a, b) => stateRank[a.state] - stateRank[b.state])
  })
  const selectedException = computed(
    () =>
      dashboard.remediationCases.find((item) => item.instructionReference.toLowerCase() === selectedReference.value.toLowerCase()) ??
      filteredCases.value[0] ??
      null,
  )
  const selectedExceptionEvidence = computed(() =>
    selectedException.value ? dashboard.outcomeEvidence.filter((item) => item.instructionReference === selectedException.value?.instructionReference) : [],
  )
  const selectedExceptionInstruction = computed(() =>
    selectedException.value
      ? dashboard.incomingInstructions.find((item) => item.reference === selectedException.value?.instructionReference) ?? null
      : null,
  )
  const selectedExceptionAttempt = computed(() =>
    selectedException.value
      ? dashboard.finalLegAttempts.find((attempt) => attempt.instructionReference === selectedException.value?.instructionReference) ?? null
      : null,
  )
  const selectedExceptionRoutePenalty = computed(() =>
    selectedException.value ? dashboard.routePenalties.find((route) => route.route === selectedException.value?.route) ?? null : null,
  )
  const selectedExceptionBackoff = computed(() => backoffForAttempt(dashboard, selectedExceptionAttempt.value))
  const selectedRequeryAttempts = computed(() =>
    selectedException.value ? dashboard.requeryAttempts.filter((attempt) => attempt.instructionReference === selectedException.value?.instructionReference) : [],
  )
  const automaticRequeryAttempts = computed(() => selectedRequeryAttempts.value.filter((attempt) => attempt.trigger === 'automatic'))
  const manualRequeryAttempts = computed(() => selectedRequeryAttempts.value.filter((attempt) => attempt.trigger === 'manual'))
  const nextAutomaticAttempt = computed(() => selectedRequeryAttempts.value.find((attempt) => attempt.completedAt === '-') ?? null)
  const automationExhausted = computed(() => {
    const item = selectedException.value
    if (!item) return false
    return item.automationStatus === 'exhausted' || item.automaticAttempts >= item.maxAutomaticAttempts
  })
  const exceptionCanSubmit = computed(() => Boolean(selectedException.value && exceptionEvidence.value.trim() && exceptionReason.value.trim()))

  watch(
    selectedException,
    () => {
      exceptionEvidence.value = ''
      exceptionReason.value = ''
      exceptionActionMessage.value = ''
      exceptionActionError.value = ''
      selectedRequeryMethod.value = selectedException.value?.requeryMethod ?? ''
    },
    { immediate: true },
  )

  async function submitExceptionAction(action: 'attach_evidence' | 'manual_complete' | 'approve_reversal' | 'requery') {
    if (!selectedException.value || !exceptionCanSubmit.value || exceptionActionPending.value) return
    const currentCase = selectedException.value
    const evidenceReference = exceptionEvidence.value.trim()
    const reason = exceptionReason.value.trim()
    const method = selectedRequeryMethod.value || currentCase.requeryMethod
    const apiReference = currentCase.apiReference ?? currentCase.instructionReference
    const apiAction = toCreditAction(action)

    exceptionActionPending.value = true
    exceptionActionError.value = ''
    exceptionActionMessage.value = ''

    try {
      await applyCreditAction(apiReference, toCreditActionPayload(apiAction, reason, evidenceReference, method))
    } catch (error) {
      exceptionActionError.value = error instanceof Error ? error.message : 'Unable to submit exception action'
      return
    } finally {
      exceptionActionPending.value = false
    }

    if (action === 'attach_evidence') {
      currentCase.makerChecker = 'Evidence attached'
      currentCase.makerCheckerState = currentCase.duplicateRisk === 'Low' ? 'not_required' : 'maker_required'
      currentCase.nextAction = 'Review captured evidence before selecting any closure action.'
      currentCase.state = currentCase.duplicateRisk === 'Low' ? 'watch' : 'degraded'
    } else if (action === 'requery') {
      currentCase.queue = 'requerying'
      currentCase.automationStatus = 'running'
      currentCase.makerChecker = 'Manual requery submitted'
      currentCase.makerCheckerState = 'not_required'
      currentCase.nextAction = 'Wait for manual requery response before any closure action.'
      currentCase.state = 'watch'
      currentCase.nextRequeryAt = 'Manual requery queued'
      dashboard.requeryAttempts.unshift({
        id: `RQ-${currentCase.instructionReference}-MAN-${manualRequeryAttempts.value.length + 1}`,
        instructionReference: currentCase.instructionReference,
        attemptNumber: selectedRequeryAttempts.value.length + 1,
        dueAt: 'Now',
        completedAt: 'Queued',
        method,
        result: `Manual requery queued with ${evidenceReference}`,
        trigger: 'manual',
        state: 'recovery',
      })
    } else if (action === 'manual_complete') {
      currentCase.queue = 'completed_outside_platform'
      currentCase.automationStatus = 'manual_only'
      currentCase.makerChecker = 'Checker review pending'
      currentCase.makerCheckerState = 'checker_pending'
      currentCase.nextAction = 'Checker must approve completed-outside-platform closure.'
      currentCase.state = 'recovery'
    } else {
      currentCase.queue = 'reversal'
      currentCase.automationStatus = 'manual_only'
      currentCase.makerChecker = 'Reversal approved'
      currentCase.makerCheckerState = 'approved'
      currentCase.nextAction = 'Post reversal to partner and close after settlement evidence lands.'
      currentCase.state = 'recovery'
    }

    dashboard.outcomeEvidence.unshift({
      instructionReference: currentCase.instructionReference,
      type: 'operator_note',
      label:
        action === 'attach_evidence'
          ? 'Repair evidence attached'
          : action === 'requery'
            ? 'Manual requery evidence'
            : action === 'manual_complete'
              ? 'Manual completion evidence'
              : 'Reversal approval evidence',
      reference: evidenceReference,
      status: reason,
      owner: actorName.value,
      state: 'recovery',
    })
    dashboard.auditEvents.unshift({
      time: 'Now',
      actor: actorName.value,
      action:
        action === 'attach_evidence'
          ? 'Evidence attached'
          : action === 'requery'
          ? 'Manual requery queued'
          : action === 'manual_complete'
            ? 'Completed outside platform submitted'
            : 'Reversal approved',
      object: currentCase.instructionReference,
      reason: `${reason} / ${evidenceReference}`,
      state: 'recovery',
    })
    exceptionActionMessage.value =
      action === 'attach_evidence'
        ? 'Evidence attached to the repair case.'
        : action === 'approve_reversal'
        ? 'Reversal approval captured with evidence.'
        : action === 'requery'
          ? 'Manual requery queued with evidence.'
          : 'Manual completion submitted for checker review.'
  }

  return {
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
  }
}

function toCreditAction(action: 'attach_evidence' | 'manual_complete' | 'approve_reversal' | 'requery'): CreditActionKind {
  if (action === 'attach_evidence') return 'attach_evidence'
  if (action === 'requery') return 'manual_requery'
  if (action === 'manual_complete') return 'mark_completed_outside_platform'
  return 'approve_reversal'
}

function toCreditActionPayload(
  action: CreditActionKind,
  note: string,
  evidenceReference: string,
  requeryMethod: string,
): CreditActionPayload {
  return {
    action,
    note,
    evidenceReference,
    reasonCode: creditActionReasonCode(action),
    requeryMethod: action === 'manual_requery' ? requeryMethod : undefined,
    customerValueDelivered: action === 'mark_completed_outside_platform',
  }
}

function creditActionReasonCode(action: CreditActionKind) {
  if (action === 'attach_evidence') return 'evidence_attached'
  if (action === 'manual_requery') return 'manual_requery_requested'
  if (action === 'mark_completed_outside_platform') return 'credited_outside_platform'
  if (action === 'approve_reversal') return 'reversal_approved'
  return 'recon_resolved'
}

export function normalizeDataState(value: unknown): UiScenario {
  if (value === 'healthy') return 'healthy'
  if (value === 'traffic-shift') return 'traffic-shift'
  if (value === 'pilot-report') return 'pilot-report'
  if (value === 'blocked') return 'blocked'
  if (value === 'stale-fx') return 'stale-fx'
  if (value === 'empty') return 'empty'
  if (value === 'permission-denied') return 'permission-denied'
  if (value === 'loading') return 'loading'
  if (value === 'api-failure') return 'api-failure'
  return 'degraded-ria'
}

export function parseMoney(value: string) {
  return Number(value.replace(/[^\d.]/g, '')) || 0
}

export function formatNgnShort(value: number) {
  if (value >= 1_000_000) return `NGN ${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1_000) return `NGN ${(value / 1_000).toFixed(0)}K`
  return `NGN ${value.toLocaleString()}`
}

export function confidenceState(confidence: OutcomeConfidence): HealthState {
  if (confidence === 'proven_credited') return 'healthy'
  if (confidence === 'safe_failed') return 'recovery'
  if (confidence === 'probably_pending') return 'watch'
  return 'degraded'
}

export function routePenaltyWidth(score: number) {
  return `${Math.min(100, Math.max(4, score))}%`
}

export function routeSlug(route: RoutePenalty) {
  return route.route.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '')
}

export function standingMeterWidth(account: { state: HealthState } | null | undefined) {
  if (!account) return '0%'
  return account.state === 'healthy' ? '76%' : account.state === 'watch' ? '48%' : '28%'
}

export function backoffForAttempt(dashboard: DashboardMock, attempt: FinalLegAttempt | null): BackoffPolicy | null {
  if (!attempt) return null
  return dashboard.backoffPolicies.find((policy) => attempt.route.toLowerCase().includes(policy.route.split(' ')[0].toLowerCase())) ?? dashboard.backoffPolicies[0] ?? null
}

function buildTimeline(dashboard: DashboardMock, inflow: IncomingInstruction | null, attempt: FinalLegAttempt | null): TimelineStep[] {
  if (!inflow) return []
  const contract = dashboard.inboundContracts.find((item) => item.id === inflow.contractId) ?? null
  const account = dashboard.standingAccounts.find((item) => item.id === contract?.standingAccountId) ?? null

  return [
    {
      label: 'Instruction received',
      owner: contract?.partner ?? inflow.origin,
      status: 'done',
      time: inflow.receivedAt,
      source: 'Partner API',
      reference: inflow.partnerReference,
    },
    {
      label: 'Contract and balance validated',
      owner: 'Fidelity standing account',
      status: 'done',
      time: inflow.receivedAt,
      source: account?.id ?? inflow.contractId,
    },
    {
      label: 'Final-leg route selected',
      owner: 'Route engine',
      status: 'done',
      time: attempt?.submittedAt ?? inflow.receivedAt,
      source: attempt?.decisionReason,
      reference: attempt?.route,
    },
    {
      label: 'Rail accepted',
      owner: attempt?.rail ?? inflow.route,
      status: attempt?.acceptedAt && attempt.acceptedAt !== '-' ? 'done' : 'current',
      time: attempt?.acceptedAt ?? 'pending',
      reference: attempt?.id,
    },
    {
      label: inflow.currentState === 'credited' ? 'Beneficiary credited' : 'Outcome proof',
      owner: inflow.owner,
      status: inflow.currentState === 'credited' || inflow.currentState === 'reconciled' ? 'done' : 'current',
      time: attempt?.creditedAt ?? 'pending',
      note: inflow.safeAction,
    },
    {
      label: 'Settlement reconciled',
      owner: 'Settlement Ops',
      status: inflow.currentState === 'reconciled' ? 'done' : 'pending',
      time: inflow.currentState === 'reconciled' ? 'matched' : 'pending',
      source: inflow.settlementBatch,
    },
  ]
}
