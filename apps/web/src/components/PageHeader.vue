<script setup lang="ts">
defineProps<{
  eyebrow: string
  title: string
  description?: string
}>()
</script>

<template>
  <header class="page-header">
    <div class="page-header__copy">
      <p class="eyebrow">{{ eyebrow }}</p>
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>
