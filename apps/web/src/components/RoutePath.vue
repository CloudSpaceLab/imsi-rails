<script setup lang="ts">
import CountryPair from './CountryPair.vue'
import ProviderMark from './ProviderMark.vue'

defineProps<{
  origin: string
  destination: string
  provider: string
  rail?: string
}>()
</script>

<template>
  <div class="route-path">
    <CountryPair :origin="origin" :destination="destination" compact />
    <span class="route-path__line" aria-hidden="true"></span>
    <ProviderMark :provider="provider" size="sm" />
    <span v-if="rail" class="route-path__rail">{{ rail }}</span>
  </div>
</template>
