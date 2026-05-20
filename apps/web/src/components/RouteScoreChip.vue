<script setup lang="ts">
defineProps<{
  score: number
  reason: string
  confidence?: string
  policyVersion?: string
}>()
</script>

<template>
  <span class="route-score" :aria-label="`Route score ${score}. ${reason}${confidence ? `. Confidence ${confidence}` : ''}`">
    <strong>{{ score }}</strong>
    <span>{{ reason }}</span>
    <small v-if="confidence || policyVersion">{{ confidence || policyVersion }}</small>
  </span>
</template>
