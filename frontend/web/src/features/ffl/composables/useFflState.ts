import { ref, readonly, computed } from 'vue'
import { getCookie, setCookie } from '@/utils/cookie'

const FFL_COOKIE = 'xffl_ffl'

interface FflState {
  seasonId: string
  roundId: string
  startDate: string
}

function readFflCookie(): FflState {
  const raw = getCookie(FFL_COOKIE)
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
const selectedClubId = ref<string>(getCookie('xffl_club_id'))
const stored = readFflCookie()
const liveSeasonId = ref<string>(stored.seasonId)
const liveRoundId = ref<string>(stored.roundId)
const liveStartDate = ref<string>(stored.startDate)

function setClub(id: string) {
  selectedClubId.value = id
  setCookie('xffl_club_id', id)
}

function setLiveRound(seasonId: string, roundId: string, startDate: string) {
  liveSeasonId.value = seasonId
  liveRoundId.value = roundId
  liveStartDate.value = startDate
  setCookie(FFL_COOKIE, JSON.stringify({ seasonId, roundId, startDate }))
}

// In-memory only (not persisted) — sticks to whichever round the user last
// navigated to, falling back to the live round. Resets on page reload.
const selectedRoundOverride = ref<string>('')
const selectedRoundId = computed(() => selectedRoundOverride.value || liveRoundId.value)

function setSelectedRound(roundId: string) {
  selectedRoundOverride.value = roundId
}

export function useFflState() {
  return {
    selectedClubId: readonly(selectedClubId),
    liveSeasonId: readonly(liveSeasonId),
    liveRoundId: readonly(liveRoundId),
    liveStartDate: readonly(liveStartDate),
    selectedRoundId,
    setClub,
    setLiveRound,
    setSelectedRound,
  }
}
