<template>
  <div>
    <Breadcrumb v-if="clubSeason" :items="breadcrumbs" />

    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="clubSeason">

      <div class="mb-6 flex items-center gap-3">
        <img :src="aflClubLogoUrl(clubSeason.club.name)" class="w-10 h-10 object-contain" />
        <h1 class="text-2xl font-bold text-text">{{ clubSeason.club.name }}</h1>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border text-left text-text-muted">
              <th class="py-2 pr-4 font-medium">Player</th>
              <th v-for="col in statCols" :key="col.key" class="py-2 px-2 font-medium text-right w-10">{{ col.label }}</th>
              <th class="py-2 pl-2 pr-3 font-medium text-right w-10 text-yellow-400">★</th>
              <th class="py-2 w-5"></th>
              <th class="py-2 pl-5 pr-3 font-medium bg-white/[0.03] border-l border-border whitespace-nowrap">FFL Club</th>
              <th class="py-2 w-5 bg-white/[0.03]"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in rows"
              :key="row.id"
              class="border-b border-border-subtle hover:bg-surface-hover group"
            >
              <td class="py-2 pr-4 font-medium">{{ row.playerName }}</td>
              <td v-for="col in statCols" :key="col.key" class="py-2 px-2 text-right tabular-nums text-text-muted">
                {{ fmtStat(row.stats?.[col.key]) }}
              </td>
              <td class="py-2 pl-2 pr-3 text-right tabular-nums text-yellow-400">{{ starStr(row.stats) }}</td>
              <td class="py-2 pr-2 w-5">
                <router-link
                  :to="{ name: 'ffl-afl-player-season', params: { aflPlayerSeasonId: row.id } }"
                  class="opacity-0 group-hover:opacity-100 transition-opacity text-text-faint hover:text-text"
                  title="View player season"
                >
                  <svg class="w-3.5 h-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M6 3H3a1 1 0 0 0-1 1v9a1 1 0 0 0 1 1h9a1 1 0 0 0 1-1v-3M9 2h5m0 0v5m0-5L7 9"/>
                  </svg>
                </router-link>
              </td>
              <td class="py-2 pl-5 pr-3 bg-white/[0.03] border-l border-border">
                <span v-if="row.fflClub" class="inline-flex items-center gap-1.5">
                  <img :src="fflClubLogoUrl(row.fflClub)" class="w-4 h-4 object-contain" />
                  <span>{{ row.fflClub }}</span>
                </span>
                <span v-else class="text-text-faint">—</span>
              </td>
              <td class="py-2 pr-2 w-5 bg-white/[0.03]">
                <router-link
                  v-if="row.fflClubSeasonId"
                  :to="{ name: 'ffl-club-season', params: { clubSeasonId: row.fflClubSeasonId } }"
                  class="opacity-0 group-hover:opacity-100 transition-opacity text-text-faint hover:text-text"
                  title="View FFL club"
                >
                  <svg class="w-3.5 h-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M6 3H3a1 1 0 0 0-1 1v9a1 1 0 0 0 1 1h9a1 1 0 0 0 1-1v-3M9 2h5m0 0v5m0-5L7 9"/>
                  </svg>
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_AFL_CLUB_SEASON } from '@/features/ffl/api/queries'
import Breadcrumb from '@/features/ffl/components/Breadcrumb.vue'
import { clubLogoUrl as aflClubLogoUrl } from '@/features/afl/utils/clubLogos'
import { clubLogoUrl as fflClubLogoUrl } from '@/features/ffl/utils/clubLogos'

const props = defineProps<{ clubSeasonId: string }>()

const { result, loading, error } = useQuery(GET_AFL_CLUB_SEASON, () => ({ id: props.clubSeasonId }))

interface StatSummary {
  goals: number; kicks: number; handballs: number
  marks: number; tackles: number; hitouts: number
}

interface FFLPlayerSeasonStub {
  id: string
  clubSeasonId: string
  club: { id: string; name: string }
  toRoundId: string | null
}

interface PlayerSeasonStub {
  id: string
  player: { id: string; name: string }
  stats: StatSummary | null
  fflPlayerSeasons: FFLPlayerSeasonStub[]
}

interface ClubSeason {
  id: string
  club: { id: string; name: string }
  season: { id: string; name: string }
  playerSeasons: PlayerSeasonStub[]
}

const clubSeason = computed(() => result.value?.aflClubSeason as ClubSeason | null ?? null)

const statCols = [
  { key: 'kicks'     as const, label: 'K' },
  { key: 'handballs' as const, label: 'H' },
  { key: 'marks'     as const, label: 'M' },
  { key: 'tackles'   as const, label: 'T' },
  { key: 'hitouts'   as const, label: 'R' },
  { key: 'goals'     as const, label: 'G' },
]

type StatKey = typeof statCols[number]['key']

const breadcrumbs = computed(() => {
  if (!clubSeason.value) return []
  return [
    { label: 'FFL', to: { name: 'home' } },
    { label: clubSeason.value.season.name, to: { name: 'afl-home' } },
    { label: clubSeason.value.club.name },
  ]
})

const rows = computed(() => {
  if (!clubSeason.value) return []
  return clubSeason.value.playerSeasons.map(ps => {
    const currentStint = ps.fflPlayerSeasons.find(s => s.toRoundId === null)
    return {
      id: ps.id,
      playerName: ps.player.name,
      stats: ps.stats,
      fflClub: currentStint?.club.name ?? null,
      fflClubSeasonId: currentStint?.clubSeasonId ?? null,
    }
  })
})

function fmtStat(val: number | null | undefined): string {
  if (val == null) return '—'
  return val.toFixed(1)
}

function starStr(s: StatSummary | null | undefined): string {
  if (!s) return '—'
  return (s.goals * 5 + s.kicks + s.handballs + s.marks * 2 + s.tackles * 4).toFixed(1)
}
</script>
