/**
 * Short label for a round pill in RoundNav.
 *
 * - Numbered home-and-away rounds show the number: "Round 5" → "5".
 * - The Opening Round shows "0" (it reads as a round-zero before Round 1).
 * - Any other round — the finals — is abbreviated to an initialism of its
 *   name so it fits the pill: "Grand Final" → "GF", "Preliminary Final" → "PF".
 */
export function roundPillLabel(name: string): string {
  if (name === 'Opening Round') return '0'

  const numbered = name.match(/^Round\s+(\d+)$/i)
  if (numbered) return numbered[1]

  return name
    .split(/\s+/)
    .filter(Boolean)
    .map((word) => word[0])
    .join('')
    .toUpperCase()
}
