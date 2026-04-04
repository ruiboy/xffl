<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="data">
      <h1 class="text-2xl font-bold mb-4">{{ data.season.name }}</h1>

      <RoundNav
        class="mb-8"
        :rounds="data.season.rounds"
        :live-round-id="data.round.id"
        :season-id="data.season.id"
      />

      <section>
        <h2 class="text-lg font-semibold text-text-heading mb-3">Ladder</h2>
        <LadderTable :ladder="data.season.ladder" />
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_LATEST_ROUND } from '../api/queries'
import LadderTable from '../components/LadderTable.vue'
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
