import { computed, watch } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_AFL_LIVE_ROUND_IDS } from '@/features/afl/api/queries'
import { GET_FFL_ROUND_IDS_BY_AFL_ROUND } from '@/features/ffl/api/queries'
import { useAflState } from '@/features/afl/composables/useAflState'
import { useFflState } from '@/features/ffl/composables/useFflState'

/**
 * Fires live round queries when the cached state is absent (cookie expired or
 * first visit). Runs from App.vue so any deep-linked page triggers the
 * bootstrap rather than only the home views.
 */
export function useLiveRoundBootstrap() {
  const { liveRoundId: aflRoundId, setLiveRound: setAflLiveRound } = useAflState()
  const { liveRoundId: fflRoundId, setLiveRound: setFflLiveRound } = useFflState()

  const aflStale = computed(() => !aflRoundId.value)
  const fflStale = computed(() => !fflRoundId.value)

  // Step 1: resolve the live AFL round IDs when stale
  const { result: aflResult } = useQuery(
    GET_AFL_LIVE_ROUND_IDS,
    undefined,
    () => ({ enabled: aflStale.value }),
  )

  watch(aflResult, (result) => {
    if (!result?.aflLiveRound) return
    const { round, startDate } = result.aflLiveRound
    setAflLiveRound(round.season.id, round.id, startDate)
  })

  const bootstrapAflRoundId = computed(() => aflResult.value?.aflLiveRound?.round?.id ?? null)

  // Step 2: resolve the corresponding FFL round IDs when stale
  const { result: fflResult } = useQuery(
    GET_FFL_ROUND_IDS_BY_AFL_ROUND,
    () => ({ aflRoundId: bootstrapAflRoundId.value }),
    () => ({ enabled: !!bootstrapAflRoundId.value && fflStale.value }),
  )

  watch(fflResult, (result) => {
    if (!result?.fflRoundByAflRound || !aflResult.value) return
    const round = result.fflRoundByAflRound
    const { startDate } = aflResult.value.aflLiveRound
    setFflLiveRound(round.season.id, round.id, startDate)
  })
}
