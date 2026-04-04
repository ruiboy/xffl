<template>
  <div>
    <h1 class="text-2xl font-bold mb-1">Team Builder</h1>
    <p class="text-text-muted mb-6">Build your lineup for the round</p>

    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="season">
      <!-- Club selector -->
      <div class="mb-6">
        <label class="text-sm font-medium text-text-muted mr-2">My club:</label>
        <select
          v-model="selectedClubSeasonId"
          class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text focus:border-active focus:outline-none"
        >
          <option v-for="cs in season.ladder" :key="cs.id" :value="cs.id">
            {{ cs.club.name }}
          </option>
        </select>
      </div>

      <template v-if="selectedClubSeason && clubMatch">
        <!-- Score projection -->
        <div class="mb-8 rounded-lg border border-border bg-surface-raised px-4 py-3">
          <div class="flex items-center justify-between mb-2">
            <h2 class="text-sm font-semibold text-text-heading">Lineup</h2>
            <span class="text-sm tabular-nums">{{ starterCount }}/22 starters, {{ benchCount }}/8 bench</span>
          </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <!-- Lineup (left 2 cols) -->
          <div class="lg:col-span-2">
            <div v-for="pos in positions" :key="pos.key" class="mb-6">
              <h3 class="text-sm font-semibold uppercase tracking-wider text-text-faint mb-2">{{ pos.label }}</h3>
              <div class="space-y-1">
                <div
                  v-for="(slot, index) in lineupSlots[pos.key]"
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
                  <div v-if="slot.player" class="flex items-center gap-2">
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
                      @click="removeFromLineup(pos.key, index)"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div class="mb-6">
              <h3 class="text-sm font-semibold uppercase tracking-wider text-text-faint mb-2">Bench ({{ benchCount }}/8)</h3>
              <div class="space-y-1">
                <div
                  v-for="(slot, index) in benchSlots"
                  :key="index"
                  class="flex items-center justify-between rounded-lg border px-4 py-2 transition-colors"
                  :class="slot.player
                    ? 'border-border bg-surface-raised'
                    : 'border-dashed border-border-subtle bg-surface'"
                >
                  <div v-if="slot.player" class="flex items-center gap-3">
                    <span class="font-medium text-text-muted">{{ slot.player.name }}</span>
                  </div>
                  <span v-else class="text-text-faint text-sm">Empty bench slot</span>
                  <button
                    v-if="slot.player"
                    class="text-xs text-red-400 hover:text-red-300 transition-colors"
                    @click="removeFromBench(index)"
                  >
                    Remove
                  </button>
                </div>
              </div>
            </div>

            <!-- Submit -->
            <button
              class="rounded-lg bg-active px-6 py-2 text-sm font-medium text-active-text hover:opacity-90 transition-opacity disabled:opacity-30 disabled:cursor-not-allowed"
              :disabled="submitting || starterCount === 0"
              @click="submitLineup"
            >
              {{ submitting ? 'Saving…' : 'Save Lineup' }}
            </button>
            <span v-if="submitMessage" class="ml-3 text-sm text-green-500">{{ submitMessage }}</span>
          </div>

          <!-- Squad panel (right col) -->
          <div>
            <h2 class="text-lg font-semibold text-text-heading mb-3">Squad ({{ availablePlayers.length }})</h2>
            <div class="space-y-1">
              <div
                v-for="player in availablePlayers"
                :key="player.id"
                class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-2"
              >
                <span class="font-medium text-sm">{{ player.name }}</span>
                <div class="flex items-center gap-1">
                  <button
                    v-for="pos in positions"
                    :key="pos.key"
                    class="rounded px-2 py-0.5 text-xs text-text-muted hover:bg-control-hover hover:text-text transition-colors"
                    :disabled="isPositionFull(pos.key)"
                    :class="{ 'opacity-30 cursor-not-allowed': isPositionFull(pos.key) }"
                    @click="addToLineup(pos.key, player)"
                  >
                    {{ pos.short }}
                  </button>
                  <button
                    class="rounded px-2 py-0.5 text-xs text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                    :disabled="benchCount >= 8"
                    :class="{ 'opacity-30 cursor-not-allowed': benchCount >= 8 }"
                    @click="addToBench(player)"
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
      <p v-else class="text-text-faint">Select a club to build your lineup.</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_TEAM_BUILDER } from '../api/queries'
import { SET_FFL_LINEUP } from '../api/mutations'

const props = defineProps<{ seasonId: string; roundId: string }>()

const positions = [
  { key: 'goals', label: 'Goals', short: 'G', count: 3 },
  { key: 'kicks', label: 'Kicks', short: 'K', count: 4 },
  { key: 'handballs', label: 'Handballs', short: 'HB', count: 4 },
  { key: 'marks', label: 'Marks', short: 'M', count: 3 },
  { key: 'tackles', label: 'Tackles', short: 'T', count: 3 },
  { key: 'hitouts', label: 'Hitouts', short: 'HO', count: 2 },
  { key: 'star', label: 'Star', short: 'S', count: 3 },
] as const

type PositionKey = typeof positions[number]['key']

interface SquadPlayer {
  id: string
  name: string
}

