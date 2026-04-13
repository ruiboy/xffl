<template>
  <div>
    <div v-if="aflLoading" class="text-text-faint">Loading…</div>
    <div v-else-if="aflError" class="text-red-400">{{ aflError.message }}</div>
    <div v-else-if="fflError" class="text-red-400">{{ fflError.message }}</div>
    <div v-else-if="aflResult && !fflRound" class="text-text-muted">
      Cannot determine round to display. Consult your admin.
    </div>
    <template v-else-if="fflRound">
      <h1 class="text-2xl font-bold mb-4">Home<span class="font-normal text-text-muted"> · {{ fflRound.season.name }}</span></h1>

      <RoundNav
        class="mb-8"
        :rounds="fflRound.season.rounds"
        :live-round-id="liveRoundId"
        :season-id="fflRound.season.id"
      />

      <section>
        <h2 class="text-lg font-semibold text-text-heading mb-3">Ladder</h2>
        <LadderTable :ladder="fflRound.season.ladder" />
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_AFL_LIVE_ROUND, GET_FFL_ROUND_BY_AFL_ROUND } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import LadderTable from '../components/LadderTable.vue'
import RoundNav from '../components/RoundNav.vue'

const { liveRoundId, setLiveRound } = useFflState()

// Step 1: get the live AFL round
const { result: aflResult, loading: aflLoading, error: aflError } = useQuery(GET_AFL_LIVE_ROUND)

const aflRoundId = computed(() => aflResult.value?.aflLiveRound?.round?.id ?? null)

// Step 2: find the FFL round linked to that AFL round (skipped until AFL round is known)
const { result: fflResult, error: fflError } = useQuery(
  GET_FFL_ROUND_BY_AFL_ROUND,
  () => ({ aflRoundId: aflRoundId.value }),
  () => ({ enabled: !!aflRoundId.value }),
)

const fflRound = computed(() => fflResult.value?.fflRoundByAflRound ?? null)

watch([fflRound, () => aflResult.value], ([round, afl]) => {
  if (!round || !afl) return
  setLiveRound(round.season.id, round.id, afl.aflLiveRound.startDate)
})
</script>
