<template>
  <div>
    <h1 class="text-2xl font-bold mb-1">Roster</h1>
    <p class="text-text-muted mb-6">Season roster and AFL stat averages</p>

    <div v-if="fflLoading" class="text-text-faint">Loading…</div>
    <div v-else-if="fflError" class="text-red-400">{{ fflError.message }}</div>
    <template v-else-if="season">
      <!-- Club selector + Manage toggle -->
      <div class="mb-6 flex items-center gap-4">
        <div>
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
        <button
          @click="managing = !managing"
          class="rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors"
          :class="managing
            ? 'border-active bg-active text-active-text'
            : 'border-border bg-surface text-text hover:bg-surface-hover'"
        >
          {{ managing ? 'Done' : 'Manage' }}
        </button>
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
                <th v-if="managing" class="py-2 px-2 font-medium text-right"></th>
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
                <td v-if="managing" class="py-2 px-2 text-right">
                  <button
                    @click="removePlayer(row.playerSeasonId)"
                    class="text-red-400 hover:text-red-300 text-xs font-medium"
                    :disabled="removingId === row.playerSeasonId"
                  >
                    {{ removingId === row.playerSeasonId ? 'Removing…' : 'Remove' }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
      <p v-else class="text-text-faint">No players on roster.</p>

      <!-- Add player search (manage mode only) -->
      <div v-if="managing" class="mt-6">
        <h2 class="text-lg font-semibold mb-2">Add Player</h2>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search AFL players by name…"
          class="w-full max-w-md rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none"
        />
        <div v-if="searchLoading" class="mt-2 text-text-faint text-sm">Searching…</div>
        <div v-else-if="searchResults.length > 0" class="mt-2 max-w-md">
          <div
            v-for="player in searchResults"
            :key="player.id"
            class="flex items-center justify-between border-b border-border-subtle py-2"
          >
            <span class="text-sm">{{ player.name }}</span>
            <button
              @click="addPlayer(player)"
              class="rounded border border-active px-2 py-0.5 text-xs font-medium text-active hover:bg-active hover:text-active-text transition-colors"
              :disabled="addingId === player.id"
            >
              {{ addingId === player.id ? 'Adding…' : 'Add' }}
            </button>
          </div>
        </div>
        <div v-else-if="searchQuery.length >= 2 && !searchLoading" class="mt-2 text-text-faint text-sm">
          No players found.
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_ROSTER, GET_AFL_PLAYER_SEASON_STATS, SEARCH_AFL_PLAYERS } from '../api/queries'
import { REMOVE_FFL_PLAYER_FROM_SEASON, ADD_FFL_ROSTER_PLAYER } from '../api/mutations'

const props = defineProps<{ seasonId: string }>()

const managing = ref(false)

// Query 1: FFL roster
const { result: fflResult, loading: fflLoading, error: fflError, refetch: refetchRoster } = useQuery(GET_FFL_ROSTER, () => ({ seasonId: props.seasonId }))

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
  aflPlayerId: string | null
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
    (r: { playerSeasonId: string; player: { name: string; aflPlayerId?: string }; aflPlayerSeasonId?: string }) => {
      const stats = r.aflPlayerSeasonId ? statsMap.value.get(r.aflPlayerSeasonId) : undefined
      return {
        playerSeasonId: r.playerSeasonId,
        name: r.player.name,
        aflPlayerId: r.player.aflPlayerId ?? null,
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

// --- Manage mode: search + add/remove ---

const searchQuery = ref('')
let searchTimeout: ReturnType<typeof setTimeout> | null = null
const debouncedQuery = ref('')

watch(searchQuery, (val) => {
  if (searchTimeout) clearTimeout(searchTimeout)
  if (val.length < 2) {
    debouncedQuery.value = ''
    return
  }
  searchTimeout = setTimeout(() => {
    debouncedQuery.value = val
  }, 300)
})

const { result: searchResult, loading: searchLoading } = useQuery(
  SEARCH_AFL_PLAYERS,
  () => ({ query: debouncedQuery.value }),
  () => ({ enabled: debouncedQuery.value.length >= 2 })
)

// Filter out players already on the roster (by AFL player ID)
const rosterAflPlayerIds = computed(() => {
  const ids = new Set<string>()
  for (const row of rosterRows.value) {
    if (row.aflPlayerId) ids.add(row.aflPlayerId)
  }
  return ids
})

const searchResults = computed(() => {
  const players = searchResult.value?.aflPlayerSearch ?? []
  return players.filter((p: { id: string }) => !rosterAflPlayerIds.value.has(p.id))
})

// Remove player
const removingId = ref<string | null>(null)
const { mutate: removePlayerMutation } = useMutation(REMOVE_FFL_PLAYER_FROM_SEASON)

async function removePlayer(playerSeasonId: string) {
  removingId.value = playerSeasonId
  try {
    await removePlayerMutation({ id: playerSeasonId })
    await refetchRoster()
  } finally {
    removingId.value = null
  }
}

// Add player
const addingId = ref<string | null>(null)
const { mutate: addPlayerMutation } = useMutation(ADD_FFL_ROSTER_PLAYER)

async function addPlayer(player: { id: string; name: string }) {
  if (!selectedClubSeasonId.value) return
  addingId.value = player.id
  try {
    await addPlayerMutation({
      input: {
        aflPlayerId: player.id,
        aflPlayerName: player.name,
        clubSeasonId: selectedClubSeasonId.value,
      },
    })
    await refetchRoster()
  } finally {
    addingId.value = null
  }
}

// Reset search when exiting manage mode
watch(managing, (val) => {
  if (!val) {
    searchQuery.value = ''
    debouncedQuery.value = ''
  }
})
</script>
