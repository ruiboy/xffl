<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <NotFound v-else-if="notFound" entity="Round" />
    <template v-else-if="round">
      <Breadcrumb :items="breadcrumbs" />

      <h1 class="text-2xl font-bold mb-6">
        {{ round.name }}<span v-if="roundStartDate" class="font-normal text-text-faint"> · {{ roundStartDate }}</span>
      </h1>

      <RoundNav
        class="mb-8"
        :rounds="season?.rounds ?? []"
        :live-round-id="liveRoundId"
        :live-start-date="liveStartDate"
      />

      <section class="mb-8">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-lg font-semibold text-text-heading">Matches</h2>
          <router-link
            v-if="round.aflRoundId && round.id === liveRoundId"
            :to="{ name: 'ffl-data-ops', query: { tab: 'afl-stats', round: round.aflRoundId } }"
            class="flex items-center gap-1.5 text-sm font-medium text-text-muted hover:text-text transition-colors"
          >
            <IconDataOps class="w-3.5 h-3.5" />
            Import Stats
          </router-link>
        </div>
        <div class="space-y-2">
          <MatchSummary
            v-for="match in round.matches"
            :key="match.id"
            :match="match"
            :to="{ name: 'ffl-match', params: { matchId: match.id } }"
            :my-club-id="selectedClubId ?? undefined"
            :build-team-to="myClubMatchId && myMatch?.id === match.id ? { name: 'ffl-club-match-edit', params: { clubMatchId: myClubMatchId } } : undefined"
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
                    <router-link :to="{ name: 'ffl-match', params: { matchId: player.matchId }, query: { highlight: player.pmId } }" class="text-sm font-medium truncate hover:text-text-muted transition-colors">{{ player.name }}</router-link>
                  </div>
                  <span class="text-sm tabular-nums font-semibold shrink-0">{{ player.score }}</span>
                </div>
              </div>
            </div>
          </template>
        </div>

      </section>


    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_ROUND } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import { useAflState } from '@/features/afl/composables/useAflState'
import { useNotFound } from '@/composables/useNotFound'
import NotFound from '@/components/NotFound.vue'
import Breadcrumb from '../components/Breadcrumb.vue'
import MatchSummary from '../components/MatchSummary.vue'
import RoundNav from '../components/RoundNav.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import IconDataOps from '@/features/data-ops/components/icons/IconDataOps.vue'

const props = defineProps<{ roundId: string }>()

const { liveRoundId, liveStartDate, selectedClubId, setSelectedRound } = useFflState()
const { setSelectedRound: setAflSelectedRound } = useAflState()
const { result, loading, error } = useQuery(GET_FFL_ROUND, () => ({ id: props.roundId }))

const round = computed(() => result.value?.fflRound ?? null)
const notFound = useNotFound(round, loading, error)
const season = computed(() => round.value?.season ?? null)

// Visiting a round makes it "stick" for cross-domain navigation (header
// links, DataOps) until the page is reloaded. The corresponding AFL round
// sticks too, so switching domains lands on the matching round.
watch(() => props.roundId, (id) => setSelectedRound(id), { immediate: true })

watch(round, (r) => {
  if (r?.aflRoundId) setAflSelectedRound(r.aflRoundId)
})

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

const breadcrumbs = computed(() => {
  if (!season.value) return []
  return [
    { label: 'FFL' },
    { label: season.value.name, to: { name: 'home' } },
  ]
})

type RoundClubMatch = { id: string; club: { id: string } }

const myMatch = computed(() => {
  if (!round.value || !selectedClubId.value) return null
  return round.value.matches.find((m: { clubMatches?: RoundClubMatch[] | null }) =>
    (m.clubMatches ?? []).some((cm) => cm.club.id === selectedClubId.value)
  ) ?? null
})

const myClubMatchId = computed(() => {
  if (!myMatch.value || !selectedClubId.value) return null
  const m = myMatch.value as { clubMatches?: RoundClubMatch[] | null }
  return (m.clubMatches ?? []).find((cm) => cm.club.id === selectedClubId.value)?.id ?? null
})

interface PlayerMatch {
  id: string
  player: { aflPlayer: { name: string } }
  position: string | null
  status: string | null
  aflStatus: string | null
  score: number
}

interface ClubMatch {
  club: { name: string }
  playerMatches: PlayerMatch[]
}

interface Match {
  id: string
  clubMatches?: ClubMatch[] | null
}

const POSITION_LABELS: Record<string, string> = {
  goals: 'Goals', kicks: 'Kicks', handballs: 'Handballs',
  marks: 'Marks', tackles: 'Tackles', hitouts: 'Hitouts', star: 'Star',
}

const TOP_SCORERS_POSITIONS = ['goals', 'kicks', 'handballs', 'marks', 'tackles', 'hitouts', 'star'] as const

type ScorerEntry = { name: string; club: string; score: number; position: string; matchId: string; pmId: string }

const topScorersByPosition = computed(() => {
  if (!round.value) return {} as Record<string, { label: string; players: ScorerEntry[] }>

  const grouped: Record<string, ScorerEntry[]> = {}
  for (const match of round.value.matches as Match[]) {
    for (const side of match.clubMatches ?? []) {
      for (const pm of side.playerMatches) {
        if (pm.aflStatus === 'played' && pm.position) {
          ;(grouped[pm.position] ??= []).push({ name: pm.player.aflPlayer.name, club: side.club.name, score: pm.score, position: pm.position, matchId: match.id, pmId: pm.id })
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
