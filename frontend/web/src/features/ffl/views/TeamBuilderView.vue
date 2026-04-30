<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="season">
      <div class="mb-6">
        <Breadcrumb v-if="currentRound" :items="breadcrumbs" />
        <div class="flex items-center">
          <h1 class="text-2xl font-bold flex items-center gap-3">
            <img v-if="selectedClubSeason" :src="clubLogoUrl(selectedClubSeason.club.name)" :alt="selectedClubSeason.club.name" class="w-10 h-10 object-contain" />
            {{ selectedClubSeason?.club.name ?? '' }}<span class="font-normal text-text-muted"> · Team Builder</span>
          </h1>
          <div class="flex items-center gap-3 ml-auto">
            <div class="flex items-center gap-1 rounded-lg border border-border px-1">
              <router-link
                v-if="prevRound"
                :to="{ name: 'ffl-team-builder', params: { seasonId: props.seasonId, roundId: prevRound.id } }"
                class="w-6 h-6 flex items-center justify-center rounded text-text-muted hover:bg-control-hover hover:text-text transition-colors text-sm"
                title="Previous round"
              >‹</router-link>
              <span v-else class="w-6 h-6 flex items-center justify-center text-text-faint text-sm opacity-30">‹</span>
              <span class="text-sm text-text-muted tabular-nums">{{ currentRound?.name }}</span>
              <router-link
                v-if="nextRound"
                :to="{ name: 'ffl-team-builder', params: { seasonId: props.seasonId, roundId: nextRound.id } }"
                class="w-6 h-6 flex items-center justify-center rounded text-text-muted hover:bg-control-hover hover:text-text transition-colors text-sm"
                title="Next round"
              >›</router-link>
              <span v-else class="w-6 h-6 flex items-center justify-center text-text-faint text-sm opacity-30">›</span>
            </div>
            <router-link
              v-if="selectedClubSeason"
              :to="{ name: 'ffl-squad', params: { seasonId: props.seasonId, clubId: selectedClubSeason.club.id } }"
              class="flex items-center gap-1.5 text-sm text-text-muted hover:text-text transition-colors"
            >
              <IconSquad class="w-4 h-4" />
              Squad
            </router-link>
          </div>
        </div>
      </div>

      <template v-if="selectedClubSeason && clubMatch">
        <div class="mb-6 flex items-center gap-4">
          <template v-if="managing">
            <button
              @click="cancelManage"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
              :disabled="submitting"
            >
              Cancel
            </button>
            <button
              @click="onSaveTeam"
              class="rounded-lg border border-active bg-active px-3 py-1.5 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              :disabled="submitting || !isDirty || !!benchValidationError"
            >
              {{ submitting ? 'Saving...' : 'Save Team' }}
            </button>
          </template>
          <button
            v-else
            @click="managing = true"
            class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
          >
            <span class="flex items-center gap-1.5">
              <IconManage class="w-3.5 h-3.5" />
              Manage
            </span>
          </button>
          <span v-if="benchValidationError" class="text-sm text-red-400">{{ benchValidationError }}</span>
          <span v-else-if="submitMessage" class="text-sm text-green-500">{{ submitMessage }}</span>
        </div>

        <!-- Summary bar -->
        <div class="mb-8 rounded-lg border border-border bg-surface-raised px-4 py-3">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-semibold text-text-heading">Team</h2>
            <div class="flex items-center gap-3">
              <span class="text-sm tabular-nums text-text-muted">{{ starterCount }}/18 starters · {{ benchCount }}/4 bench</span>
              <span class="text-sm font-semibold tabular-nums">{{ grandTotal }}</span>
            </div>
          </div>
        </div>

        <div class="grid gap-8" :class="managing ? 'grid-cols-1 sm:grid-cols-2' : 'grid-cols-1'">
          <!-- Team (left col) -->
          <div>

            <!-- Starter position groups -->
            <div v-for="pos in positions" :key="pos.key" class="mb-6">
              <div class="flex items-center justify-between mb-2">
                <h3 class="text-sm font-semibold text-text-faint">
                  {{ pos.label }}<span v-if="positionTotal(pos.key) > 0" class="font-normal ml-3">({{ positionTotal(pos.key) }})</span>
                </h3>
              </div>
              <div class="space-y-1">
                <div
                  v-for="(slot, index) in teamSlots[pos.key]"
                  :key="index"
                  class="flex items-center justify-between rounded-lg border px-4 py-2 transition-colors"
                  :class="slot.player
                    ? 'border-border bg-surface-raised'
                    : 'border-dashed border-border-subtle bg-surface'"
                >
                  <div v-if="slot.player" class="flex items-center gap-3">
                    <span v-if="pos.key === 'star'" class="text-yellow-400 text-xs">★</span>
                    <div v-if="managing">
                      <div class="font-medium text-sm">{{ slot.player.name }}</div>
                      <div v-if="slot.player.club" class="text-xs text-text-muted">{{ slot.player.club }}</div>
                    </div>
                    <div v-else class="flex items-baseline gap-2">
                      <span class="font-medium text-sm">{{ slot.player.name }}</span>
                      <span v-if="slot.player.club" class="text-xs text-text-muted">{{ slot.player.club }}</span>
                    </div>
                  </div>
                  <span v-else class="text-text-faint text-sm">Empty slot</span>
                  <div v-if="slot.player && managing" class="flex items-center gap-2">
                    <button
                      v-for="target in positions.filter(p => p.key !== pos.key)"
                      :key="target.key"
                      class="rounded px-1.5 py-0.5 text-xs transition-colors"
                      :disabled="isPositionFull(target.key)"
                      :class="[
                        isPositionFull(target.key) ? 'opacity-30 cursor-not-allowed' : '',
                        target.key === 'star' ? 'text-yellow-400 hover:bg-control-hover hover:text-yellow-300' : 'text-text-faint hover:bg-control-hover hover:text-text'
                      ]"
                      :title="`Move to ${target.label}`"
                      @click="moveToPosition(pos.key, index, target.key)"
                    >
                      {{ target.short }}
                    </button>
                    <button
                      aria-label="Remove"
                      class="text-xs text-red-400 hover:text-red-300 transition-colors"
                      @click="removeFromTeam(pos.key, index)"
                    >
                      <IconBin class="w-3.5 h-3.5" />
                    </button>
                  </div>
                  <div v-else-if="slot.player" class="flex items-center shrink-0 w-44">
                    <span class="w-16 shrink-0">
                      <StatusBadge v-if="slot.player.status" :status="slot.player.status" />
                    </span>
                    <template v-if="slot.player.status === 'played'">
                      <span class="flex-1 text-right text-xs tabular-nums text-text-faint pr-2">{{ positionFormula(pos.key, slot.player.score ?? 0) ?? '' }}</span>
                      <span class="w-8 text-right text-sm tabular-nums text-text shrink-0">{{ slot.player.score }}</span>
                    </template>
                  </div>
                </div>
              </div>
            </div>

            <!-- Bench -->
            <div class="mb-6">
              <h3 class="text-sm font-semibold text-text-faint mb-2">Bench</h3>

              <div v-for="(slot, index) in benchDualSlots" :key="index" class="mb-1">
                <div
                  class="flex items-center justify-between rounded-lg border px-4 py-2 transition-colors"
                  :class="[
                    slot.player ? 'border-border bg-surface-raised' : 'border-dashed border-border-subtle bg-surface',
                    recentlyClearedSlot === index ? '!border-orange-400' : ''
                  ]"
                >
                  <!-- Left: name -->
                  <div class="flex items-center gap-3 min-w-0">
                    <div v-if="slot.player" class="flex items-baseline gap-2" :class="managing ? 'flex-col gap-0' : ''">
                      <span class="font-medium text-sm text-text-muted">{{ slot.player.name }}</span>
                      <span v-if="slot.player.club" class="text-xs text-text-muted">{{ slot.player.club }}</span>
                    </div>
                    <span v-else class="text-text-faint text-sm">Empty slot</span>
                  </div>
                  <!-- Right: selectors + remove (manage) or read-only tags -->
                  <div class="flex items-center gap-2 ml-4 shrink-0">
                    <template v-if="slot.player && managing">
                      <select
                        class="text-xs rounded bg-control text-text px-1 py-0.5 border border-border"
                        :value="slot.positions[0] ?? ''"
                        @change="setBenchPosition(index, 0, ($event.target as HTMLSelectElement).value)"
                        aria-label="Position 1"
                      >
                        <option value=""></option>
                        <option v-for="pos in positions" :key="pos.key" :value="pos.key">
                          {{ pos.short }}{{ isBenchPositionUsed(pos.key, index, 0) ? ' ·' : '' }}
                        </option>
                      </select>
                      <select
                        v-if="slot.positions[0] !== 'star'"
                        class="text-xs rounded bg-control text-text px-1 py-0.5 border border-border"
                        :value="slot.positions[1] ?? ''"
                        @change="setBenchPosition(index, 1, ($event.target as HTMLSelectElement).value)"
                        aria-label="Position 2"
                      >
                        <option value=""></option>
                        <option v-for="pos in nonStarPositions" :key="pos.key" :value="pos.key">
                          {{ pos.short }}{{ isBenchPositionUsed(pos.key, index, 1) ? ' ·' : '' }}
                        </option>
                      </select>
                      <button
                        aria-label="Remove"
                        class="text-xs text-red-400 hover:text-red-300 transition-colors"
                        @click="removeBenchDual(index)"
                      >
                        <svg class="w-3.5 h-3.5" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
                          <path d="M2 3.5h10M5.5 3.5V2.5a.5.5 0 01.5-.5h2a.5.5 0 01.5.5v1M6 6.5v4M8 6.5v4M3 3.5l.7 7.5a.5.5 0 00.5.5h5.6a.5.5 0 00.5-.5L11 3.5"/>
                        </svg>
                      </button>
                    </template>
                    <template v-else-if="slot.player">
                      <div class="flex items-center w-44 shrink-0">
                        <span class="w-16 shrink-0">
                          <StatusBadge v-if="slot.player.status" :status="slot.player.status" />
                        </span>
                        <div class="flex items-center gap-1 flex-1 justify-end">
                          <template v-if="slot.positions[0]">
                            <span class="text-xs bg-control rounded px-1.5 py-0.5 text-text-muted">
                              {{ positionShort(slot.positions[0]) }}<template v-if="interchangePosition === slot.positions[0]"> · Int</template>
                            </span>
                          </template>
                          <template v-if="slot.positions[1]">
                            <span class="text-xs bg-control rounded px-1.5 py-0.5 text-text-muted">
                              {{ positionShort(slot.positions[1]) }}<template v-if="interchangePosition === slot.positions[1]"> · Int</template>
                            </span>
                          </template>
                          <span v-if="slot.player.status === 'played'" class="w-8 text-right text-sm tabular-nums text-text shrink-0">{{ slot.player.score }}</span>
                        </div>
                      </div>
                    </template>
                  </div>
                </div>
              </div>

              <!-- Interchange -->
              <div v-if="managing" class="mt-3 flex items-center gap-2 justify-end">
                <span class="text-xs text-text-faint">Interchange</span>
                <select
                  class="text-xs rounded bg-control text-text px-1 py-0.5 border border-border"
                  aria-label="Interchange"
                  :value="interchangePosition ?? ''"
                  @change="setInterchange(($event.target as HTMLSelectElement).value)"
                >
                  <option value=""></option>
                  <option v-for="pos in positions" :key="pos.key" :value="pos.key">{{ pos.short }}</option>
                </select>
              </div>
            </div>

          </div>

          <!-- Squad panel (right col, manage mode only) -->
          <div v-if="managing">
            <h2 class="text-lg font-semibold text-text-heading mb-3">Squad ({{ availablePlayers.length }})</h2>
            <div class="space-y-1">
              <div
                v-for="player in availablePlayers"
                :key="player.id"
                class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-2"
              >
                <div>
                  <div class="font-medium text-sm">{{ player.name }}</div>
                  <div v-if="player.club" class="text-xs text-text-muted">{{ player.club }}</div>
                </div>
                <div class="flex items-center gap-1">
                  <!-- Position buttons (starters) -->
                  <button
                    v-for="pos in positions"
                    :key="pos.key"
                    class="rounded px-2 py-0.5 text-xs transition-colors"
                    :class="[
                      isPositionFull(pos.key) ? 'opacity-30 cursor-not-allowed' : '',
                      pos.key === 'star' ? 'text-yellow-400 hover:bg-control-hover hover:text-yellow-300' : 'text-text-muted hover:bg-control-hover hover:text-text'
                    ]"
                    :disabled="isPositionFull(pos.key)"
                    @click="addToTeam(pos.key, player)"
                  >
                    {{ pos.short }}
                  </button>
                  <span class="w-px h-4 bg-border mx-0.5 shrink-0"></span>
                  <!-- Bench -->
                  <button
                    class="rounded px-2 py-0.5 text-xs text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                    :disabled="benchDualFull"
                    :class="{ 'opacity-30 cursor-not-allowed': benchDualFull }"
                    title="Add to bench"
                    @click="addBenchDual(player)"
                  >
                    B
                  </button>
                </div>
              </div>
              <p v-if="availablePlayers.length === 0" class="text-sm text-text-faint">All players assigned</p>
            </div>
          </div>
        </div>
      </template>
      <p v-else class="text-text-faint">No club selected. Choose a club in the nav bar.</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_TEAM_BUILDER } from '../api/queries'
