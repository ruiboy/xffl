<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="data">
      <Breadcrumb :items="[{ label: 'AFL' }, { label: data.season.name, to: { name: 'afl-home' } }]" />
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
            :to="{ name: 'afl-match', params: { seasonId: props.seasonId, matchId: match.id } }"
          />
        </div>
      </section>

      <section v-if="topPlayerStats.length > 0" class="mb-8">
        <h2 class="text-lg font-semibold text-text-heading mb-4">Top Players</h2>
        <div class="grid grid-cols-2 md:grid-cols-3 gap-6">
          <TopPlayers
            v-for="stat in topPlayerStats"
            :key="stat.key"
            :label="stat.label"
            :players="stat.players"
          />
        </div>
      </section>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_ROUND } from '../api/queries'
import { useAflState } from '../composables/useAflState'
import Breadcrumb from '../components/Breadcrumb.vue'
import MatchSummary from '../components/MatchSummary.vue'
import RoundNav from '../components/RoundNav.vue'
import TopPlayers from '../components/TopPlayers.vue'

const props = defineProps<{ seasonId: string; roundId: string }>()

const { liveRoundId } = useAflState()
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

const roundStartDate = computed(() => {
  if (!data.value) return null
  const times = data.value.round.matches
    .map((m: { startTime?: string | null }) => m.startTime)
    .filter((t): t is string => !!t)
    .map(t => new Date(t))
  if (!times.length) return null
  const earliest = new Date(Math.min(...times.map(t => t.getTime())))
  const day = earliest.getDate()
  const month = earliest.toLocaleDateString('en-AU', { month: 'short' })
  const year = String(earliest.getFullYear()).slice(-2)
  return `${day} ${month} '${year}`
})

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
