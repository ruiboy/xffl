import { describe, it, expect } from 'vitest'
import { starScore, trendDir, TREND_PCT, TREND_MIN_ABS, type StatSummary } from './playerStats'
import { POSITION_MULTIPLIERS } from './position'

describe('starScore', () => {
  const s = (over: Partial<StatSummary>): StatSummary => ({
    goals: 0, kicks: 0, handballs: 0, marks: 0, tackles: 0, hitouts: 0, ...over,
  })

  it('weights each stat by its position multiplier', () => {
    expect(starScore(s({ goals: 2 }))).toBe(2 * POSITION_MULTIPLIERS.goals)
    expect(starScore(s({ kicks: 10 }))).toBe(10 * POSITION_MULTIPLIERS.kicks)
    expect(starScore(s({ handballs: 4 }))).toBe(4 * POSITION_MULTIPLIERS.handballs)
    expect(starScore(s({ marks: 3 }))).toBe(3 * POSITION_MULTIPLIERS.marks)
    expect(starScore(s({ tackles: 5 }))).toBe(5 * POSITION_MULTIPLIERS.tackles)
  })

  it('excludes hitouts', () => {
    expect(starScore(s({ hitouts: 40 }))).toBe(0)
  })

  it('sums a full stat line', () => {
    // 2*5 + 10 + 4 + 3*2 + 5*4 = 50
    expect(starScore(s({ goals: 2, kicks: 10, handballs: 4, marks: 3, tackles: 5, hitouts: 40 }))).toBe(50)
  })
})

describe('trendDir', () => {
  it('is null when form equals season', () => {
    expect(trendDir(10, 10)).toBeNull()
  })

  it('fires up exactly at the percentage threshold', () => {
    const season = 10
    const atThreshold = season + season * TREND_PCT
    expect(trendDir(atThreshold, season)).toBe('up')
    expect(trendDir(atThreshold - 0.01, season)).toBeNull()
  })

  it('fires down exactly at the percentage threshold', () => {
    const season = 10
    const atThreshold = season - season * TREND_PCT
    expect(trendDir(atThreshold, season)).toBe('down')
    expect(trendDir(atThreshold + 0.01, season)).toBeNull()
  })

  it('requires the absolute minimum for small season values', () => {
    // 10% of 2 is 0.2, below TREND_MIN_ABS — deviation must reach the absolute floor
    const season = 2
    expect(trendDir(season + TREND_MIN_ABS - 0.01, season)).toBeNull()
    expect(trendDir(season + TREND_MIN_ABS, season)).toBe('up')
    expect(trendDir(season - TREND_MIN_ABS, season)).toBe('down')
  })

  it('is null for zero season and zero form', () => {
    expect(trendDir(0, 0)).toBeNull()
  })

  it('fires on a zero season average once the absolute floor is met', () => {
    expect(trendDir(TREND_MIN_ABS, 0)).toBe('up')
    expect(trendDir(TREND_MIN_ABS - 0.01, 0)).toBeNull()
  })
})
