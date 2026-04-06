import { ref, readonly } from 'vue'

function getCookie(name: string): string {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? decodeURIComponent(match[2]) : ''
}

function setCookie(name: string, value: string) {
  const expires = new Date()
  expires.setDate(expires.getDate() + 30)
  document.cookie = `${name}=${encodeURIComponent(value)};expires=${expires.toUTCString()};path=/`
}

// Module-level singletons — shared across all component instances
const selectedClubId = ref<string>(getCookie('xffl_club_id'))
const currentSeasonId = ref<string>(getCookie('xffl_season_id'))
const currentRoundId = ref<string>(getCookie('xffl_round_id'))

function setClub(id: string) {
  selectedClubId.value = id
  setCookie('xffl_club_id', id)
}

function setCurrentSeason(seasonId: string, roundId: string) {
  currentSeasonId.value = seasonId
  setCookie('xffl_season_id', seasonId)
  currentRoundId.value = roundId
  setCookie('xffl_round_id', roundId)
}

export function useFflState() {
  return {
    selectedClubId: readonly(selectedClubId),
    currentSeasonId: readonly(currentSeasonId),
    currentRoundId: readonly(currentRoundId),
    setClub,
    setCurrentSeason,
  }
}
