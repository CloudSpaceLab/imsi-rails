<script setup lang="ts">
import type { Component } from 'vue'
import type { HealthState } from '../types'

withDefaults(
  defineProps<{
    label: string
    value: string | number
    detail?: string
    tone?: HealthState | 'brand'
    icon?: Component
    size?: 'sm' | 'md'
  }>(),
  {
    tone: 'brand',
    size: 'md',
  },
)
</script>

<template>
  <article class="kpi-tile" :class="[`kpi-tile--${tone}`, `kpi-tile--${size}`]">
    <div class="kpi-tile__head">
      <span class="eyebrow">{{ label }}</span>
      <component :is="icon" v-if="icon" :size="16" aria-hidden="true" />
    </div>
    <strong>{{ value }}</strong>
    <p v-if="detail">{{ detail }}</p>
  </article>
</template>
