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
import { router } from '../router'

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
    expect(country.text()).not.toContain('DE')
    expect(country.text()).not.toContain('NG')
    expect(country.find('.country-flag--de').exists()).toBe(true)
    expect(country.find('.country-flag--ng').exists()).toBe(true)

    const table = mount(DataTable, {
      props: { empty: true, emptyTitle: 'No rows' },
      slots: { default: '<table><tbody></tbody></table>' },
    })
    expect(table.text()).toContain('No rows')
  })
})

describe('premium dashboard workflows', () => {
  it('provides scenario fixtures for empty and failure states', () => {
    expect(getDashboardMock('empty').transactions).toHaveLength(0)
    expect(getDashboardMock('empty').reconciliation).toHaveLength(0)
    expect(getDashboardMock('api-failure').viewState).toBe('error')
    expect(getDashboardMock('permission-denied').auditEvents).toHaveLength(0)
  })

  it('renders all primary pages from the shared shell', async () => {
    const wrapper = await mountApp()
    const pageTitles = ['Control Room', 'Transactions', 'Routes', 'Policy', 'Incidents', 'Rates & costs', 'Reconcile', 'Providers', 'Audit']

    for (const title of pageTitles) {
      const button = wrapper.findAll('button.nav-item').find((item) => item.text().includes(title))
      expect(button, `missing nav item ${title}`).toBeTruthy()
      await button?.trigger('click')
      await flushPromises()
      expect(wrapper.text()).toContain(title)
    }
  })

  it('keeps transactions dense, searchable, and traceable', async () => {
    const wrapper = await mountApp()
    await wrapper.findAll('button.nav-item').find((item) => item.text().includes('Transactions'))?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Find, filter, trace')
    expect(wrapper.text()).toContain('Trace detail')
    expect(wrapper.text()).toContain('Select a transfer')
    expect(wrapper.text()).toContain('IMSI-txn_000000000001')

    await wrapper.get('input[aria-label="Search transactions"]').setValue('RMT-UK-55Q8')
    expect(wrapper.text()).toContain('IMSI-txn_000000031822')
    expect(wrapper.text()).not.toContain('IMSI-txn_000000000014')
  })

  it('makes rates comparable from a visible currency baseline', async () => {
    const wrapper = await mountApp()
    await wrapper.findAll('button.nav-item').find((item) => item.text().includes('Rates & costs'))?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('USD baseline')
    expect(wrapper.find('select[aria-label="Base currency"]').exists()).toBe(true)
    expect(wrapper.find('select[aria-label="Comparison currency"]').exists()).toBe(true)
  })

  it('shows policy save/reset behavior for editable thresholds', async () => {
    const wrapper = await mountApp()
    await wrapper.findAll('button.nav-item').find((item) => item.text().includes('Policy'))?.trigger('click')
    await flushPromises()

    const saveThresholds = wrapper.findAll('button').find((button) => button.text().includes('Save thresholds'))
    expect(saveThresholds?.attributes('disabled')).toBeDefined()

    const thresholdInput = wrapper.findAll('input[type="number"]').at(0)
    await thresholdInput?.setValue(120)

    const enabledSaveThresholds = wrapper.findAll('button').find((button) => button.text().includes('Save thresholds'))
    expect(enabledSaveThresholds?.attributes('disabled')).toBeUndefined()
  })

  it('supports dashboard context controls and KPI drilldowns', async () => {
    const wrapper = await mountApp()
    expect(wrapper.find('select[aria-label="Dashboard provider"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Operations picture')
    expect(wrapper.text()).toContain('Sample operational data')
    expect(wrapper.text()).not.toContain('mock snapshot')

    await wrapper.findAll('.kpi-tile--clickable').at(0)?.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/transactions')
    expect(router.currentRoute.value.query.scenario).toBe('degraded-ria')
    expect(router.currentRoute.value.query.currency).toBe('USD')
  })

  it('keeps data state quiet but URL-backed', async () => {
    const wrapper = await mountApp('/?scenario=healthy')
    await flushPromises()

    const dataState = wrapper.get('select[aria-label="Data state"]')
    expect((dataState.element as HTMLSelectElement).value).toBe('healthy')
    expect(wrapper.text()).toContain('No route action needed')
    expect(wrapper.text()).not.toContain('scenario')

    await dataState.setValue('traffic-shift')
    await flushPromises()
    expect(router.currentRoute.value.query.scenario).toBe('traffic-shift')
    expect(wrapper.text()).toContain('Traffic shift is reducing failed credits')
  })

  it('cleans page-specific query state when navigating between work areas', async () => {
    const wrapper = await mountApp('/transactions?timing=Stalled+only&currency=NGN&scenario=traffic-shift')
    await flushPromises()

    await wrapper.findAll('button.nav-item').find((item) => item.text().includes('Rates & costs'))?.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/rates')
    expect(router.currentRoute.value.query.currency).toBe('NGN')
    expect(router.currentRoute.value.query.scenario).toBe('traffic-shift')
    expect(router.currentRoute.value.query.timing).toBeUndefined()
  })

  it('shows transaction reporting controls and compact trace expansion', async () => {
    const wrapper = await mountApp('/transactions')
    await flushPromises()

    expect(wrapper.text()).toContain('Export CSV')
    expect(wrapper.find('select[aria-label="Rows per page"]').exists()).toBe(true)
    await wrapper.find('tbody tr.click-row').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Open full trace')
    expect(wrapper.text()).toContain('Elapsed')
    expect(wrapper.text()).toContain('Owner')
    expect(wrapper.text()).not.toContain('References')
    expect(wrapper.text()).not.toContain('Route decision')
    expect(wrapper.text()).not.toContain('QA limit')
  })

  it('surfaces maker-checker policy controls', async () => {
    const wrapper = await mountApp('/policy')
    await flushPromises()

    expect(wrapper.text()).toContain('Policy inventory')
    expect(wrapper.text()).toContain('Create corridor policy')
    expect(wrapper.text()).toContain('pending approval')
    expect(wrapper.text()).toContain('Activate')
  })
})
