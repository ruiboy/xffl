import { ref, readonly, computed } from 'vue'

interface AflState {
  seasonId: string
  roundId: string
  startDate: string
}

const COOKIE_NAME = 'xffl_afl'

function getCookieRaw(name: string): string {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? decodeURIComponent(match[2]) : ''
}

function readCookie(): AflState {
  const raw = getCookieRaw(COOKIE_NAME)
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

function writeCookie(state: AflState) {
  const expires = new Date()
  expires.setHours(24, 0, 0, 0)
  document.cookie = `${COOKIE_NAME}=${encodeURIComponent(JSON.stringify(state))};expires=${expires.toUTCString()};path=/`
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
  writeCookie({ seasonId, roundId, startDate })
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
