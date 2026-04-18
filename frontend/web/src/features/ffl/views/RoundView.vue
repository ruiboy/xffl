<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="data">
      <Breadcrumb :items="breadcrumbs" />

      <h1 class="text-2xl font-bold mb-6">
        {{ data.round.name }}<span v-if="roundStartDate" class="font-normal text-text-faint"> · {{ roundStartDate }}</span>
      </h1>

      <RoundNav
        class="mb-8"
        :rounds="data.season.rounds"
        :live-round-id="liveRoundId"
        :season-id="props.seasonId"
      />

      <section class="mb-8">
        <h2 class="text-lg font-semibold text-text-heading mb-3">Matches</h2>
        <div class="space-y-2">
          <MatchSummary
            v-for="match in data.round.matches"
            :key="match.id"
            :match="match"
            :to="{ name: 'ffl-match', params: { seasonId: props.seasonId, matchId: match.id } }"
            :my-club-id="selectedClubId ?? undefined"
            :build-team-to="myMatch?.id === match.id ? { name: 'ffl-team-builder', params: { seasonId: props.seasonId, roundId: props.roundId } } : undefined"
          />
        </div>
      </section>

      <section v-if="topScorers.length > 0" class="mb-8">
        <h2 class="text-lg font-semibold text-text-heading mb-3">Top Fantasy Scorers</h2>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border text-left text-text-muted">
                <th class="py-2 pr-4 font-medium w-8">#</th>
                <th class="py-2 pr-4 font-medium">Player</th>
                <th class="py-2 px-2 font-medium">Club</th>
                <th class="py-2 px-2 font-medium">Position</th>
                <th class="py-2 px-2 font-medium text-right">Score</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(player, index) in topScorers"
                :key="index"
                class="border-b border-border-subtle hover:bg-surface-hover"
              >
                <td class="py-2 pr-4 tabular-nums text-text-faint">{{ index + 1 }}</td>
                <td class="py-2 pr-4 font-medium">{{ player.name }}</td>
                <td class="py-2 px-2 text-text-muted">{{ player.club }}</td>
                <td class="py-2 px-2 text-text-muted capitalize">{{ player.position ?? '—' }}</td>
                <td class="py-2 px-2 text-right tabular-nums font-semibold">{{ player.score }}</td>
              </tr>
            </tbody>
          </table>
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
import { GET_FFL_SEASON } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import { useAflState } from '../../afl/composables/useAflState'
import Breadcrumb from '../components/Breadcrumb.vue'
import MatchSummary from '../components/MatchSummary.vue'
import RoundNav from '../components/RoundNav.vue'

const props = defineProps<{ seasonId: string; roundId: string }>()

const { liveRoundId, selectedClubId } = useFflState()
const { liveSeasonId: aflSeasonId } = useAflState()
const { result, loading, error } = useQuery(GET_FFL_SEASON, () => ({ id: props.seasonId }))

const data = computed(() => {
  const season = result.value?.fflSeason
  if (!season) return null
  const round = season.rounds.find((r: { id: string }) => r.id === props.roundId)
  if (!round) return null
  return { season, round }
})

const roundStartDate = computed(() => {
  if (!data.value) return null
  const times = (data.value.round.matches as Array<{ startTime?: string | null }>)
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
  const aflRoundId = data.value?.round.aflRoundId
  if (!aflRoundId || !aflSeasonId.value) return null
  return { name: 'afl-round', params: { seasonId: aflSeasonId.value, roundId: aflRoundId } }
})

const breadcrumbs = computed(() => {
  if (!data.value) return []
  return [
    { label: 'FFL' },
    { label: data.value.season.name, to: { name: 'home' } },
  ]
})

const myMatch = computed(() => {
  if (!data.value || !selectedClubId.value) return null
  return data.value.round.matches.find((m: { homeClubMatch?: { club: { id: string } } | null; awayClubMatch?: { club: { id: string } } | null }) =>
    m.homeClubMatch?.club.id === selectedClubId.value ||
    m.awayClubMatch?.club.id === selectedClubId.value
  ) ?? null
})

interface PlayerMatch {
  player: { name: string }
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

const topScorers = computed(() => {
  if (!data.value) return []

  const all: { name: string; club: string; position: string | null; score: number }[] = []
  for (const match of data.value.round.matches as Match[]) {
    for (const side of [match.homeClubMatch, match.awayClubMatch]) {
      if (!side) continue
      for (const pm of side.playerMatches) {
        if (pm.status === 'played') {
          all.push({ name: pm.player.name, club: side.club.name, position: pm.position, score: pm.score })
        }
      }
    }
  }

  return all.sort((a, b) => b.score - a.score).slice(0, 10)
})
</script>
