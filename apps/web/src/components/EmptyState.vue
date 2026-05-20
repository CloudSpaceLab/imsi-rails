<script setup lang="ts">
import type { Component } from 'vue'

withDefaults(
  defineProps<{
    title: string
    description?: string
    icon?: Component
    tone?: 'neutral' | 'success' | 'warning'
  }>(),
  {
    tone: 'neutral',
  },
)
</script>

<template>
  <div class="empty-state" :class="`empty-state--${tone}`">
    <span v-if="icon" class="empty-state__icon" aria-hidden="true">
      <component :is="icon" :size="20" />
    </span>
    <strong>{{ title }}</strong>
    <p v-if="description">{{ description }}</p>
    <slot />
  </div>
</template>
