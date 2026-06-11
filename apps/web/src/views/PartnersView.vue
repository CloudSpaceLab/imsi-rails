<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ArrowLeft, ArrowRight, CheckCircle2, PlusCircle, Send } from '@lucide/vue'
import DashboardChart from '../components/DashboardChart.vue'
import ActionBar from '../components/ActionBar.vue'
import HealthBadge from '../components/HealthBadge.vue'
import Panel from '../components/Panel.vue'
import UiButton from '../components/UiButton.vue'
import type { DashboardMock, HealthState, ImtoIntegrationMode, ImtoPartner, ImtoRiskRating, NetworkDirectoryEntry } from '../types'

const props = defineProps<{
  dashboard: DashboardMock
  actorName?: string
}>()

const localPartners = ref<ImtoPartner[]>([])
const partners = computed(() => [...props.dashboard.imtoPartners, ...localPartners.value])
const activeCount = computed(() => partners.value.filter((partner) => partner.status === 'active').length)
const onboardingCount = computed(() => partners.value.filter((partner) => partner.status === 'onboarding' || partner.status === 'draft').length)
const pendingApprovals = computed(() => partners.value.filter((partner) => partner.approvalState === 'checker_pending').length)

const statusLabels: Record<ImtoPartner['status'], string> = {
  active: 'Active',
  onboarding: 'Onboarding',
  suspended: 'Suspended',
  draft: 'Draft',
}
const approvalLabels: Record<ImtoPartner['approvalState'], string> = {
  approved: 'Approved',
  checker_pending: 'Checker approval pending',
  maker_draft: 'Maker draft',
  rejected: 'Rejected',
}

function partnerTone(partner: ImtoPartner): HealthState {
  if (partner.status === 'suspended') return 'blocked'
  if (partner.status === 'onboarding' || partner.status === 'draft') return 'watch'
  return partner.state
}

// Partner detail drilldown
const selectedPartnerId = ref('')
const selectedPartner = computed(() => partners.value.find((partner) => partner.id === selectedPartnerId.value) ?? null)
const selectedSla = computed(() =>
  selectedPartner.value ? props.dashboard.inboundSla.find((row) => row.partner === selectedPartner.value?.name) ?? null : null,
)
const selectedSlaHistory = computed(() =>
  selectedPartner.value
    ? props.dashboard.visuals.partnerSlaHistory.find((series) => series.partner === selectedPartner.value?.name) ?? null
    : null,
)
const selectedTrendLabels = computed(() => selectedSlaHistory.value?.daily.slice(-30).map((point) => point.label.slice(5)) ?? [])
const selectedTrendDatasets = computed(() =>
  selectedSlaHistory.value
    ? [
        {
          label: 'SLA compliance',
          values: selectedSlaHistory.value.daily.slice(-30).map((point) => point.value),
          color: '#0a66ff',
          fill: true,
        },
        {
          label: `Target ${selectedSlaHistory.value.target}%`,
          values: selectedTrendLabels.value.map(() => selectedSlaHistory.value?.target ?? 97),
          color: '#b42318',
        },
      ]
    : [],
)
const dueDiligenceStateMap: Record<'complete' | 'pending' | 'overdue', HealthState> = {
  complete: 'healthy',
  pending: 'watch',
  overdue: 'degraded',
}

// Onboarding wizard
const wizardSteps = [
  { id: 'profile', label: 'Partner profile' },
  { id: 'contract', label: 'Contract and SLA' },
  { id: 'integration', label: 'Integration' },
  { id: 'compliance', label: 'Compliance' },
  { id: 'review', label: 'Review and submit' },
] as const

const wizardOpen = ref(false)
const wizardStep = ref(0)
const wizardError = ref('')
const submittedPartnerId = ref('')

const form = reactive({
  name: '',
  country: '',
  corridor: '',
  riskRating: 'medium' as ImtoRiskRating,
  creditSla: '90s',
  settlementCurrency: '',
  prefundingModel: 'prefunded' as ImtoPartner['prefundingModel'],
  contractEnd: '',
  integrationMode: 'REST' as ImtoIntegrationMode,
  iso20022Ready: false,
  bic: '',
  adapterProfile: '',
  apiCredentialReference: '',
  certificateReference: '',
  licenceVerified: false,
  amlReviewed: false,
  sanctionsCalibrated: false,
})

