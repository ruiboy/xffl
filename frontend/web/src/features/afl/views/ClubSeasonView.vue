<template>
  <div>
    <Breadcrumb v-if="clubSeason" :items="breadcrumbs" />

    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <NotFound v-else-if="notFound" entity="Club season" />
    <template v-else-if="clubSeason">

      <div class="mb-6 flex items-center justify-between gap-3">
        <div class="flex items-center gap-3">
          <img :src="aflClubLogoUrl(clubSeason.club.name)" class="w-10 h-10 object-contain" />
          <h1 class="text-2xl font-bold text-text">{{ clubSeason.club.name }}</h1>
        </div>
        <StatSourceToggle />
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border text-left text-text-muted">
              <th class="py-2 pr-4 font-medium cursor-pointer select-none" :class="sortKey === 'playerName' ? 'text-sky-400' : 'hover:text-text'" @click="sortBy('playerName')">Player</th>
              <th class="py-2 px-2 font-medium text-right cursor-pointer select-none" :class="sortKey === 'games' ? 'text-sky-400' : 'hover:text-text'" @click="sortBy('games')">Gms</th>
              <th v-for="col in statCols" :key="col.key" class="py-2 px-2 font-medium text-right cursor-pointer select-none" :class="sortKey === col.key ? 'text-sky-400' : 'hover:text-text'" @click="sortBy(col.key)">{{ POSITION_LABEL[col.key] }}</th>
              <th class="py-2 pl-2 pr-3 font-medium text-right cursor-pointer select-none" :class="sortKey === 'star' ? 'text-sky-400' : 'hover:text-text'" @click="sortBy('star')">Star <span class="text-yellow-400">★</span></th>
              <th class="py-2 pl-3 pr-3 font-medium bg-white/[0.03] border-l border-border whitespace-nowrap cursor-pointer select-none" :class="sortKey === 'fflClub' ? 'text-sky-400' : 'hover:text-text'" @click="sortBy('fflClub')">FFL Club</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in sortedRows"
              :key="row.id"
              class="border-b border-border-subtle hover:bg-surface-hover"
            >
              <td class="py-2 pr-4 font-medium">
                <router-link
                  :to="{ name: 'ffl-afl-player-season', params: { aflPlayerSeasonId: row.id } }"
                  class="hover:text-text-muted transition-colors"
                >{{ row.playerName }}</router-link>
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
              <td class="py-2 pl-3 pr-3 bg-white/[0.03] border-l border-border whitespace-nowrap">
                <router-link
                  v-if="row.fflClubSeasonId"
                  :to="{ name: 'ffl-club-season', params: { clubSeasonId: row.fflClubSeasonId } }"
                  class="inline-flex items-center gap-1.5 hover:text-text-muted transition-colors"
                >
                  <img :src="fflClubLogoUrl(row.fflClub!)" class="w-4 h-4 object-contain" />
                  <span>{{ row.fflClub }}</span>
                </router-link>
                <span v-else-if="row.fflClub" class="inline-flex items-center gap-1.5">
                  <img :src="fflClubLogoUrl(row.fflClub)" class="w-4 h-4 object-contain" />
                  <span>{{ row.fflClub }}</span>
                </span>
                <span v-else class="text-text-faint">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { useNotFound } from '@/composables/useNotFound'
import NotFound from '@/components/NotFound.vue'
import { GET_AFL_CLUB_SEASON } from '@/features/ffl/api/queries'
import Breadcrumb from '@/features/ffl/components/Breadcrumb.vue'
import { clubLogoUrl as aflClubLogoUrl } from '@/features/afl/utils/clubLogos'
import { clubLogoUrl as fflClubLogoUrl } from '@/features/ffl/utils/clubLogos'
import { useTheme } from '@/composables/useTheme'
import { heatStyle } from '@/utils/heatmap'
import { statCols, starScore, type StatSummary, type StatKey } from '@/features/ffl/utils/playerStats'
import { POSITION_LABEL } from '@/features/ffl/utils/position'
import StatCell from '@/features/ffl/components/StatCell.vue'
import StatSourceToggle from '@/features/ffl/components/StatSourceToggle.vue'
import { useStatSource } from '@/features/ffl/composables/useStatSource'

