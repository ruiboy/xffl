import { POSITION_MULTIPLIERS } from './position'

export const LAST_N = 5

export interface StatSummary {
  goals: number
  kicks: number
  handballs: number
  marks: number
  tackles: number
  hitouts: number
}

export type StatKey = 'kicks' | 'handballs' | 'marks' | 'tackles' | 'hitouts' | 'goals'

export const statCols: { key: StatKey; label: string }[] = [
  { key: 'kicks',     label: 'K' },
  { key: 'handballs', label: 'H' },
  { key: 'marks',     label: 'M' },
  { key: 'tackles',   label: 'T' },
  { key: 'hitouts',   label: 'R' },
  { key: 'goals',     label: 'G' },
]

// Star position excludes hitouts: goals×5 + kicks + handballs + marks×2 + tackles×4
export function starScore(s: StatSummary): number {
  return s.goals * POSITION_MULTIPLIERS.goals +
    s.kicks * POSITION_MULTIPLIERS.kicks +
    s.handballs * POSITION_MULTIPLIERS.handballs +
    s.marks * POSITION_MULTIPLIERS.marks +
    s.tackles * POSITION_MULTIPLIERS.tackles
}

export function fmtStat(val: number | null | undefined): string {
  if (val == null) return '—'
  return val.toFixed(1)
}

// Trend thresholds: form must deviate from season by at least TREND_PCT of the
// season value, and by at least TREND_MIN_ABS absolute, to count as a trend.
export const TREND_PCT = 0.10
export const TREND_MIN_ABS = 0.5

// Direction of form (last-N) vs season average, or null when within thresholds.
export function trendDir(form: number, season: number): 'up' | 'down' | null {
  const threshold = Math.max(TREND_MIN_ABS, TREND_PCT * Math.abs(season))
  if (form - season >= threshold) return 'up'
  if (season - form >= threshold) return 'down'
  return null
}
