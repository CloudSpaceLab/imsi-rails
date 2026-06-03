<script setup lang="ts">
defineProps<{
  updated: string
  stale?: boolean
  mode?: string
}>()
</script>

<template>
  <span class="freshness" :class="{ 'freshness--stale': stale }">
    <span aria-hidden="true"></span>
    {{ stale ? 'Feed delayed' : 'Feed updated' }} - {{ updated }}<small v-if="mode">{{ mode }}</small>
  </span>
</template>
