export interface AflClubMatchInfo {
  matchId: string
  dataStatus: string
}

export type AflClubMatchMap = Record<string, AflClubMatchInfo>

export interface AflRoundForMap {
  id: string
  matches: Array<{
    id: string
    dataStatus: string
    homeClubMatch: { club: { id: string } } | null
    awayClubMatch: { club: { id: string } } | null
  }>
}

export function buildAflClubMatchMap(aflRound: AflRoundForMap | null | undefined): AflClubMatchMap {
  if (!aflRound) return {}
  const map: AflClubMatchMap = {}
  for (const match of aflRound.matches) {
    const info: AflClubMatchInfo = { matchId: match.id, dataStatus: match.dataStatus }
    if (match.homeClubMatch?.club?.id) map[match.homeClubMatch.club.id] = info
    if (match.awayClubMatch?.club?.id) map[match.awayClubMatch.club.id] = info
  }
  return map
}

export function derivePlayerStatus(dataStatus: string | null | undefined, score: number | null | undefined): 'played' | 'dnp' | 'named' {
  if (!dataStatus || dataStatus === 'no_data') return 'named'
  return score != null ? 'played' : 'dnp'
}

export function showScore(dataStatus: string | null | undefined, score: number | null | undefined): boolean {
  return score != null && !!dataStatus && dataStatus !== 'no_data'
}

export function aflMatchRoute(info: AflClubMatchInfo | null | undefined): { name: string; params: { matchId: string } } | null {
  if (!info) return null
  return { name: 'afl-match', params: { matchId: info.matchId } }
}
