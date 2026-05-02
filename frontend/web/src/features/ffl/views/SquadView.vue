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

    <!-- Manage toolbar — only for the selected club -->
    <div v-if="isMyClub" class="mb-6 flex items-center gap-3">
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
      <button
        v-if="managing"
        @click="openAddSearch"
        class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
      >
        + Add Player
      </button>
      <span v-if="saveMessage" class="text-sm text-green-500">{{ saveMessage }}</span>
    </div>

    <div v-if="squadLoading" class="text-text-faint">Loading...</div>
    <div v-else-if="squadError" class="text-red-400">{{ squadError.message }}</div>
    <template v-else>
      <div v-if="players.length > 0" class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border text-left text-text-muted">
              <th class="py-2 pr-4 font-medium">Player</th>
              <th class="py-2 pr-4 font-medium">Club</th>
              <th class="py-2">
                <div class="flex gap-0.5">
                  <span
                    v-for="round in rounds"
                    :key="round.id"
                    class="w-5 text-center text-[10px] text-text-faint font-normal"
                  >{{ roundLabel(round.name) }}</span>
                </div>
              </th>
              <th v-if="isMyClub && managing" class="py-2 px-2"></th>
            </tr>
          </thead>
          <tbody>
            <template v-for="(group, gi) in groupedPlayers" :key="group.pos ?? 'bench'">
              <tr v-if="gi > 0"><td colspan="3" class="pt-3"></td></tr>
              <template v-for="row in group.players" :key="row.id">
                <tr
                  :class="[
                    managing ? 'cursor-pointer hover:bg-surface-hover' : '',
                    expandedId === row.id ? '' : 'border-b border-border-subtle'
                  ]"
                  @click="managing && toggleRow(row)"
                >
                  <td class="py-2 pr-4 font-medium">{{ row.player.name }}</td>
                  <td class="py-2 pr-4 text-xs text-text-muted">{{ row.aflPlayerSeason?.clubSeason?.club?.name ?? '—' }}</td>
                  <td class="py-2">
                    <div class="flex gap-0.5">
                      <span
                        v-for="round in rounds"
                        :key="round.id"
                        class="w-5 text-center text-xs font-mono inline-block"
                        :class="roundColor(row.id, round.id)"
                      >{{ roundLetter(row.id, round.id) }}</span>
                    </div>
                  </td>
                  <td v-if="isMyClub && managing" class="py-2 px-2 text-right" @click.stop>
                    <button
                      @click="openRemoveModal(row.id, row.player.name, row.aflPlayerSeason?.clubSeason?.club?.name ?? '')"
                      aria-label="Remove"
                      class="text-red-400 hover:text-red-300 transition-colors"
                    >
                      <IconBin class="w-3.5 h-3.5" />
                    </button>
                  </td>
                </tr>
                <tr v-if="expandedId === row.id" class="border-b border-border-subtle">
                  <td colspan="4" class="pb-3 pt-1 px-0">
                    <div class="flex gap-4">
                      <div class="flex flex-col gap-2 text-xs shrink-0">
                        <div><div class="text-text-muted">From</div><div class="text-text">{{ roundName(row.fromRoundId) }}</div></div>
                        <div><div class="text-text-muted">To</div><div class="text-text">{{ roundName(row.toRoundId) }}</div></div>
                        <div><div class="text-text-muted">Cost</div><div class="text-text">{{ formatCost(row.costCents) }}</div></div>
                      </div>
                      <div class="flex-1 flex flex-col" @click.stop>
                        <textarea
                          v-model="expandedNotes"
                          rows="3"
                          placeholder="Add notes..."
                          class="flex-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none resize-none"
                        />
                        <div class="flex justify-end mt-2">
                          <button
                            v-if="expandedDirty"
                            @click="saveExpanded"
                            :disabled="expandedSubmitting"
                            class="rounded-lg px-3 py-1.5 text-xs font-medium bg-active text-active-text hover:opacity-90 transition-colors disabled:opacity-40"
                          >{{ expandedSubmitting ? '…' : 'Save' }}</button>
                        </div>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </template>
            <template v-if="tradedPlayers.length > 0">
              <tr><td colspan="4" class="pt-4 pb-1">
                <button
                  @click="showTraded = !showTraded"
                  class="flex items-center gap-1.5 text-xs font-medium text-text-faint uppercase tracking-wide hover:text-text-muted transition-colors"
                >
                  <span>{{ showTraded ? '▾' : '▸' }}</span>
                  Traded ({{ tradedPlayers.length }})
                </button>
              </td></tr>
              <template v-if="showTraded" v-for="row in tradedPlayers" :key="row.id">
                <tr
                  class="opacity-40"
                  :class="[
                    managing ? 'cursor-pointer hover:opacity-60' : '',
                    expandedId === row.id ? '' : 'border-b border-border-subtle'
                  ]"
                  @click="managing && toggleRow(row)"
                >
                  <td class="py-2 pr-4 font-medium">{{ row.player.name }}</td>
                  <td class="py-2 pr-4 text-xs text-text-muted">{{ row.aflPlayerSeason?.clubSeason?.club?.name ?? '—' }}</td>
                  <td class="py-2">
                    <div class="flex gap-0.5">
                      <span
                        v-for="round in rounds"
                        :key="round.id"
                        class="w-5 text-center text-xs font-mono inline-block"
                        :class="roundColor(row.id, round.id)"
                      >{{ roundLetter(row.id, round.id) }}</span>
                    </div>
                  </td>
                  <td v-if="isMyClub && managing" class="py-2 px-2"></td>
                </tr>
                <tr v-if="expandedId === row.id" class="border-b border-border-subtle opacity-40">
                  <td colspan="4" class="pb-3 pt-1 px-0">
                    <div class="flex gap-4">
                      <div class="flex flex-col gap-2 text-xs shrink-0">
                        <div><div class="text-text-muted">From</div><div class="text-text">{{ roundName(row.fromRoundId) }}</div></div>
                        <div><div class="text-text-muted">To</div><div class="text-text">{{ roundName(row.toRoundId) }}</div></div>
                        <div><div class="text-text-muted">Cost</div><div class="text-text">{{ formatCost(row.costCents) }}</div></div>
                      </div>
                      <div class="flex-1 flex flex-col" @click.stop>
                        <textarea
                          v-model="expandedNotes"
                          rows="3"
                          placeholder="Add notes..."
                          class="flex-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none resize-none"
                        />
                        <div class="flex justify-end mt-2">
                          <button
                            v-if="expandedDirty"
                            @click="saveExpanded"
                            :disabled="expandedSubmitting"
                            class="rounded-lg px-3 py-1.5 text-xs font-medium bg-active text-active-text hover:opacity-90 transition-colors disabled:opacity-40"
                          >{{ expandedSubmitting ? '…' : 'Save' }}</button>
                        </div>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </template>
          </tbody>
        </table>
      </div>
      <p v-else class="text-text-faint">No players on squad.</p>
    </template>

    <Teleport to="body">
      <!-- Add player: search dialog (step 1) -->
      <div v-if="addStep === 'search'" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/60" @click="cancelAddSearch" />
        <div class="relative z-10 w-96 h-[30rem] flex flex-col rounded-xl border border-border bg-surface-raised p-6 shadow-2xl">
          <h3 class="text-base font-semibold text-text mb-4 shrink-0">Add Player</h3>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by name..."
            class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none mb-3 shrink-0"
            autofocus
          />
          <div class="flex-1 overflow-y-auto min-h-0 -mx-1 px-1">
            <div v-if="searchLoading" class="text-text-faint text-sm py-2">Searching...</div>
            <div v-else-if="searchResults.length > 0">
              <div
                v-for="node in searchResults"
                :key="node.id"
                class="flex items-center justify-between border-b border-border-subtle py-2"
              >
                <div>
                  <div class="text-sm text-text">{{ node.player.name }}</div>
                  <div class="text-xs text-text-muted">{{ node.clubSeason.club.name }}</div>
                </div>
                <button
                  @click="selectAddPlayer(node)"
                  class="rounded border border-active px-2 py-0.5 text-xs font-medium text-active hover:bg-active hover:text-active-text transition-colors"
                >Add</button>
              </div>
            </div>
            <div v-else-if="searchQuery.length >= 2 && !searchLoading" class="text-text-faint text-sm py-2">
              No players found.
            </div>
          </div>
          <div class="flex justify-end mt-4 shrink-0">
            <button
              @click="cancelAddSearch"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >Cancel</button>
          </div>
        </div>
      </div>

      <!-- Add player: confirm round dialog (step 2) -->
      <div v-if="addStep === 'confirm' && pendingAddNode" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/60" @click="cancelAddConfirm" />
        <div class="relative z-10 w-80 rounded-xl border border-border bg-surface-raised p-6 shadow-2xl">
          <h3 class="text-base font-semibold text-text mb-1">Add Player</h3>
          <p class="text-sm font-medium text-text mb-0.5">{{ pendingAddNode.player.name }}</p>
          <p class="text-xs text-text-muted mb-4">{{ pendingAddNode.clubSeason.club.name }}</p>
          <label class="text-xs text-text-muted block mb-1">From round</label>
          <select
            v-model="addRoundId"
            class="w-full rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text mb-5 focus:outline-none focus:border-active"
          >
            <option v-for="r in rounds" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
          <div class="flex gap-2 justify-end">
            <button
              @click="cancelAddConfirm"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >Back</button>
            <button
              @click="confirmAdd"
              :disabled="modalSubmitting || !addRoundId"
              class="rounded-lg px-3 py-1.5 text-sm font-medium bg-active text-active-text hover:opacity-90 transition-colors disabled:opacity-40"
            >{{ modalSubmitting ? '…' : 'Add' }}</button>
          </div>
        </div>
      </div>

      <!-- Remove player modal -->
      <div v-if="modal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/60" @click="closeModal" />
        <div class="relative z-10 w-80 rounded-xl border border-border bg-surface-raised p-6 shadow-2xl">
          <h3 class="text-base font-semibold text-text mb-1">Remove Player</h3>
          <p class="text-sm font-medium text-text mb-0.5">{{ modal.playerName }}</p>
          <p class="text-xs text-text-muted mb-4">{{ modal.playerClub }}</p>
          <label class="text-xs text-text-muted block mb-1">Round</label>
          <select
            v-model="modal.roundId"
            class="w-full rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text mb-5 focus:outline-none focus:border-active"
          >
            <option v-for="r in rounds" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
          <div class="flex gap-2 justify-end">
            <button
              @click="closeModal"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >Cancel</button>
            <button
              @click="confirmModal"
              :disabled="modalSubmitting || !modal.roundId"
              class="rounded-lg px-3 py-1.5 text-sm font-medium bg-red-600 text-white hover:bg-red-500 transition-colors disabled:opacity-40"
            >{{ modalSubmitting ? '…' : 'Remove' }}</button>
          </div>
        </div>
      </div>

    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_CLUB_SEASON, GET_AFL_PLAYER_SEASONS, GET_FFL_SEASON_POSITIONS } from '../api/queries'
