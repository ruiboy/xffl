import { ref, readonly } from 'vue'

const FFL_COOKIE = 'xffl_ffl'

interface FflState {
  seasonId: string
  roundId: string
  startDate: string
}

function getCookie(name: string): string {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? decodeURIComponent(match[2]) : ''
}

function setCookie(name: string, value: string) {
  const expires = new Date()
  expires.setHours(expires.getHours() + 24)
  document.cookie = `${name}=${encodeURIComponent(value)};expires=${expires.toUTCString()};path=/`
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

export function useFflState() {
  return {
    selectedClubId: readonly(selectedClubId),
    liveSeasonId: readonly(liveSeasonId),
    liveRoundId: readonly(liveRoundId),
    liveStartDate: readonly(liveStartDate),
    setClub,
    setLiveRound,
  }
}
