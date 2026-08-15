<template>
  <section v-if="datasets.length > 0" class="mt-8 border-t border-border pt-6">
    <h2 class="text-lg font-semibold text-text-heading">Form vs season — all stats</h2>
    <p class="text-xs text-text-faint mb-3">
      Each stat normalized to the squad's season spread, so kicks and hitouts share one scale.
      Above the line is trending up, below is trending down.
    </p>
    <div class="h-80">
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
  Legend,
} from 'chart.js'
import { useTheme } from '@/composables/useTheme'
import { chartAxisColors, categoricalColor } from '@/utils/chartColors'
import { statCols, trendDir, LAST_N, type StatSummary, type StatKey } from '../utils/playerStats'
import { POSITION_LABEL } from '../utils/position'

ChartJS.register(LineController, PointElement, LineElement, LinearScale, Tooltip, Legend)

interface PlayerStats {
  id: string
  playerName: string
  statsAll: StatSummary | null
  statsLastN: StatSummary | null
}

const props = defineProps<{ players: PlayerStats[] }>()
const { isDark } = useTheme()

interface StatPoint {
  name: string
  x: number
  y: number
  seasonRaw: number
  formRaw: number
  trend: 'up' | 'down' | null
}

function meanStd(vals: number[]): { mean: number; std: number } {
  const mean = vals.reduce((a, b) => a + b, 0) / vals.length
  const variance = vals.reduce((a, b) => a + (b - mean) ** 2, 0) / vals.length
  return { mean, std: Math.sqrt(variance) || 1 }
}

const validPlayers = computed(() => props.players.filter((p) => p.statsAll && p.statsLastN))

// One series per stat, each z-scored against the squad's own season spread
// for that stat — season and last-N both expressed in "how far from the
// squad's typical season value" units, so a kicks point and a hitouts point
// are directly comparable on the same axes.
interface StatSeries {
  key: StatKey
  label: string
  color: string
  points: StatPoint[]
}

const series = computed<StatSeries[]>(() => {
  const players = validPlayers.value
  if (players.length === 0) return []
  return statCols.map((col, i) => {
    const seasonVals = players.map((p) => p.statsAll![col.key])
    const { mean, std } = meanStd(seasonVals)
    const points: StatPoint[] = players.map((p) => {
      const seasonRaw = p.statsAll![col.key]
      const formRaw = p.statsLastN![col.key]
      return {
        name: p.playerName,
        x: (seasonRaw - mean) / std,
        y: (formRaw - mean) / std,
        seasonRaw,
        formRaw,
        trend: trendDir(formRaw, seasonRaw),
      }
    })
    return { key: col.key, label: POSITION_LABEL[col.key], color: categoricalColor(i, isDark.value), points }
  })
})

const domain = computed(() => {
  const vals = series.value.flatMap((s) => s.points.flatMap((p) => [p.x, p.y]))
  if (vals.length === 0) return { min: -3, max: 3 }
  const min = Math.min(...vals)
  const max = Math.max(...vals)
  const pad = (max - min) * 0.08
  return { min: Math.floor(min - pad), max: Math.ceil(max + pad) }
})

const datasets = computed(() => {
  const { min, max } = domain.value
  const { grid } = chartAxisColors(isDark.value)
  const parity = {
    type: 'line' as const,
    label: 'Parity',
    data: [{ x: min, y: min }, { x: max, y: max }],
    borderColor: grid,
    borderWidth: 1,
    pointRadius: 0,
    fill: false,
    order: 99,
  }
  const statSets = series.value.map((s) => ({
    type: 'scatter' as const,
    label: s.label,
    data: s.points.map((p) => ({ x: p.x, y: p.y })),
    backgroundColor: s.color,
    pointBorderWidth: 0,
    pointRadius: 5,
    pointHoverRadius: 7,
    order: 1,
    statLabel: s.label,
    players: s.points,
  }))
  return [parity, ...statSets]
})

const chartData = computed(() => ({ datasets: datasets.value }))

const chartOptions = computed(() => {
  const { grid, tick } = chartAxisColors(isDark.value)
  return {
    responsive: true,
    maintainAspectRatio: false,
    // Scatter's default interaction mode ('point') groups every point near
    // the cursor across all datasets into one tooltip, mixing different
    // players together. 'nearest' + intersect keeps it to the single point
    // actually under the cursor.
    interaction: { mode: 'nearest' as const, intersect: true },
    plugins: {
      legend: {
        position: 'top' as const,
        labels: { color: tick, boxWidth: 10, boxHeight: 10 },
      },
      tooltip: {
        filter: (item: { datasetIndex: number }) => datasets.value[item.datasetIndex].label !== 'Parity',
        callbacks: {
          title: (items: { datasetIndex: number; dataIndex: number }[]) => {
            const item = items[0]
            const dataset = datasets.value[item.datasetIndex] as { players: StatPoint[] }
            return dataset.players[item.dataIndex]?.name ?? ''
          },
          label: (ctx: { datasetIndex: number; dataIndex: number }) => {
            const dataset = datasets.value[ctx.datasetIndex] as { statLabel: string; players: StatPoint[] }
            const p = dataset.players[ctx.dataIndex]
            if (!p) return ''
            const trendLabel = p.trend === 'up' ? 'trending up' : p.trend === 'down' ? 'trending down' : 'steady'
            return `${dataset.statLabel}: ${p.seasonRaw.toFixed(1)} → ${p.formRaw.toFixed(1)} (${trendLabel})`
          },
        },
      },
    },
    scales: {
      x: {
        min: domain.value.min,
        max: domain.value.max,
        title: { display: true, text: 'Season (relative to squad)', color: tick },
        grid: { color: grid },
        ticks: { color: tick },
      },
      y: {
        min: domain.value.min,
        max: domain.value.max,
        title: { display: true, text: `Last ${LAST_N} (relative to squad)`, color: tick },
        grid: { color: grid },
        ticks: { color: tick },
      },
    },
  }
})
</script>
