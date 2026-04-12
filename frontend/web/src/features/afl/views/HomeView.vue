<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="data">
      <h1 class="text-2xl font-bold mb-4">Home<span class="font-normal text-text-muted"> · {{ data.season.name }}</span></h1>

      <RoundNav
        class="mb-8"
        :rounds="data.season.rounds"
        :live-round-id="liveRoundId"
        :live-round-status="liveRoundStatus"
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
import { computed, watch } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_AFL_LIVE_ROUND } from '../api/queries'
import { useAflState } from '../composables/useAflState'
import LadderTable from '../components/LadderTable.vue'
import RoundNav from '../components/RoundNav.vue'

const { liveRoundId, liveRoundStatus, setLiveRound } = useAflState()
const { result, loading, error } = useQuery(GET_AFL_LIVE_ROUND)

const data = computed(() => {
  const live = result.value?.aflLiveRound
  if (!live) return null
  return {
    round: { id: live.round.id, name: live.round.name, matches: live.round.matches },
    season: live.round.season,
    status: live.status,
  }
})

watch(data, (d) => {
  if (d) setLiveRound(d.season.id, d.round.id, d.status)
})
</script>
