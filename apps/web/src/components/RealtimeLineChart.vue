<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import 'uplot/dist/uPlot.min.css'
import type { DashboardTimeseriesPoint } from '../types'

const props = defineProps<{
  points: DashboardTimeseriesPoint[]
  metric: 'processed_count' | 'volume' | 'sla_rate' | 'p95_seconds'
  label: string
  values?: number[]
  color?: string
  height?: number
}>()

const host = ref<HTMLElement | null>(null)
let chart: { destroy: () => void } | null = null
let renderId = 0

const values = computed(() => {
  if (props.values?.length === props.points.length) return props.values.map((value) => Number(value || 0))
  return props.points.map((point) => Number(point[props.metric] ?? 0))
})
const maxValue = computed(() => Math.max(...values.value, 1))
const chartColor = computed(() => props.color ?? '#56d6ff')
const chartHeight = computed(() => props.height ?? 180)
const svgPoints = computed(() =>
  values.value
    .map((value, index) => {
      const x = props.points.length <= 1 ? 0 : (index / (props.points.length - 1)) * 100
      const y = 34 - (value / maxValue.value) * 30
      return `${x},${y}`
    })
    .join(' '),
)

async function renderUPlot() {
  if (import.meta.env.MODE === 'test' || !host.value || props.points.length === 0) return
  const id = ++renderId
  const { default: uPlot } = await import('uplot')
  if (id !== renderId || !host.value) return
  chart?.destroy()
  host.value.replaceChildren()
  const timestamps = props.points.map((point) => Math.floor(new Date(point.time).getTime() / 1000))
  const series = values.value
  chart = new uPlot(
    {
      width: Math.max(host.value.clientWidth, 320),
      height: chartHeight.value,
      cursor: { drag: { x: false, y: false } },
      legend: { show: false },
      scales: { x: { time: true } },
      series: [{}, { label: props.label, stroke: chartColor.value, width: 2 }],
      axes: [{ stroke: '#7f8d9b', grid: { stroke: 'rgba(148, 163, 184, 0.12)' } }, { stroke: '#7f8d9b' }],
    },
    [timestamps, series],
    host.value,
  )
}

onMounted(() => nextTick(renderUPlot))
onBeforeUnmount(() => {
  renderId += 1
  chart?.destroy()
})
watch(() => [props.points, props.metric, props.values, props.label, props.color], () => nextTick(renderUPlot), { deep: true })
</script>

<template>
  <div class="realtime-chart" :aria-label="label" :style="{ color: chartColor }">
    <div ref="host" class="realtime-chart__canvas" :style="{ minHeight: `${chartHeight}px` }"></div>
    <svg class="realtime-chart__fallback" viewBox="0 0 100 38" role="img" :aria-label="label">
      <polyline :points="svgPoints" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" />
    </svg>
  </div>
</template>
