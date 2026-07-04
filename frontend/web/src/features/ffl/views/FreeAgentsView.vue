<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-text">Free Agents</h1>
    </div>

    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else>
      <!-- Tab navigation -->
      <div class="flex gap-1 mb-6 border-b border-border">
        <button
          v-for="sec in sections"
          :key="sec.key"
          @click="activeTab = sec.key"
          class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px"
          :class="activeTab === sec.key
            ? 'border-active text-active'
            : 'border-transparent text-text-muted hover:text-text'"
        >{{ sec.title }}</button>
      </div>

      <!-- Tab content -->
      <template v-for="sec in sections" :key="sec.key">
        <div v-if="activeTab === sec.key">
          <p class="mb-3 text-xs text-text-faint">Season avg <span class="text-green-400">↑</span><span class="text-red-400">↓</span> Last {{ LAST_N }} avg — top 20</p>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border text-text-muted">
                  <th class="py-2 pr-4 text-left font-medium">Player</th>
                  <th class="py-2 pr-4 text-left font-medium whitespace-nowrap">Club</th>
                  <th class="py-2 px-2 text-right font-medium">Gms</th>
                  <th
                    v-for="col in statCols" :key="col.key"
                    class="py-2 px-2 text-right font-medium"
                    :class="col.key === sec.key ? 'text-text' : ''"
                  >{{ col.label }}</th>
                  <th
                    class="py-2 pl-2 pr-3 text-right font-medium"
                    :class="sec.key === 'star' ? 'text-yellow-400' : 'text-text-muted opacity-60'"
                  >★</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in sec.rows"
                  :key="row.id"
                  class="border-b border-border-subtle hover:bg-surface-hover"
                >
                  <td class="py-2 pr-4 font-medium">
                    <router-link
                      :to="{ name: 'ffl-afl-player-season', params: { aflPlayerSeasonId: row.id } }"
                      class="hover:text-text-muted transition-colors"
                    >{{ row.playerName }}</router-link>
                  </td>
                  <td class="py-2 pr-4 text-xs text-text-muted whitespace-nowrap">
                    <router-link
                      v-if="row.clubSeasonId"
                      :to="{ name: 'ffl-afl-club-season', params: { clubSeasonId: row.clubSeasonId } }"
                      class="inline-flex items-center gap-1.5 hover:text-text transition-colors"
                    >
                      <img :src="clubLogoUrl(row.clubName)" class="w-4 h-4 object-contain" />
                      {{ row.clubName }}
                    </router-link>
                    <span v-else>{{ row.clubName ?? '—' }}</span>
                  </td>
                  <td class="py-2 px-2 text-right tabular-nums text-text-muted text-xs">{{ row.games ?? '—' }}</td>
                  <td v-for="col in statCols" :key="col.key" class="py-2 px-2 tabular-nums">
                    <StatCell
                      :avg="row.statsAll?.[col.key]"
                      :last3="row.statsLastN?.[col.key]"
                      :avg-style="statHeat(row.statsAll?.[col.key], col.key)"
                      :last3-style="statHeat(row.statsLastN?.[col.key], col.key)"
                    />
                  </td>
                  <td class="py-2 pl-2 pr-3 tabular-nums">
                    <StatCell
                      :avg="row.statsAll ? starScore(row.statsAll) : null"
                      :last3="row.statsLastN ? starScore(row.statsLastN) : null"
                      :avg-style="statHeat(row.statsAll ? starScore(row.statsAll) : null, 'star')"
                      :last3-style="statHeat(row.statsLastN ? starScore(row.statsLastN) : null, 'star')"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { useTheme } from '@/composables/useTheme'
import { heatStyle } from '@/utils/heatmap'
import { statCols, starScore, type StatSummary, type StatKey, LAST_N } from '../utils/playerStats'
import { POSITION_LABEL } from '../utils/position'
import StatCell from '../components/StatCell.vue'
import { GET_FREE_AGENTS } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import { clubLogoUrl } from '@/features/afl/utils/clubLogos'

const { liveSeasonId } = useFflState()
const { isDark } = useTheme()

const { result, loading, error } = useQuery(
  GET_FREE_AGENTS,
  () => ({ fflSeasonId: liveSeasonId.value }),
  () => ({ enabled: !!liveSeasonId.value }),
)

type SectionKey = StatKey | 'star'

const activeTab = ref<SectionKey>('kicks')

interface PlayerSeasonRaw {
  id: string
  player: { id: string; name: string }
  clubSeason: { id: string; club: { id: string; name: string } } | null
  statsAll: (StatSummary & { games?: number }) | null
  statsLastN: StatSummary | null
  fflPlayerSeasons: { toRoundId: string | null }[]
}

interface MappedRow {
  id: string
  playerName: string
  clubName: string | null
  clubSeasonId: string | null
  games: number | null
  statsAll: StatSummary | null
  statsLastN: StatSummary | null
}

const rows = computed((): MappedRow[] => {
  const nodes: PlayerSeasonRaw[] = result.value?.fflSeason?.aflSeason?.playerSeasons?.nodes ?? []
  return nodes
    .filter(n => n.fflPlayerSeasons.length === 0 || !n.fflPlayerSeasons.some(s => s.toRoundId === null))
    .map(n => ({
      id: n.id,
      playerName: n.player.name,
      clubName: n.clubSeason?.club.name ?? null,
      clubSeasonId: n.clubSeason?.id ?? null,
      games: n.statsAll?.games ?? null,
      statsAll: n.statsAll,
      statsLastN: n.statsLastN,
    }))
})

// Global heatmap ranges — consistent colours across all tabs
const columnRange = computed(() => {
  const range = {} as Record<SectionKey, { min: number; max: number }>
  for (const col of statCols) {
    const vals = rows.value.map(r => r.statsLastN?.[col.key]).filter((v): v is number => v != null)
    if (vals.length) range[col.key] = { min: Math.min(...vals), max: Math.max(...vals) }
  }
  const starVals = rows.value.map(r => r.statsLastN ? starScore(r.statsLastN) : null).filter((v): v is number => v != null)
  if (starVals.length) range['star'] = { min: Math.min(...starVals), max: Math.max(...starVals) }
  return range
})

function statHeat(value: number | null | undefined, key: SectionKey): Record<string, string> {
  if (value == null) return {}
  const r = columnRange.value[key]
  if (!r) return {}
  return heatStyle(value, r.min, r.max, isDark.value)
}

function sortVal(row: MappedRow, key: SectionKey): number {
  if (key === 'star') {
    return row.statsLastN ? starScore(row.statsLastN) : row.statsAll ? starScore(row.statsAll) : -1
  }
  return row.statsLastN?.[key] ?? row.statsAll?.[key] ?? -1
}

const sections = computed(() => {
  const defs: { key: SectionKey; title: string }[] = [
    ...statCols.map(c => ({ key: c.key as SectionKey, title: POSITION_LABEL[c.key] ?? c.label })),
    { key: 'star', title: POSITION_LABEL['star'] ?? 'Star' },
  ]
  return defs.map(def => ({
    ...def,
    rows: [...rows.value]
      .sort((a, b) => sortVal(b, def.key) - sortVal(a, def.key))
      .slice(0, 20),
  }))
})
</script>
