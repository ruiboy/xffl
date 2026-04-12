import { ref, readonly } from 'vue'

interface AflState {
  seasonId: string
  roundId: string
  roundStatus: string
}

const COOKIE_NAME = 'xffl_afl'

function getCookieRaw(name: string): string {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? decodeURIComponent(match[2]) : ''
}

function readCookie(): AflState {
  const raw = getCookieRaw(COOKIE_NAME)
  if (!raw) return { seasonId: '', roundId: '', roundStatus: '' }
  try {
    const parsed = JSON.parse(raw)
    return {
      seasonId: parsed.seasonId ?? '',
      roundId: parsed.roundId ?? '',
      roundStatus: parsed.roundStatus ?? '',
    }
  } catch {
    return { seasonId: '', roundId: '', roundStatus: '' }
  }
}

function writeCookie(state: AflState) {
  const expires = new Date()
  expires.setDate(expires.getDate() + 30)
  document.cookie = `${COOKIE_NAME}=${encodeURIComponent(JSON.stringify(state))};expires=${expires.toUTCString()};path=/`
}

// Module-level singletons — shared across all component instances
const stored = readCookie()
const liveSeasonId = ref<string>(stored.seasonId)
const liveRoundId = ref<string>(stored.roundId)
const liveRoundStatus = ref<string>(stored.roundStatus)

function setLiveRound(seasonId: string, roundId: string, roundStatus: string) {
  liveSeasonId.value = seasonId
  liveRoundId.value = roundId
  liveRoundStatus.value = roundStatus
  writeCookie({ seasonId, roundId, roundStatus })
}

export function useAflState() {
  return {
    liveSeasonId: readonly(liveSeasonId),
    liveRoundId: readonly(liveRoundId),
    liveRoundStatus: readonly(liveRoundStatus),
    setLiveRound,
  }
}
