<script setup lang="ts">
defineProps<{
  score: number
  reason: string
}>()
</script>

<template>
  <span class="route-score" :aria-label="`Route score ${score}. ${reason}`">
    <strong>{{ score }}</strong>
    <span>{{ reason }}</span>
  </span>
</template>