import { SET_FFL_TEAM } from '../api/mutations'
import Breadcrumb from '../components/Breadcrumb.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { positionFormula } from '../utils/position'
import IconSquad from '../components/icons/IconSquad.vue'
import IconManage from '../components/icons/IconManage.vue'
import IconBin from '../components/icons/IconBin.vue'
import { useFflState } from '../composables/useFflState'

const props = defineProps<{ seasonId: string; roundId: string }>()

const positions = [
  { key: 'goals',     label: 'Goals',     short: 'G',  count: 3 },
  { key: 'kicks',     label: 'Kicks',     short: 'K',  count: 4 },
  { key: 'handballs', label: 'Handballs', short: 'H',  count: 4 },
  { key: 'marks',     label: 'Marks',     short: 'M',  count: 2 },
  { key: 'tackles',   label: 'Tackles',   short: 'T',  count: 2 },
  { key: 'hitouts',   label: 'Hitouts',   short: 'R',  count: 2 },
  { key: 'star',      label: 'Star',      short: '★',  count: 1 },
] as const

type PositionKey = typeof positions[number]['key']
type NonStarPositionKey = Exclude<PositionKey, 'star'>

const nonStarPositions = positions.filter(p => p.key !== 'star')

interface SquadPlayer {
  id: string
  name: string
  club: string | null
  score: number | null
  status: string | null
}

