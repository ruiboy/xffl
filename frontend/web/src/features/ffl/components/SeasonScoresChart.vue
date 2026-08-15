<template>
  <section v-if="rounds.length > 0" class="mt-8 border-t border-border pt-6">
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-lg font-semibold text-text-heading">Scores by round</h2>
      <div class="flex items-center rounded border border-border overflow-hidden text-[10px] text-text-faint">
        <button
          class="px-1.5 py-0.5 transition-colors"
          :class="viewMode === 'chart' ? 'bg-control text-text' : 'hover:text-text'"
          @click="viewMode = 'chart'"
        >Chart</button>
        <button
          class="px-1.5 py-0.5 transition-colors"
          :class="viewMode === 'table' ? 'bg-control text-text' : 'hover:text-text'"
          @click="viewMode = 'table'"
        >Table</button>
      </div>
    </div>

    <div v-if="viewMode === 'chart'" class="h-80">
      <Line :data="chartData" :options="chartOptions" />
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border text-left text-text-muted">
            <th class="py-2 pr-4 font-medium">Round</th>
            <th v-for="club in sortedClubs" :key="club.id" class="py-2 px-2 font-medium text-right">
              {{ club.club.name }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(round, i) in rounds" :key="round.id" class="border-b border-border-subtle">
            <td class="py-1.5 pr-4 text-text-faint">{{ round.name }}</td>
            <td
              v-for="club in sortedClubs"
              :key="club.id"
              class="py-1.5 px-2 text-right tabular-nums"
              :style="cellStyle(perClubEntries[club.id][i].kind)"
            >
              {{ formatCell(perClubEntries[club.id][i]) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  Legend,
} from 'chart.js'
import { GET_FFL_SEASON_ROUND_HISTORY } from '../api/queries'
import { useTheme } from '@/composables/useTheme'
import {
  deriveClubRoundEntries,
  RESULT_COLORS,
  clubColor,
  chartAxisColors,
  chartSurfaceColor,
  type SeasonRound,
  type ClubRoundEntry,
  type RoundKind,
} from '../utils/roundHistory'

ChartJS.register(LineElement, PointElement, CategoryScale, LinearScale, Tooltip, Legend)

interface Club {
  id: string
  club: { id: string; name: string }
}

const props = defineProps<{ seasonId: string; clubs: Club[] }>()
const { isDark } = useTheme()
const viewMode = ref<'chart' | 'table'>('chart')
const hoveredIndex = ref<number | null>(null)

// Club colors come back either as a hex string (palette fallback) or an
// rgba(...) string (named colors from clubColors.ts) — handle both rather
// than assuming hex, since appending hex digits to an rgba() string makes
// an invalid color that canvas silently ignores (keeping the previous
// draw's color, which is what caused the cross-line color bleed).
function withAlpha(color: string, alpha: number): string {
  if (color.startsWith('#')) {
    return `${color}${Math.round(alpha * 255).toString(16).padStart(2, '0')}`
  }
  const match = color.match(/rgba?\(([^)]+)\)/)
  if (!match) return color
  const [r, g, b] = match[1].split(',').map((s) => s.trim())
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

const { result } = useQuery(
  GET_FFL_SEASON_ROUND_HISTORY,
  () => ({ id: props.seasonId }),
)

const rounds = computed<SeasonRound[]>(() => result.value?.fflSeason?.rounds ?? [])

// Alphabetical by name (not current ladder rank) so a club's legend/tooltip
// position and table column stay put as standings shift round to round.
// Color is keyed by name too (clubColors.ts), so this doesn't affect it.
const sortedClubs = computed(() => [...props.clubs].sort((a, b) => a.club.name.localeCompare(b.club.name)))

const perClubEntries = computed<Record<string, ClubRoundEntry[]>>(() => {
  const map: Record<string, ClubRoundEntry[]> = {}
  for (const club of sortedClubs.value) {
    map[club.id] = deriveClubRoundEntries(rounds.value, club.id)
  }
  return map
})

function cellStyle(kind: RoundKind): Record<string, string> {
  if (kind === 'pending') return {}
  return { color: RESULT_COLORS[kind] }
}

function formatCell(entry: ClubRoundEntry): string {
  if (entry.kind === 'pending') return '—'
  if (entry.kind === 'bye' || entry.kind === 'superbye') return 'Bye'
  return String(entry.score)
}

interface ClubDataset {
  label: string
  data: (number | null)[]
  borderColor: string
  backgroundColor: string
  borderWidth: number
  tension: number
  spanGaps: boolean
  pointRadius: number
  pointHoverRadius: number
  pointBorderWidth: number
  pointBorderColor: string
  pointBackgroundColor: string[]
  clubEntries: ClubRoundEntry[]
}

const chartData = computed(() => {
  const ring = chartSurfaceColor(isDark.value)
  const datasets: ClubDataset[] = sortedClubs.value.map((club, i) => {
    const entries = perClubEntries.value[club.id]
    const color = clubColor(club.club.name, i, isDark.value)
    const isHovered = hoveredIndex.value === i
    const isDimmed = hoveredIndex.value != null && !isHovered
    return {
      label: club.club.name,
      data: entries.map((e) => e.score),
      borderColor: isDimmed ? withAlpha(color, 0.2) : color,
      backgroundColor: color,
      borderWidth: isHovered ? 3 : 2,
      tension: 0,
      spanGaps: false,
      pointRadius: 4,
      pointHoverRadius: 6,
      pointBorderWidth: 2,
      pointBorderColor: ring,
      pointBackgroundColor: entries.map((e) => isDimmed ? withAlpha(RESULT_COLORS[e.kind], 0.2) : RESULT_COLORS[e.kind]),
      clubEntries: entries,
    }
  })
  return {
    labels: rounds.value.map((r) => r.name),
    datasets,
  }
})

const chartOptions = computed(() => {
  const { grid, tick } = chartAxisColors(isDark.value)
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: 'index' as const, intersect: false },
    plugins: {
      // The identity channel for 4 lines: converging end-of-line labels get
      // unreadable when scores bunch together, so club color is carried by
      // the legend instead — swatch AND label text both in the club's line
      // color, so a club is identifiable at a glance without hovering.
      legend: {
        position: 'top' as const,
        onHover: (_event: unknown, legendItem: { datasetIndex?: number }) => {
          hoveredIndex.value = legendItem.datasetIndex ?? null
        },
        onLeave: () => {
          hoveredIndex.value = null
        },
        labels: {
          color: tick,
          usePointStyle: true,
          pointStyle: 'line' as const,
          boxWidth: 24,
          generateLabels: (chart: ChartJS) => {
            const base = ChartJS.defaults.plugins.legend.labels.generateLabels(chart)
            return base.map((item) => ({
              ...item,
              fontColor: chart.data.datasets[item.datasetIndex as number].borderColor as string,
            }))
          },
        },
      },
      tooltip: {
        // Suppressed while a legend item is hovered (hoveredIndex is driven
        // entirely by legend onHover/onLeave) — otherwise Chart.js's
        // nearest-index interaction fires the data tooltip just from the
        // cursor being over the legend row, above the actual plot.
        enabled: hoveredIndex.value === null,
        mode: 'index' as const,
        intersect: false,
        filter: (item: { datasetIndex: number; dataIndex: number }) => {
          const dataset = chartData.value.datasets[item.datasetIndex]
          return dataset.clubEntries[item.dataIndex].score != null
        },
        callbacks: {
          label: (ctx: { datasetIndex: number; dataIndex: number; dataset: { label?: string } }) => {
            const dataset = chartData.value.datasets[ctx.datasetIndex]
            const e = dataset.clubEntries[ctx.dataIndex]
            if (e.kind === 'bye' || e.kind === 'superbye') return `${dataset.label}: Bye`
            return `${dataset.label}: ${e.score} vs ${e.opponent} (${e.kind})`
          },
        },
      },
    },
    scales: {
      x: {
        offset: false,
        grid: { color: grid },
        ticks: {
          color: tick,
          autoSkip: true,
          maxRotation: 0,
          // Abbreviate only the rendered tick — the tooltip title still
          // uses the full round name from the underlying label.
          callback: (_value: number, index: number) => {
            const name = rounds.value[index]?.name
            return name === 'Grand Final' ? 'GF' : name
          },
        },
      },
      y: { grid: { color: grid }, ticks: { color: tick }, beginAtZero: true },
    },
  }
})
</script>
