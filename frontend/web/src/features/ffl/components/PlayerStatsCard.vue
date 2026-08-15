<template>
  <div ref="wrap" class="relative inline-block" @mouseenter="onEnter" @mouseleave="show = false">
    <slot />
    <!-- Teleported to body + fixed positioning so the card escapes both
         overflow-clipping ancestors (overflow-x-auto wrappers) and stacking
         contexts (sticky table columns). -->
    <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="show && aflPlayerSeasonId"
        class="fixed z-50 w-max min-w-72 rounded-2xl border border-border bg-surface shadow-lg pointer-events-none overflow-hidden"
        :style="cardStyle"
      >
        <div v-if="loading" class="px-4 py-3 text-xs text-text-faint">Loading...</div>
        <template v-else-if="data">
          <div class="px-4 py-3">
            <div v-if="starSparkline.length > 1" class="flex items-center justify-between mb-2">
              <span class="text-[10px] text-text-faint uppercase tracking-wide">Star, by round</span>
              <Sparkline :values="starSparkline" :color="sparklineColor" />
            </div>
            <table class="w-full tabular-nums text-xs">
              <thead>
                <tr class="text-text-faint">
                  <th class="text-left font-normal pb-2 pr-4"></th>
                  <th
                    v-for="col in cardCols"
                    :key="col.key"
                    class="text-right font-normal pb-2 px-2"
                    :class="col.key === 'star' ? 'text-yellow-400/70' : ''"
                  >{{ col.label }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="seasonAvg" class="text-text-muted">
                  <td class="text-text-faint py-1 pr-4 whitespace-nowrap">Season ({{ seasonAvg.games }})</td>
                  <td
                    v-for="col in cardCols"
                    :key="col.key"
                    class="text-right py-1 px-2"
                    :class="col.key === 'star' ? 'text-yellow-400/70' : ''"
                  >{{ fmt(statOf(seasonAvg, col.key)) }}</td>
                </tr>
                <tr v-if="lastN" class="text-text font-medium">
                  <td class="text-text-faint font-normal py-1 pr-4 whitespace-nowrap">Last {{ LAST_N }}</td>
                  <td
                    v-for="col in cardCols"
                    :key="col.key"
                    class="text-right py-1 px-2 whitespace-nowrap"
                    :class="col.key === 'star' ? 'text-yellow-400' : ''"
                  ><span v-if="lastNTrend(col.key) === 'up'" class="text-[10px] font-normal text-green-400 mr-0.5">↑</span><span v-else-if="lastNTrend(col.key) === 'down'" class="text-[10px] font-normal text-red-400 mr-0.5">↓</span>{{ fmt(statOf(lastN, col.key)) }}</td>
                </tr>
                <tr v-if="roundRows.length">
                  <td :colspan="cardCols.length + 1"><div class="h-px bg-border-subtle my-1" /></td>
                </tr>
                <tr v-for="m in roundRows" :key="m.id" class="text-text-muted">
                  <td class="text-text-faint py-0.5 pr-4 whitespace-nowrap">{{ m.clubMatch.match.round.name }}</td>
                  <template v-if="m.status !== 'dnp'">
                    <td
                      v-for="col in cardCols"
                      :key="col.key"
                      class="text-right py-0.5 px-2"
                      :style="roundHeat(m, col.key)"
                    >{{ statOf(m, col.key) }}</td>
                  </template>
                  <td v-else :colspan="cardCols.length" class="text-right py-0.5 px-2 text-text-faint">DNP</td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>
    </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_PLAYER_STATS_CARD } from '../api/queries'
import { starScore, statCols, trendDir, LAST_N, type StatSummary } from '../utils/playerStats'
import { heatStyle } from '@/utils/heatmap'
import { useTheme } from '@/composables/useTheme'
import Sparkline from '@/components/Sparkline.vue'

const props = defineProps<{
  name: string
  club: string | null
  aflStatus: string | null
  aflPlayerSeasonId: string | null
  aflRoundId: string | null
}>()

const show = ref(false)
const wrap = ref<HTMLElement | null>(null)
// Viewport-fixed placement, computed on hover: open on whichever side of the
// trigger has more room, capped to that space, clamped inside the viewport.
const cardStyle = ref<Record<string, string>>({})

