<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border text-left text-text-muted">
          <th class="py-2 pr-4 font-medium w-8">#</th>
          <th class="py-2 pr-4 font-medium">Club</th>
          <th class="py-2 px-2 font-medium text-right">P</th>
          <th class="py-2 px-2 font-medium text-right">W</th>
          <th class="py-2 px-2 font-medium text-right">L</th>
          <th class="py-2 px-2 font-medium text-right">D</th>
          <th v-if="showExtra" class="py-2 px-2 font-medium text-right" title="Extra points (superbye bonus)">+</th>
          <th class="py-2 px-2 font-medium text-right">F</th>
          <th class="py-2 px-2 font-medium text-right">A</th>
          <th class="py-2 px-2 font-medium text-right">%</th>
          <th class="py-2 px-2 font-medium text-right">Pts</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(entry, index) in ladder"
          :key="entry.id"
          class="border-b border-border-subtle hover:bg-surface-hover"
        >
          <td class="py-2 pr-4 tabular-nums text-text-faint">{{ index + 1 }}</td>
          <td class="py-2 pr-4 font-medium">
            <router-link
              :to="{ name: 'ffl-club-season', params: { clubSeasonId: entry.id } }"
              class="flex items-center gap-2 hover:text-active transition-colors"
            >
              <img :src="clubLogoUrl(entry.club.name)" :alt="entry.club.name" class="w-6 h-6 object-contain" />
              {{ entry.club.name }}
            </router-link>
          </td>
          <td class="py-2 px-2 text-right tabular-nums">{{ entry.played }}</td>
          <td class="py-2 px-2 text-right tabular-nums" :style="heat(entry.won, 'won')">{{ entry.won }}</td>
          <td class="py-2 px-2 text-right tabular-nums" :style="heat(entry.lost, 'lost')">{{ entry.lost }}</td>
          <td class="py-2 px-2 text-right tabular-nums">{{ entry.drawn }}</td>
          <td v-if="showExtra" class="py-2 px-2 text-right tabular-nums" :style="heat(entry.extraPoints ?? 0, 'extraPoints')">{{ entry.extraPoints || '' }}</td>
          <td class="py-2 px-2 text-right tabular-nums" :style="heat(entry.for, 'for')">{{ entry.for }}</td>
          <td class="py-2 px-2 text-right tabular-nums" :style="heat(entry.against, 'against')">{{ entry.against }}</td>
          <td class="py-2 px-2 text-right tabular-nums font-semibold" :style="heat(entry.percentage, 'percentage')">{{ entry.percentage.toFixed(1) }}</td>
          <td class="py-2 px-2 text-right tabular-nums font-semibold" :style="heat(entry.premiershipPoints, 'premiershipPoints')">{{ entry.premiershipPoints }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { heatStyle } from '@/utils/heatmap'
import { useTheme } from '@/composables/useTheme'

interface LadderEntry {
  id: string
  club: { id: string; name: string }
  played: number
  won: number
  lost: number
  drawn: number
  for: number
  against: number
  percentage: number
  premiershipPoints: number
  extraPoints?: number
}

const props = defineProps<{ ladder: LadderEntry[] }>()
const { isDark } = useTheme()

// Only surface the extra-points column for seasons that actually award them.
const showExtra = computed(() => props.ladder.some((e) => (e.extraPoints ?? 0) > 0))

// Heatmap columns. heatStyle highlights higher values, so lower-is-better columns
// (L, A) are mirrored to keep "brighter = better". P and D are left untinted.
type HeatKey = 'won' | 'lost' | 'for' | 'against' | 'percentage' | 'premiershipPoints' | 'extraPoints'
const LOWER_IS_BETTER: Record<HeatKey, boolean> = {
  won: false,
  lost: true,
  for: false,
  against: true,
  percentage: false,
  premiershipPoints: false,
  extraPoints: false,
}

const ranges = computed(() => {
  const r = {} as Record<HeatKey, { min: number; max: number }>
  for (const key of Object.keys(LOWER_IS_BETTER) as HeatKey[]) {
    const vals = props.ladder.map((e) => e[key] ?? 0)
    r[key] = { min: Math.min(...vals), max: Math.max(...vals) }
  }
  return r
})

function heat(value: number, key: HeatKey): Record<string, string> {
  const { min, max } = ranges.value[key]
  const v = LOWER_IS_BETTER[key] ? min + max - value : value
  return heatStyle(v, min, max, isDark.value)
}
</script>