interface Slot {
  player: SquadPlayer | null
}

interface BenchDualSlot {
  player: SquadPlayer | null
  positions: [PositionKey | null, NonStarPositionKey | null]
}

const { selectedClubId } = useFflState()
const managing = ref(false)

// Data loading
const { result, loading, error } = useQuery(
  GET_FFL_TEAM_BUILDER,
  () => ({ seasonId: props.seasonId }),
  { errorPolicy: 'all' },
)

const season = computed(() => result.value?.fflSeason ?? null)

const selectedClubSeason = computed(() =>
  season.value?.ladder.find((cs: { club: { id: string } }) => cs.club.id === selectedClubId.value) ?? null
)

const currentRound = computed(() =>
  season.value?.rounds.find((r: { id: string }) => r.id === props.roundId) ?? null
)

const breadcrumbs = computed(() => {
  if (!season.value || !currentRound.value) return []
  const crumbs: { label: string; to?: object }[] = [
    { label: 'FFL' },
    { label: season.value.name, to: { name: 'home' } },
    { label: currentRound.value.name, to: { name: 'ffl-round', params: { seasonId: props.seasonId, roundId: props.roundId } } },
  ]
  if (currentMatch.value) {
    const home = currentMatch.value.homeClubMatch?.club.name ?? '?'
    const away = currentMatch.value.awayClubMatch?.club.name ?? '?'
    crumbs.push({ label: `${home} v ${away}`, to: { name: 'ffl-match', params: { seasonId: props.seasonId, matchId: currentMatch.value.id } } })
  }
  return crumbs
})

