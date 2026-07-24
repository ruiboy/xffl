// Per-club accent colours, used to tint match rows. Defaults are each club's
// canonical guernsey/logo colour (a single distinctive hue per club). Unknown
// clubs return null and get no tint.
const CLUB_COLORS: Record<string, string> = {
  'Adelaide Crows': '#002b5c', // navy
  'Brisbane Lions': '#7c2529', // maroon
  'Carlton Blues': '#001e3c', // navy blue
  'Collingwood Magpies': '#c8cdd0', // black & white → light grey (black is invisible on a dark row)
  'Essendon Bombers': '#cc2031', // red
  'Fremantle Dockers': '#4a1a7b', // purple
  'Geelong Cats': '#041e42', // navy
  'Gold Coast Suns': '#d6001c', // red
  'Greater Western Sydney Giants': '#f26522', // orange
  'Hawthorn Hawks': '#7a4a21', // brown & gold
  'Melbourne Demons': '#0a1d3b', // navy
  'North Melbourne Kangaroos': '#003da5', // royal blue
  'Port Adelaide Power': '#008ea0', // teal
  'Richmond Tigers': '#ffd200', // yellow
  'St Kilda Saints': '#ed0f05', // red
  'Sydney Swans': '#e1251b', // red
  'West Coast Eagles': '#002f87', // blue
  'Western Bulldogs': '#14377d', // blue
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