import { REMOVE_FFL_PLAYER_FROM_SEASON, ADD_FFL_SQUAD_PLAYER, UPDATE_FFL_PLAYER_SEASON } from '../api/mutations'
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
const showTraded = ref(false)

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

const byLastName = (a: { player: { name: string } }, b: { player: { name: string } }) => {
  const lastA = a.player.name.split(' ').pop()?.toLowerCase() ?? ''
  const lastB = b.player.name.split(' ').pop()?.toLowerCase() ?? ''
  return lastA.localeCompare(lastB)
}

const players = computed(() => {
  const nodes = clubSeason.value?.players?.nodes ?? []
  return [...nodes].sort(byLastName)
})

const activePlayers = computed(() => players.value.filter((p: PlayerSeasonRow) => !p.toRoundId))
const tradedPlayers = computed(() => [...players.value.filter((p: PlayerSeasonRow) => !!p.toRoundId)].sort(byLastName))

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
  type P = typeof activePlayers.value[number]
  const buckets = new Map<string | null, P[]>(
    [...POSITION_ORDER.map(p => [p, []] as [string, P[]]), [null, []]]
  )
  for (const p of activePlayers.value) {
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
  GET_AFL_PLAYER_SEASONS,
  () => ({ seasonId: props.seasonId, query: debouncedQuery.value }),
  () => ({ enabled: debouncedQuery.value.length >= 2 })
)

