<template>
  <div class="relative inline-block" @mouseenter="onEnter" @mouseleave="show = false">
    <slot />
    <Transition name="fade">
      <div
        v-if="show && aflPlayerSeasonId"
        class="absolute z-50 left-0 top-full mt-1.5 w-96 rounded-2xl border border-border bg-surface shadow-lg pointer-events-none"
      >
        <div v-if="loading" class="px-4 py-3 text-xs text-text-faint">Loading...</div>
        <template v-else-if="data">
          <div class="px-4 pt-3 pb-3">
            <div class="flex items-center gap-2 mb-3">
              <span class="text-sm font-medium text-text">{{ name }}</span>
              <span v-if="aflStatus === 'bye'" class="text-xs rounded px-1.5 py-0.5 bg-violet-500/15 text-violet-400">Bye</span>
            </div>
            <table class="w-full tabular-nums text-xs">
              <thead>
                <tr class="text-text-faint">
                  <th class="text-left font-normal pb-2 pr-4"></th>
                  <th class="text-right font-normal pb-2 px-2">G</th>
                  <th class="text-right font-normal pb-2 px-2">K</th>
                  <th class="text-right font-normal pb-2 px-2">H</th>
                  <th class="text-right font-normal pb-2 px-2">M</th>
                  <th class="text-right font-normal pb-2 px-2">T</th>
                  <th class="text-right font-normal pb-2 px-2">R</th>
                  <th class="text-right font-normal pb-2 pl-4 text-yellow-400/70">★</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="data.aflPlayerSeason?.seasonAvg" class="text-text-muted">
                  <td class="text-text-faint py-1.5 pr-4">Season</td>
                  <td class="text-right py-1.5 px-2">{{ fmt(data.aflPlayerSeason.seasonAvg.goals) }}</td>
                  <td class="text-right py-1.5 px-2">{{ fmt(data.aflPlayerSeason.seasonAvg.kicks) }}</td>
                  <td class="text-right py-1.5 px-2">{{ fmt(data.aflPlayerSeason.seasonAvg.handballs) }}</td>
                  <td class="text-right py-1.5 px-2">{{ fmt(data.aflPlayerSeason.seasonAvg.marks) }}</td>
                  <td class="text-right py-1.5 px-2">{{ fmt(data.aflPlayerSeason.seasonAvg.tackles) }}</td>
                  <td class="text-right py-1.5 px-2">{{ fmt(data.aflPlayerSeason.seasonAvg.hitouts) }}</td>
                  <td class="text-right py-1.5 pl-4 text-yellow-400/70">{{ fmt(starScore(data.aflPlayerSeason.seasonAvg)) }}</td>
                </tr>
                <tr v-if="data.aflPlayerSeason?.last3" class="text-text font-medium">
                  <td class="text-text-faint font-normal py-1.5 pr-4">Last 3</td>
                  <td class="text-right py-1.5 px-2" :class="up(data.aflPlayerSeason.last3.goals, data.aflPlayerSeason.seasonAvg?.goals)">{{ fmt(data.aflPlayerSeason.last3.goals) }}</td>
                  <td class="text-right py-1.5 px-2" :class="up(data.aflPlayerSeason.last3.kicks, data.aflPlayerSeason.seasonAvg?.kicks)">{{ fmt(data.aflPlayerSeason.last3.kicks) }}</td>
                  <td class="text-right py-1.5 px-2" :class="up(data.aflPlayerSeason.last3.handballs, data.aflPlayerSeason.seasonAvg?.handballs)">{{ fmt(data.aflPlayerSeason.last3.handballs) }}</td>
                  <td class="text-right py-1.5 px-2" :class="up(data.aflPlayerSeason.last3.marks, data.aflPlayerSeason.seasonAvg?.marks)">{{ fmt(data.aflPlayerSeason.last3.marks) }}</td>
                  <td class="text-right py-1.5 px-2" :class="up(data.aflPlayerSeason.last3.tackles, data.aflPlayerSeason.seasonAvg?.tackles)">{{ fmt(data.aflPlayerSeason.last3.tackles) }}</td>
                  <td class="text-right py-1.5 px-2" :class="up(data.aflPlayerSeason.last3.hitouts, data.aflPlayerSeason.seasonAvg?.hitouts)">{{ fmt(data.aflPlayerSeason.last3.hitouts) }}</td>
                  <td class="text-right py-1.5 pl-4" :class="up(starScore(data.aflPlayerSeason.last3), data.aflPlayerSeason.seasonAvg ? starScore(data.aflPlayerSeason.seasonAvg) : undefined) || 'text-yellow-400'">{{ fmt(starScore(data.aflPlayerSeason.last3)) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_PLAYER_STATS_CARD } from '../api/queries'

const props = defineProps<{
  name: string
  club: string | null
  aflStatus: string | null
  aflPlayerSeasonId: string | null
  aflRoundId: string | null
}>()

const show = ref(false)

const { result: data, loading } = useQuery(
  GET_PLAYER_STATS_CARD,
  () => ({ id: props.aflPlayerSeasonId, aflRoundId: props.aflRoundId }),
  () => ({ enabled: show.value && !!props.aflPlayerSeasonId }),
)

function onEnter() {
  if (props.aflPlayerSeasonId) show.value = true
}


interface StatRow { goals: number; kicks: number; handballs: number; marks: number; tackles: number; hitouts: number }

function starScore(s: StatRow): number {
  return s.goals * 5 + s.kicks + s.handballs + s.marks * 2 + s.tackles * 4
}

function fmt(v: number): string {
  return v % 1 === 0 ? String(v) : v.toFixed(1)
}

function up(a: number, b: number | undefined): string {
  return b !== undefined && a > b ? 'text-green-400' : ''
}
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
