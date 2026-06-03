import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { routeForScreen } from '../router'
import type { RoutePenalty, ScreenId } from '../types'
import { normalizeDataState, routeSlug } from './useInboundTower'

export function useTowerRouting() {
  const route = useRoute()
  const router = useRouter()
  const currentScenario = computed(() => normalizeDataState(route.query.scenario))

  function globalQuery() {
    return {
      scenario: currentScenario.value,
    }
  }

  function activate(screen: ScreenId) {
    void router.push({ path: routeForScreen(screen).path, query: globalQuery() })
  }

  function openPath(path: string, query: Record<string, string | undefined> = {}) {
    void router.push({ path, query: { ...globalQuery(), ...withoutEmpty(query) } })
  }

  function openInflow(reference: string) {
    openPath(`/inflows/${encodeURIComponent(reference)}`)
  }

  function openException(reference: string) {
    openPath(`/exceptions/${encodeURIComponent(reference)}`)
  }

  function openRoute(routePenalty: RoutePenalty) {
    openPath(`/routes/${routeSlug(routePenalty)}`)
  }

  return {
    route,
    router,
    currentScenario,
    globalQuery,
    activate,
    openPath,
    openInflow,
    openException,
    openRoute,
  }
}

function withoutEmpty(query: Record<string, string | undefined>) {
  return Object.fromEntries(Object.entries(query).filter(([, value]) => value))
}
