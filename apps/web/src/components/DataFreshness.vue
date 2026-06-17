<script setup lang="ts">
defineProps<{
  updated: string
  stale?: boolean
}>()
</script>

<template>
  <span class="freshness" :class="{ 'freshness--stale': stale }">
    <span aria-hidden="true"></span>
    {{ stale ? 'Stale data' : 'Live' }} - {{ updated }}
  </span>
</template>
