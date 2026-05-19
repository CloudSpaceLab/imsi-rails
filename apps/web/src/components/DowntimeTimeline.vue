<script setup lang="ts">
import HealthBadge from './HealthBadge.vue'
import type { DowntimeEvent } from '../types'

defineProps<{
  events: DowntimeEvent[]
}>()
</script>

<template>
  <section class="panel" id="downtime" aria-labelledby="downtime-title">
    <header class="panel__header">
      <div>
        <h2 id="downtime-title">Downtime Timeline</h2>
        <p>Ria / EU -> Nigeria</p>
      </div>
      <HealthBadge state="watch" window="live" />
    </header>

    <ol class="downtime" aria-label="Downtime and operator timeline">
      <li v-for="event in events" :key="`${event.time}-${event.title}`" class="downtime__item" :class="`downtime__item--${event.state}`">
        <time>{{ event.time }}</time>
        <div>
          <strong>{{ event.title }}</strong>
          <span>{{ event.actor }}</span>
          <p>{{ event.detail }}</p>
        </div>
      </li>
    </ol>
  </section>
</template>
