export interface AflClubMatchInfo {
  matchId: string
  statsImportStatus: string
  seasonId: string
}

export type AflClubMatchMap = Record<string, AflClubMatchInfo>

export interface AflRoundForMap {
  id: string
  season: { id: string }
  matches: Array<{
    id: string
    statsImportStatus: string
    homeClubMatch: { club: { id: string } } | null
    awayClubMatch: { club: { id: string } } | null
  }>
}

export function buildAflClubMatchMap(aflRound: AflRoundForMap | null | undefined): AflClubMatchMap {
  if (!aflRound) return {}
  const map: AflClubMatchMap = {}
  for (const match of aflRound.matches) {
    const info: AflClubMatchInfo = { matchId: match.id, statsImportStatus: match.statsImportStatus, seasonId: aflRound.season.id }
    if (match.homeClubMatch?.club?.id) map[match.homeClubMatch.club.id] = info
    if (match.awayClubMatch?.club?.id) map[match.awayClubMatch.club.id] = info
  }
  return map
}

export function derivePlayerStatus(statsImportStatus: string | null | undefined, score: number | null | undefined): 'played' | 'dnp' | 'named' {
  if (!statsImportStatus || statsImportStatus === 'no_data') return 'named'
  return score != null ? 'played' : 'dnp'
}

export function showScore(statsImportStatus: string | null | undefined, score: number | null | undefined): boolean {
  return score != null && !!statsImportStatus && statsImportStatus !== 'no_data'
}

export function aflMatchRoute(info: AflClubMatchInfo | null | undefined): { name: string; params: { seasonId: string; matchId: string } } | null {
  if (!info) return null
  return { name: 'afl-match', params: { seasonId: info.seasonId, matchId: info.matchId } }
}