// Network directory quick-start: BIC or institution name resolves a SWIFT
// directory entry, which preconfigures integration with a standard adapter
// profile so no per-partner integration build is needed.
const directoryQuery = ref('')
const directoryMatch = ref<NetworkDirectoryEntry | null>(null)
const directoryError = ref('')

function lookupDirectory() {
  const query = directoryQuery.value.trim().toLowerCase()
  if (!query) {
    directoryError.value = 'Enter a BIC or institution name to search the network directory.'
    return
  }
  const match = props.dashboard.networkDirectory.find(
    (entry) => entry.bic.toLowerCase() === query || entry.institution.toLowerCase().includes(query),
  )
  if (!match) {
    directoryMatch.value = null
    directoryError.value = 'No directory match. Continue with manual entry or check the BIC.'
    return
  }
  directoryError.value = ''
  directoryMatch.value = match
  form.name = match.institution
  form.country = match.country
  form.corridor = match.corridor
  form.riskRating = match.defaultRiskRating
  form.creditSla = match.defaultCreditSla
  form.settlementCurrency = match.settlementCurrency
  form.integrationMode = 'NETWORK'
  form.iso20022Ready = match.messagingStandard === 'ISO 20022'
  form.bic = match.bic
  form.adapterProfile = match.adapterProfile
}

function clearDirectoryMatch() {
  directoryMatch.value = null
  directoryQuery.value = ''
  directoryError.value = ''
  form.integrationMode = 'REST'
  form.bic = ''
  form.adapterProfile = ''
}

function integrationLabel(partner: Pick<ImtoPartner, 'integrationMode' | 'iso20022Ready' | 'bic'>): string {
  if (partner.integrationMode === 'NETWORK') {
    return `SWIFT network${partner.bic ? ` / ${partner.bic}` : ''}${partner.iso20022Ready ? ' / ISO 20022' : ' / MT103'}`
  }
  return `${partner.integrationMode}${partner.iso20022Ready ? ' / ISO 20022' : ''}`
}

const dueDiligencePreview = computed(() => [
  { item: 'Regulatory licence verification', status: form.licenceVerified ? 'complete' : 'pending' } as const,
  { item: 'AML programme review', status: form.amlReviewed ? 'complete' : 'pending' } as const,
  { item: 'Sanctions screening calibration', status: form.sanctionsCalibrated ? 'complete' : 'pending' } as const,
])

function openWizard() {
  wizardOpen.value = true
  wizardStep.value = 0
  wizardError.value = ''
  submittedPartnerId.value = ''
  clearDirectoryMatch()
}

function validateStep(): string {
  if (wizardStep.value === 0) {
    if (!form.name.trim()) return 'Partner name is required.'
    if (!form.country.trim()) return 'Country is required.'
    if (!form.corridor.trim()) return 'Corridor is required.'
  }
  if (wizardStep.value === 1) {
    if (!form.settlementCurrency.trim()) return 'Settlement currency is required.'
    if (!form.contractEnd) return 'Contract end date is required.'
  }
  if (wizardStep.value === 2 && form.integrationMode !== 'NETWORK') {
    if (!form.apiCredentialReference.trim()) return 'API credential reference is required.'
    if (form.integrationMode !== 'SFTP' && !form.certificateReference.trim()) return 'Certificate reference is required for API integrations.'
  }
  return ''
}

function nextStep() {
  const error = validateStep()
  if (error) {
    wizardError.value = error
    return
  }
  wizardError.value = ''
  if (wizardStep.value < wizardSteps.length - 1) wizardStep.value += 1
}

function previousStep() {
  wizardError.value = ''
  if (wizardStep.value > 0) wizardStep.value -= 1
}

