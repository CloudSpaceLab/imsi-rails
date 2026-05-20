<script setup lang="ts">
import type { TimelineStep } from '../types'

defineProps<{
  steps: TimelineStep[]
}>()
</script>

<template>
  <ol class="timeline" aria-label="Transaction timeline">
    <li v-for="step in steps" :key="`${step.label}-${step.time}`" class="timeline__item" :class="`timeline__item--${step.status}`">
      <span class="timeline__marker" aria-hidden="true"></span>
      <div>
        <strong>{{ step.label }}</strong>
        <span>{{ step.owner }}</span>
        <small v-if="step.source || step.reference">{{ step.source }}<template v-if="step.reference"> / {{ step.reference }}</template></small>
        <p v-if="step.note">{{ step.note }}</p>
      </div>
      <time>
        {{ step.time }}
        <small v-if="step.duration">{{ step.duration }}</small>
      </time>
    </li>
  </ol>
</template>
