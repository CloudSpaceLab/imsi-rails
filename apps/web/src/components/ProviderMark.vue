<script setup lang="ts">
import { computed } from 'vue'
import { Banknote, CircleDollarSign, Globe2, Landmark, Network, ShieldCheck, Smartphone } from '@lucide/vue'
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
const badgeIcon = computed(() => {
  switch (identity.value.category) {
    case 'B2B payout network':
      return Network
    case 'Digital IMTO':
      return Smartphone
    case 'Legacy IMTO':
      return Landmark
    case 'Pan-African rail':
      return Globe2
    case 'Local payout rail':
      return Banknote
    case 'Operational queue':
      return ShieldCheck
    default:
      return CircleDollarSign
  }
})
</script>

<template>
  <span class="provider-mark" :class="`provider-mark--${size}`">
    <span class="provider-mark__badge" :style="{ '--provider-color': identity.color }" :title="identity.category" aria-hidden="true">
      <component :is="badgeIcon" :size="size === 'lg' ? 21 : size === 'sm' ? 14 : 17" :stroke-width="2.2" />
    </span>
    <span v-if="showName" class="provider-mark__copy">
      <strong>{{ identity.name }}</strong>
      <small v-if="showCategory">{{ identity.category }}</small>
    </span>
  </span>
</template>
