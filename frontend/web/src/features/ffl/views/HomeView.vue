<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="data">
      <div class="mb-6">
        <h1 class="text-2xl font-bold">{{ data.season.name }}</h1>
        <p class="text-text-muted">{{ data.round.name }}</p>
      </div>

      <section class="mb-8">
        <h2 class="text-lg font-semibold text-text-heading mb-3">Ladder</h2>
        <LadderTable :ladder="data.season.ladder" />
      </section>

      <section class="mb-8">
        <h2 class="text-lg font-semibold text-text-heading mb-3">Matches</h2>
        <div class="space-y-2">
          <MatchSummary
            v-for="match in data.round.matches"
            :key="match.id"
            :match="match"
            :to="{ name: 'ffl-match', params: { seasonId: data.season.id, matchId: match.id } }"
          />
        </div>
      </section>

      <section>
        <h2 class="text-lg font-semibold text-text-heading mb-3">Rounds</h2>
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
import { computed, watch } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_LATEST_ROUND } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import LadderTable from '../components/LadderTable.vue'
import MatchSummary from '../components/MatchSummary.vue'
import RoundNav from '../components/RoundNav.vue'

const { setCurrentSeason } = useFflState()
const { result, loading, error } = useQuery(GET_FFL_LATEST_ROUND)

const data = computed(() => {
  const round = result.value?.fflLatestRound
  if (!round) return null
  return {
    round: { id: round.id, name: round.name, matches: round.matches },
    season: round.season,
  }
})

watch(data, (d) => {
  if (d) setCurrentSeason(d.season.id, d.round.id)
})
</script>
