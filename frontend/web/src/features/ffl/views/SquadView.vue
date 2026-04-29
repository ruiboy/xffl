<template>
  <div>
    <Breadcrumb v-if="clubSeason" :items="breadcrumbs" />
    <div class="mb-6 flex items-center">
      <h1 class="text-2xl font-bold flex items-center gap-3">
        <img v-if="clubSeason" :src="clubLogoUrl(clubSeason.club.name)" :alt="clubSeason.club.name" class="w-10 h-10 object-contain" />
        {{ clubSeason?.club.name ?? '' }}
      </h1>
      <router-link
        v-if="isMyClub && liveRoundId"
        :to="{ name: 'ffl-team-builder', params: { seasonId: props.seasonId, roundId: liveRoundId } }"
        class="ml-auto flex items-center gap-1.5 text-sm text-text-muted hover:text-text transition-colors"
      >
        <IconTeamBuilder class="w-4 h-4" />
        Team Builder
      </router-link>
    </div>

    <!-- Manage toggle + Team Builder link — only for the selected club -->
    <div v-if="isMyClub" class="mb-6 flex items-center gap-4">
      <button
        @click="managing = !managing"
        class="rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors"
        :class="managing
          ? 'border-active bg-active text-active-text'
          : 'border-border bg-surface text-text hover:bg-surface-hover'"
      >
        <span class="flex items-center gap-1.5">
          <IconManage v-if="!managing" class="w-3.5 h-3.5" />
          {{ managing ? 'Done' : 'Manage' }}
        </span>
      </button>
      <span v-if="saveMessage" class="text-sm text-green-500">{{ saveMessage }}</span>
    </div>

    <div v-if="squadLoading" class="text-text-faint">Loading...</div>
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
                  <th class="py-2 pr-4 font-medium">Club</th>
                  <th v-if="!managing" class="py-2">
                    <div class="flex gap-0.5">
                      <span
                        v-for="round in rounds"
                        :key="round.id"
                        class="w-5 text-center text-[10px] text-text-faint font-normal"
                      >{{ roundLabel(round.name) }}</span>
                    </div>
                  </th>
                  <th v-if="isMyClub && managing" class="py-2 px-2 font-medium text-right"></th>
                </tr>
              </thead>
              <tbody>
                <template v-for="(group, gi) in groupedPlayers" :key="group.pos ?? 'bench'">
                  <tr v-if="gi > 0"><td colspan="2" class="pt-3"></td></tr>
                  <tr
                    v-for="row in group.players"
                    :key="row.id"
                    class="border-b border-border-subtle hover:bg-surface-hover"
                  >
                    <td class="py-2 pr-4 font-medium">{{ row.player.name }}</td>
                    <td class="py-2 pr-4 text-xs text-text-muted">{{ row.aflPlayerSeason?.clubSeason?.club?.name ?? '—' }}</td>
                    <td v-if="!managing" class="py-2">
                      <div class="flex gap-0.5">
                        <span
                          v-for="round in rounds"
                          :key="round.id"
                          class="w-5 text-center text-xs font-mono inline-block"
                          :class="roundColor(row.id, round.id)"
                        >{{ roundLetter(row.id, round.id) }}</span>
                      </div>
                    </td>
                    <td v-if="isMyClub && managing" class="py-2 px-2 text-right">
                      <button
                        @click="removePlayer(row.id)"
                        aria-label="Remove"
                        class="text-red-400 hover:text-red-300 transition-colors disabled:opacity-40"
                        :disabled="removingId === row.id"
                      >
                        <IconBin class="w-3.5 h-3.5" />
                      </button>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
          <p v-else class="text-text-faint">No players on squad.</p>
        </div>

        <!-- Add player search (manage mode only) -->
        <div v-if="isMyClub && managing" class="w-72 shrink-0">
          <h2 class="text-lg font-semibold mb-2">Add Player</h2>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search AFL players by name..."
            class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none"
          />
          <div v-if="searchLoading" class="mt-2 text-text-faint text-sm">Searching...</div>
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
                {{ addingId === player.id ? 'Adding...' : 'Add' }}
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
import { GET_FFL_CLUB_SEASON, SEARCH_AFL_PLAYERS, GET_FFL_SEASON_POSITIONS } from '../api/queries'
import { REMOVE_FFL_PLAYER_FROM_SEASON, ADD_FFL_SQUAD_PLAYER } from '../api/mutations'
import { useFflState } from '../composables/useFflState'
import Breadcrumb from '../components/Breadcrumb.vue'
import IconTeamBuilder from '../components/icons/IconTeamBuilder.vue'
import IconManage from '../components/icons/IconManage.vue'
import IconBin from '../components/icons/IconBin.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { POSITION_LETTERS, POSITION_COLORS, POSITION_ORDER, POSITION_LABEL, primaryPosition, type RoundEntry } from '../utils/position'

