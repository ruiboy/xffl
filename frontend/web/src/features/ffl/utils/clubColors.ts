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
// or null if the club has no colour defined. Colours that would be nearly invisible
// against the current theme's row background (e.g. navy in dark mode, pale colours
// in light mode) are nudged toward the midpoint so every club's tint stays visible
// in both themes.
export function clubColorRgba(name: string | undefined | null, alpha: number, isDark = false): string | null {
  const hex = clubColor(name)
  if (!hex) return null
  let r = parseInt(hex.slice(1, 3), 16)
  let g = parseInt(hex.slice(3, 5), 16)
  let b = parseInt(hex.slice(5, 7), 16)
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255
  if (isDark && luminance < 0.35) {
    const boost = 0.4
    r += (255 - r) * boost; g += (255 - g) * boost; b += (255 - b) * boost
  } else if (!isDark && luminance > 0.75) {
    const dim = 0.4
    r *= (1 - dim); g *= (1 - dim); b *= (1 - dim)
  }
  return `rgba(${Math.round(r)}, ${Math.round(g)}, ${Math.round(b)}, ${alpha})`
}
