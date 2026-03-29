<template>
  <div>
    <div v-if="loading" class="text-gray-400">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="data">
      <h1 class="text-2xl font-bold mb-1">{{ data.round.name }}</h1>
      <p class="text-gray-400 mb-6">{{ data.season.name }}</p>

      <section class="mb-8">
        <h2 class="text-lg font-semibold text-gray-300 mb-3">Matches</h2>
        <div class="space-y-2">
          <MatchSummary
            v-for="match in data.round.matches"
            :key="match.id"
            :match="match"
            :to="{ name: 'match', params: { seasonId: props.seasonId, matchId: match.id } }"
          />
        </div>
      </section>

      <section v-if="topPlayerStats.length > 0" class="mb-8">
        <h2 class="text-lg font-semibold text-gray-300 mb-4">Top Players</h2>
        <div class="grid grid-cols-2 md:grid-cols-3 gap-6">
          <TopPlayers
            v-for="stat in topPlayerStats"
            :key="stat.key"
            :label="stat.label"
            :players="stat.players"
          />
        </div>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-gray-300 mb-3">Rounds</h2>
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
import { GET_ROUND } from '../api/queries'
import MatchSummary from '../components/MatchSummary.vue'
import RoundNav from '../components/RoundNav.vue'
import TopPlayers from '../components/TopPlayers.vue'

const props = defineProps<{ seasonId: string; roundId: string }>()

const { result, loading, error } = useQuery(GET_ROUND, () => ({ seasonId: props.seasonId }))

const data = computed(() => {
  const season = result.value?.aflSeason
  if (!season) return null
  const round = season.rounds.find((r: { id: string }) => r.id === props.roundId)
  if (!round) return null
  return { season, round }
})

const statCategories = [
  { key: 'kicks', label: 'Kicks' },
  { key: 'handballs', label: 'Handballs' },
  { key: 'marks', label: 'Marks' },
  { key: 'hitouts', label: 'Hitouts' },
  { key: 'tackles', label: 'Tackles' },
  { key: 'goals', label: 'Goals' },
] as const

interface PlayerMatch {
  player: { name: string }
  kicks: number
  handballs: number
  marks: number
  hitouts: number
  tackles: number
  goals: number
  [key: string]: unknown
}

interface ClubMatch {
  club: { name: string }
  playerMatches: PlayerMatch[]
}

interface Match {
  homeClubMatch?: ClubMatch | null
  awayClubMatch?: ClubMatch | null
}

const topPlayerStats = computed(() => {
  if (!data.value) return []

  const allPlayers: { name: string; club: string; stats: PlayerMatch }[] = []
  for (const match of data.value.round.matches as Match[]) {
    for (const side of [match.homeClubMatch, match.awayClubMatch]) {
      if (!side) continue
      for (const pm of side.playerMatches) {
        allPlayers.push({ name: pm.player.name, club: side.club.name, stats: pm })
      }
    }
  }

  return statCategories.map(cat => {
    const sorted = [...allPlayers]
      .sort((a, b) => (b.stats[cat.key] as number) - (a.stats[cat.key] as number))
      .slice(0, 5)
    return {
      key: cat.key,
      label: cat.label,
      players: sorted.map(p => ({
        name: p.name,
        club: p.club,
        value: p.stats[cat.key] as number,
      })),
    }
  })
})
</script>
