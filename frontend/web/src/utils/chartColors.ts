export type RoundKind = 'win' | 'loss' | 'draw' | 'bye' | 'superbye' | 'pending'

// Result is a fixed status scale (win/loss/draw/bye), reserved meaning —
// never reassigned to identify a club. Matches the color already used for
// win/loss/bye badges elsewhere (StatusBadge.vue, MatchSideHeader.vue).
export const RESULT_COLORS: Record<RoundKind, string> = {
  win: '#22c55e',
  loss: '#ef4444',
  draw: '#94a3b8',
  bye: '#8b5cf6',
  superbye: '#8b5cf6',
  pending: '#475569',
}

// Club identity palette — 8 fixed hues in a CVD-validated order (never
// cycled, never reassigned by rank). Slots 6 (green) and 8 (red) sit close
// in hue to the win/loss result colors above; fine while a chart has ≤5
// series (only slots 1-4 in use), worth revisiting if one grows past that.
export const CATEGORICAL_PALETTE: { light: string; dark: string }[] = [
  { light: '#2a78d6', dark: '#3987e5' }, // blue
  { light: '#eb6834', dark: '#d95926' }, // orange
  { light: '#1baf7a', dark: '#199e70' }, // aqua
  { light: '#eda100', dark: '#c98500' }, // yellow
  { light: '#e87ba4', dark: '#d55181' }, // magenta
  { light: '#008300', dark: '#008300' }, // green
  { light: '#4a3aa7', dark: '#9085e9' }, // violet
  { light: '#e34948', dark: '#e66767' }, // red
]

export function categoricalColor(index: number, isDark: boolean): string {
  const slot = CATEGORICAL_PALETTE[index % CATEGORICAL_PALETTE.length]
  return isDark ? slot.dark : slot.light
}

export function chartAxisColors(isDark: boolean): { grid: string; tick: string } {
  return {
    grid: isDark ? '#334155' : '#e2e8f0',
    tick: isDark ? '#94a3b8' : '#64748b',
  }
}

export function chartSurfaceColor(isDark: boolean): string {
  return isDark ? '#030712' : '#f9fafb'
}
