<script setup lang="ts">
import { computed, ref } from 'vue'
import { Download, FileText } from '@lucide/vue'
import Panel from '../components/Panel.vue'
import UiButton from '../components/UiButton.vue'
import type { DashboardMock } from '../types'

const props = defineProps<{
  dashboard: DashboardMock
}>()

type ReportCategory = 'Operational' | 'Regulatory' | 'Executive'
type ReportFormat = 'csv' | 'json'

type ReportDefinition = {
  id: string
  name: string
  category: ReportCategory
  description: string
  build: () => Array<Record<string, string | number | boolean>>
}

type GeneratedReport = {
  id: string
  name: string
  range: string
  format: ReportFormat
  generatedAt: string
  rows: number
  url: string
  filename: string
}

const rangeOptions = ['Today', 'Last 7 days', 'Last 30 days', 'Last 90 days'] as const
const selectedRange = ref<(typeof rangeOptions)[number]>('Last 30 days')
const selectedFormat = ref<ReportFormat>('csv')
const generatedReports = ref<GeneratedReport[]>([])

const reportCatalog = computed<ReportDefinition[]>(() => [
  {
    id: 'transactions',
    name: 'Transaction report',
    category: 'Operational',
    description: 'All inbound remittance instructions with lifecycle status.',
    build: () =>
      props.dashboard.incomingInstructions.map((item) => ({
        reference: item.reference,
        partnerReference: item.partnerReference,
        origin: item.origin,
        beneficiary: item.beneficiary,
        destinationBank: item.destinationBank,
        amount: item.amount,
        status: item.currentState,
        receivedAt: item.receivedAt,
      })),
  },
  {
    id: 'failed-transactions',
    name: 'Failed transaction report',
    category: 'Operational',
    description: 'Failed, reversed, and unresolved instructions with repair-case linkage.',
    build: () =>
      props.dashboard.remediationCases.map((item) => ({
        caseId: item.id,
        instruction: item.instructionReference,
        queue: item.queue,
        route: item.route,
        valueAtRisk: item.valueAtRisk,
        age: item.age,
        owner: item.staffAssignment?.staffName ?? item.owner,
        nextAction: item.nextAction,
      })),
  },
  {
    id: 'settlement',
    name: 'Settlement report',
    category: 'Operational',
    description: 'Partner settlement positions, prefunding, and credit facilities.',
    build: () =>
      props.dashboard.standingAccounts.map((account) => ({
        account: account.id,
        partner: account.partner,
        currency: account.currency,
        prefundBalance: account.prefundBalance,
        availableLimit: account.availableLimit,
        creditLimit: account.creditFacility?.limit ?? '-',
        creditUtilized: account.creditFacility?.utilized ?? '-',
        projectedExhaustion: account.projectedExhaustion,
      })),
  },
  {
    id: 'reconciliation',
    name: 'Reconciliation report',
    category: 'Operational',
    description: 'Ledger, transit account, and provider statement matches.',
    build: () =>
      props.dashboard.reconciliationMatches.map((match) => ({
        ...match,
      })),
  },
  {
    id: 'compliance',
    name: 'Compliance and audit report',
    category: 'Regulatory',
    description: 'Compliance holds, decisions pending, and audit evidence status.',
    build: () =>
      props.dashboard.complianceHolds.map((hold) => ({
        reference: hold.reference,
        partner: hold.partner,
        type: hold.type,
        beneficiary: hold.beneficiary,
        amount: hold.amount,
        age: hold.age,
        decisionSla: hold.decisionSla,
        owner: hold.owner,
        state: hold.state,
      })),
  },
  {
    id: 'partner-performance',
    name: 'Partner performance report',
    category: 'Executive',
    description: 'Per-IMTO SLA compliance, breach rates, and value at risk.',
    build: () =>
      props.dashboard.inboundSla.map((row) => ({
        contract: row.contractId,
        partner: row.partner,
        corridor: row.corridor,
        creditSla: row.creditSla,
        p95Credit: row.p95Credit,
        breachRate: row.breachRate,
        agingBreaches: row.agingBreaches,
        valueAtRisk: row.valueAtRisk,
        trend: row.trend,
      })),
  },
  {
    id: 'business-trends',
    name: 'Business trends report',
    category: 'Executive',
    description: 'Volume, value, and SLA completion trend for the selected range.',
    build: () =>
      props.dashboard.visuals.partnerSlaHistory.flatMap((series) =>
        series.daily.map((point) => ({
          partner: series.partner,
          date: point.label,
          slaCompliance: point.value,
          target: series.target,
        })),
      ),
  },
])

