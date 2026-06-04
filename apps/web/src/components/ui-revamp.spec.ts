import { describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import App from '../App.vue'
import CountryPair from './CountryPair.vue'
import DataTable from './DataTable.vue'
import EmptyState from './EmptyState.vue'
import HealthBadge from './HealthBadge.vue'
import Panel from './Panel.vue'
import ProviderMark from './ProviderMark.vue'
import UiButton from './UiButton.vue'
import { getDashboardMock } from '../services/mockDashboard'
import { router, screenRoutes } from '../router'

async function mountApp(path = '/') {
  await router.push(path)
  await router.isReady()
  const wrapper = mount(App, { global: { plugins: [router] } })
  await flushPromises()
  return wrapper
}

describe('premium dashboard primitives', () => {
  it('renders health metadata accessibly', () => {
    const wrapper = mount(HealthBadge, {
      props: {
        state: 'degraded',
        trigger: 'P95 target breached',
        window: '15 min',
        updated: '14:32 UTC',
      },
    })

    expect(wrapper.attributes('aria-label')).toContain('Degraded')
    expect(wrapper.attributes('aria-label')).toContain('P95 target breached')
    expect(wrapper.text()).toContain('Degraded')
  })

  it('prevents disabled button activation', () => {
    const wrapper = mount(UiButton, {
      props: { disabled: true },
      slots: { default: 'Save draft' },
    })

    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
  })

  it('renders shared panel, empty, provider, country, and table primitives', () => {
    const panel = mount(Panel, {
      props: { title: 'Provider comparison', eyebrow: 'Rates', accent: 'healthy' },
      slots: { default: '<p>Ready</p>' },
    })
    expect(panel.text()).toContain('Provider comparison')

    const empty = mount(EmptyState, {
      props: { title: 'No settlement breaks', description: 'Everything matched.' },
    })
    expect(empty.text()).toContain('No settlement breaks')

    const provider = mount(ProviderMark, { props: { provider: 'Thunes', showCategory: true } })
    expect(provider.text()).toContain('Thunes')
    expect(provider.text()).toContain('B2B payout network')

    const country = mount(CountryPair, { props: { origin: 'Germany', destination: 'Nigeria' } })
    expect(country.text()).toContain('Germany')
    expect(country.text()).toContain('Nigeria')
    expect(country.find('.country-flag--de').exists()).toBe(true)
    expect(country.find('.country-flag--ng').exists()).toBe(true)

    const table = mount(DataTable, {
      props: { empty: true, emptyTitle: 'No rows' },
      slots: { default: '<table><tbody></tbody></table>' },
    })
    expect(table.text()).toContain('No rows')
  })
})

describe('inbound settlement tower workflows', () => {
  it('keeps fixtures aligned to the inbound settlement model', () => {
    const dashboard = getDashboardMock()
    const currentStates = dashboard.incomingInstructions.map((item) => item.currentState)
    const exhaustedCase = dashboard.remediationCases.find((item) => item.instructionReference === 'INF-10002')

    expect(currentStates).toEqual(expect.arrayContaining(['cooldown', 'requerying', 'manual_remediation']))
    expect(dashboard.requeryAttempts.length).toBeGreaterThan(0)
    expect(dashboard.requeryAttempts.every((attempt) => attempt.trigger === 'automatic')).toBe(true)
    expect(dashboard.remediationCases.every((item) => item.requeryMethod && item.automaticAttempts <= item.maxAutomaticAttempts)).toBe(true)
    expect(dashboard.incomingInstructions.find((item) => item.currentState === 'cooldown')?.safeAction).toContain('Freeze duplicate action')
    expect(dashboard.requeryAttempts.find((attempt) => attempt.instructionReference === 'INF-10003' && attempt.completedAt === '-')?.result).toBe('Due now')
    expect(dashboard.routeDecisions.find((item) => item.instructionReference === 'INF-10002')?.rejectedRoutes.length).toBeGreaterThan(0)
    expect(dashboard.evidenceRequirements.find((item) => item.instructionReference === 'INF-10002')?.sourceTable).toBe('outcome_evidence')
    expect(dashboard.reconciliationMatches.find((item) => item.instructionReference === 'INF-10002')?.sourceTable).toBe('reconciliation_matches')
    expect(dashboard.routeHealthWindows.find((item) => item.route === 'NIP final leg')?.sourceTable).toBe('route_health_windows')
    expect(dashboard.integrationHealth.map((item) => item.serviceName)).toContain('Switching API')
    expect(dashboard.caseActionSteps.map((item) => item.action)).toEqual(
      expect.arrayContaining(['attach_evidence', 'manual_requery', 'mark_completed_outside_platform', 'approve_reversal']),
    )
    expect(exhaustedCase?.automationStatus).toBe('exhausted')
    expect(exhaustedCase?.apiReference).toBe('CRD-90288')
    expect(dashboard.requeryAttempts.filter((attempt) => attempt.instructionReference === exhaustedCase?.instructionReference)).toHaveLength(
      exhaustedCase?.maxAutomaticAttempts ?? 0,
    )
    expect(dashboard.featureSwitches.every((feature) => feature.enabled === feature.defaultEnabled)).toBe(true)
    expect(getDashboardMock('empty').incomingInstructions).toHaveLength(0)
    expect(getDashboardMock('empty').integrationHealth).toHaveLength(0)
    expect(getDashboardMock('api-failure').viewState).toBe('error')
  })

  it('exposes only five top-level screens', async () => {
    const wrapper = await mountApp()
    const labels = screenRoutes.map((screen) => screen.label)

    expect(labels).toEqual(['Command Center', 'Inflows', 'Routes', 'Exceptions', 'Settings'])
    for (const label of labels) {
      const button = wrapper.findAll('button.nav-item').find((item) => item.text().includes(label))
      expect(button, `missing nav item ${label}`).toBeTruthy()
      await button?.trigger('click')
      await flushPromises()
      expect(wrapper.text()).toContain(label)
    }

    const primaryNav = wrapper.find('.primary-nav').text()
    expect(primaryNav).not.toContain('Rates & costs')
    expect(primaryNav).not.toContain('Providers')
    expect(primaryNav).not.toContain('Audit')
    expect(primaryNav).not.toContain('Incoming credits')
    expect(primaryNav).not.toContain('Inbound SLA')
  })

  it('answers live inbound risk and next action on the first screen', async () => {
    const wrapper = await mountApp('/')

    expect(wrapper.text()).toContain('Primary dashboard')
    expect(wrapper.text()).toContain('Failed/reversed local transfers')
    expect(wrapper.text()).toContain('Switching API status')
    expect(wrapper.text()).toContain('Local providers working')
    expect(wrapper.text()).toContain('International partner SLAs')
    expect(wrapper.text()).toContain('SLA completion trend')
    expect(wrapper.text()).toContain('Exception mix')
    expect(wrapper.text()).toContain('8-hour window')
    expect(wrapper.text()).toContain('Switching API telemetry')
    expect(wrapper.text()).toContain('Transactions and middleware calls to fix')
    expect(wrapper.text()).toContain('Local provider performance')
    expect(wrapper.text()).toContain('International banking partner SLAs')
    expect(wrapper.text()).toContain('Moniepoint leads NIP, Interswitch, and Paystack')
    expect(wrapper.text()).toContain('Paystack')
    expect(wrapper.text()).toContain('Interswitch final leg')
    expect(wrapper.text()).toContain('Hellenic Remit (Greece)')
    expect(wrapper.text()).toContain('Work exceptions')
    expect(wrapper.text()).toContain('Do not reroute')
  })

  it('launches focused queues from command center cards', async () => {
    let wrapper = await mountApp('/')
    await wrapper.findAll('.ops-signal').find((button) => button.text().includes('Failed/reversed local transfers'))?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/exceptions')
    expect(router.currentRoute.value.query).toMatchObject({ queue: 'all', focus: 'failed' })

    wrapper.unmount()
    wrapper = await mountApp('/')
    await wrapper.findAll('.ops-signal').find((button) => button.text().includes('Local providers working'))?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/routes')
    expect(router.currentRoute.value.query.focus).toBe('provider-performance')

    wrapper.unmount()
    wrapper = await mountApp('/')
    await wrapper.findAll('.ops-signal').find((button) => button.text().includes('International partner SLAs'))?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/settings')
    expect(router.currentRoute.value.query).toMatchObject({ tab: 'sla', focus: 'breaches' })

    wrapper.unmount()
    wrapper = await mountApp('/')
    await wrapper.findAll('.ops-signal').find((button) => button.text().includes('Switching API status'))?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/settings')
    expect(router.currentRoute.value.query).toMatchObject({ tab: 'integrations', focus: 'api-health' })
  })

  it('keeps inflows searchable and traceable by final-leg evidence', async () => {
    const wrapper = await mountApp('/inflows')

    expect(wrapper.text()).toContain('Inflow search')
    expect(wrapper.text()).toContain('Credit proof lanes')
    expect(wrapper.text()).toContain('Manual proof')
    expect(wrapper.text()).toContain('Incoming instructions')
    expect(wrapper.text()).toContain('All routes')
    expect(wrapper.text()).toContain('All SLA ages')
    expect(wrapper.text()).toContain('All outcomes')
    expect(wrapper.text()).toContain('INF-10001')
    expect(wrapper.text()).toContain('INF-10002')

    await wrapper.get('input[aria-label="Search inflows"]').setValue('BX-ENG-77118')
    await flushPromises()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(1)
    expect(rows[0].text()).toContain('INF-10002')
    expect(rows[0].text()).not.toContain('INF-10001')
  })

  it('opens critical table drilldowns from keyboard-operable rows', async () => {
    let wrapper = await mountApp('/inflows')
    const inflowRow = wrapper.get('tr[role="button"][aria-label="Open inflow INF-10002"]')

    expect(inflowRow.attributes('tabindex')).toBe('0')
    await inflowRow.trigger('keydown', { key: 'Enter' })
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/inflows/INF-10002')
    expect(wrapper.text()).toContain('Tunde Bakare')

    wrapper.unmount()
    wrapper = await mountApp('/routes')
    const routeRow = wrapper.get('tr[role="button"][aria-label="Open route Moniepoint final leg"]')

    expect(routeRow.attributes('tabindex')).toBe('0')
    await routeRow.trigger('keydown', { key: 'Enter' })
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/routes/moniepoint-final-leg')
    expect(wrapper.text()).toContain('Fastest P95')
  })

  it('opens an inflow trace with backoff and evidence', async () => {
    const wrapper = await mountApp('/inflows/INF-10002')

    expect(wrapper.text()).toContain('Inflow detail')
    expect(wrapper.text()).toContain('Summary')
    expect(wrapper.text()).toContain('Route decision')
    expect(wrapper.text()).toContain('Timeline')
    expect(wrapper.text()).toContain('Evidence')
    expect(wrapper.text()).toContain('Requery')
    expect(wrapper.text()).toContain('Reconciliation')
    expect(wrapper.text()).toContain('Audit')
    expect(wrapper.text()).toContain('Settlement batch')
    expect(wrapper.text()).toContain('Safe action')
    expect(wrapper.text()).toContain('Do not reroute')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Route decision'))?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.query.tab).toBe('route')
    expect(wrapper.text()).toContain('Final-leg route pressure')
    expect(wrapper.text()).toContain('NIP final leg')
    expect(wrapper.text()).toContain('Interswitch final leg')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Timeline'))?.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Final-leg route selected')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Evidence'))?.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('NIP-TRF-88421')
    expect(wrapper.text()).toContain('Core ledger posting')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Requery'))?.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Backoff ladder')
    expect(wrapper.text()).toContain('T+5m')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Reconciliation'))?.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Provider file says paid')
  })

  it('requires evidence before manual exception actions', async () => {
    const wrapper = await mountApp('/exceptions/INF-10002?tab=closure')

    expect(wrapper.text()).toContain('Case detail')
    expect(wrapper.text()).toContain('Attach evidence')
    expect(wrapper.text()).not.toContain('Mark completed outside platform')
    expect(wrapper.text()).toContain('Evidence and reason are required before')
    const submitButton = () => wrapper.findAll('button').find((button) => button.text().includes('Submit action'))
    expect(submitButton()?.attributes('disabled')).toBeDefined()

    await wrapper.get('input[aria-label="Evidence reference"]').setValue('NIP-TRF-88421')
    await wrapper.get('textarea[aria-label="Resolution reason"]').setValue('Verified NIP session against suspense ledger before any closure action.')
    await flushPromises()

    expect(submitButton()?.attributes('disabled')).toBeUndefined()
    await submitButton()?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Evidence attached to the repair case')
    expect(wrapper.text()).toContain('Audit event captured')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Audit'))?.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Evidence attached')
    expect(wrapper.text()).toContain('NIP-TRF-88421')
  })

  it('surfaces exhausted backoff and maker-checker guardrails', async () => {
    const wrapper = await mountApp('/exceptions?queue=exhausted')

    expect(router.currentRoute.value.query.queue).toBe('exhausted')
    expect(wrapper.text()).toContain('Automation exhausted')
    expect(wrapper.text()).toContain('INF-10002')

    await wrapper.findAll('.remediation-queue button').find((button) => button.text().includes('INF-10002'))?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/exceptions/INF-10002')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Backoff/requery'))?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Backoff/requery')
    expect(wrapper.text()).toContain('Closure readiness')
    expect(wrapper.text()).toContain('Evidence blocked')
    expect(wrapper.text()).toContain('Requery active')
    expect(wrapper.text()).toContain('Checker review')
    expect(wrapper.text()).toContain('Reversal ready')
    expect(wrapper.text()).toContain('Automatic attempts')
    expect(wrapper.text()).toContain('3 / 3')
    expect(wrapper.text()).toContain('Automation exhausted')
    expect(wrapper.text()).toContain('Manual only')
    expect(wrapper.text()).toContain('3 automatic / 0 manual')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Closure'))?.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Maker evidence required')
    expect(wrapper.text()).toContain('Manual completion needs ledger/session evidence')
  })

  it('mutates exception actions with audit evidence', async () => {
    const wrapper = await mountApp('/exceptions/INF-10004?tab=closure')

    expect(wrapper.text()).toContain('Approve reversal')
    await wrapper.findAll('.action-choice-grid button').find((button) => button.text().includes('Approve reversal'))?.trigger('click')
    await flushPromises()
    await wrapper.get('input[aria-label="Evidence reference"]').setValue('REV-ISW-55092')
    await wrapper.get('textarea[aria-label="Resolution reason"]').setValue('Final failed-safe status confirms the beneficiary was not credited.')
    await wrapper.findAll('button').find((button) => button.text().includes('Submit action'))?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Reversal approval captured with evidence')
    expect(wrapper.text()).toContain('Audit event captured')
    expect(wrapper.text()).toContain('Reversal approved')

    await wrapper.findAll('.detail-tabs button').find((button) => button.text().includes('Audit'))?.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('REV-ISW-55092')
  })

  it('shows route penalties and opens a route detail workspace', async () => {
    const wrapper = await mountApp('/routes')

    expect(wrapper.text()).toContain('Route matrix')
    expect(wrapper.text()).toContain('Payment provider scorecards')
    expect(wrapper.text()).toContain('Traffic instruction')
    expect(wrapper.text()).toContain('Route traffic controls')
    expect(wrapper.text()).toContain('Maintain direct credits')
    expect(wrapper.text()).toContain('Contain new traffic')
    expect(wrapper.text()).toContain('Increase eligible traffic')
    expect(wrapper.text()).toContain('Paystack final leg')
    expect(wrapper.text()).toContain('NIP final leg')
    expect(wrapper.text()).toContain('Penalty explanation')

    const routeButton = wrapper.findAll('.route-health-strip button').find((button) => button.text().includes('Moniepoint final leg'))
    expect(routeButton).toBeTruthy()
    await routeButton?.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/routes/moniepoint-final-leg')
    expect(wrapper.text()).toContain('Moniepoint final leg')
    expect(wrapper.text()).toContain('Fastest P95')
  })

  it('opens linked route work items from the route detail workspace', async () => {
    let wrapper = await mountApp('/routes/nip-final-leg?tab=cases')

    expect(wrapper.text()).toContain('CASE-INF-10002')
    expect(wrapper.text()).toContain('INF-10005')

    await wrapper.findAll('.linked-work-grid button').find((button) => button.text().includes('INF-10005'))?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/inflows/INF-10005')

    wrapper.unmount()
    wrapper = await mountApp('/routes/nip-final-leg?tab=cases')
    await wrapper.findAll('.linked-work-grid button').find((button) => button.text().includes('CASE-INF-10002'))?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/exceptions/INF-10002')
  })

  it('opens settings sections from URL deep links', async () => {
    let wrapper = await mountApp('/settings?tab=integrations')

    expect(wrapper.text()).toContain('Switching API and provider endpoints')
    expect(wrapper.findAll('.settings-tabs button').find((button) => button.text().includes('Integrations'))?.classes()).toContain('is-selected')

    wrapper.unmount()
    wrapper = await mountApp('/settings?tab=sla')
    expect(wrapper.text()).toContain('SLA definitions')
    expect(wrapper.text()).toContain('Backoff policy')

    wrapper.unmount()
    wrapper = await mountApp('/settings?tab=audit')
    expect(wrapper.text()).toContain('Audit trail and exports')
  })

  it('keeps advanced modules behind feature switches', async () => {
    const wrapper = await mountApp('/settings')

    await wrapper.findAll('.settings-tabs button').find((button) => button.text().includes('Feature controls'))?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Debit and refund rail operations')
    expect(wrapper.text()).toContain('Disabled for operators')
    expect(wrapper.findAll('input[type="checkbox"]').every((input) => !(input.element as HTMLInputElement).checked)).toBe(true)

    await wrapper.findAll('input[type="checkbox"]').at(2)?.setValue(true)
    await flushPromises()

    expect(wrapper.text()).toContain('Available to administrators')
  })

  it('redirects legacy work areas into the five-screen model', async () => {
    const wrapper = await mountApp('/credits/INF-10002')
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/inflows/INF-10002')
    expect(wrapper.text()).toContain('Inflows')
    expect(wrapper.text()).toContain('INF-10002')
  })

  it('keeps data state URL-backed', async () => {
    const wrapper = await mountApp('/?scenario=healthy')

    const dataState = wrapper.get('select[aria-label="Monitoring feed"]')
    expect((dataState.element as HTMLSelectElement).value).toBe('healthy')

    await dataState.setValue('traffic-shift')
    await flushPromises()

    expect(router.currentRoute.value.query.scenario).toBe('traffic-shift')
  })
})
