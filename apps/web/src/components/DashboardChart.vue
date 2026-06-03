<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  ArcElement,
  BarController,
  BarElement,
  CategoryScale,
  Chart,
  DoughnutController,
  Filler,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
} from 'chart.js'

type ChartKind = 'line' | 'bar' | 'doughnut'

type ChartDataset = {
  label: string
  values: number[]
  color?: string
  fill?: boolean
}

const props = withDefaults(
  defineProps<{
    kind: ChartKind
    labels: string[]
    datasets: ChartDataset[]
    summary: string
    unit?: string
    height?: number
  }>(),
  {
    unit: '',
    height: 224,
  },
)
const emit = defineEmits<{
  drilldown: []
}>()

Chart.register(
  ArcElement,
  BarController,
  BarElement,
  CategoryScale,
  DoughnutController,
  Filler,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
)

const palette = ['#0a66ff', '#067647', '#b54708', '#b42318', '#6941c6', '#475467']
const canvas = ref<HTMLCanvasElement | null>(null)
let chart: Chart | null = null
let renderId = 0

const maxValue = computed(() =>
  Math.max(
    ...props.datasets.flatMap((dataset) => dataset.values.map((value) => Number(value || 0))),
    1,
  ),
)

const fallbackRows = computed(() =>
  props.labels.map((label, index) => ({
    label,
    value: props.datasets[0]?.values[index] ?? 0,
    width: `${Math.max(4, ((props.datasets[0]?.values[index] ?? 0) / maxValue.value) * 100)}%`,
    color: props.datasets[0]?.color ?? palette[index % palette.length],
  })),
)

function colorFor(index: number) {
  return props.datasets[index]?.color ?? palette[index % palette.length]
}

function destroyChart() {
  renderId += 1
  const target = canvas.value
  if (target) Chart.getChart(target)?.destroy()
  chart?.destroy()
  chart = null
}

async function renderChart() {
  if (import.meta.env.MODE === 'test' || !canvas.value || props.labels.length === 0) return

  const id = ++renderId
  await nextTick()
  if (id !== renderId || !canvas.value) return
  const target = canvas.value

  Chart.getChart(target)?.destroy()
  chart?.destroy()
  chart = null

  const isDoughnut = props.kind === 'doughnut'
  const suffix = props.unit

  chart = new Chart(target, {
    type: props.kind,
    data: {
      labels: props.labels,
      datasets: props.datasets.map((dataset, index) => {
        const color = colorFor(index)
        return {
          label: dataset.label,
          data: dataset.values,
          borderColor: color,
          backgroundColor: isDoughnut
            ? props.labels.map((_, labelIndex) => palette[labelIndex % palette.length])
            : dataset.fill
              ? `${color}1f`
              : `${color}d9`,
          borderWidth: props.kind === 'line' ? 2 : 0,
          pointRadius: props.kind === 'line' ? 2.5 : 0,
          pointHoverRadius: 4,
          tension: 0.36,
          fill: Boolean(dataset.fill),
          borderRadius: props.kind === 'bar' ? 5 : 0,
          maxBarThickness: 28,
          cutout: isDoughnut ? '66%' : undefined,
        }
      }),
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: { duration: 450 },
      interaction: { intersect: false, mode: 'index' },
      plugins: {
        legend: {
          display: props.datasets.length > 1 || isDoughnut,
          position: 'bottom',
          labels: {
            usePointStyle: true,
            boxWidth: 7,
            boxHeight: 7,
            color: '#475467',
            font: { size: 11, family: 'Inter, Segoe UI, Arial, sans-serif' },
          },
        },
        tooltip: {
          callbacks: {
            label(context) {
              const label = context.dataset.label ? `${context.dataset.label}: ` : ''
              const value = Number(context.parsed.y ?? context.parsed ?? 0)
              return `${label}${value.toLocaleString()}${suffix}`
            },
          },
        },
      },
      scales: isDoughnut
        ? {}
        : {
            x: {
              grid: { display: false },
              ticks: { color: '#667085', font: { size: 11 } },
              border: { display: false },
            },
            y: {
              beginAtZero: true,
              grid: { color: 'rgba(16, 24, 40, 0.08)' },
              ticks: {
                color: '#667085',
                font: { size: 11 },
                callback(value) {
                  return `${Number(value).toLocaleString()}${suffix}`
                },
              },
              border: { display: false },
            },
          },
    },
  })
}

onMounted(renderChart)
onBeforeUnmount(destroyChart)
watch(() => [props.kind, props.labels, props.datasets, props.unit], renderChart, { deep: true })
</script>

<template>
  <figure class="dashboard-chart" :aria-label="summary" role="button" tabindex="0" @click="emit('drilldown')" @keydown.enter.prevent="emit('drilldown')" @keydown.space.prevent="emit('drilldown')">
    <div class="dashboard-chart__canvas" :style="{ minHeight: `${height}px` }">
      <canvas ref="canvas" :aria-label="summary"></canvas>
    </div>
    <figcaption>{{ summary }}</figcaption>
    <div class="dashboard-chart__fallback" aria-hidden="true">
      <span v-for="row in fallbackRows" :key="row.label">
        <small>{{ row.label }}</small>
        <i :style="{ width: row.width, background: row.color }"></i>
        <strong>{{ row.value.toLocaleString() }}{{ unit }}</strong>
      </span>
    </div>
  </figure>
</template>