const squadAflPlayerIds = computed(() => {
  const ids = new Set<string>()
  for (const row of players.value) {
    if (row.player.aflPlayerId) ids.add(row.player.aflPlayerId)
  }
  return ids
})

interface AFLPlayerSeasonNode {
  id: string
  player: { id: string; name: string }
  clubSeason: { club: { name: string } }
}

const searchResults = computed((): AFLPlayerSeasonNode[] => {
  const nodes: AFLPlayerSeasonNode[] = searchResult.value?.fflSeason?.aflSeason?.playerSeasons?.nodes ?? []
  return nodes.filter(n => !squadAflPlayerIds.value.has(n.player.id))
})

// Saved flash
const saveMessage = ref('')
let saveMessageTimer: ReturnType<typeof setTimeout> | null = null
function flashSaved() {
  if (saveMessageTimer) clearTimeout(saveMessageTimer)
  saveMessage.value = 'Saved'
  saveMessageTimer = setTimeout(() => { saveMessage.value = '' }, 3000)
}

// Remove modal
interface RemoveModalState { playerSeasonId: string; playerName: string; playerClub: string; roundId: string }
const modal = ref<RemoveModalState | null>(null)
const modalSubmitting = ref(false)

const { mutate: removePlayerMutation } = useMutation(REMOVE_FFL_PLAYER_FROM_SEASON)
const { mutate: addPlayerMutation } = useMutation(ADD_FFL_SQUAD_PLAYER)
const { mutate: updatePlayerSeasonMutation } = useMutation(UPDATE_FFL_PLAYER_SEASON)

