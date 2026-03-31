<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="data">
      <h1 class="text-2xl font-bold mb-1">{{ data.round.name }}</h1>
      <p class="text-text-muted mb-6">{{ data.season.name }}</p>

      <section class="mb-8">
        <h2 class="text-lg font-semibold text-text-heading mb-3">Matches</h2>
        <div class="space-y-2">
          <MatchSummary
            v-for="match in data.round.matches"
            :key="match.id"
            :match="match"
            :to="{ name: 'ffl-match', params: { seasonId: props.seasonId, matchId: match.id } }"
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

      <section>
        <h2 class="text-lg font-semibold text-text-heading mb-3">Rounds</h2>
        <RoundNav
          :rounds="data.season.rounds"
          :current-round-id="data.round.id"
          :season-id="props.seasonId"
        />
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_SEASON } from '../api/queries'
import MatchSummary from '../components/MatchSummary.vue'
import RoundNav from '../components/RoundNav.vue'

const props = defineProps<{ seasonId: string; roundId: string }>()

const { result, loading, error } = useQuery(GET_FFL_SEASON, () => ({ id: props.seasonId }))

const data = computed(() => {
  const season = result.value?.fflSeason
  if (!season) return null
  const round = season.rounds.find((r: { id: string }) => r.id === props.roundId)
  if (!round) return null
  return { season, round }
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
