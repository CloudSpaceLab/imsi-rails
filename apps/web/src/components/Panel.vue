<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    eyebrow?: string
    tone?: 'default' | 'inset' | 'glow'
    accent?: 'healthy' | 'watch' | 'degraded' | 'blocked' | 'recovery' | 'stale' | 'unknown'
    loading?: boolean
  }>(),
  {
    tone: 'default',
    loading: false,
  },
)
</script>

<template>
  <section class="kit-panel" :class="[`kit-panel--${tone}`, accent ? `kit-panel--${accent}` : '', { 'is-loading': loading }]">
    <header v-if="title || eyebrow || $slots.actions" class="kit-panel__header">
      <div class="kit-panel__title">
        <p v-if="eyebrow" class="eyebrow">
          <span v-if="accent" class="eyebrow-dot" aria-hidden="true"></span>
          {{ eyebrow }}
        </p>
        <h2 v-if="title">{{ title }}</h2>
      </div>
      <div v-if="$slots.actions" class="kit-panel__actions">
        <slot name="actions" />
      </div>
    </header>
    <div class="kit-panel__body">
      <slot />
    </div>
  </section>
</template>
