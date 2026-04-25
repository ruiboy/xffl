// Canonical position order for display and grouping.
export const POSITION_ORDER = ['goals', 'kicks', 'handballs', 'marks', 'tackles', 'hitouts', 'star'] as const
export type PositionKey = typeof POSITION_ORDER[number]

export const POSITION_LABEL: Record<string, string> = {
  goals: 'Goals', kicks: 'Kicks', handballs: 'Handballs',
  marks: 'Marks', tackles: 'Tackles', hitouts: 'Hitouts', star: 'Star',
}

export const POSITION_LETTERS: Record<string, string> = {
  goals: 'G', kicks: 'K', handballs: 'H',
  marks: 'M', tackles: 'T', hitouts: 'R', star: '★',
}

export const POSITION_COLORS: Record<string, string> = {
  goals:     'text-orange-400',
  kicks:     'text-blue-400',
  handballs: 'text-emerald-400',
  marks:     'text-purple-400',
  tackles:   'text-red-400',
  hitouts:   'text-cyan-400',
  star:      'text-yellow-400',
}

// Max starter slots per position — mirrors services/ffl/internal/domain/player_match.go
export const POSITION_SLOTS: Record<string, number> = {
  goals: 3, kicks: 4, handballs: 4,
  marks: 2, tackles: 2, hitouts: 2, star: 1,
}

// Scoring multipliers per position — mirrors services/ffl/internal/domain/player_match.go
export const POSITION_MULTIPLIERS: Record<string, number> = {
  goals: 5, kicks: 1, handballs: 1,
  marks: 2, tackles: 4, hitouts: 1, star: 1,
}

// Returns a formula string like "3×5" for positions with multiplier > 1, otherwise null.
export function positionFormula(position: string, score: number): string | null {
  const multiplier = POSITION_MULTIPLIERS[position]
  if (!multiplier || multiplier <= 1) return null
  return `${Math.round(score / multiplier)}×${multiplier}`
}

export interface RoundRef { id: string }
export interface RoundEntry { position: string | null; isBench: boolean }

// Returns the recency-weighted primary position for a player across all rounds.
// Starter appearances in later rounds carry more weight (roundIndex + 1).
// Returns null if the player has no starter appearances.
export function primaryPosition(
  playerSeasonId: string,
  playerRoundMap: Map<string, Map<string, RoundEntry>>,
  rounds: RoundRef[],
): string | null {
  const entries = playerRoundMap.get(playerSeasonId)
  if (!entries) return null
  const tally: Record<string, number> = {}
  rounds.forEach((round, idx) => {
    const e = entries.get(round.id)
    if (!e || e.isBench || !e.position) return
    tally[e.position] = (tally[e.position] ?? 0) + (idx + 1)
  })
  const ranked = Object.entries(tally)
  if (!ranked.length) return null
  return ranked.sort((a, b) => b[1] - a[1])[0][0]
}
