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

export function starScore(s: StatSummary): number {
  return statCols.reduce((sum, col) => sum + s[col.key] * (POSITION_MULTIPLIERS[col.key] ?? 1), 0)
}

export function fmtStat(val: number | null | undefined): string {
  if (val == null) return '—'
  return val.toFixed(1)
}
