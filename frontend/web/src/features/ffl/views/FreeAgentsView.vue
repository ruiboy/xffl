<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold text-text">Free Agents</h1>
      <StatSourceToggle />
    </div>

    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else>
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
              >
                <button
                  class="transition-colors"
                  :class="col.key === activeKey ? 'text-sky-400' : 'hover:text-text'"
                  :title="`Top 20 by ${POSITION_LABEL[col.key]}, ranked on ${rankBasisLabel} average`"
                  @click="activeKey = col.key"
                >{{ POSITION_LABEL[col.key] }}</button>
              </th>
              <th class="py-2 pl-2 pr-3 text-right font-medium">
                <button
                  class="transition-colors"
                  :class="activeKey === 'star' ? 'text-sky-400' : 'text-text-muted hover:text-text'"
                  :title="`Top 20 by Star score, ranked on ${rankBasisLabel} average`"
                  @click="activeKey = 'star'"
                >Star <span class="text-yellow-400">★</span></button>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in topRows"
              :key="row.id"
              class="border-b border-border-subtle hover:bg-surface-hover"
            >
                  <td class="py-2 pr-4 font-medium">
                    <PlayerStatsCard :name="row.playerName" :club="row.clubName" :afl-status="null" :afl-player-season-id="row.id" :afl-round-id="null">
                      <router-link
                        :to="{ name: 'ffl-afl-player-season', params: { aflPlayerSeasonId: row.id } }"
                        class="hover:text-text-muted transition-colors"
                      >{{ row.playerName }}</router-link>
                    </PlayerStatsCard>
                  </td>
                  <td class="py-2 pr-4 text-xs text-text-muted whitespace-nowrap">
                    <router-link
                      v-if="row.clubSeasonId"
                      :to="{ name: 'ffl-afl-club-season', params: { clubSeasonId: row.clubSeasonId } }"
                      class="inline-flex items-center gap-1.5 hover:text-text transition-colors"
                    >
                      <img :src="clubLogoUrl(row.clubName ?? '')" class="w-4 h-4 object-contain" />
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
import PlayerStatsCard from '../components/PlayerStatsCard.vue'
import StatSourceToggle from '../components/StatSourceToggle.vue'
import { useStatSource } from '../composables/useStatSource'
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

// Stat that ranks the top-20 cut — set by clicking a column header.
const activeKey = ref<SectionKey>('kicks')

// Which average ranks the cut (and scales the heatmap) — global toggle shared
// with the other stats pages.
const { statSource } = useStatSource()

const rankBasisLabel = computed(() => statSource.value === 'form' ? `last ${LAST_N}` : 'season')

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

// Global heatmap ranges — consistent colours across all columns, scaled to the
// active rank basis.
const columnRange = computed(() => {
  const src = (r: MappedRow) => statSource.value === 'form' ? r.statsLastN : r.statsAll
  const range = {} as Record<SectionKey, { min: number; max: number }>
  for (const col of statCols) {
    const vals = rows.value.map(r => src(r)?.[col.key]).filter((v): v is number => v != null)
    if (vals.length) range[col.key] = { min: Math.min(...vals), max: Math.max(...vals) }
  }
  const starVals = rows.value.map(r => { const s = src(r); return s ? starScore(s) : null }).filter((v): v is number => v != null)
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
  const [primary, fallback] = statSource.value === 'form'
    ? [row.statsLastN, row.statsAll]
    : [row.statsAll, row.statsLastN]
  if (key === 'star') {
    return primary ? starScore(primary) : fallback ? starScore(fallback) : -1
  }
  return primary?.[key] ?? fallback?.[key] ?? -1
}

const topRows = computed(() =>
  [...rows.value]
    .sort((a, b) => sortVal(b, activeKey.value) - sortVal(a, activeKey.value))
    .slice(0, 20),
)
</script>
