<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="season">
      <div class="mb-6">
        <h1 class="text-2xl font-bold mb-1">{{ selectedClubSeason?.club.name ?? '' }}</h1>
        <p class="text-text-muted">Build your team for the round</p>
      </div>

      <template v-if="selectedClubSeason && clubMatch">
        <div class="mb-6 flex items-center gap-4">
          <button
            @click="onManageToggle"
            class="rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors"
            :class="managing
              ? 'border-active bg-active text-active-text'
              : 'border-border bg-surface text-text hover:bg-surface-hover'"
            :disabled="submitting"
          >
            {{ submitting ? 'Saving…' : managing ? 'Done' : 'Manage' }}
          </button>
          <span v-if="submitMessage" class="text-sm text-green-500">{{ submitMessage }}</span>
        </div>

        <!-- Summary bar -->
        <div class="mb-8 rounded-lg border border-border bg-surface-raised px-4 py-3">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-semibold text-text-heading">Team</h2>
            <span class="text-sm tabular-nums">{{ starterCount }}/18 starters · {{ benchCount }}/4 bench</span>
          </div>
        </div>

        <div class="grid gap-8" :class="managing ? 'grid-cols-1 lg:grid-cols-3' : 'grid-cols-1'">
          <!-- Team (left 2 cols) -->
          <div class="lg:col-span-2">

            <!-- Starter position groups -->
            <div v-for="pos in positions" :key="pos.key" class="mb-6">
              <div class="flex items-center justify-between mb-2">
                <h3 class="text-sm font-semibold uppercase tracking-wider text-text-faint">{{ pos.label }}</h3>
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
                    <span class="font-medium">{{ slot.player.name }}</span>
                  </div>
                  <span v-else class="text-text-faint text-sm">Empty slot</span>
                  <div v-if="slot.player && managing" class="flex items-center gap-2">
                    <button
                      v-for="target in positions.filter(p => p.key !== pos.key)"
                      :key="target.key"
                      class="rounded px-1.5 py-0.5 text-xs text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                      :disabled="isPositionFull(target.key)"
                      :class="{ 'opacity-30 cursor-not-allowed': isPositionFull(target.key) }"
                      :title="`Move to ${target.label}`"
                      @click="moveToPosition(pos.key, index, target.key)"
                    >
                      {{ target.short }}
                    </button>
                    <button
                      class="text-xs text-red-400 hover:text-red-300 transition-colors"
                      @click="removeFromTeam(pos.key, index)"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Bench -->
            <div class="mb-6">
              <h3 class="text-sm font-semibold uppercase tracking-wider text-text-faint mb-2">Bench</h3>

              <!-- Backup Star -->
              <div class="mb-1">
                <div
                  class="flex items-center justify-between rounded-lg border px-4 py-2 transition-colors"
                  :class="benchStarSlot.player
                    ? 'border-border bg-surface-raised'
                    : 'border-dashed border-border-subtle bg-surface'"
                >
                  <div class="flex items-center gap-3 min-w-0">
                    <span class="text-xs text-text-faint shrink-0">★</span>
                    <span v-if="benchStarSlot.player" class="font-medium text-text-muted">{{ benchStarSlot.player.name }}</span>
                    <span v-else class="text-text-faint text-sm">Backup Star</span>
                  </div>
                  <div v-if="managing" class="flex items-center gap-2 ml-2 shrink-0">
                    <label class="flex items-center gap-1 text-xs text-text-faint cursor-pointer">
                      <input
                        type="checkbox"
                        class="accent-active"
                        :checked="interchangePosition === 'star'"
                        :disabled="!benchStarSlot.player"
                        @change="toggleInterchange('star')"
                      />
                      IC
                    </label>
                    <button
                      v-if="benchStarSlot.player"
                      class="text-xs text-red-400 hover:text-red-300 transition-colors"
                      @click="benchStarSlot.player = null"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              </div>

              <!-- Dual-position bench slots -->
              <div
                v-for="(slot, index) in benchDualSlots"
                :key="index"
                class="mb-1"
              >
                <div
                  class="flex items-center justify-between rounded-lg border px-4 py-2 transition-colors"
                  :class="slot.player
                    ? 'border-border bg-surface-raised'
                    : 'border-dashed border-border-subtle bg-surface'"
                >
                  <div class="flex items-center gap-3 min-w-0 flex-1">
                    <span class="text-xs text-text-faint shrink-0">B{{ index + 1 }}</span>
                    <span v-if="slot.player" class="font-medium text-text-muted">{{ slot.player.name }}</span>
                    <span v-else class="text-text-faint text-sm">Empty bench slot</span>
                    <!-- Dual-position selectors (manage mode only) -->
                    <div v-if="slot.player && managing" class="flex items-center gap-1 ml-2">
                      <select
                        class="text-xs rounded bg-control text-text px-1 py-0.5 border border-border"
                        :value="slot.positions[0] ?? ''"
                        @change="setBenchPosition(index, 0, ($event.target as HTMLSelectElement).value)"
                        aria-label="Backup position 1"
                      >
                        <option value="">— pos 1 —</option>
                        <option
                          v-for="pos in nonStarPositions"
                          :key="pos.key"
                          :value="pos.key"
                          :disabled="isBenchPositionUsed(pos.key, index, 0)"
                        >{{ pos.label }}</option>
                      </select>
                      <select
                        class="text-xs rounded bg-control text-text px-1 py-0.5 border border-border"
                        :value="slot.positions[1] ?? ''"
                        @change="setBenchPosition(index, 1, ($event.target as HTMLSelectElement).value)"
                        aria-label="Backup position 2"
                      >
                        <option value="">— pos 2 —</option>
                        <option
                          v-for="pos in nonStarPositions"
                          :key="pos.key"
                          :value="pos.key"
                          :disabled="isBenchPositionUsed(pos.key, index, 1)"
                        >{{ pos.label }}</option>
                      </select>
                    </div>
                    <!-- Read-only position display -->
                    <div v-else-if="slot.player && (slot.positions[0] || slot.positions[1])" class="flex items-center gap-1 ml-2">
                      <span v-if="slot.positions[0]" class="text-xs bg-control rounded px-1.5 py-0.5 text-text-muted">{{ slot.positions[0] }}</span>
                      <span v-if="slot.positions[1]" class="text-xs bg-control rounded px-1.5 py-0.5 text-text-muted">{{ slot.positions[1] }}</span>
                    </div>
                  </div>
                  <div v-if="managing" class="flex items-center gap-2 ml-2 shrink-0">
                    <label v-if="slot.player" class="flex items-center gap-1 text-xs text-text-faint cursor-pointer">
                      <input
                        type="checkbox"
                        class="accent-active"
                        :checked="interchangePosition === benchDualInterchangeKey(index)"
                        :disabled="!slot.positions[0] && !slot.positions[1]"
                        @change="toggleInterchange(benchDualInterchangeKey(index))"
                      />
                      IC
                    </label>
                    <button
                      v-if="slot.player"
                      class="text-xs text-red-400 hover:text-red-300 transition-colors"
                      @click="removeBenchDual(index)"
                    >
                      Remove
                    </button>
                  </div>
                </div>
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
                <span class="font-medium text-sm">{{ player.name }}</span>
                <div class="flex items-center gap-1">
                  <!-- Position buttons (starters) -->
                  <button
                    v-for="pos in positions"
                    :key="pos.key"
                    class="rounded px-2 py-0.5 text-xs text-text-muted hover:bg-control-hover hover:text-text transition-colors"
                    :disabled="isPositionFull(pos.key)"
                    :class="{ 'opacity-30 cursor-not-allowed': isPositionFull(pos.key) }"
                    @click="addToTeam(pos.key, player)"
                  >
                    {{ pos.short }}
                  </button>
                  <!-- Backup star -->
                  <button
                    class="rounded px-2 py-0.5 text-xs text-yellow-400 hover:bg-control-hover hover:text-yellow-300 transition-colors"
                    :disabled="!!benchStarSlot.player"
                    :class="{ 'opacity-30 cursor-not-allowed': !!benchStarSlot.player }"
                    title="Add as Backup Star"
                    @click="addBenchStar(player)"
                  >
                    ★
                  </button>
                  <!-- Dual-position bench -->
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
import { useFflState } from '../composables/useFflState'

