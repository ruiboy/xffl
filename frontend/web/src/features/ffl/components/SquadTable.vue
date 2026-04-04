<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border text-left text-text-muted">
          <th class="py-2 pr-4 font-medium">Player</th>
          <th class="py-2 px-2 font-medium">Position</th>
          <th class="py-2 px-2 font-medium">Status</th>
          <th class="py-2 px-2 font-medium text-right">Score</th>
          <th class="py-2 px-2 font-medium w-8"></th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="pm in starters"
          :key="pm.id"
          class="border-b border-border-subtle hover:bg-surface-hover"
        >
          <td class="py-2 pr-4 font-medium">{{ pm.player.name }}</td>
          <td class="py-2 px-2 capitalize text-text-muted">{{ pm.position ?? '—' }}</td>
          <td class="py-2 px-2">
            <StatusBadge :status="pm.status" />
          </td>
          <td class="py-2 px-2 text-right tabular-nums font-semibold">{{ pm.score }}</td>
          <td class="py-2 px-2"></td>
        </tr>
        <tr v-if="bench.length > 0">
          <td colspan="5" class="pt-4 pb-2 text-xs font-semibold uppercase tracking-wider text-text-faint">
            Bench
          </td>
        </tr>
        <tr
          v-for="pm in bench"
          :key="pm.id"
          class="border-b border-border-subtle hover:bg-surface-hover"
        >
          <td class="py-2 pr-4 font-medium text-text-muted">{{ pm.player.name }}</td>
          <td class="py-2 px-2 text-text-faint text-xs">
            <span v-if="pm.backupPositions">backup: {{ pm.backupPositions }}</span>
            <span v-else-if="pm.interchangePosition">interchange: {{ pm.interchangePosition }}</span>
          </td>
          <td class="py-2 px-2">
            <StatusBadge :status="pm.status" />
          </td>
          <td class="py-2 px-2 text-right tabular-nums">{{ pm.score }}</td>
          <td class="py-2 px-2">
            <span v-if="isSubActivated(pm)" class="text-xs text-green-500" title="Substitution activated">SUB</span>
            <span v-else-if="isInterchangeActivated(pm)" class="text-xs text-blue-500" title="Interchange activated">INT</span>
          </td>
        </tr>
      </tbody>
      <tfoot>
        <tr class="border-t border-border font-semibold">
          <td class="py-2 pr-4">Total</td>
          <td></td>
          <td></td>
          <td class="py-2 px-2 text-right tabular-nums">{{ total }}</td>
          <td></td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import StatusBadge from './StatusBadge.vue'

interface PlayerMatch {
  id: string
  player: { name: string }
  position: string | null
  status: string | null
  backupPositions: string | null
  interchangePosition: string | null
  score: number
}

const props = defineProps<{ playerMatches: PlayerMatch[] }>()

const isBench = (pm: PlayerMatch) => pm.backupPositions != null || pm.interchangePosition != null

const starters = computed(() => props.playerMatches.filter(pm => !isBench(pm)))
const bench = computed(() => props.playerMatches.filter(pm => isBench(pm)))

const isSubActivated = (pm: PlayerMatch) => {
  if (!pm.backupPositions || pm.status !== 'played') return false
  // A bench sub activates when a starter at a matching position has DNP'd
  const backupPos = pm.backupPositions.split(',').map(p => p.trim())
  return starters.value.some(s => s.status === 'dnp' && backupPos.includes(s.position ?? ''))
}

const isInterchangeActivated = (pm: PlayerMatch) => {
  if (!pm.interchangePosition || pm.status !== 'played') return false
  // An interchange activates when the bench player outscored the starter at that position
  const starter = starters.value.find(s => s.position === pm.interchangePosition && s.status === 'played')
  return starter != null && pm.score > starter.score
}

const total = computed(() => props.playerMatches.reduce((sum, pm) => sum + pm.score, 0))
</script>
