<script setup lang="ts">
import { computed } from 'vue'
import { healthStates } from '../copy'
import type { HealthState } from '../types'

const props = defineProps<{
  state: HealthState
  window?: string
  updated?: string
  trigger?: string
}>()

const label = computed(() => healthStates[props.state])
const accessibleLabel = computed(() => {
  const parts: string[] = [label.value]
  if (props.trigger) parts.push(props.trigger)
  if (props.window) parts.push(props.window)
  if (props.updated) parts.push(`updated ${props.updated}`)
  return parts.join(', ')
})
</script>

<template>
  <span class="health-badge" :class="`health-badge--${state}`" :aria-label="accessibleLabel">
    <span class="health-badge__dot" aria-hidden="true"></span>
    <span>{{ label }}</span>
    <small v-if="window">{{ window }}</small>
  </span>
</template>
