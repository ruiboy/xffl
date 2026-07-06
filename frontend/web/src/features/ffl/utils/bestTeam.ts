// Optimal team assignment via the Hungarian algorithm (Kuhn–Munkres with
// potentials, O(n²m)). Sizes here are tiny (18 slots × ~30 players), so the
// exact solution is effectively instant.

// Classic assignment solver: cost matrix with rows ≤ cols, returns for each
// row the column it is matched to, minimising total cost.
function hungarian(cost: number[][]): number[] {
  const n = cost.length
  const m = cost[0]?.length ?? 0
  const INF = Number.POSITIVE_INFINITY
  const u = new Array<number>(n + 1).fill(0)
  const v = new Array<number>(m + 1).fill(0)
  const p = new Array<number>(m + 1).fill(0) // p[j] = row matched to column j (1-based)
  const way = new Array<number>(m + 1).fill(0)

  for (let i = 1; i <= n; i++) {
    p[0] = i
    let j0 = 0
    const minv = new Array<number>(m + 1).fill(INF)
    const used = new Array<boolean>(m + 1).fill(false)
    do {
      used[j0] = true
      const i0 = p[j0]
      let delta = INF
      let j1 = 0
      for (let j = 1; j <= m; j++) {
        if (used[j]) continue
        const cur = cost[i0 - 1][j - 1] - u[i0] - v[j]
        if (cur < minv[j]) {
          minv[j] = cur
          way[j] = j0
        }
        if (minv[j] < delta) {
          delta = minv[j]
          j1 = j
        }
      }
      for (let j = 0; j <= m; j++) {
        if (used[j]) {
          u[p[j]] += delta
          v[j] -= delta
        } else {
          minv[j] -= delta
        }
      }
      j0 = j1
    } while (p[j0] !== 0)
    do {
      const j1 = way[j0]
      p[j0] = p[j1]
      j0 = j1
    } while (j0)
  }

  const ans = new Array<number>(n).fill(-1)
  for (let j = 1; j <= m; j++) {
    if (p[j] > 0) ans[p[j] - 1] = j - 1
  }
  return ans
}

// Maximises total value over a [slot][player] value matrix. Returns for each
// slot the chosen player index, or -1 when there are fewer players than slots
// and the slot stays empty. Each player is used at most once.
export function solveBestAssignment(values: number[][]): number[] {
  const nSlots = values.length
  const nPlayers = values[0]?.length ?? 0
  if (nSlots === 0 || nPlayers === 0) return new Array<number>(nSlots).fill(-1)

  if (nSlots <= nPlayers) {
    return hungarian(values.map(row => row.map(v => -v)))
  }

  // Fewer players than slots: solve the transposed problem (every player gets
  // their best slot), then map back.
  const costT: number[][] = []
  for (let pIdx = 0; pIdx < nPlayers; pIdx++) {
    costT.push(values.map(row => -row[pIdx]))
  }
  const playerToSlot = hungarian(costT)
  const ans = new Array<number>(nSlots).fill(-1)
  playerToSlot.forEach((slot, player) => {
    if (slot >= 0) ans[slot] = player
  })
  return ans
}