function defaultRoundId(): string {
  return liveRoundId.value || (rounds.value.at(-1)?.id ?? '')
}

function openRemoveModal(playerSeasonId: string, playerName: string, playerClub: string) {
  modal.value = { playerSeasonId, playerName, playerClub, roundId: defaultRoundId() }
}

function closeModal() {
  if (!modalSubmitting.value) modal.value = null
}

async function confirmModal() {
  if (!modal.value) return
  modalSubmitting.value = true
  try {
    await removePlayerMutation({ id: modal.value.playerSeasonId, toRoundId: modal.value.roundId })
    await refetchSquad()
    flashSaved()
    modal.value = null
  } finally {
    modalSubmitting.value = false
  }
}

// Add player: two-step flow (search → confirm)
const addStep = ref<null | 'search' | 'confirm'>(null)
const pendingAddNode = ref<AFLPlayerSeasonNode | null>(null)
const addRoundId = ref<string>('')

function openAddSearch() {
  addStep.value = 'search'
  searchQuery.value = ''
  debouncedQuery.value = ''
  addRoundId.value = defaultRoundId()
}

function selectAddPlayer(node: AFLPlayerSeasonNode) {
  pendingAddNode.value = node
  addStep.value = 'confirm'
}

function cancelAddConfirm() {
  addStep.value = 'search'
}

function cancelAddSearch() {
  addStep.value = null
  pendingAddNode.value = null
  searchQuery.value = ''
  debouncedQuery.value = ''
}

async function confirmAdd() {
  if (!pendingAddNode.value || !clubSeasonId.value) return
  modalSubmitting.value = true
  try {
    await addPlayerMutation({
      input: {
        aflPlayerId: pendingAddNode.value.player.id,
        aflPlayerName: pendingAddNode.value.player.name,
        clubSeasonId: clubSeasonId.value,
        aflPlayerSeasonId: pendingAddNode.value.id,
        ...(addRoundId.value ? { fromRoundId: addRoundId.value } : {}),
      },
    })
    await refetchSquad()
    flashSaved()
    addStep.value = null
    pendingAddNode.value = null
  } finally {
    modalSubmitting.value = false
  }
}

// Inline row expansion
interface PlayerSeasonRow {
  id: string
  player: { id: string; name: string; aflPlayerId: string }
  aflPlayerSeason?: { clubSeason?: { club?: { name: string } } } | null
  fromRoundId?: string | null
  toRoundId?: string | null
  notes?: string | null
  costCents?: number | null
}

const expandedId = ref<string | null>(null)
const expandedNotes = ref('')
const expandedSubmitting = ref(false)
const expandedDirty = computed(() => {
  const row = players.value.find((p: PlayerSeasonRow) => p.id === expandedId.value)
  return row != null && expandedNotes.value !== (row.notes ?? '')
})

function roundName(id: string | null | undefined): string {
  if (!id) return '—'
  return rounds.value.find((r: { id: string; name: string }) => r.id === id)?.name ?? '—'
}

function formatCost(cents: number | null | undefined): string {
  if (cents == null) return '—'
  return `$${(cents / 100).toFixed(2)}`
}

function toggleRow(row: PlayerSeasonRow) {
  if (expandedId.value === row.id) {
    expandedId.value = null
    expandedNotes.value = ''
  } else {
    expandedId.value = row.id
    expandedNotes.value = row.notes ?? ''
  }
}

async function saveExpanded() {
  if (!expandedId.value) return
  expandedSubmitting.value = true
  try {
    await updatePlayerSeasonMutation({
      input: {
        id: expandedId.value,
        notes: expandedNotes.value || null,
      },
    })
    await refetchSquad()
    flashSaved()
  } finally {
    expandedSubmitting.value = false
  }
}

watch(managing, (val) => {
  if (!val) {
    cancelAddSearch()
    expandedId.value = null
    expandedNotes.value = ''
  }
})

watch(isMyClub, (val) => {
  if (!val) managing.value = false
})
</script>
