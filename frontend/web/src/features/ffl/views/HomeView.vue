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
import { GET_FFL_LATEST_ROUND, GET_AFL_LIVE_ROUND } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import LadderTable from '../components/LadderTable.vue'
import RoundNav from '../components/RoundNav.vue'

const { liveRoundId, liveRoundStatus, setLiveRound } = useFflState()
const { result, loading, error } = useQuery(GET_FFL_LATEST_ROUND)
const { result: aflResult } = useQuery(GET_AFL_LIVE_ROUND)

const data = computed(() => {
  const round = result.value?.fflLatestRound
  if (!round) return null
  return {
    round: { id: round.id, name: round.name, matches: round.matches },
    season: round.season,
  }
})

watch([data, () => aflResult.value], ([d, afl]) => {
  if (!d || !afl) return
  const aflRoundId = afl.aflLiveRound.round.id
  const status = afl.aflLiveRound.status
  const fflRound = d.season.rounds.find(
    (r: { id: string; aflRoundId?: string | null }) => r.aflRoundId === aflRoundId
  )
  const roundId = fflRound?.id ?? d.round.id
  setLiveRound(d.season.id, roundId, status)
})
</script>