const prevRound = computed(() => {
  const rounds = season.value?.rounds ?? []
  const idx = rounds.findIndex((r: { id: string }) => r.id === props.roundId)
  return idx > 0 ? rounds[idx - 1] : null
})

const nextRound = computed(() => {
  const rounds = season.value?.rounds ?? []
  const idx = rounds.findIndex((r: { id: string }) => r.id === props.roundId)
  return idx >= 0 && idx < rounds.length - 1 ? rounds[idx + 1] : null
})

const currentMatch = computed(() => {
  if (!season.value || !selectedClubSeason.value) return null
  const round = season.value.rounds.find((r: { id: string }) => r.id === props.roundId)
  if (!round) return null
  const clubId = selectedClubSeason.value.club.id
  return round.matches.find((m: { homeClubMatch?: { club: { id: string } } | null; awayClubMatch?: { club: { id: string } } | null }) =>
    m.homeClubMatch?.club.id === clubId || m.awayClubMatch?.club.id === clubId
  ) ?? null
})

const clubMatch = computed(() => {
  if (!currentMatch.value || !selectedClubSeason.value) return null
  const clubId = selectedClubSeason.value.club.id
  const m = currentMatch.value
  if (m.homeClubMatch?.club.id === clubId) return m.homeClubMatch
  if (m.awayClubMatch?.club.id === clubId) return m.awayClubMatch
  return null
})

