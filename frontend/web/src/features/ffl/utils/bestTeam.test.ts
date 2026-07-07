import { describe, it, expect } from 'vitest'
import { solveBestAssignment } from './bestTeam'

// Total value of an assignment against its value matrix.
function totalValue(values: number[][], assignment: number[]): number {
  return assignment.reduce((sum, p, slot) => sum + (p >= 0 ? values[slot][p] : 0), 0)
}

// Exhaustive optimal total for small matrices — brute-force all permutations.
function bruteForceBest(values: number[][]): number {
  const nSlots = values.length
  const nPlayers = values[0]?.length ?? 0
  let best = 0
  const used = new Array<boolean>(nPlayers).fill(false)
  const walk = (slot: number, acc: number) => {
    if (slot === nSlots) { best = Math.max(best, acc); return }
    // leave slot empty only when players can run out
    if (nPlayers < nSlots) walk(slot + 1, acc)
    for (let p = 0; p < nPlayers; p++) {
      if (used[p]) continue
      used[p] = true
      walk(slot + 1, acc + values[slot][p])
      used[p] = false
    }
  }
  walk(0, 0)
  return best
}

describe('solveBestAssignment', () => {
  it('assigns each slot its player in the trivial diagonal case', () => {
    const values = [
      [5, 0, 0],
      [0, 5, 0],
      [0, 0, 5],
    ]
    expect(solveBestAssignment(values)).toEqual([0, 1, 2])
  })

  it('uses each player at most once', () => {
    const values = [
      [9, 8],
      [9, 1],
    ]
    const assignment = solveBestAssignment(values)
    const used = assignment.filter(p => p >= 0)
    expect(new Set(used).size).toBe(used.length)
  })

  it('beats greedy: the shared-best player goes where the alternative is worst', () => {
    // Player 0 is best at both slots. Greedy gives slot 0 player 0 (10) and
    // slot 1 player 1 (1) = 11. Optimal is player 1 on slot 0 (8) and player 0
    // on slot 1 (9) = 17.
    const values = [
      [10, 8],
      [9, 1],
    ]
    const assignment = solveBestAssignment(values)
    expect(totalValue(values, assignment)).toBe(17)
  })

  it('matches brute force on a dense 4×6 matrix', () => {
    const values = [
      [7, 3, 9, 2, 4, 8],
      [5, 8, 1, 9, 2, 3],
      [6, 6, 6, 1, 9, 2],
      [2, 9, 4, 8, 3, 7],
    ]
    const assignment = solveBestAssignment(values)
    expect(totalValue(values, assignment)).toBe(bruteForceBest(values))
  })

  it('leaves slots empty when there are fewer players than slots, still optimally', () => {
    const values = [
      [9, 1],
      [8, 7],
      [1, 6],
    ]
    const assignment = solveBestAssignment(values)
    expect(assignment.filter(p => p === -1)).toHaveLength(1)
    expect(totalValue(values, assignment)).toBe(bruteForceBest(values))
  })

  it('handles empty inputs', () => {
    expect(solveBestAssignment([])).toEqual([])
    expect(solveBestAssignment([[]])).toEqual([-1])
  })

  it('is stable on the team-shaped case: 18 slots from 30 players', () => {
    // Deterministic pseudo-random matrix; assert validity + local optimality
    // (no single swap or unused-player substitution improves the total).
    const nSlots = 18
    const nPlayers = 30
    let seed = 42
    const rand = () => (seed = (seed * 1103515245 + 12345) % 2147483648) / 2147483648
    const values = Array.from({ length: nSlots }, () =>
      Array.from({ length: nPlayers }, () => Math.round(rand() * 50)),
    )
    const assignment = solveBestAssignment(values)

    expect(assignment).toHaveLength(nSlots)
    const used = assignment.filter(p => p >= 0)
    expect(used).toHaveLength(nSlots)
    expect(new Set(used).size).toBe(nSlots)

    const total = totalValue(values, assignment)
    const unused = Array.from({ length: nPlayers }, (_, p) => p).filter(p => !used.includes(p))
    for (let a = 0; a < nSlots; a++) {
      // swapping any two slots' players must not improve the total
      for (let b = a + 1; b < nSlots; b++) {
        const swapped = [...assignment]
        ;[swapped[a], swapped[b]] = [swapped[b], swapped[a]]
        expect(totalValue(values, swapped)).toBeLessThanOrEqual(total)
      }
      // substituting any unused player must not improve the total
      for (const p of unused) {
        const subbed = [...assignment]
        subbed[a] = p
        expect(totalValue(values, subbed)).toBeLessThanOrEqual(total)
      }
    }
  })
})