const props = defineProps<{ seasonId: string; roundId: string }>()

const positions = [
  { key: 'goals',     label: 'Goals',     short: 'G',  count: 3 },
  { key: 'kicks',     label: 'Kicks',     short: 'K',  count: 4 },
  { key: 'handballs', label: 'Handballs', short: 'HB', count: 4 },
  { key: 'marks',     label: 'Marks',     short: 'M',  count: 2 },
  { key: 'tackles',   label: 'Tackles',   short: 'T',  count: 2 },
  { key: 'hitouts',   label: 'Hitouts',   short: 'HO', count: 2 },
  { key: 'star',      label: 'Star',      short: 'S',  count: 1 },
] as const

type PositionKey = typeof positions[number]['key']
type NonStarPositionKey = Exclude<PositionKey, 'star'>

const nonStarPositions = positions.filter(p => p.key !== 'star')

interface SquadPlayer {
  id: string
  name: string
}

interface Slot {
  player: SquadPlayer | null
}

interface BenchStarSlot {
  player: SquadPlayer | null
}

interface BenchDualSlot {
  player: SquadPlayer | null
  positions: [NonStarPositionKey | null, NonStarPositionKey | null]
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

const clubMatch = computed(() => {
  if (!season.value || !selectedClubSeason.value) return null
  const round = season.value.rounds.find((r: { id: string }) => r.id === props.roundId)
  if (!round) return null
  const clubId = selectedClubSeason.value.club.id
  for (const match of round.matches) {
    if (match.homeClubMatch?.club.id === clubId) return match.homeClubMatch
    if (match.awayClubMatch?.club.id === clubId) return match.awayClubMatch
  }
  return null
})

const squad = computed<SquadPlayer[]>(() => {
  if (!selectedClubSeason.value) return []
  return selectedClubSeason.value.players.nodes.map((r: { id: string; player: { name: string } }) => ({
    id: r.id,
    name: r.player.name,
  }))
})

// ── Team state ──────────────────────────────────────────────────────────────

const createSlots = (count: number): Slot[] => Array.from({ length: count }, () => ({ player: null }))

const teamSlots = ref<Record<PositionKey, Slot[]>>(
  Object.fromEntries(positions.map(p => [p.key, createSlots(p.count)])) as Record<PositionKey, Slot[]>
)

const benchStarSlot = ref<BenchStarSlot>({ player: null })

const benchDualSlots = ref<BenchDualSlot[]>([
  { player: null, positions: [null, null] },
  { player: null, positions: [null, null] },
  { player: null, positions: [null, null] },
])

// Interchange: the position key for which the bench player may freely swap.
// For the star slot this is 'star'; for dual slots it's derived from the first chosen position.
const interchangePosition = ref<string | null>(null)

// Track the match ID we last loaded from to avoid Apollo cache updates wiping local edits.
const initializedMatchId = ref<string | null>(null)

function resetTeamState() {
  for (const pos of positions) {
    teamSlots.value[pos.key] = createSlots(pos.count)
  }
  benchStarSlot.value = { player: null }
  benchDualSlots.value = [
    { player: null, positions: [null, null] },
    { player: null, positions: [null, null] },
    { player: null, positions: [null, null] },
  ]
  interchangePosition.value = null
}

// Load existing team from server data — only when the match changes, not on every Apollo cache update.
// { immediate: true } ensures this fires on component remount when Apollo cache already has data
// (without it, watch only fires on changes — a cache hit on remount produces no change event).
watch(clubMatch, (cm) => {
  if (!cm) return
  if (cm.id === initializedMatchId.value) return  // already initialised for this match; don't reset local edits
  initializedMatchId.value = cm.id

  resetTeamState()
  if (!cm.playerMatches) return

  let dualIndex = 0
  for (const pm of cm.playerMatches) {
    const player: SquadPlayer = { id: pm.playerSeasonId, name: pm.player.name }
    const isBench = pm.backupPositions != null || pm.interchangePosition != null

    if (!isBench) {
      // Starter
      const posSlots = teamSlots.value[pm.position as PositionKey]
      if (posSlots) {
        const slot = posSlots.find((s: Slot) => !s.player)
        if (slot) slot.player = player
      }
    } else if (pm.backupPositions === 'star') {
      // Backup star
      benchStarSlot.value.player = player
      if (pm.interchangePosition) interchangePosition.value = pm.interchangePosition
    } else if (pm.backupPositions && dualIndex < 3) {
      // Dual-position bench
      const parts = pm.backupPositions.split(',').map((p: string) => p.trim()) as NonStarPositionKey[]
      benchDualSlots.value[dualIndex].player = player
      benchDualSlots.value[dualIndex].positions = [parts[0] ?? null, parts[1] ?? null]
      if (pm.interchangePosition) interchangePosition.value = pm.interchangePosition
      dualIndex++
    }
  }
}, { immediate: true })

// ── Computed helpers ──────────────────────────────────────────────────────────

const assignedPlayerIds = computed(() => {
  const ids = new Set<string>()
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (slot.player) ids.add(slot.player.id)
    }
  }
  if (benchStarSlot.value.player) ids.add(benchStarSlot.value.player.id)
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

const benchCount = computed(() => {
  let count = benchStarSlot.value.player ? 1 : 0
  count += benchDualSlots.value.filter(s => s.player).length
  return count
})

const benchDualFull = computed(() => benchDualSlots.value.every(s => s.player !== null))

const isPositionFull = (key: PositionKey) =>
  teamSlots.value[key].every(s => s.player !== null)

// Returns all non-star positions already claimed by another bench dual slot (optionally excluding a specific slot+side).
function isBenchPositionUsed(posKey: string, slotIndex: number, sideIndex: number): boolean {
  for (let i = 0; i < benchDualSlots.value.length; i++) {
    const slot = benchDualSlots.value[i]
    for (let j = 0; j < 2; j++) {
      if (i === slotIndex && j === sideIndex) continue
      if (slot.positions[j as 0 | 1] === posKey) return true
    }
  }
  return false
}

// Derive an interchange key for a dual bench slot from its assigned positions.
function benchDualInterchangeKey(index: number): string | null {
  const slot = benchDualSlots.value[index]
  return slot.positions[0] ?? slot.positions[1] ?? null
}

// ── Team management ─────────────────────────────────────────────────────────

function addToTeam(key: PositionKey, player: SquadPlayer) {
  const slot = teamSlots.value[key].find(s => !s.player)
  if (slot) slot.player = player
}

function removeFromTeam(key: PositionKey, index: number) {
  teamSlots.value[key][index].player = null
}

function moveToPosition(fromKey: PositionKey, fromIndex: number, toKey: PositionKey) {
  const player = teamSlots.value[fromKey][fromIndex].player
  if (!player) return
  const toSlot = teamSlots.value[toKey].find(s => !s.player)
  if (!toSlot) return
  teamSlots.value[fromKey][fromIndex].player = null
  toSlot.player = player
}

function addBenchStar(player: SquadPlayer) {
  if (benchStarSlot.value.player) return
  benchStarSlot.value.player = player
}

function addBenchDual(player: SquadPlayer) {
  const slot = benchDualSlots.value.find(s => !s.player)
  if (slot) slot.player = player
}

function removeBenchDual(index: number) {
  benchDualSlots.value[index].player = null
  benchDualSlots.value[index].positions = [null, null]
  // Clear interchange if it pointed to this slot
  const key = benchDualInterchangeKey(index)
  if (key && interchangePosition.value === key) interchangePosition.value = null
}

function setBenchPosition(slotIndex: number, sideIndex: 0 | 1, value: string) {
  benchDualSlots.value[slotIndex].positions[sideIndex] = (value || null) as NonStarPositionKey | null
}

function toggleInterchange(key: string | null) {
  if (!key) return
  interchangePosition.value = interchangePosition.value === key ? null : key
}

// ── Submit ────────────────────────────────────────────────────────────────────

const { mutate: setTeam } = useMutation(SET_FFL_TEAM, () => ({
  refetchQueries: [{ query: GET_FFL_TEAM_BUILDER, variables: { seasonId: props.seasonId } }],
  awaitRefetchQueries: true,
}))
const submitting = ref(false)
const submitMessage = ref('')

async function onManageToggle() {
  if (managing.value) {
    await submitTeam()
  }
  managing.value = !managing.value
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

  // Backup star
  if (benchStarSlot.value.player) {
    const entry: (typeof players)[number] = {
      playerSeasonId: benchStarSlot.value.player.id,
      position: 'star',
      backupPositions: 'star',
    }
    if (interchangePosition.value === 'star') entry.interchangePosition = 'star'
    players.push(entry)
  }

  // Dual-position bench
  for (const slot of benchDualSlots.value) {
    if (!slot.player) continue
    const [p1, p2] = slot.positions
    const bp = [p1, p2].filter(Boolean).join(',')
    const entry: (typeof players)[number] = {
      playerSeasonId: slot.player.id,
      position: p1 ?? p2 ?? 'goals',
      backupPositions: bp || undefined,
    }
    // Attach interchange if this slot's first position matches the interchange position
    const icKey = benchDualInterchangeKey(benchDualSlots.value.indexOf(slot))
    if (icKey && interchangePosition.value === icKey) {
      entry.interchangePosition = interchangePosition.value
    }
    players.push(entry)
  }

  try {
    await setTeam({ input: { clubMatchId: clubMatch.value.id, players } })
    submitMessage.value = 'Team saved!'
    setTimeout(() => { submitMessage.value = '' }, 3000)
  } catch (e) {
    submitMessage.value = 'Failed to save team'
  } finally {
    submitting.value = false
  }
}
</script>
