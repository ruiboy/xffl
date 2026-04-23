// Scoring multipliers per position — mirrors services/ffl/internal/domain/player_match.go
export const POSITION_MULTIPLIERS: Record<string, number> = {
  goals:     5,
  kicks:     1,
  handballs: 1,
  marks:     2,
  tackles:   4,
  hitouts:   1,
  star:      1,
}

// Returns a formula string like "3×5" for positions with multiplier > 1, otherwise null.
export function positionFormula(position: string, score: number): string | null {
  const multiplier = POSITION_MULTIPLIERS[position]
  if (!multiplier || multiplier <= 1) return null
  const count = Math.round(score / multiplier)
  return `${count}×${multiplier}`
}
