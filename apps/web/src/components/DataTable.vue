<script setup lang="ts">
import EmptyState from './EmptyState.vue'

withDefaults(
  defineProps<{
    empty?: boolean
    loading?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    empty: false,
    loading: false,
    emptyTitle: 'No results',
    emptyDescription: 'No rows match the current filters.',
  },
)
</script>

<template>
  <div class="data-table" :aria-busy="loading || undefined">
    <div class="data-table__scroll">
      <slot />
    </div>
    <EmptyState v-if="empty && !loading" :title="emptyTitle" :description="emptyDescription" />
    <div v-if="loading" class="table-loading" aria-label="Loading rows">
      <span v-for="index in 5" :key="index"></span>
    </div>
  </div>
</template>
