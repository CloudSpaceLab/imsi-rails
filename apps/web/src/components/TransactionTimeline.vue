<script setup lang="ts">
import type { TimelineStep } from '../types'

defineProps<{
  steps: TimelineStep[]
}>()
</script>

<template>
  <ol class="timeline" aria-label="Transaction timeline">
    <li v-for="step in steps" :key="step.label" class="timeline__item" :class="`timeline__item--${step.status}`">
      <span class="timeline__marker" aria-hidden="true"></span>
      <div>
        <strong>{{ step.label }}</strong>
        <span>{{ step.owner }}</span>
      </div>
      <time>{{ step.time }}</time>
    </li>
  </ol>
</template>

