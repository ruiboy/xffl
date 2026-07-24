// Shared logic for the "on-field team size" indicator shown next to a team's score.

// A full FFL team is 18 on-field players — the denominator for the played count.
export const TEAM_SIZE = 18

export interface CountablePlayerMatch {
  status?: string | null
  aflStatus?: string | null
  backupPositions?: string | null
  interchangePosition?: string | null
}

export interface CountableClubMatch {
  dataStatus?: string | null
  playerMatches?: CountablePlayerMatch[] | null
}

// playedCount counts a team's on-field contributors — 18 is a full team. It counts
// played or bye starters plus any bench player subbed/interchanged on, and excludes
// un-activated bench and DNPs.
export function playedCount(playerMatches: CountablePlayerMatch[] | null | undefined): number {
  if (!playerMatches) return 0
  return playerMatches.filter((pm) => {
    if (pm.status === 'subbed_in' || pm.status === 'interchanged_in') return true // activated bench
    if (pm.backupPositions != null || pm.interchangePosition != null) return false // un-activated bench
    return pm.aflStatus === 'played' || pm.aflStatus === 'bye'
  }).length
}

// progressCount is the count to display as an in-progress indicator (rendered with a
// trailing *). It returns 0 once the side is final — the team is settled, so callers
// drop the indicator — and 0 when there is nothing to show.
export function progressCount(clubMatch: CountableClubMatch | null | undefined): number {
  if (!clubMatch || clubMatch.dataStatus === 'final') return 0
  return playedCount(clubMatch.playerMatches)
}