const categories: ReportCategory[] = ['Operational', 'Regulatory', 'Executive']
const catalogByCategory = computed(() =>
  categories.map((category) => ({
    category,
    reports: reportCatalog.value.filter((report) => report.category === category),
  })),
)

function toCsv(rows: Array<Record<string, string | number | boolean>>) {
  if (!rows.length) return ''
  const headers = Object.keys(rows[0])
  const escape = (value: string | number | boolean) => {
    const text = String(value)
    return /[",\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
  }
  return [headers.join(','), ...rows.map((row) => headers.map((key) => escape(row[key] ?? '')).join(','))].join('\n')
}

function generateReport(definition: ReportDefinition) {
  const rows = definition.build()
  const format = selectedFormat.value
  const payload = format === 'csv' ? toCsv(rows) : JSON.stringify(rows, null, 2)
  const blob = new Blob([payload], { type: format === 'csv' ? 'text/csv' : 'application/json' })
  const filename = `${definition.id}-${selectedRange.value.replace(/\s+/g, '-').toLowerCase()}.${format}`
  generatedReports.value.unshift({
    id: `${definition.id}-${generatedReports.value.length + 1}`,
    name: definition.name,
    range: selectedRange.value,
    format,
    generatedAt: new Date().toLocaleTimeString(),
    rows: rows.length,
    url: URL.createObjectURL(blob),
    filename,
  })
}
</script>

<template>
  <section class="screen-stack">
    <Panel title="Report parameters" eyebrow="Applies to every generated report" accent="healthy">
      <div class="report-controls">
        <label>Time range
          <select v-model="selectedRange">
            <option v-for="option in rangeOptions" :key="option" :value="option">{{ option }}</option>
          </select>
        </label>
        <label>Format
          <select v-model="selectedFormat">
            <option value="csv">CSV</option>
            <option value="json">JSON</option>
          </select>
        </label>
      </div>
    </Panel>

    <Panel v-for="group in catalogByCategory" :key="group.category" :title="`${group.category} reports`" accent="healthy">
      <div class="report-catalog">
        <article v-for="report in group.reports" :key="report.id" class="report-row">
          <FileText :size="17" aria-hidden="true" />
          <span>
            <strong>{{ report.name }}</strong>
            <small>{{ report.description }}</small>
          </span>
          <UiButton size="sm" @click="generateReport(report)">Generate</UiButton>
        </article>
      </div>
    </Panel>

    <Panel title="Generated reports" eyebrow="This session" accent="healthy">
      <div v-if="generatedReports.length" class="report-catalog">
        <article v-for="report in generatedReports" :key="report.id" class="report-row">
          <Download :size="17" aria-hidden="true" />
          <span>
            <strong>{{ report.name }}</strong>
            <small>{{ report.range }} / {{ report.rows }} rows / {{ report.format.toUpperCase() }} / {{ report.generatedAt }}</small>
          </span>
          <a class="report-download" :href="report.url" :download="report.filename">Download</a>
        </article>
      </div>
      <article v-else class="quiet-empty">
        <FileText :size="18" aria-hidden="true" />
        <span>
          <strong>No reports generated</strong>
          <small>Select a report above to generate.</small>
        </span>
      </article>
    </Panel>
  </section>
</template>
