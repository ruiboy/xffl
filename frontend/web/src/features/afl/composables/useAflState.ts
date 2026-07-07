import { ref, readonly, computed } from 'vue'
import { getCookie, setCookie } from '@/utils/cookie'

interface AflState {
  seasonId: string
  roundId: string
  startDate: string
}

const COOKIE_NAME = 'xffl_afl'

function readCookie(): AflState {
  const raw = getCookie(COOKIE_NAME)
  if (!raw) return { seasonId: '', roundId: '', startDate: '' }
  try {
    const parsed = JSON.parse(raw)
    return {
      seasonId: parsed.seasonId ?? '',
      roundId: parsed.roundId ?? '',
      startDate: parsed.startDate ?? '',
    }
  } catch {
    return { seasonId: '', roundId: '', startDate: '' }
  }
}

// Module-level singletons — shared across all component instances
const stored = readCookie()
const liveSeasonId = ref<string>(stored.seasonId)
const liveRoundId = ref<string>(stored.roundId)
const liveStartDate = ref<string>(stored.startDate)

function setLiveRound(seasonId: string, roundId: string, startDate: string) {
  liveSeasonId.value = seasonId
  liveRoundId.value = roundId
  liveStartDate.value = startDate
  setCookie(COOKIE_NAME, JSON.stringify({ seasonId, roundId, startDate }))
}

// In-memory only (not persisted) — sticks to whichever round the user last
// navigated to, falling back to the live round. Resets on page reload.
const selectedRoundOverride = ref<string>('')
const selectedRoundId = computed(() => selectedRoundOverride.value || liveRoundId.value)

function setSelectedRound(roundId: string) {
  selectedRoundOverride.value = roundId
}

export function useAflState() {
  return {
    liveSeasonId: readonly(liveSeasonId),
    liveRoundId: readonly(liveRoundId),
    liveStartDate: readonly(liveStartDate),
    selectedRoundId,
    setLiveRound,
    setSelectedRound,
  }
}
