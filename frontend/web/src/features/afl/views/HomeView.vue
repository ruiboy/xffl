<template>
  <div>
    <div v-if="loading" class="text-gray-400">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="data">
      <h1 class="text-2xl font-bold mb-1">{{ data.season.name }}</h1>
      <p class="text-gray-500 mb-6">{{ data.round.name }}</p>

      <section class="mb-8">
        <h2 class="text-lg font-semibold text-gray-700 mb-3">Ladder</h2>
        <LadderTable :ladder="data.season.ladder" />
      </section>

      <section class="mb-8">
        <h2 class="text-lg font-semibold text-gray-700 mb-3">Matches</h2>
        <div class="space-y-2">
          <MatchSummary
            v-for="match in data.round.matches"
            :key="match.id"
            :match="match"
            :to="{ name: 'match', params: { seasonId: data.season.id, matchId: match.id } }"
          />
        </div>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-gray-700 mb-3">Rounds</h2>
        <RoundNav
          :rounds="data.season.rounds"
          :current-round-id="data.round.id"
          :season-id="data.season.id"
        />
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_LATEST_ROUND } from '../api/queries'
import LadderTable from '../components/LadderTable.vue'
import MatchSummary from '../components/MatchSummary.vue'
import RoundNav from '../components/RoundNav.vue'

const { result, loading, error } = useQuery(GET_LATEST_ROUND)

const data = computed(() => {
  const round = result.value?.aflLatestRound
  if (!round) return null
  return {
    round: { id: round.id, name: round.name, matches: round.matches },
    season: round.season,
  }
})
</script>
