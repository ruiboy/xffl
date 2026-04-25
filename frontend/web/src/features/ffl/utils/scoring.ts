export const POSITION_LETTERS: Record<string, string> = {
  goals:     'G',
  kicks:     'K',
  handballs: 'H',
  marks:     'M',
  tackles:   'T',
  hitouts:   'R',
  star:      '★',
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
