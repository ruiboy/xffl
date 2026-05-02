<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="round">
      <Breadcrumb :items="breadcrumbs" />

      <h1 class="text-2xl font-bold mb-6">
        {{ round.name }}<span v-if="roundStartDate" class="font-normal text-text-faint"> · {{ roundStartDate }}</span>
      </h1>

      <RoundNav
        class="mb-8"
        :rounds="season?.rounds ?? []"
        :live-round-id="liveRoundId"
        :season-id="props.seasonId"
      />

      <section class="mb-8">
        <h2 class="text-lg font-semibold text-text-heading mb-3">Matches</h2>
        <div class="space-y-2">
          <MatchSummary
            v-for="match in round.matches"
            :key="match.id"
            :match="match"
            :to="{ name: 'ffl-match', params: { seasonId: props.seasonId, matchId: match.id } }"
            :my-club-id="selectedClubId ?? undefined"
            :build-team-to="myMatch?.id === match.id ? { name: 'ffl-team-builder', params: { seasonId: props.seasonId, roundId: props.roundId } } : undefined"
          />
        </div>
      </section>

      <section v-if="Object.keys(topScorersByPosition).length > 0" class="mb-8">
        <h2 class="text-lg font-semibold text-text-heading mb-4">Top Scorers</h2>
        <div class="grid grid-cols-4 gap-6">
          <template v-for="pos in TOP_SCORERS_POSITIONS" :key="pos">
            <div
              v-if="topScorersByPosition[pos]"
              class=""
            >
              <p class="text-sm font-semibold text-text-faint mb-2">{{ POSITION_LABELS[pos] }}</p>
              <div class="space-y-1">
                <div
                  v-for="(player, i) in topScorersByPosition[pos].players.slice(0, 4)"
                  :key="i"
                  class="flex items-center justify-between gap-2"
                >
                  <div class="flex items-center gap-2 min-w-0">
                    <img :src="clubLogoUrl(player.club)" :alt="player.club" class="w-5 h-5 object-contain shrink-0" />
                    <span class="text-sm font-medium truncate">{{ player.name }}</span>
                  </div>
                  <span class="text-sm tabular-nums font-semibold shrink-0">{{ player.score }}</span>
                </div>
              </div>
            </div>
          </template>
        </div>

      </section>

      <div v-if="aflRoundTo" class="mt-8">
        <router-link :to="aflRoundTo" class="text-sm text-text-muted hover:text-text transition-colors">
          AFL Round ↗
        </router-link>
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_ROUND } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import { useAflState } from '../../afl/composables/useAflState'
import Breadcrumb from '../components/Breadcrumb.vue'
import MatchSummary from '../components/MatchSummary.vue'
import RoundNav from '../components/RoundNav.vue'
import { clubLogoUrl } from '../utils/clubLogos'

const props = defineProps<{ seasonId: string; roundId: string }>()


const { liveRoundId, selectedClubId } = useFflState()
const { liveSeasonId: aflSeasonId } = useAflState()
const { result, loading, error } = useQuery(GET_FFL_ROUND, () => ({ id: props.roundId }))

const round = computed(() => result.value?.fflRound ?? null)
const season = computed(() => round.value?.season ?? null)

const roundStartDate = computed(() => {
  if (!round.value) return null
  const times = (round.value.matches as Array<{ startTime?: string | null }>)
    .map(m => m.startTime)
    .filter((t): t is string => !!t)
    .map((t: string) => new Date(t))
  if (!times.length) return null
  const earliest = new Date(Math.min(...times.map((t: Date) => t.getTime())))
  const day = earliest.getDate()
  const month = earliest.toLocaleDateString('en-AU', { month: 'short' })
  const year = String(earliest.getFullYear()).slice(-2)
  return `${day} ${month} '${year}`
})

const aflRoundTo = computed(() => {
  const aflRoundId = round.value?.aflRoundId
  if (!aflRoundId || !aflSeasonId.value) return null
  return { name: 'afl-round', params: { seasonId: aflSeasonId.value, roundId: aflRoundId } }
})

const breadcrumbs = computed(() => {
  if (!season.value) return []
  return [
    { label: 'FFL' },
    { label: season.value.name, to: { name: 'home' } },
  ]
})

const myMatch = computed(() => {
  if (!round.value || !selectedClubId.value) return null
  return round.value.matches.find((m: { homeClubMatch?: { club: { id: string } } | null; awayClubMatch?: { club: { id: string } } | null }) =>
    m.homeClubMatch?.club.id === selectedClubId.value ||
    m.awayClubMatch?.club.id === selectedClubId.value
  ) ?? null
})

interface PlayerMatch {
  player: { aflPlayer: { name: string } }
  position: string | null
  status: string | null
  score: number
}

interface ClubMatch {
  club: { name: string }
  playerMatches: PlayerMatch[]
}

interface Match {
  homeClubMatch?: ClubMatch | null
  awayClubMatch?: ClubMatch | null
}

const POSITION_LABELS: Record<string, string> = {
  goals: 'Goals', kicks: 'Kicks', handballs: 'Handballs',
  marks: 'Marks', tackles: 'Tackles', hitouts: 'Hitouts', star: 'Star',
}

const TOP_SCORERS_POSITIONS = ['goals', 'kicks', 'handballs', 'marks', 'tackles', 'hitouts', 'star'] as const

type ScorerEntry = { name: string; club: string; score: number; position: string }

const topScorersByPosition = computed(() => {
  if (!round.value) return {} as Record<string, { label: string; players: ScorerEntry[] }>

  const grouped: Record<string, ScorerEntry[]> = {}
  for (const match of round.value.matches as Match[]) {
    for (const side of [match.homeClubMatch, match.awayClubMatch]) {
      if (!side) continue
      for (const pm of side.playerMatches) {
        if (pm.status === 'played' && pm.position) {
          ;(grouped[pm.position] ??= []).push({ name: pm.player.aflPlayer.name, club: side.club.name, score: pm.score, position: pm.position })
        }
      }
    }
  }

  const result: Record<string, { label: string; players: ScorerEntry[] }> = {}

  for (const pos of TOP_SCORERS_POSITIONS) {
    if (grouped[pos]?.length) {
      result[pos] = {
        label: POSITION_LABELS[pos],
        players: grouped[pos].sort((a, b) => b.score - a.score),
      }
    }
  }
  return result
})
</script>