const squad = computed<SquadPlayer[]>(() => {
  if (!selectedClubSeason.value) return []
  return selectedClubSeason.value.players.nodes.map((r: { id: string; player: { name: string }; aflPlayerSeason?: { clubSeason?: { club?: { name: string } } } }) => ({
    id: r.id,
    name: r.player.name,
    club: r.aflPlayerSeason?.clubSeason?.club?.name ?? null,
    score: null,
    status: null,
  }))
})

// ── Team state ──────────────────────────────────────────────────────────────

const createSlots = (count: number): Slot[] => Array.from({ length: count }, () => ({ player: null }))

const teamSlots = ref<Record<PositionKey, Slot[]>>(
  Object.fromEntries(positions.map(p => [p.key, createSlots(p.count)])) as Record<PositionKey, Slot[]>
)

const benchDualSlots = ref<BenchDualSlot[]>([
  { player: null, positions: [null, null] },
  { player: null, positions: [null, null] },
  { player: null, positions: [null, null] },
  { player: null, positions: [null, null] },
])

// The position that acts as the free interchange slot.
const interchangePosition = ref<string | null>(null)

// Highlight recently-stolen bench slot index (orange border flash).
const recentlyClearedSlot = ref<number | null>(null)
let clearHighlightTimer: ReturnType<typeof setTimeout> | null = null

// Track the match ID we last loaded from to avoid Apollo cache updates wiping local edits.
const initializedMatchId = ref<string | null>(null)

// Dirty tracking — snapshot taken after load or save; compared to detect unsaved changes.
const isDirty = ref(false)

function takeSnapshot() {
  isDirty.value = false
}

function markDirty() {
  isDirty.value = true
}

function resetTeamState() {
  for (const pos of positions) {
    teamSlots.value[pos.key] = createSlots(pos.count)
  }
  benchDualSlots.value = [
    { player: null, positions: [null, null] },
    { player: null, positions: [null, null] },
    { player: null, positions: [null, null] },
    { player: null, positions: [null, null] },
  ]
  interchangePosition.value = null
}

