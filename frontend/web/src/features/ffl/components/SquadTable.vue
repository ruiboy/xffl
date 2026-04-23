<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border text-left text-text-muted">
          <th class="py-2 pr-4 font-medium">Player</th>
          <th class="py-2 px-2 font-medium">Status</th>
          <th class="py-2 px-2 font-medium text-right">Score</th>
        </tr>
      </thead>
      <tbody>
        <template v-for="(group, i) in starterGroups" :key="group.position">
          <tr>
            <td colspan="3" class="pb-1 text-xs font-semibold uppercase tracking-wider text-text-faint" :class="i === 0 ? 'pt-3' : 'pt-5'">
              <span>{{ group.label }}</span>
              <span class="pl-3">({{ group.total }})</span>
            </td>
          </tr>
          <tr
            v-for="pm in group.players"
            :key="pm.id"
            class="border-b border-border-subtle hover:bg-surface-hover"
          >
            <td class="py-2 pr-4 font-medium">{{ pm.player.name }}</td>
            <td class="py-2 px-2"><StatusBadge :status="pm.status" /></td>
            <td class="py-2 px-2 text-right tabular-nums font-semibold">{{ pm.score }}</td>
          </tr>
        </template>

        <template v-if="bench.length > 0">
          <tr>
            <td colspan="3" class="pt-5 pb-2 text-xs font-semibold uppercase tracking-wider text-text-faint">Bench</td>
          </tr>
          <tr
            v-for="pm in bench"
            :key="pm.id"
            class="border-b border-border-subtle hover:bg-surface-hover"
          >
            <td class="py-2 pr-4 font-medium text-text-muted">{{ pm.player.name }}</td>
            <td class="py-2 px-2"><StatusBadge :status="pm.status" />
              <span v-if="isSubActivated(pm)" class="text-xs text-green-500" title="Substitution activated">SUB</span>
              <span v-else-if="isInterchangeActivated(pm)" class="text-xs text-blue-500" title="Interchange activated">INT</span>
            </td>
            <td class="py-2 px-2 text-right tabular-nums">{{ pm.score }}</td>
          </tr>
        </template>
      </tbody>
      <tfoot>
        <tr class="border-t border-border font-semibold">
          <td class="py-2 pr-4">Total</td>
          <td></td>
          <td class="py-2 px-2 text-right tabular-nums">{{ total }}</td>
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

const POSITION_ORDER = ['goals', 'kicks', 'handballs', 'marks', 'tackles', 'hitouts', 'star'] as const
type PositionKey = typeof POSITION_ORDER[number]

const POSITION_LABELS: Record<PositionKey, string> = {
  goals:     'Goals',
  kicks:     'Kicks',
  handballs: 'Handballs',
  marks:     'Marks',
  tackles:   'Tackles',
  hitouts:   'Hitouts',
  star:      'Star',
}

const isBench = (pm: PlayerMatch) => pm.backupPositions != null || pm.interchangePosition != null

const starters = computed(() => props.playerMatches.filter(pm => !isBench(pm)))
const bench    = computed(() => props.playerMatches.filter(pm => isBench(pm)))

const starterGroups = computed(() => {
  const grouped: Partial<Record<string, PlayerMatch[]>> = {}
  for (const pm of starters.value) {
    const key = pm.position ?? 'unknown'
    ;(grouped[key] ??= []).push(pm)
  }
  return POSITION_ORDER
    .filter(pos => grouped[pos]?.length)
    .map(pos => ({
      position: pos,
      label: POSITION_LABELS[pos],
      players: grouped[pos]!,
      total: grouped[pos]!.reduce((sum, pm) => sum + pm.score, 0),
    }))
})

const isSubActivated = (pm: PlayerMatch) => {
  if (!pm.backupPositions || pm.status !== 'played') return false
  const backupPos = pm.backupPositions.split(',').map(p => p.trim())
  return starters.value.some(s => s.status === 'dnp' && backupPos.includes(s.position ?? ''))
}

const isInterchangeActivated = (pm: PlayerMatch) => {
  if (!pm.interchangePosition || pm.status !== 'played') return false
  const starter = starters.value.find(s => s.position === pm.interchangePosition && s.status === 'played')
  return starter != null && pm.score > starter.score
}

const total = computed(() => props.playerMatches.reduce((sum, pm) => sum + pm.score, 0))
</script>
