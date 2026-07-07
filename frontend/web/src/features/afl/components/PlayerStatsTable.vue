<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border text-left text-text-muted">
          <th class="py-2 pr-4 font-medium">Player</th>
          <th v-for="col in preDisposalCols" :key="col.key" class="py-2 px-2 font-medium text-right w-16">
            {{ col.label }}
          </th>
          <th class="py-2 px-2 font-medium text-right w-16">D</th>
          <th v-for="col in postDisposalCols" :key="col.key" class="py-2 px-2 font-medium text-right w-16">
            {{ col.label }}
          </th>
          <th class="py-2 px-2 font-medium text-right w-16">Pts</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="pm in clubMatch.playerMatches"
          :key="pm.id"
          :id="`pm-${pm.id}`"
          class="border-b border-border-subtle hover:bg-surface-hover"
          :class="{ 'highlight-pulse': pm.id === props.highlightPmId }"
        >
          <td class="py-2 pr-4 font-medium">
            <router-link
              :to="{ name: 'ffl-afl-player-season', params: { aflPlayerSeasonId: pm.playerSeasonId } }"
              class="hover:text-active transition-colors"
            >{{ pm.player.name }}</router-link>
          </td>
          <td v-for="col in preDisposalCols" :key="col.key" class="py-1 px-1 text-right">
            <input
              v-if="!readonly"
              type="number"
              :value="pm[col.key]"
              min="0"
              class="w-14 rounded bg-transparent px-1 py-1 text-right text-text tabular-nums hover:bg-control focus:bg-control focus:outline-none focus:ring-1 focus:ring-control-ring"
              @change="onStatChange(pm, col.key, $event)"
            />
            <span v-else class="tabular-nums px-1" :style="heatColor(pm[col.key], col.key)">{{ pm[col.key] }}</span>
          </td>
          <td class="py-2 px-2 text-right tabular-nums text-text-muted">{{ pm.disposals }}</td>
          <td v-for="col in postDisposalCols" :key="col.key" class="py-1 px-1 text-right">
            <input
              v-if="!readonly"
              type="number"
              :value="pm[col.key]"
              min="0"
              class="w-14 rounded bg-transparent px-1 py-1 text-right text-text tabular-nums hover:bg-control focus:bg-control focus:outline-none focus:ring-1 focus:ring-control-ring"
              @change="onStatChange(pm, col.key, $event)"
            />
            <span v-else class="tabular-nums px-1" :style="heatColor(pm[col.key], col.key)">{{ pm[col.key] }}</span>
          </td>
          <td class="py-2 px-2 text-right tabular-nums text-text-muted">{{ pm.score }}</td>
        </tr>
      </tbody>
      <tfoot>
        <tr class="border-t border-border-strong font-semibold text-text-heading">
          <td class="py-2 pr-4">Totals</td>
          <td v-for="col in preDisposalCols" :key="col.key" class="py-2 px-2 text-right tabular-nums">
            {{ totals[col.key] }}
          </td>
          <td class="py-2 px-2 text-right tabular-nums">{{ totals.disposals }}</td>
          <td v-for="col in postDisposalCols" :key="col.key" class="py-2 px-2 text-right tabular-nums">
            {{ totals[col.key] }}
          </td>
          <td class="py-2 px-2 text-right tabular-nums">{{ totals.score }}</td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from '@/composables/useTheme'
import { heatStyle } from '@/utils/heatmap'

interface PlayerMatch {
  id: string
  playerSeasonId: string
  player: { id: string; name: string }
  status: string
  kicks: number
  handballs: number
  marks: number
  hitouts: number
  tackles: number
  goals: number
  behinds: number
  disposals: number
  score: number
}

interface ClubMatch {
  id: string
  club: { id: string; name: string }
  playerMatches: PlayerMatch[]
}

const { isDark } = useTheme()

const props = withDefaults(defineProps<{
  clubMatch: ClubMatch
  readonly?: boolean
  highlightPmId?: string | null
}>(), {
  readonly: false,
  highlightPmId: null,
})

const emit = defineEmits<{
  update: [input: { playerSeasonId: string; clubMatchId: string; [key: string]: unknown }]
}>()

const preDisposalCols = [
  { key: 'kicks' as const, label: 'K' },
  { key: 'handballs' as const, label: 'H' },
]

const postDisposalCols = [
  { key: 'marks' as const, label: 'M' },
  { key: 'hitouts' as const, label: 'R' },
  { key: 'tackles' as const, label: 'T' },
  { key: 'goals' as const, label: 'G' },
  { key: 'behinds' as const, label: 'B' },
]

const statColumns = [...preDisposalCols, ...postDisposalCols]

type StatKey = typeof statColumns[number]['key']

const heatKeys = ['kicks', 'handballs', 'marks', 'hitouts', 'tackles', 'goals', 'behinds', 'disposals', 'score'] as const
type HeatKey = typeof heatKeys[number]

const columnRange = computed(() => {
  const range = {} as Record<HeatKey, { min: number; max: number }>
  for (const key of heatKeys) {
    const vals = props.clubMatch.playerMatches.map(pm => pm[key])
    range[key] = { min: Math.min(...vals), max: Math.max(...vals) }
  }
  return range
})

function heatColor(value: number, key: HeatKey): Record<string, string> {
  const r = columnRange.value[key]
  if (!r) return {}
  return heatStyle(value, r.min, r.max, isDark.value)
}

const totals = computed(() => {
  const keys = [...statColumns.map(c => c.key), 'disposals' as const, 'score' as const]
  const sums: Record<string, number> = {}
  for (const key of keys) {
    sums[key] = props.clubMatch.playerMatches.reduce((sum, pm) => sum + pm[key], 0)
  }
  return sums
})

function onStatChange(pm: PlayerMatch, key: StatKey, event: Event) {
  const target = event.target as HTMLInputElement
  const value = parseInt(target.value, 10)
  if (isNaN(value) || value < 0) return
  if (value === pm[key]) return

  emit('update', {
    playerSeasonId: pm.playerSeasonId,
    clubMatchId: props.clubMatch.id,
    [key]: value,
  })
}
</script>

<style scoped>
@keyframes highlight-pulse {
  0%, 30% { background-color: rgb(34 197 94 / 0.3); }
  100% { background-color: transparent; }
}
.highlight-pulse {
  animation: highlight-pulse 7.3s ease-out forwards;
}
</style>