function loadTeamFromMatch(cm: NonNullable<typeof clubMatch.value>) {
  resetTeamState()
  takeSnapshot()
  if (!cm.playerMatches) return

  let dualIndex = 0
  for (const pm of cm.playerMatches) {
    const player: SquadPlayer = { id: pm.playerSeasonId, name: pm.player.name, club: squad.value.find(s => s.id === pm.playerSeasonId)?.club ?? null, score: pm.score ?? null, status: pm.status ?? null }
    const isBench = pm.backupPositions != null || pm.interchangePosition != null

    if (!isBench) {
      const posSlots = teamSlots.value[pm.position as PositionKey]
      if (posSlots) {
        const slot = posSlots.find((s: Slot) => !s.player)
        if (slot) slot.player = player
      }
    } else if (dualIndex < 4) {
      if (pm.backupPositions === 'star') {
        benchDualSlots.value[dualIndex].player = player
        benchDualSlots.value[dualIndex].positions = ['star', null]
      } else if (pm.backupPositions) {
        const parts = pm.backupPositions.split(',').map((p: string) => p.trim()) as NonStarPositionKey[]
        benchDualSlots.value[dualIndex].player = player
        benchDualSlots.value[dualIndex].positions = [parts[0] ?? null, parts[1] ?? null]
      }
      if (pm.interchangePosition) interchangePosition.value = pm.interchangePosition
      dualIndex++
    }
  }
}

// Load existing team from server data — only when the match changes, not on every Apollo cache update.
// { immediate: true } ensures this fires on component remount when Apollo cache already has data
// (without it, watch only fires on changes — a cache hit on remount produces no change event).
watch(clubMatch, (cm) => {
  if (!cm) return
  if (cm.id === initializedMatchId.value) return  // already initialised for this match; don't reset local edits
  initializedMatchId.value = cm.id
  loadTeamFromMatch(cm)
}, { immediate: true })

// ── Computed helpers ──────────────────────────────────────────────────────────

const assignedPlayerIds = computed(() => {
  const ids = new Set<string>()
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (slot.player) ids.add(slot.player.id)
    }
  }
  for (const slot of benchDualSlots.value) {
    if (slot.player) ids.add(slot.player.id)
  }
  return ids
})

const availablePlayers = computed(() =>
  squad.value.filter(p => !assignedPlayerIds.value.has(p.id))
)

const starterCount = computed(() => {
  let count = 0
  for (const pos of positions) {
    count += teamSlots.value[pos.key].filter(s => s.player).length
  }
  return count
})

const benchCount = computed(() => benchDualSlots.value.filter(s => s.player).length)

const positionTotal = (key: PositionKey): number =>
  teamSlots.value[key].reduce((sum: number, s: Slot) => sum + (s.player?.score ?? 0), 0)

const grandTotal = computed(() => clubMatch.value?.score ?? 0)

const benchDualFull = computed(() => benchDualSlots.value.every(s => s.player !== null))

const benchValidationError = computed<string | null>(() => {
  for (const slot of benchDualSlots.value) {
    if (!slot.player) continue
    const [p1, p2] = slot.positions
    if (!p1) return 'Each bench player must have a position assigned'
    if (p1 !== 'star' && !p2) return 'Non-star bench players need two backup positions'
  }
  const filledCount = benchDualSlots.value.filter(s => s.player).length
  if (filledCount > 1 && !interchangePosition.value) return 'Choose an interchange position'
  return null
})

const isPositionFull = (key: PositionKey) =>
  teamSlots.value[key].every(s => s.player !== null)

// Returns true if posKey is already used by another bench slot (excluding slotIndex+sideIndex).
function isBenchPositionUsed(posKey: string, slotIndex: number, sideIndex: number): boolean {
  for (let i = 0; i < benchDualSlots.value.length; i++) {
    const slot = benchDualSlots.value[i]
    for (const j of [0, 1] as const) {
      if (i === slotIndex && j === sideIndex) continue
      if (slot.positions[j] === posKey) return true
    }
  }
  return false
}

function positionShort(key: string): string {
  return positions.find(p => p.key === key)?.short ?? key
}

// ── Team management ─────────────────────────────────────────────────────────

function addToTeam(key: PositionKey, player: SquadPlayer) {
  const slot = teamSlots.value[key].find(s => !s.player)
  if (slot) { slot.player = player; markDirty() }
}

