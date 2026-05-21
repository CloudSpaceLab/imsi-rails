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
    clickable?: boolean
  }>(),
  {
    tone: 'brand',
    size: 'md',
    clickable: false,
  },
)

defineEmits<{
  click: [event: MouseEvent]
}>()
</script>

<template>
  <component
    :is="clickable ? 'button' : 'article'"
    :type="clickable ? 'button' : undefined"
    class="kpi-tile"
    :class="[`kpi-tile--${tone}`, `kpi-tile--${size}`, { 'kpi-tile--clickable': clickable }]"
    @click="$emit('click', $event)"
  >
    <div class="kpi-tile__head">
      <span class="eyebrow">{{ label }}</span>
      <component :is="icon" v-if="icon" :size="16" aria-hidden="true" />
    </div>
    <strong>{{ value }}</strong>
    <p v-if="detail">{{ detail }}</p>
  </component>
</template>
