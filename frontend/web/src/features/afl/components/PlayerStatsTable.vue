<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-gray-200 text-left text-gray-500">
          <th class="py-2 pr-4 font-medium">Player</th>
          <th v-for="col in statColumns" :key="col.key" class="py-2 px-2 font-medium text-right w-16">
            {{ col.label }}
          </th>
          <th class="py-2 px-2 font-medium text-right w-16">D</th>
          <th class="py-2 px-2 font-medium text-right w-16">SC</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="pm in clubMatch.playerMatches"
          :key="pm.id"
          class="border-b border-gray-100 hover:bg-gray-50"
        >
          <td class="py-2 pr-4 font-medium">{{ pm.player.name }}</td>
          <td v-for="col in statColumns" :key="col.key" class="py-1 px-1 text-right">
            <input
              v-if="!readonly"
              type="number"
              :value="pm[col.key]"
              min="0"
              class="w-14 rounded bg-transparent px-1 py-1 text-right text-gray-900 tabular-nums hover:bg-gray-100 focus:bg-gray-100 focus:outline-none focus:ring-1 focus:ring-gray-300"
              @change="onStatChange(pm, col.key, $event)"
            />
            <span v-else class="tabular-nums text-gray-900 px-1">{{ pm[col.key] }}</span>
          </td>
          <td class="py-2 px-2 text-right tabular-nums text-gray-500">{{ pm.disposals }}</td>
          <td class="py-2 px-2 text-right tabular-nums text-gray-500">{{ pm.score }}</td>
        </tr>
      </tbody>
      <tfoot>
        <tr class="border-t border-gray-300 font-semibold text-gray-700">
          <td class="py-2 pr-4">Totals</td>
          <td v-for="col in statColumns" :key="col.key" class="py-2 px-2 text-right tabular-nums">
            {{ totals[col.key] }}
          </td>
          <td class="py-2 px-2 text-right tabular-nums">{{ totals.disposals }}</td>
          <td class="py-2 px-2 text-right tabular-nums">{{ totals.score }}</td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface PlayerMatch {
  id: string
  playerSeasonId: string
  player: { id: string; name: string }
  kicks: number
  handballs: number
  marks: number
  hitouts: number
  tackles: number
  goals: number
  behinds: number
  disposals: number
  score: number
}

interface ClubMatch {
  id: string
  club: { id: string; name: string }
  playerMatches: PlayerMatch[]
}

const props = withDefaults(defineProps<{
  clubMatch: ClubMatch
  readonly?: boolean
}>(), {
  readonly: false,
})

const emit = defineEmits<{
  update: [input: { playerSeasonId: string; clubMatchId: string; [key: string]: unknown }]
}>()

const statColumns = [
  { key: 'kicks' as const, label: 'K' },
  { key: 'handballs' as const, label: 'HB' },
  { key: 'marks' as const, label: 'M' },
  { key: 'hitouts' as const, label: 'HO' },
  { key: 'tackles' as const, label: 'T' },
  { key: 'goals' as const, label: 'G' },
  { key: 'behinds' as const, label: 'B' },
]

type StatKey = typeof statColumns[number]['key']

const totals = computed(() => {
  const keys = [...statColumns.map(c => c.key), 'disposals' as const, 'score' as const]
  const sums: Record<string, number> = {}
  for (const key of keys) {
    sums[key] = props.clubMatch.playerMatches.reduce((sum, pm) => sum + pm[key], 0)
  }
  return sums
})

function onStatChange(pm: PlayerMatch, key: StatKey, event: Event) {
  const target = event.target as HTMLInputElement
  const value = parseInt(target.value, 10)
  if (isNaN(value) || value < 0) return
  if (value === pm[key]) return

  emit('update', {
    playerSeasonId: pm.playerSeasonId,
    clubMatchId: props.clubMatch.id,
    [key]: value,
  })
}
</script>