function submitForApproval() {
  const id = `IMTO-${String(100 + partners.value.length + 1)}`
  localPartners.value.push({
    id,
    name: form.name.trim(),
    country: form.country.trim(),
    corridor: form.corridor.trim(),
    status: 'onboarding',
    riskRating: form.riskRating,
    integrationMode: form.integrationMode,
    iso20022Ready: form.iso20022Ready,
    bic: form.bic || undefined,
    onboardingChannel: form.integrationMode === 'NETWORK' ? 'network' : 'manual',
    creditSla: form.creditSla,
    settlementCurrency: form.settlementCurrency.trim(),
    prefundingModel: form.prefundingModel,
    contractEnd: form.contractEnd,
    onboardingStage: 'Awaiting checker approval',
    approvalState: 'checker_pending',
    dueDiligence: dueDiligencePreview.value.map((entry) => ({ ...entry })),
    state: 'watch',
  })
  submittedPartnerId.value = id
  wizardOpen.value = false
}
</script>

<template>
  <section class="screen-stack">
    <section class="metric-grid metric-grid--three" aria-label="Partner summary">
      <article class="metric-card">
        <small>Active IMTOs</small>
        <strong>{{ activeCount }}</strong>
      </article>
      <article class="metric-card">
        <small>In onboarding</small>
        <strong>{{ onboardingCount }}</strong>
      </article>
      <article class="metric-card">
        <small>Pending checker approvals</small>
        <strong>{{ pendingApprovals }}</strong>
      </article>
    </section>

    <aside v-if="submittedPartnerId" class="state-note state-note--watch">
      <CheckCircle2 :size="16" aria-hidden="true" />
      <span>{{ submittedPartnerId }} submitted. Checker approval pending.</span>
    </aside>

    <Panel title="IMTO partners" eyebrow="Directory and onboarding status" accent="healthy">
      <template v-if="!wizardOpen && selectedPartner">
        <button type="button" class="sidebar-link" @click="selectedPartnerId = ''">
          <ArrowLeft :size="13" aria-hidden="true" />
          Back to directory
        </button>
        <div class="partner-detail">
          <header>
            <span>
              <strong>{{ selectedPartner.name }}</strong>
              <small>{{ selectedPartner.corridor }} / {{ selectedPartner.id }}</small>
            </span>
            <HealthBadge :state="partnerTone(selectedPartner)" />
          </header>
          <dl class="wizard-review">
            <div><dt>Status</dt><dd>{{ statusLabels[selectedPartner.status] }} / {{ selectedPartner.onboardingStage }}</dd></div>
            <div><dt>Approval</dt><dd>{{ approvalLabels[selectedPartner.approvalState] }}</dd></div>
            <div><dt>Risk</dt><dd class="partner-risk" :data-risk="selectedPartner.riskRating">{{ selectedPartner.riskRating }}</dd></div>
            <div><dt>Integration</dt><dd>{{ integrationLabel(selectedPartner) }}</dd></div>
            <div><dt>Credit SLA</dt><dd>{{ selectedPartner.creditSla }}</dd></div>
            <div><dt>Settlement</dt><dd>{{ selectedPartner.settlementCurrency }} / {{ selectedPartner.prefundingModel }}</dd></div>
            <div><dt>Contract end</dt><dd>{{ selectedPartner.contractEnd }}</dd></div>
            <div v-if="selectedPartner.onboardingChannel === 'network'"><dt>Routing</dt><dd>Auto-enrolled on corridor template, no custom build</dd></div>
            <div v-if="selectedSla"><dt>Breach rate</dt><dd>{{ selectedSla.breachRate }} / oldest {{ selectedSla.oldestBreach }}</dd></div>
            <div v-if="selectedSla"><dt>Value at risk</dt><dd>{{ selectedSla.valueAtRisk }}</dd></div>
          </dl>
          <section v-if="selectedSlaHistory" class="partner-detail__chart">
            <h4>SLA compliance, last 30 days</h4>
            <DashboardChart
              kind="line"
              :labels="selectedTrendLabels"
              :datasets="selectedTrendDatasets"
              unit="%"
              :height="180"
              :summary="`Daily SLA compliance for ${selectedPartner.name} against the ${selectedSlaHistory.target}% target.`"
            />
          </section>
          <section class="partner-detail__diligence">
            <h4>Due diligence</h4>
            <ul class="sla-standing-list">
              <li v-for="entry in selectedPartner.dueDiligence" :key="entry.item">
                <strong>{{ entry.item }}</strong>
                <small>{{ entry.status }}</small>
                <HealthBadge :state="dueDiligenceStateMap[entry.status]" />
              </li>
            </ul>
          </section>
        </div>
      </template>

      <template v-else-if="!wizardOpen">
        <div class="partner-directory">
          <button v-for="partner in partners" :key="partner.id" type="button" class="partner-row" @click="selectedPartnerId = partner.id">
            <span class="partner-row__identity">
              <strong>{{ partner.name }}</strong>
              <small>{{ partner.corridor }} / {{ partner.settlementCurrency }}</small>
            </span>
            <dl>
              <div><dt>Status</dt><dd>{{ statusLabels[partner.status] }}</dd></div>
              <div><dt>Risk</dt><dd class="partner-risk" :data-risk="partner.riskRating">{{ partner.riskRating }}</dd></div>
              <div><dt>Integration</dt><dd>{{ integrationLabel(partner) }}</dd></div>
              <div><dt>Credit SLA</dt><dd>{{ partner.creditSla }}</dd></div>
              <div><dt>Contract end</dt><dd>{{ partner.contractEnd }}</dd></div>
            </dl>
            <span class="partner-row__approval">
              <small>{{ approvalLabels[partner.approvalState] }}</small>
              <small>{{ partner.onboardingStage }}</small>
            </span>
            <HealthBadge :state="partnerTone(partner)" />
          </button>
        </div>
        <ActionBar>
          <UiButton @click="openWizard">
            <PlusCircle :size="15" aria-hidden="true" />
            Onboard new IMTO
          </UiButton>
        </ActionBar>
      </template>

      <template v-else>
        <ol class="wizard-steps" aria-label="Onboarding steps">
          <li
            v-for="(step, index) in wizardSteps"
            :key="step.id"
            :class="{ 'is-active': index === wizardStep, 'is-done': index < wizardStep }"
          >
            <span>{{ index + 1 }}</span>
            {{ step.label }}
          </li>
        </ol>

        <form class="wizard-form" @submit.prevent="nextStep">
          <template v-if="wizardStep === 0">
            <div class="directory-lookup">
              <label>Network directory quick start
                <span class="directory-lookup__row">
                  <input v-model="directoryQuery" type="text" placeholder="BIC (e.g. WISEGB2LXXX) or institution name" @keydown.enter.prevent="lookupDirectory" />
                  <UiButton variant="secondary" @click="lookupDirectory">Look up</UiButton>
                </span>
              </label>
              <p v-if="directoryError" class="wizard-error" role="alert">{{ directoryError }}</p>
              <aside v-if="directoryMatch" class="state-note state-note--healthy directory-lookup__match">
                <CheckCircle2 :size="16" aria-hidden="true" />
                <span>
                  {{ directoryMatch.bic }} · {{ directoryMatch.institution }} · {{ directoryMatch.messagingStandard }} · {{ directoryMatch.adapterProfile }}
                </span>
                <button type="button" class="sidebar-link" @click="clearDirectoryMatch">Clear</button>
              </aside>
            </div>
            <label>Partner name<input v-model="form.name" type="text" placeholder="e.g. Maple Transfer" /></label>
            <label>Country<input v-model="form.country" type="text" placeholder="e.g. Canada" /></label>
            <label>Corridor<input v-model="form.corridor" type="text" placeholder="e.g. Canada -> Nigeria" /></label>
            <label>Risk classification
              <select v-model="form.riskRating">
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </label>
          </template>

          <template v-else-if="wizardStep === 1">
            <label>Credit SLA
              <select v-model="form.creditSla">
                <option value="30s">30s</option>
                <option value="90s">90s</option>
                <option value="120s">120s</option>
              </select>
            </label>
            <label>Settlement currency<input v-model="form.settlementCurrency" type="text" placeholder="e.g. CAD / NGN" /></label>
            <label>Prefunding model
              <select v-model="form.prefundingModel">
                <option value="prefunded">Prefunded</option>
                <option value="credit-line">Credit line</option>
                <option value="hybrid">Hybrid</option>
              </select>
            </label>
            <label>Contract end date<input v-model="form.contractEnd" type="date" /></label>
          </template>

          <template v-else-if="wizardStep === 2">
            <template v-if="form.integrationMode === 'NETWORK'">
              <dl class="wizard-review">
                <div><dt>Integration</dt><dd>SWIFT network rails</dd></div>
                <div><dt>BIC</dt><dd>{{ form.bic }}</dd></div>
                <div><dt>Messaging</dt><dd>{{ form.iso20022Ready ? 'ISO 20022' : 'MT103' }}</dd></div>
                <div><dt>Adapter profile</dt><dd>{{ form.adapterProfile }}</dd></div>
              </dl>
              <p class="wizard-note">Standard adapter · no credentials required</p>
              <UiButton variant="secondary" @click="clearDirectoryMatch">Switch to manual integration</UiButton>
            </template>
            <template v-else>
              <label>Integration mode
                <select v-model="form.integrationMode">
                  <option value="REST">REST API</option>
                  <option value="SOAP">SOAP</option>
                  <option value="SFTP">SFTP file exchange</option>
                </select>
              </label>
              <label>API credential reference<input v-model="form.apiCredentialReference" type="text" placeholder="Vault reference, not the secret" /></label>
              <label>Certificate reference<input v-model="form.certificateReference" type="text" placeholder="mTLS certificate reference" /></label>
              <label class="wizard-check"><input v-model="form.iso20022Ready" type="checkbox" /> ISO 20022 message support</label>
            </template>
          </template>

          <template v-else-if="wizardStep === 3">
            <label class="wizard-check"><input v-model="form.licenceVerified" type="checkbox" /> Regulatory licence verified</label>
            <label class="wizard-check"><input v-model="form.amlReviewed" type="checkbox" /> AML programme reviewed</label>
            <label class="wizard-check"><input v-model="form.sanctionsCalibrated" type="checkbox" /> Sanctions screening calibrated</label>
            <p class="wizard-note">Unchecked items remain open on the partner record.</p>
          </template>

          <template v-else>
            <dl class="wizard-review">
              <div><dt>Partner</dt><dd>{{ form.name }} ({{ form.country }})</dd></div>
              <div><dt>Corridor</dt><dd>{{ form.corridor }}</dd></div>
              <div><dt>Risk</dt><dd>{{ form.riskRating }}</dd></div>
              <div><dt>Credit SLA</dt><dd>{{ form.creditSla }}</dd></div>
              <div><dt>Settlement</dt><dd>{{ form.settlementCurrency }} / {{ form.prefundingModel }}</dd></div>
              <div><dt>Contract end</dt><dd>{{ form.contractEnd }}</dd></div>
              <div><dt>Integration</dt><dd>{{ integrationLabel(form) }}</dd></div>
              <div v-if="form.integrationMode !== 'NETWORK'"><dt>Credentials</dt><dd>{{ form.apiCredentialReference }}</dd></div>
              <div v-else><dt>Adapter profile</dt><dd>{{ form.adapterProfile }}</dd></div>
              <div v-for="entry in dueDiligencePreview" :key="entry.item"><dt>{{ entry.item }}</dt><dd>{{ entry.status }}</dd></div>
            </dl>
            <p class="wizard-note">Checker approval required before integration testing.</p>
          </template>
        </form>

        <p v-if="wizardError" class="wizard-error" role="alert">{{ wizardError }}</p>

        <ActionBar>
          <UiButton variant="secondary" @click="wizardOpen = false">Cancel</UiButton>
          <UiButton v-if="wizardStep > 0" variant="secondary" @click="previousStep">
            <ArrowLeft :size="15" aria-hidden="true" />
            Back
          </UiButton>
          <UiButton v-if="wizardStep < wizardSteps.length - 1" @click="nextStep">
            Next
            <ArrowRight :size="15" aria-hidden="true" />
          </UiButton>
          <UiButton v-else @click="submitForApproval">
            <Send :size="15" aria-hidden="true" />
            Submit for checker approval
          </UiButton>
        </ActionBar>
      </template>
    </Panel>
  </section>
</template>