const props = defineProps<{ clubSeasonId: string }>()

const { result, loading, error } = useQuery(GET_AFL_CLUB_SEASON, () => ({ id: props.clubSeasonId }))

interface FFLPlayerSeasonStub {
  id: string
  clubSeasonId: string
  club: { id: string; name: string }
  toRoundId: string | null
}

interface PlayerSeasonStub {
  id: string
  player: { id: string; name: string }
  statsAll: (StatSummary & { games?: number }) | null
  statsLastN: StatSummary | null
  fflPlayerSeasons: FFLPlayerSeasonStub[]
}

interface ClubSeason {
  id: string
  club: { id: string; name: string }
  season: { id: string; name: string }
  playerSeasons: PlayerSeasonStub[]
}

const clubSeason = computed(() => result.value?.aflClubSeason as ClubSeason | null ?? null)
const notFound = useNotFound(clubSeason, loading, error)

const { isDark } = useTheme()
const { statSource } = useStatSource()

type SectionKey = StatKey | 'star'

const breadcrumbs = computed(() => {
  if (!clubSeason.value) return []
  return [
    { label: 'FFL', to: { name: 'home' } },
    { label: clubSeason.value.season.name, to: { name: 'afl-home' } },
    { label: clubSeason.value.club.name },
  ]
})

interface Row {
  id: string
  playerName: string
  statsAll: (StatSummary & { games?: number }) | null
  statsLastN: StatSummary | null
  games: number | null
  fflClub: string | null
  fflClubSeasonId: string | null
}

const rows = computed((): Row[] => {
  if (!clubSeason.value) return []
  return clubSeason.value.playerSeasons.map(ps => {
    const currentStint = ps.fflPlayerSeasons.find(s => s.toRoundId === null)
    return {
      id: ps.id,
      playerName: ps.player.name,
      statsAll: ps.statsAll,
      statsLastN: ps.statsLastN,
      games: ps.statsAll?.games ?? null,
      fflClub: currentStint?.club.name ?? null,
      fflClubSeasonId: currentStint?.clubSeasonId ?? null,
    }
  })
})

// Heatmap ranges scale to the active source (season vs last-5), consistent per column.
const columnRange = computed(() => {
  const src = (r: Row) => statSource.value === 'form' ? r.statsLastN : r.statsAll
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

// ── Sorting ──────────────────────────────────────────────────────────────────
type SortKey = 'playerName' | 'games' | StatKey | 'star' | 'fflClub'
const textKeys: SortKey[] = ['playerName', 'fflClub']

const sortKey = ref<SortKey | null>(null)
const sortDir = ref<'asc' | 'desc'>('desc')

function sortBy(key: SortKey) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = textKeys.includes(key) ? 'asc' : 'desc' // names A→Z, numbers high→low
  }
}

// Stats sort on the active source (season vs last-5), falling back to the other.
function sortValue(row: Row, key: SortKey): number | string | null {
  if (key === 'playerName') return row.playerName
  if (key === 'fflClub') return row.fflClub
  if (key === 'games') return row.games
  const primary = statSource.value === 'form' ? row.statsLastN : row.statsAll
  const fallback = statSource.value === 'form' ? row.statsAll : row.statsLastN
  if (key === 'star') return primary ? starScore(primary) : fallback ? starScore(fallback) : null
  return primary?.[key] ?? fallback?.[key] ?? null
}

const sortedRows = computed(() => {
  const key = sortKey.value
  if (!key) return rows.value
  const dir = sortDir.value === 'asc' ? 1 : -1
  return [...rows.value].sort((a, b) => {
    const av = sortValue(a, key)
    const bv = sortValue(b, key)
    if (av == null && bv == null) return 0
    if (av == null) return 1 // missing values last, either direction
    if (bv == null) return -1
    if (typeof av === 'string' && typeof bv === 'string') return av.localeCompare(bv) * dir
    return ((av as number) - (bv as number)) * dir
  })
})
</script>
