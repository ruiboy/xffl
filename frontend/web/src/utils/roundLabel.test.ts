import { describe, it, expect } from 'vitest'
import { roundPillLabel } from './roundLabel'

describe('roundPillLabel', () => {
  it.each([
    ['Round 1', '1'],
    ['Round 24', '24'],
    ['Opening Round', '0'],
    ['Qualifying Final', 'QF'],
    ['Elimination Final', 'EF'],
    ['Semi Final', 'SF'],
    ['Preliminary Final', 'PF'],
    ['Grand Final', 'GF'],
    ['Wildcard Round', 'WR'],
  ])('%s → %s', (input, expected) => {
    expect(roundPillLabel(input)).toBe(expected)
  })
})
