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
</script>

<template>
  <span class="country-pair" :class="{ 'country-pair--compact': compact }">
    <span>
      <b>{{ originCountry.flag }}</b>
      <strong>{{ compact ? originCountry.shortName : originCountry.name }}</strong>
    </span>
    <i aria-hidden="true"></i>
    <span>
      <b>{{ destinationCountry.flag }}</b>
      <strong>{{ compact ? destinationCountry.shortName : destinationCountry.name }}</strong>
    </span>
  </span>
</template>
