<script setup lang="ts">
import { AlertTriangle, CheckCircle2, Clock3, LockKeyhole, WifiOff } from '@lucide/vue'
import { computed } from 'vue'
import type { UiScenario } from '../types'

const props = defineProps<{
  scenario: UiScenario
}>()

const content = computed(() => {
  if (props.scenario === 'loading') {
    return { icon: Clock3, title: 'Loading data', detail: 'Refreshing routes, transfers, and settlement.' }
  }
  if (props.scenario === 'api-failure') {
    return { icon: WifiOff, title: 'Feed unavailable', detail: 'Showing the last verified operational state.' }
  }
  if (props.scenario === 'permission-denied') {
    return { icon: LockKeyhole, title: 'Limited access', detail: 'Some details are hidden for this role.' }
  }
  if (props.scenario === 'empty') {
    return { icon: CheckCircle2, title: 'No active work', detail: 'No degraded routes or breaks match this view.' }
  }
  if (props.scenario === 'stale-fx') {
    return { icon: AlertTriangle, title: 'Stale FX', detail: 'Some rates are excluded from routing.' }
  }
  return null
})
</script>

<template>
  <aside v-if="content" class="state-banner" :class="`state-banner--${scenario}`">
    <component :is="content.icon" :size="18" aria-hidden="true" />
    <div>
      <strong>{{ content.title }}</strong>
      <p>{{ content.detail }}</p>
    </div>
  </aside>
</template>
