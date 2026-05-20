<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost' | 'plain' | 'danger' | 'icon' | 'choice' | 'segmented'
    size?: 'sm' | 'md' | 'lg'
    type?: 'button' | 'submit' | 'reset'
    selected?: boolean
    disabled?: boolean
    loading?: boolean
    ariaLabel?: string
    title?: string
  }>(),
  {
    variant: 'primary',
    size: 'md',
    type: 'button',
    selected: false,
    disabled: false,
    loading: false,
  },
)

defineEmits<{
  click: [event: MouseEvent]
}>()
</script>

<template>
  <button
    :type="type"
    class="ui-button"
    :class="[`ui-button--${variant}`, `ui-button--${size}`, { 'is-selected': selected, 'is-loading': loading }]"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
    :aria-label="ariaLabel"
    :title="title"
    @click="$emit('click', $event)"
  >
    <span v-if="loading" class="ui-button__spinner" aria-hidden="true"></span>
    <slot />
  </button>
</template>
