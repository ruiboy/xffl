// Experimental per-club accent colours, used to tint match rows. Only a handful
// of clubs have a colour yet; unknown clubs return null and get no tint.
const CLUB_COLORS: Record<string, string> = {
  Ruiboys: '#4169e1', // royal blue
  'The Howling Cows': '#cc5500', // burnt orange (a.k.a. THC)
  THC: '#cc5500', // alias, in case an abbreviated name is used
  Slashers: '#2ecc40', // bright green
  Cheetahs: '#1e3a8a', // navy blue
}

export function clubColor(name: string | undefined | null): string | null {
  if (!name) return null
  return CLUB_COLORS[name] ?? null
}

// clubColorRgba returns the club's colour as an rgba() string at the given alpha,
// or null if the club has no colour defined.
export function clubColorRgba(name: string | undefined | null, alpha: number): string | null {
  const hex = clubColor(name)
  if (!hex) return null
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}