interface Slot {
  player: SquadPlayer | null
}

// Data loading
const { result, loading, error } = useQuery(GET_FFL_TEAM_BUILDER, () => ({ seasonId: props.seasonId }))

const season = computed(() => result.value?.fflSeason ?? null)

const selectedClubSeasonId = ref<string>('')

// Auto-select first club when data loads
watch(season, (s) => {
  if (s && s.ladder.length > 0 && !selectedClubSeasonId.value) {
    selectedClubSeasonId.value = s.ladder[0].id
  }
})

const selectedClubSeason = computed(() =>
  season.value?.ladder.find((cs: { id: string }) => cs.id === selectedClubSeasonId.value) ?? null
)

// Find the club match for the selected club in the current round
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

// Squad from club season
const squad = computed<SquadPlayer[]>(() => {
  if (!selectedClubSeason.value) return []
  return selectedClubSeason.value.players.nodes.map((r: { id: string; player: { name: string } }) => ({
    id: r.id,
    name: r.player.name,
  }))
})

// Lineup state
const createSlots = (count: number): Slot[] => Array.from({ length: count }, () => ({ player: null }))

const lineupSlots = ref<Record<PositionKey, Slot[]>>(
  Object.fromEntries(positions.map(p => [p.key, createSlots(p.count)])) as Record<PositionKey, Slot[]>
)

const benchSlots = ref<Slot[]>(createSlots(8))

// Load existing lineup from club match player matches
watch(clubMatch, (cm) => {
  // Reset slots
  for (const pos of positions) {
    lineupSlots.value[pos.key] = createSlots(pos.count)
  }
  benchSlots.value = createSlots(8)

  if (!cm?.playerMatches) return

  for (const pm of cm.playerMatches) {
    const player: SquadPlayer = { id: pm.playerSeasonId, name: pm.player.name }
    const isBench = pm.backupPositions != null || pm.interchangePosition != null

    if (isBench) {
      const slot = benchSlots.value.find((s: Slot) => !s.player)
      if (slot) slot.player = player
    } else if (pm.position) {
      const posSlots = lineupSlots.value[pm.position as PositionKey]
      if (posSlots) {
        const slot = posSlots.find((s: Slot) => !s.player)
        if (slot) slot.player = player
      }
    }
  }
})

const assignedPlayerSeasonIds = computed(() => {
  const ids = new Set<string>()
  for (const pos of positions) {
    for (const slot of lineupSlots.value[pos.key]) {
      if (slot.player) ids.add(slot.player.id)
    }
  }
  for (const slot of benchSlots.value) {
    if (slot.player) ids.add(slot.player.id)
  }
  return ids
})

const availablePlayers = computed(() =>
  squad.value.filter(p => !assignedPlayerSeasonIds.value.has(p.id))
)

const starterCount = computed(() => {
  let count = 0
  for (const pos of positions) {
    count += lineupSlots.value[pos.key].filter(s => s.player).length
  }
  return count
})

const benchCount = computed(() => benchSlots.value.filter(s => s.player).length)

const isPositionFull = (key: PositionKey) =>
  lineupSlots.value[key].every(s => s.player !== null)

// Lineup management
function addToLineup(key: PositionKey, player: SquadPlayer) {
  const slot = lineupSlots.value[key].find(s => !s.player)
  if (slot) slot.player = player
}

function removeFromLineup(key: PositionKey, index: number) {
  lineupSlots.value[key][index].player = null
}

function moveToPosition(fromKey: PositionKey, fromIndex: number, toKey: PositionKey) {
  const player = lineupSlots.value[fromKey][fromIndex].player
  if (!player) return
  const toSlot = lineupSlots.value[toKey].find(s => !s.player)
  if (!toSlot) return
  lineupSlots.value[fromKey][fromIndex].player = null
  toSlot.player = player
}

function addToBench(player: SquadPlayer) {
  const slot = benchSlots.value.find(s => !s.player)
  if (slot) slot.player = player
}

function removeFromBench(index: number) {
  benchSlots.value[index].player = null
}

// Submit lineup
const { mutate: setLineup } = useMutation(SET_FFL_LINEUP)
const submitting = ref(false)
const submitMessage = ref('')

async function submitLineup() {
  if (!clubMatch.value) return
  submitting.value = true
  submitMessage.value = ''

  const players: { playerSeasonId: string; position: string; backupPositions?: string; interchangePosition?: string }[] = []

  for (const pos of positions) {
    for (const slot of lineupSlots.value[pos.key]) {
      if (slot.player) {
        players.push({ playerSeasonId: slot.player.id, position: pos.key })
      }
    }
  }

  for (const slot of benchSlots.value) {
    if (slot.player) {
      // Bench players get first available position as backup
      players.push({ playerSeasonId: slot.player.id, position: 'kicks', backupPositions: 'kicks' })
    }
  }

  try {
    await setLineup({ input: { clubMatchId: clubMatch.value.id, players } })
    submitMessage.value = 'Lineup saved!'
  } catch (e) {
    submitMessage.value = 'Failed to save lineup'
  } finally {
    submitting.value = false
  }
}
</script>
