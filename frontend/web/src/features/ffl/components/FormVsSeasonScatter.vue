<template>
  <section v-if="points.length > 0" class="mt-8 border-t border-border pt-6">
    <div class="flex items-center justify-between mb-1">
      <h2 class="text-lg font-semibold text-text-heading">Form vs season</h2>
      <div class="flex items-center gap-2">
        <span class="text-xs text-text-faint">Position</span>
        <div class="flex items-center rounded border border-border overflow-hidden text-[10px]">
          <button
            v-for="col in statCols"
            :key="col.key"
            class="px-1.5 py-0.5 transition-colors"
            :class="statKey === col.key ? 'bg-control text-text' : 'text-text-faint hover:text-text'"
            :title="POSITION_LABEL[col.key]"
            @click="emit('update:statKey', col.key)"
          >{{ col.label }}</button>
          <button
            class="px-1.5 py-0.5 transition-colors"
            :class="statKey === 'star' ? 'bg-control text-text' : 'text-text-faint hover:text-text'"
            title="Star"
            @click="emit('update:statKey', 'star')"
          >★</button>
        </div>
      </div>
    </div>
    <p class="text-xs text-text-faint mb-3">
      {{ statLabel }} — last {{ LAST_N }} vs season average. Above the line is trending up, below is trending down.
    </p>
    <div class="h-72">
      <Scatter :data="chartData" :options="chartOptions" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Scatter } from 'vue-chartjs'
import {
  Chart as ChartJS,
  LineController,
  PointElement,
  LineElement,
  LinearScale,
  Tooltip,
} from 'chart.js'
import { useTheme } from '@/composables/useTheme'
import { chartAxisColors } from '@/utils/chartColors'
import { starScore, trendDir, statCols, LAST_N, type StatSummary, type StatKey } from '../utils/playerStats'
import { POSITION_LABEL } from '../utils/position'

ChartJS.register(LineController, PointElement, LineElement, LinearScale, Tooltip)

interface PlayerStats {
  id: string
  playerName: string
  statsAll: StatSummary | null
  statsLastN: StatSummary | null
}

const props = defineProps<{ players: PlayerStats[]; statKey: StatKey | 'star' }>()
const emit = defineEmits<{ 'update:statKey': [value: StatKey | 'star'] }>()
const { isDark } = useTheme()

const statLabel = computed(() => POSITION_LABEL[props.statKey])

function valueOf(s: StatSummary): number {
  return props.statKey === 'star' ? starScore(s) : s[props.statKey]
}

const TREND_COLOR = { up: '#22c55e', down: '#ef4444', neutral: '#94a3b8' }

interface Point {
  id: string
  name: string
  season: number
  form: number
  trend: 'up' | 'down' | null
}

const points = computed<Point[]>(() =>
  props.players
    .filter((p) => p.statsAll && p.statsLastN)
    .map((p) => {
      const season = valueOf(p.statsAll!)
      const form = valueOf(p.statsLastN!)
      return { id: p.id, name: p.playerName, season, form, trend: trendDir(form, season) }
    }),
)

const domain = computed(() => {
  const vals = points.value.flatMap((p) => [p.season, p.form])
  if (vals.length === 0) return { min: 0, max: 100 }
  const min = Math.min(0, ...vals)
  const max = Math.max(...vals)
  const pad = (max - min) * 0.08
  return { min: Math.floor(min - pad), max: Math.ceil(max + pad) }
})

const chartData = computed(() => {
  const { min, max } = domain.value
  const { grid } = chartAxisColors(isDark.value)
  return {
    datasets: [
      {
        type: 'line' as const,
        label: 'Parity',
        data: [{ x: min, y: min }, { x: max, y: max }],
        borderColor: grid,
        borderWidth: 1,
        pointRadius: 0,
        fill: false,
        order: 2,
      },
      {
        type: 'scatter' as const,
        label: 'Players',
        data: points.value.map((p) => ({ x: p.season, y: p.form })),
        backgroundColor: points.value.map((p) => TREND_COLOR[p.trend ?? 'neutral']),
        pointRadius: 5,
        pointHoverRadius: 7,
        pointBorderWidth: 0,
        order: 1,
        players: points.value,
      },
    ],
  }
})

const chartOptions = computed(() => {
  const { grid, tick } = chartAxisColors(isDark.value)
  return {
    responsive: true,
    maintainAspectRatio: false,
    // See AllStatsFormScatter.vue — scatter's default 'point' interaction
    // mode groups every nearby point into one tooltip; 'nearest' keeps it
    // to the single point under the cursor.
    interaction: { mode: 'nearest' as const, intersect: true },
    plugins: {
      legend: { display: false },
      tooltip: {
        filter: (item: { datasetIndex: number }) => item.datasetIndex === 1,
        callbacks: {
          title: (items: { datasetIndex: number; dataIndex: number }[]) => {
            const item = items[0]
            const dataset = chartData.value.datasets[item.datasetIndex] as { players: Point[] }
            return dataset.players[item.dataIndex]?.name ?? ''
          },
          label: (ctx: { dataIndex: number }) => {
            const p = points.value[ctx.dataIndex]
            return p ? [`Season: ${p.season.toFixed(1)}`, `Last ${LAST_N}: ${p.form.toFixed(1)}`] : ''
          },
        },
      },
    },
    scales: {
      x: {
        min: domain.value.min,
        max: domain.value.max,
        title: { display: true, text: `${statLabel.value} — season average`, color: tick },
        grid: { color: grid },
        ticks: { color: tick },
      },
      y: {
        min: domain.value.min,
        max: domain.value.max,
        title: { display: true, text: `${statLabel.value} — last ${LAST_N} average`, color: tick },
        grid: { color: grid },
        ticks: { color: tick },
      },
    },
  }
})
</script>
