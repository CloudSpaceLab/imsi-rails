<script setup lang="ts">
import { computed } from 'vue'
import { healthStates } from '../copy'
import type { HealthState } from '../types'

const props = defineProps<{
  state: HealthState
  window?: string
  updated?: string
}>()

const label = computed(() => healthStates[props.state])
</script>

<template>
  <span class="health-badge" :class="`health-badge--${state}`" :aria-label="`${label}${window ? `, ${window}` : ''}${updated ? `, updated ${updated}` : ''}`">
    <span class="health-badge__dot" aria-hidden="true"></span>
    {{ label }}
  </span>
</template>

