// Shared logic for the "on-field team size" indicator shown next to a team's score.

// A full FFL team is 18 on-field players — the denominator for the played count.
export const TEAM_SIZE = 18

export interface CountablePlayerMatch {
  status?: string | null
  aflStatus?: string | null
  backupPositions?: string | null
  interchangePosition?: string | null
  position?: string | null
}

export interface CountableClubMatch {
  dataStatus?: string | null
  playerMatches?: CountablePlayerMatch[] | null
}

// contributes reports whether a player counts toward the on-field team: a played or
// bye starter, or a bench player subbed/interchanged on. Un-activated bench and DNPs
// do not count.
function contributes(pm: CountablePlayerMatch): boolean {
  if (pm.status === 'subbed_in' || pm.status === 'interchanged_in') return true // activated bench
  if (pm.backupPositions != null || pm.interchangePosition != null) return false // un-activated bench
  return pm.aflStatus === 'played' || pm.aflStatus === 'bye'
}

// playedCount counts a team's on-field contributors — 18 is a full team.
export function playedCount(playerMatches: CountablePlayerMatch[] | null | undefined): number {
  if (!playerMatches) return 0
  return playerMatches.filter(contributes).length
}

// hasStar reports whether the team's star (the 'star' position) is among the
// counted on-field contributors.
export function hasStar(clubMatch: CountableClubMatch | null | undefined): boolean {
  return (clubMatch?.playerMatches ?? []).some((pm) => pm.position === 'star' && contributes(pm))
}

// progressCount is the count to display as an in-progress indicator (rendered with a
// trailing *). It returns 0 once the side is final — the team is settled, so callers
// drop the indicator — and 0 when there is nothing to show.
export function progressCount(clubMatch: CountableClubMatch | null | undefined): number {
  if (!clubMatch || clubMatch.dataStatus === 'final') return 0
  return playedCount(clubMatch.playerMatches)
}
