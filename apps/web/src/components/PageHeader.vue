<script setup lang="ts">
defineProps<{
  eyebrow?: string
  title: string
  description?: string
  breadcrumbs?: Array<{
    label: string
    path?: string
  }>
}>()
</script>

<template>
  <header class="page-header">
    <div class="page-header__copy">
      <nav v-if="breadcrumbs?.length" class="breadcrumb-nav" aria-label="Breadcrumb">
        <ol>
          <li v-for="(crumb, index) in breadcrumbs" :key="`${crumb.label}-${index}`">
            <RouterLink v-if="crumb.path && index < breadcrumbs.length - 1" :to="crumb.path">{{ crumb.label }}</RouterLink>
            <span v-else>{{ crumb.label }}</span>
          </li>
        </ol>
      </nav>
      <p v-if="eyebrow" class="eyebrow">{{ eyebrow }}</p>
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>
