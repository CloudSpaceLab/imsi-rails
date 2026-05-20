<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'plain' | 'icon' | 'choice'
    size?: 'sm' | 'md'
    type?: 'button' | 'submit' | 'reset'
    selected?: boolean
    ariaLabel?: string
    title?: string
  }>(),
  {
    variant: 'primary',
    size: 'md',
    type: 'button',
    selected: false,
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
    :class="[`ui-button--${variant}`, `ui-button--${size}`, { 'is-selected': selected }]"
    :aria-label="ariaLabel"
    :title="title"
    @click="$emit('click', $event)"
  >
    <slot />
  </button>
</template>
