<template>
  <div>
    <h1 class="text-2xl font-bold mb-1">Team Builder</h1>
    <p class="text-text-muted mb-6">Build your lineup for the round</p>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Lineup (left 2 cols) -->
      <div class="lg:col-span-2">
        <h2 class="text-lg font-semibold text-text-heading mb-3">Lineup ({{ starterCount }}/22)</h2>

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
              <div class="flex items-center gap-2">
                <button
                  v-if="slot.player"
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
      </div>

      <!-- Roster panel (right col) -->
      <div>
        <h2 class="text-lg font-semibold text-text-heading mb-3">Roster ({{ availablePlayers.length }})</h2>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

defineProps<{ seasonId: string; roundId: string }>()

interface Player {
  id: string
  name: string
}

interface Slot {
  player: Player | null
}

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

// Stub roster — will be replaced with real data from API
const roster = ref<Player[]>([
  { id: '1', name: 'Player 1' },
  { id: '2', name: 'Player 2' },
  { id: '3', name: 'Player 3' },
  { id: '4', name: 'Player 4' },
  { id: '5', name: 'Player 5' },
  { id: '6', name: 'Player 6' },
  { id: '7', name: 'Player 7' },
  { id: '8', name: 'Player 8' },
  { id: '9', name: 'Player 9' },
  { id: '10', name: 'Player 10' },
])

const createSlots = (count: number): Slot[] => Array.from({ length: count }, () => ({ player: null }))

const lineupSlots = ref<Record<PositionKey, Slot[]>>(
  Object.fromEntries(positions.map(p => [p.key, createSlots(p.count)])) as Record<PositionKey, Slot[]>
)

const benchSlots = ref<Slot[]>(createSlots(8))

const assignedPlayerIds = computed(() => {
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
  roster.value.filter(p => !assignedPlayerIds.value.has(p.id))
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

function addToLineup(key: PositionKey, player: Player) {
  const slot = lineupSlots.value[key].find(s => !s.player)
  if (slot) slot.player = player
}

function removeFromLineup(key: PositionKey, index: number) {
  lineupSlots.value[key][index].player = null
}

function addToBench(player: Player) {
  const slot = benchSlots.value.find(s => !s.player)
  if (slot) slot.player = player
}

function removeFromBench(index: number) {
  benchSlots.value[index].player = null
}
</script>
