import { clubColorRgba } from './clubColors'
import {
  RESULT_COLORS,
  categoricalColor,
  chartAxisColors,
  chartSurfaceColor,
  type RoundKind,
} from '@/utils/chartColors'

export { RESULT_COLORS, chartAxisColors, chartSurfaceColor, type RoundKind }

export interface ClubMatchEntry {
  clubSeasonId: string
  // 'home' | 'away' | 'bye' | 'superbye'
  side: string
  score: number
  dataStatus: string
  club: { name: string }
}

export interface RoundMatch {
  id: string
  clubMatches: ClubMatchEntry[]
}

export interface SeasonRound {
  id: string
  name: string
  matches: RoundMatch[]
}

export interface ClubRoundEntry {
  roundId: string
  roundName: string
  matchId: string | null
  kind: RoundKind
  score: number | null
  opponent: string | null
}

// One entry per round, aligned to `rounds` — a round this club has no data
// for yet (fixture not generated, or not yet final) comes back `pending`
// with a null score, so callers can key a chart's x-axis on `rounds` and
// have every series line up without gaps.
//
// Win/loss/draw is derived by comparing each side's own score directly,
// gated on `dataStatus === 'final'` — not from the match's derived `result`
// field, which can be final (scores locked, counted on the ladder) without
// `result` ever having been backfilled. Comparing scores is always correct
// once both sides are final.
export function deriveClubRoundEntries(rounds: SeasonRound[], clubSeasonId: string): ClubRoundEntry[] {
  return rounds.map((round) => {
    const match = round.matches.find((m) => m.clubMatches.some((cm) => cm.clubSeasonId === clubSeasonId))
    if (!match) {
      return { roundId: round.id, roundName: round.name, matchId: null, kind: 'pending' as const, score: null, opponent: null }
    }

    const own = match.clubMatches.find((cm) => cm.clubSeasonId === clubSeasonId)!
    const isFinal = own.dataStatus === 'final'

    if (own.side === 'bye' || own.side === 'superbye') {
      return {
        roundId: round.id,
        roundName: round.name,
        matchId: match.id,
        kind: isFinal ? (own.side as RoundKind) : 'pending',
        score: isFinal ? own.score : null,
        opponent: null,
      }
    }

    const opp = match.clubMatches.find((cm) => cm.clubSeasonId !== clubSeasonId)
    const played = isFinal && opp?.dataStatus === 'final'

    let kind: RoundKind = 'pending'
    if (played && opp) {
      if (own.score === opp.score) kind = 'draw'
      else kind = own.score > opp.score ? 'win' : 'loss'
    }

    return {
      roundId: round.id,
      roundName: round.name,
      matchId: match.id,
      kind,
      score: kind === 'pending' ? null : own.score,
      opponent: opp?.club.name ?? null,
    }
  })
}

// The named club colors from clubColors.ts (the canonical identity used
// elsewhere, e.g. match-row tints) take priority — chosen identity beats an
// arbitrary CVD-safe slot. clubColorRgba's own dark-mode boost (built for
// exactly this: a dark color like navy going near-invisible on a dark
// surface) is reused here at full opacity for a legible line. Clubs with no
// defined color fall back to the automatic index-based palette slot.
export function clubColor(clubName: string, index: number, isDark: boolean): string {
  const named = clubColorRgba(clubName, 1, isDark)
  if (named) return named
  return categoricalColor(index, isDark)
}
