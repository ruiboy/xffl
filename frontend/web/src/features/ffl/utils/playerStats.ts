import { POSITION_MULTIPLIERS } from './position'

export const LAST_N = 3

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