const props = defineProps<{ seasonId: string; clubId: string }>()

const { selectedClubId, liveRoundId } = useFflState()
const managing = ref(false)

const isMyClub = computed(() => !!selectedClubId.value && props.clubId === selectedClubId.value)

// Squad query — driven by route clubId
const { result: squadResult, loading: squadLoading, error: squadError, refetch: refetchSquad } = useQuery(
  GET_FFL_CLUB_SEASON,
  () => ({ seasonId: props.seasonId, clubId: props.clubId }),
)

const clubSeason = computed(() => squadResult.value?.fflClubSeason ?? null)
const clubSeasonId = computed(() => clubSeason.value?.id ?? '')

const breadcrumbs = computed(() => {
  if (!clubSeason.value) return []
  return [
    { label: 'FFL' },
    { label: clubSeason.value.season.name, to: { name: 'home' } },
    { label: clubSeason.value.club.name },
  ]
})

const players = computed(() => {
  const nodes = clubSeason.value?.players?.nodes ?? []
  return [...nodes].sort((a, b) => {
    const lastA = a.player.name.split(' ').pop()?.toLowerCase() ?? ''
    const lastB = b.player.name.split(' ').pop()?.toLowerCase() ?? ''
    return lastA.localeCompare(lastB)
  })
})

// --- Round history ---

const { result: seasonResult } = useQuery(GET_FFL_SEASON_POSITIONS, () => ({ id: props.seasonId }))

const rounds = computed(() => seasonResult.value?.fflSeason?.rounds ?? [])

const playerRoundMap = computed((): Map<string, Map<string, RoundEntry>> => {
  const map = new Map<string, Map<string, RoundEntry>>()
  for (const round of rounds.value) {
    for (const match of round.matches ?? []) {
      for (const side of [match.homeClubMatch, match.awayClubMatch]) {
        if (!side) continue
        for (const pm of side.playerMatches ?? []) {
          const isBench = pm.backupPositions != null || pm.interchangePosition != null
          if (!map.has(pm.playerSeasonId)) map.set(pm.playerSeasonId, new Map())
          map.get(pm.playerSeasonId)!.set(round.id, { position: pm.position ?? null, isBench })
        }
      }
    }
  }
  return map
})

function roundLetter(playerSeasonId: string, roundId: string): string {
  const e = playerRoundMap.value.get(playerSeasonId)?.get(roundId)
  if (!e) return '–'
  if (e.isBench) return 'B'
  return e.position ? (POSITION_LETTERS[e.position] ?? '?') : '–'
}

function roundColor(playerSeasonId: string, roundId: string): string {
  const e = playerRoundMap.value.get(playerSeasonId)?.get(roundId)
  if (!e) return 'text-text-faint'
  if (e.isBench) return 'text-text-muted'
  return e.position ? (POSITION_COLORS[e.position] ?? 'text-text') : 'text-text-faint'
}

function roundLabel(name: string): string {
  return name.replace(/\D+/g, '')
}

// --- Position grouping (recency-weighted) ---

const groupedPlayers = computed(() => {
  type P = typeof players.value[number]
  const buckets = new Map<string | null, P[]>(
    [...POSITION_ORDER.map(p => [p, []] as [string, P[]]), [null, []]]
  )
  for (const p of players.value) {
    const pos = primaryPosition(p.id, playerRoundMap.value, rounds.value)
    buckets.get(POSITION_ORDER.includes(pos as typeof POSITION_ORDER[number]) ? pos : null)!.push(p)
  }
  return [...POSITION_ORDER, null].flatMap(pos => {
    const group = buckets.get(pos) ?? []
    if (!group.length) return []
    return [{ pos, label: pos ? POSITION_LABEL[pos] : 'Bench / Unassigned', players: group }]
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

// Saved flash
const saveMessage = ref('')
let saveMessageTimer: ReturnType<typeof setTimeout> | null = null
function flashSaved() {
  if (saveMessageTimer) clearTimeout(saveMessageTimer)
  saveMessage.value = 'Saved'
  saveMessageTimer = setTimeout(() => { saveMessage.value = '' }, 3000)
}

// Remove player
const removingId = ref<string | null>(null)
const { mutate: removePlayerMutation } = useMutation(REMOVE_FFL_PLAYER_FROM_SEASON)

async function removePlayer(playerSeasonId: string) {
  removingId.value = playerSeasonId
  try {
    await removePlayerMutation({ id: playerSeasonId })
    await refetchSquad()
    flashSaved()
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
    flashSaved()
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

watch(isMyClub, (val) => {
  if (!val) managing.value = false
})
</script>
