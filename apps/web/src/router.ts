import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import type { ScreenId } from './types'

export const screenRoutes: Array<{ id: ScreenId; path: string; label: string }> = [
  { id: 'control', path: '/', label: 'Control Room' },
  { id: 'transactions', path: '/transactions', label: 'Transactions' },
  { id: 'corridors', path: '/routes', label: 'Routes' },
  { id: 'policy', path: '/policy', label: 'Policy' },
  { id: 'incidents', path: '/incidents', label: 'Incidents' },
  { id: 'fx', path: '/rates', label: 'Rates & costs' },
  { id: 'reconciliation', path: '/reconcile', label: 'Reconcile' },
  { id: 'providers', path: '/providers', label: 'Providers' },
  { id: 'audit', path: '/audit', label: 'Audit' },
]

const routes: RouteRecordRaw[] = screenRoutes.map((screen) => ({
  path: screen.path,
  name: screen.id,
  component: { template: '<span />' },
  meta: { screen: screen.id, label: screen.label },
}))

routes.push({
  path: '/routes/:routeId',
  name: 'route-detail',
  component: { template: '<span />' },
  meta: { screen: 'corridors', label: 'Route detail' },
})

routes.push({
  path: '/policy/new',
  name: 'policy-new',
  component: { template: '<span />' },
  meta: { screen: 'policy', label: 'New policy' },
})

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

export function routeForScreen(screen: ScreenId) {
  return screenRoutes.find((route) => route.id === screen) ?? screenRoutes[0]
}