// Width estimate for horizontal clamping — the card sizes to its content
// (w-max), so this errs generous to keep the right edge on screen.
const CARD_W = 448

const { result: data, loading } = useQuery(
  GET_PLAYER_STATS_CARD,
  () => ({ id: props.aflPlayerSeasonId, aflRoundId: props.aflRoundId }),
  () => ({ enabled: show.value && !!props.aflPlayerSeasonId }),
)

function onEnter() {
  if (!props.aflPlayerSeasonId) return
  const rect = wrap.value?.getBoundingClientRect()
  if (rect) {
    const below = window.innerHeight - rect.bottom
    const above = rect.top
    const placeAbove = below < 440 && above > below
    const maxH = Math.max(160, (placeAbove ? above : below) - 24)
    const left = Math.max(8, Math.min(rect.left, window.innerWidth - CARD_W - 8))
    cardStyle.value = {
      left: `${left}px`,
      maxWidth: `${window.innerWidth - left - 8}px`,
      maxHeight: `${maxH}px`,
      ...(placeAbove
        ? { bottom: `${window.innerHeight - rect.top + 6}px` }
        : { top: `${rect.bottom + 6}px` }),
    }
  }
  show.value = true
}

const cardCols = [...statCols, { key: 'star', label: '★' }] as const
type CardKey = typeof cardCols[number]['key']

interface CardMatch {
  id: string
  status: string | null
  goals: number
  kicks: number
  handballs: number
  marks: number
  tackles: number
  hitouts: number
  clubMatch: { match: { id: string; round: { id: string; name: string } } }
}

const seasonAvg = computed<(StatSummary & { games: number }) | null>(() => data.value?.aflPlayerSeason?.seasonAvg ?? null)
const lastN = computed<StatSummary | null>(() => data.value?.aflPlayerSeason?.lastN ?? null)

// Most recent round first.
const roundRows = computed<CardMatch[]>(() => {
  const ms: CardMatch[] = data.value?.aflPlayerSeason?.matches ?? []
  return [...ms].sort((a, b) => parseInt(b.clubMatch.match.round.id) - parseInt(a.clubMatch.match.round.id))
})

function statOf(s: StatSummary, key: CardKey): number {
  return key === 'star' ? starScore(s) : s[key]
}

const { isDark } = useTheme()

// Per-column heatmap across the played round rows.
const roundHeatRanges = computed(() => {
  const played = roundRows.value.filter(m => m.status !== 'dnp')
  const range = {} as Record<CardKey, { min: number; max: number }>
  for (const col of cardCols) {
    const vals = played.map(m => statOf(m, col.key))
    if (vals.length) range[col.key] = { min: Math.min(...vals), max: Math.max(...vals) }
  }
  return range
})

function roundHeat(m: CardMatch, key: CardKey): Record<string, string> {
  const r = roundHeatRanges.value[key]
  if (!r) return {}
  return heatStyle(statOf(m, key), r.min, r.max, isDark.value)
}

// Shared trend semantics (utils/playerStats) for the Last N row's arrows.
function lastNTrend(key: CardKey): 'up' | 'down' | null {
  if (!lastN.value || !seasonAvg.value) return null
  return trendDir(statOf(lastN.value, key), statOf(seasonAvg.value, key))
}

function fmt(v: number): string {
  return v % 1 === 0 ? String(v) : v.toFixed(1)
}

// Chronological (oldest first) star score across played rounds — roundRows
// itself is most-recent-first for the table, so this reverses it.
const starSparkline = computed<number[]>(() =>
  [...roundRows.value]
    .filter((m) => m.status !== 'dnp')
    .reverse()
    .map((m) => starScore(m)),
)

const sparklineColor = computed(() => {
  if (!lastN.value || !seasonAvg.value) return '#94a3b8'
  const dir = trendDir(starScore(lastN.value), starScore(seasonAvg.value))
  return dir === 'up' ? '#22c55e' : dir === 'down' ? '#ef4444' : '#94a3b8'
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.1s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
