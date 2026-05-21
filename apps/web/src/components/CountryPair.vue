<script setup lang="ts">
import { computed } from 'vue'
import { getCountryIdentity } from '../services/identity'

const props = defineProps<{
  origin: string
  destination: string
  compact?: boolean
}>()

const originCountry = computed(() => getCountryIdentity(props.origin))
const destinationCountry = computed(() => getCountryIdentity(props.destination))
const flagClass = (code: string) => `country-flag--${code.toLowerCase()}`
</script>

<template>
  <span
    class="country-pair"
    :class="{ 'country-pair--compact': compact }"
    :aria-label="`${originCountry.name} to ${destinationCountry.name}`"
  >
    <span>
      <span class="country-flag" :class="flagClass(originCountry.code)" :title="originCountry.name" aria-hidden="true"></span>
      <strong>{{ compact ? originCountry.shortName : originCountry.name }}</strong>
    </span>
    <i aria-hidden="true"></i>
    <span>
      <span class="country-flag" :class="flagClass(destinationCountry.code)" :title="destinationCountry.name" aria-hidden="true"></span>
      <strong>{{ compact ? destinationCountry.shortName : destinationCountry.name }}</strong>
    </span>
  </span>
</template>