function removeFromTeam(key: PositionKey, index: number) {
  teamSlots.value[key][index].player = null
  markDirty()
}

function moveToPosition(fromKey: PositionKey, fromIndex: number, toKey: PositionKey) {
  const player = teamSlots.value[fromKey][fromIndex].player
  if (!player) return
  const toSlot = teamSlots.value[toKey].find(s => !s.player)
  if (!toSlot) return
  teamSlots.value[fromKey][fromIndex].player = null
  toSlot.player = player
  markDirty()
}

function addBenchDual(player: SquadPlayer) {
  const slot = benchDualSlots.value.find(s => !s.player)
  if (slot) { slot.player = player; markDirty() }
}

function removeBenchDual(index: number) {
  benchDualSlots.value[index].player = null
  benchDualSlots.value[index].positions = [null, null]
  markDirty()
}

function setBenchPosition(slotIndex: number, sideIndex: 0 | 1, value: string) {
  const slot = benchDualSlots.value[slotIndex]
  // Steal position from any other slot that already has it, and flash that slot
  if (value) {
    for (let i = 0; i < benchDualSlots.value.length; i++) {
      const other = benchDualSlots.value[i]
      if (other.positions[0] === value && !(i === slotIndex && sideIndex === 0)) {
        other.positions[0] = null
        flashClearedSlot(i)
      } else if (other.positions[1] === value && !(i === slotIndex && sideIndex === 1)) {
        other.positions[1] = null
        flashClearedSlot(i)
      }
    }
  }
  if (sideIndex === 0) {
    slot.positions[0] = (value || null) as PositionKey | null
    if (value === 'star') slot.positions[1] = null
  } else {
    slot.positions[1] = (value || null) as NonStarPositionKey | null
  }
  markDirty()
}

function flashClearedSlot(index: number) {
  if (clearHighlightTimer) clearTimeout(clearHighlightTimer)
  recentlyClearedSlot.value = index
  clearHighlightTimer = setTimeout(() => { recentlyClearedSlot.value = null }, 2000)
}

function setInterchange(value: string) {
  interchangePosition.value = value || null
  markDirty()
}

// ── Submit ────────────────────────────────────────────────────────────────────

const { mutate: setTeam } = useMutation(SET_FFL_TEAM, () => ({
  refetchQueries: [{ query: GET_FFL_TEAM_BUILDER, variables: { seasonId: props.seasonId } }],
  awaitRefetchQueries: true,
}))
const submitting = ref(false)
const submitMessage = ref('')

async function onSaveTeam() {
  await submitTeam()
  managing.value = false
}

function cancelManage() {
  if (clubMatch.value) loadTeamFromMatch(clubMatch.value)
  managing.value = false
}

async function submitTeam() {
  if (!clubMatch.value) return
  submitting.value = true
  submitMessage.value = ''

  const players: {
    playerSeasonId: string
    position: string
    backupPositions?: string
    interchangePosition?: string
  }[] = []

  // Starters
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (slot.player) {
        players.push({ playerSeasonId: slot.player.id, position: pos.key })
      }
    }
  }

  // Bench slots
  for (const slot of benchDualSlots.value) {
    if (!slot.player) continue
    const [p1, p2] = slot.positions
    const isStar = p1 === 'star'
    const bp = isStar ? 'star' : [p1, p2].filter(Boolean).join(',')
    const entry: (typeof players)[number] = {
      playerSeasonId: slot.player.id,
      position: p1 ?? p2 ?? 'goals',
      backupPositions: bp || undefined,
    }
    if (interchangePosition.value && (p1 === interchangePosition.value || p2 === interchangePosition.value)) {
      entry.interchangePosition = interchangePosition.value ?? undefined
    }
    players.push(entry)
  }

  try {
    await setTeam({ input: { clubMatchId: clubMatch.value.id, players } })
    takeSnapshot()
    submitMessage.value = 'Saved'
    setTimeout(() => { submitMessage.value = '' }, 3000)
  } catch (e) {
    submitMessage.value = 'Failed to save team'
  } finally {
    submitting.value = false
  }
}
</script>
