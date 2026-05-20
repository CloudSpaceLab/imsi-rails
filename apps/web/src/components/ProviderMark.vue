<script setup lang="ts">
import { computed } from 'vue'
import { getProviderIdentity } from '../services/identity'

const props = withDefaults(
  defineProps<{
    provider: string
    showName?: boolean
    showCategory?: boolean
    size?: 'sm' | 'md' | 'lg'
  }>(),
  {
    showName: true,
    showCategory: false,
    size: 'md',
  },
)

const identity = computed(() => getProviderIdentity(props.provider))
</script>

<template>
  <span class="provider-mark" :class="`provider-mark--${size}`">
    <span class="provider-mark__badge" :style="{ '--provider-color': identity.color }">
      {{ identity.mark }}
    </span>
    <span v-if="showName" class="provider-mark__copy">
      <strong>{{ identity.name }}</strong>
      <small v-if="showCategory">{{ identity.category }}</small>
    </span>
  </span>
</template>
