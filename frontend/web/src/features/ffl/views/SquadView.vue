<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold mb-1">{{ clubSeason?.club.name ?? '' }}</h1>
      <p class="text-text-muted">{{ clubSeason?.season.name ? clubSeason.season.name + ' Squad' : 'Squad' }}</p>
    </div>

    <!-- Manage toggle -->
    <div class="mb-6 flex items-center gap-4">
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

    <div v-if="squadLoading" class="text-text-faint">Loading…</div>
    <div v-else-if="squadError" class="text-red-400">{{ squadError.message }}</div>
    <template v-else>
      <div class="flex gap-8 items-start">
        <!-- Player list -->
        <div class="flex-1 min-w-0">
          <div v-if="players.length > 0" class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border text-left text-text-muted">
                  <th class="py-2 pr-4 font-medium">Player</th>
                  <th v-if="managing" class="py-2 px-2 font-medium text-right"></th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in players"
                  :key="row.id"
                  class="border-b border-border-subtle hover:bg-surface-hover"
                >
                  <td class="py-2 pr-4 font-medium">{{ row.player.name }}</td>
                  <td v-if="managing" class="py-2 px-2 text-right">
                    <button
                      @click="removePlayer(row.id)"
                      class="text-red-400 hover:text-red-300 text-xs font-medium"
                      :disabled="removingId === row.id"
                    >
                      {{ removingId === row.id ? 'Removing…' : 'Remove' }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <p v-else class="text-text-faint">No players on squad.</p>
        </div>

        <!-- Add player search (manage mode only) -->
        <div v-if="managing" class="w-72 shrink-0">
          <h2 class="text-lg font-semibold mb-2">Add Player</h2>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search AFL players by name…"
            class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none"
          />
          <div v-if="searchLoading" class="mt-2 text-text-faint text-sm">Searching…</div>
          <div v-else-if="searchResults.length > 0" class="mt-2">
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
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_CLUB_SEASON, SEARCH_AFL_PLAYERS } from '../api/queries'
import { REMOVE_FFL_PLAYER_FROM_SEASON, ADD_FFL_SQUAD_PLAYER } from '../api/mutations'
import { useFflState } from '../composables/useFflState'

const props = defineProps<{ seasonId: string }>()

const { selectedClubId } = useFflState()
const managing = ref(false)

// Squad query — driven by global club selection
const { result: squadResult, loading: squadLoading, error: squadError, refetch: refetchSquad } = useQuery(
  GET_FFL_CLUB_SEASON,
  () => ({ seasonId: props.seasonId, clubId: selectedClubId.value }),
  () => ({ enabled: !!selectedClubId.value })
)

const clubSeason = computed(() => squadResult.value?.fflClubSeason ?? null)
const clubSeasonId = computed(() => clubSeason.value?.id ?? '')

const players = computed(() => {
  const nodes = clubSeason.value?.players?.nodes ?? []
  return [...nodes].sort((a, b) => {
    const lastA = a.player.name.split(' ').pop()?.toLowerCase() ?? ''
    const lastB = b.player.name.split(' ').pop()?.toLowerCase() ?? ''
    return lastA.localeCompare(lastB)
  })
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

const squadAflPlayerIds = computed(() => {
  const ids = new Set<string>()
  for (const row of players.value) {
    if (row.player.aflPlayerId) ids.add(row.player.aflPlayerId)
  }
  return ids
})

const searchResults = computed(() => {
  const results = searchResult.value?.aflPlayerSearch ?? []
  return results.filter((p: { id: string }) => !squadAflPlayerIds.value.has(p.id))
})

// Remove player
const removingId = ref<string | null>(null)
const { mutate: removePlayerMutation } = useMutation(REMOVE_FFL_PLAYER_FROM_SEASON)

async function removePlayer(playerSeasonId: string) {
  removingId.value = playerSeasonId
  try {
    await removePlayerMutation({ id: playerSeasonId })
    await refetchSquad()
  } finally {
    removingId.value = null
  }
}

// Add player
const addingId = ref<string | null>(null)
const { mutate: addPlayerMutation } = useMutation(ADD_FFL_SQUAD_PLAYER)

async function addPlayer(player: { id: string; name: string }) {
  if (!clubSeasonId.value) return
  addingId.value = player.id
  try {
    await addPlayerMutation({
      input: {
        aflPlayerId: player.id,
        aflPlayerName: player.name,
        clubSeasonId: clubSeasonId.value,
      },
    })
    await refetchSquad()
  } finally {
    addingId.value = null
  }
}

watch(managing, (val) => {
  if (!val) {
    searchQuery.value = ''
    debouncedQuery.value = ''
  }
})

watch(selectedClubId, () => {
  managing.value = false
})
</script>
