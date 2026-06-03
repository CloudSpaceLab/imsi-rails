import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import type { ScreenId } from './types'

export const screenRoutes: Array<{ id: ScreenId; path: string; label: string }> = [
  { id: 'command', path: '/', label: 'Command Center' },
  { id: 'inflows', path: '/inflows', label: 'Inflows' },
  { id: 'routes', path: '/routes', label: 'Routes' },
  { id: 'exceptions', path: '/exceptions', label: 'Exceptions' },
  { id: 'settings', path: '/settings', label: 'Settings' },
]

const placeholder = { template: '<span />' }

const routes: RouteRecordRaw[] = screenRoutes.map((screen) => ({
  path: screen.path,
  name: screen.id,
  component: placeholder,
  meta: { screen: screen.id, label: screen.label },
}))

routes.push(
  {
    path: '/inflows/:reference',
    name: 'inflow-detail',
    component: placeholder,
    meta: { screen: 'inflows', label: 'Inflow detail' },
  },
  {
    path: '/routes/:routeId',
    name: 'route-detail',
    component: placeholder,
    meta: { screen: 'routes', label: 'Route detail' },
  },
  {
    path: '/exceptions/:reference',
    name: 'exception-detail',
    component: placeholder,
    meta: { screen: 'exceptions', label: 'Exception detail' },
  },
)

const legacyRedirects: Array<{ from: string; to: string }> = [
  { from: '/transactions', to: '/inflows' },
  { from: '/transactions/:reference', to: '/inflows/:reference' },
  { from: '/credits', to: '/inflows' },
  { from: '/credits/:reference', to: '/inflows/:reference' },
  { from: '/reconcile', to: '/exceptions' },
  { from: '/reconcile/:reference', to: '/exceptions/:reference' },
  { from: '/incidents', to: '/exceptions' },
  { from: '/incidents/:reference', to: '/exceptions/:reference' },
  { from: '/policy', to: '/settings' },
  { from: '/policy/:section', to: '/settings' },
  { from: '/rates', to: '/settings' },
  { from: '/providers', to: '/routes' },
  { from: '/providers/:providerId', to: '/routes' },
  { from: '/contracts', to: '/settings' },
  { from: '/contracts/:contractId', to: '/settings' },
  { from: '/audit', to: '/settings' },
  { from: '/inbound-sla', to: '/settings' },
]

legacyRedirects.forEach((redirect) => {
  routes.push({
    path: redirect.from,
    redirect: (to) => ({
      path: interpolatePath(redirect.to, to.params),
      query: to.query,
    }),
  })
})

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

export function routeForScreen(screen: ScreenId) {
  return screenRoutes.find((route) => route.id === screen) ?? screenRoutes[0]
}

function interpolatePath(path: string, params: Record<string, string | string[]>) {
  return path.replace(/:([A-Za-z0-9_]+)/g, (_, key: string) => {
    const value = params[key]
    return encodeURIComponent(Array.isArray(value) ? value[0] ?? '' : value ?? '')
  })
}
