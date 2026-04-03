<template>
  <div>
    <h1 class="text-2xl font-bold mb-1">Roster</h1>
    <p class="text-text-muted mb-6">Season roster and AFL stat averages</p>

    <div v-if="fflLoading" class="text-text-faint">Loading…</div>
    <div v-else-if="fflError" class="text-red-400">{{ fflError.message }}</div>
    <template v-else-if="season">
      <!-- Club selector -->
      <div class="mb-6">
        <label class="text-sm font-medium text-text-muted mr-2">Club:</label>
        <select
          v-model="selectedClubSeasonId"
          class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text focus:border-active focus:outline-none"
        >
          <option v-for="cs in season.ladder" :key="cs.id" :value="cs.id">
            {{ cs.club.name }}
          </option>
        </select>
      </div>

      <template v-if="rosterRows.length > 0">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border text-left text-text-muted">
                <th class="py-2 pr-4 font-medium">Player</th>
                <th class="py-2 px-2 font-medium text-right">GP</th>
                <th class="py-2 px-2 font-medium text-right">G</th>
                <th class="py-2 px-2 font-medium text-right">K</th>
                <th class="py-2 px-2 font-medium text-right">HB</th>
                <th class="py-2 px-2 font-medium text-right">M</th>
                <th class="py-2 px-2 font-medium text-right">T</th>
                <th class="py-2 px-2 font-medium text-right">HO</th>
                <th class="py-2 px-2 font-medium text-right">B</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in rosterRows"
                :key="row.playerSeasonId"
                class="border-b border-border-subtle hover:bg-surface-hover"
              >
                <td class="py-2 pr-4 font-medium">{{ row.name }}</td>
                <td class="py-2 px-2 text-right tabular-nums">{{ row.gamesPlayed ?? '—' }}</td>
                <td class="py-2 px-2 text-right tabular-nums" :class="row.avgGoals != null ? '' : 'text-text-faint'">
                  {{ row.avgGoals ?? '—' }}
                </td>
                <td class="py-2 px-2 text-right tabular-nums" :class="row.avgKicks != null ? '' : 'text-text-faint'">
                  {{ row.avgKicks ?? '—' }}
                </td>
                <td class="py-2 px-2 text-right tabular-nums" :class="row.avgHandballs != null ? '' : 'text-text-faint'">
                  {{ row.avgHandballs ?? '—' }}
                </td>
                <td class="py-2 px-2 text-right tabular-nums" :class="row.avgMarks != null ? '' : 'text-text-faint'">
                  {{ row.avgMarks ?? '—' }}
                </td>
                <td class="py-2 px-2 text-right tabular-nums" :class="row.avgTackles != null ? '' : 'text-text-faint'">
                  {{ row.avgTackles ?? '—' }}
                </td>
                <td class="py-2 px-2 text-right tabular-nums" :class="row.avgHitouts != null ? '' : 'text-text-faint'">
                  {{ row.avgHitouts ?? '—' }}
                </td>
                <td class="py-2 px-2 text-right tabular-nums" :class="row.avgBehinds != null ? '' : 'text-text-faint'">
                  {{ row.avgBehinds ?? '—' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
      <p v-else class="text-text-faint">No players on roster.</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_ROSTER, GET_AFL_PLAYER_SEASON_STATS } from '../api/queries'

const props = defineProps<{ seasonId: string }>()

// Query 1: FFL roster
const { result: fflResult, loading: fflLoading, error: fflError } = useQuery(GET_FFL_ROSTER, () => ({ seasonId: props.seasonId }))

const season = computed(() => fflResult.value?.fflSeason ?? null)

const selectedClubSeasonId = ref<string>('')

watch(season, (s) => {
  if (s && s.ladder.length > 0 && !selectedClubSeasonId.value) {
    selectedClubSeasonId.value = s.ladder[0].id
  }
})

const selectedClubSeason = computed(() =>
  season.value?.ladder.find((cs: { id: string }) => cs.id === selectedClubSeasonId.value) ?? null
)

// Collect AFL player season IDs from the selected club's roster
const aflPlayerSeasonIds = computed<string[]>(() => {
  if (!selectedClubSeason.value) return []
  return selectedClubSeason.value.roster
    .map((r: { aflPlayerSeasonId?: string }) => r.aflPlayerSeasonId)
    .filter((id: string | undefined): id is string => id != null)
})

// Query 2: AFL stats (only when we have IDs)
const { result: aflResult } = useQuery(
  GET_AFL_PLAYER_SEASON_STATS,
  () => ({ ids: aflPlayerSeasonIds.value }),
  () => ({ enabled: aflPlayerSeasonIds.value.length > 0 })
)

// Build a map of AFL player season ID → stats
const statsMap = computed(() => {
  const map = new Map<string, {
    gamesPlayed: number
    avgGoals: number
    avgKicks: number
    avgHandballs: number
    avgMarks: number
    avgTackles: number
    avgHitouts: number
    avgBehinds: number
  }>()
  const stats = aflResult.value?.aflPlayerSeasonStats
  if (!stats) return map
  for (const s of stats) {
    map.set(s.playerSeasonId, s)
  }
  return map
})

interface RosterRow {
  playerSeasonId: string
  name: string
  gamesPlayed: number | null
  avgGoals: string | null
  avgKicks: string | null
  avgHandballs: string | null
  avgMarks: string | null
  avgTackles: string | null
  avgHitouts: string | null
  avgBehinds: string | null
}

const rosterRows = computed<RosterRow[]>(() => {
  if (!selectedClubSeason.value) return []

  const rows: RosterRow[] = selectedClubSeason.value.roster.map(
    (r: { playerSeasonId: string; player: { name: string }; aflPlayerSeasonId?: string }) => {
      const stats = r.aflPlayerSeasonId ? statsMap.value.get(r.aflPlayerSeasonId) : undefined
      return {
        playerSeasonId: r.playerSeasonId,
        name: r.player.name,
        gamesPlayed: stats?.gamesPlayed ?? null,
        avgGoals: stats ? stats.avgGoals.toFixed(1) : null,
        avgKicks: stats ? stats.avgKicks.toFixed(1) : null,
        avgHandballs: stats ? stats.avgHandballs.toFixed(1) : null,
        avgMarks: stats ? stats.avgMarks.toFixed(1) : null,
        avgTackles: stats ? stats.avgTackles.toFixed(1) : null,
        avgHitouts: stats ? stats.avgHitouts.toFixed(1) : null,
        avgBehinds: stats ? stats.avgBehinds.toFixed(1) : null,
      }
    }
  )

  // Sort by last name
  rows.sort((a, b) => {
    const lastA = a.name.split(' ').pop()?.toLowerCase() ?? ''
    const lastB = b.name.split(' ').pop()?.toLowerCase() ?? ''
    return lastA.localeCompare(lastB)
  })

  return rows
})
</script>
