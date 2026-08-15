import { clubColorRgba } from './clubColors'

export type RoundKind = 'win' | 'loss' | 'draw' | 'bye' | 'superbye' | 'pending'

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

// Result is a fixed status scale (win/loss/draw/bye), reserved meaning —
// never reassigned to identify a club. Matches the color already used for
// win/loss/bye badges elsewhere (StatusBadge.vue, MatchSideHeader.vue).
export const RESULT_COLORS: Record<RoundKind, string> = {
  win: '#22c55e',
  loss: '#ef4444',
  draw: '#94a3b8',
  bye: '#8b5cf6',
  superbye: '#8b5cf6',
  pending: '#475569',
}

// Club identity palette — 8 fixed hues in a CVD-validated order (never
// cycled, never reassigned by rank). Slots 6 (green) and 8 (red) sit close
// in hue to the win/loss result colors above; fine while the league has
// ≤5 clubs (only slots 1-4 in use), worth revisiting if it grows past that.
export const CATEGORICAL_PALETTE: { light: string; dark: string }[] = [
  { light: '#2a78d6', dark: '#3987e5' }, // blue
  { light: '#eb6834', dark: '#d95926' }, // orange
  { light: '#1baf7a', dark: '#199e70' }, // aqua
  { light: '#eda100', dark: '#c98500' }, // yellow
  { light: '#e87ba4', dark: '#d55181' }, // magenta
  { light: '#008300', dark: '#008300' }, // green
  { light: '#4a3aa7', dark: '#9085e9' }, // violet
  { light: '#e34948', dark: '#e66767' }, // red
]

// The named club colors from clubColors.ts (the canonical identity used
// elsewhere, e.g. match-row tints) take priority — chosen identity beats an
// arbitrary CVD-safe slot. clubColorRgba's own dark-mode boost (built for
// exactly this: a dark color like navy going near-invisible on a dark
// surface) is reused here at full opacity for a legible line. Clubs with no
// defined color fall back to the automatic index-based palette slot.
export function clubColor(clubName: string, index: number, isDark: boolean): string {
  const named = clubColorRgba(clubName, 1, isDark)
  if (named) return named
  const slot = CATEGORICAL_PALETTE[index % CATEGORICAL_PALETTE.length]
  return isDark ? slot.dark : slot.light
}

export function chartAxisColors(isDark: boolean): { grid: string; tick: string } {
  return {
    grid: isDark ? '#334155' : '#e2e8f0',
    tick: isDark ? '#94a3b8' : '#64748b',
  }
}

export function chartSurfaceColor(isDark: boolean): string {
  return isDark ? '#030712' : '#f9fafb'
}
