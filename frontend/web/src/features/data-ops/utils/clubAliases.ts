// Folds the different spellings a club is written under onto one identity, so the
// importer UIs treat "THC" and "The Howling Cows" as the same club — mirroring the
// backend `clubalias` package. Keep the two in sync when adding aliases.

// Each row lists aliases of one club; the first entry is canonical.
const GROUPS: string[][] = [
  ['THC', 'The Howling Cows'],
]

const canonicalByNorm = new Map<string, string>()
for (const group of GROUPS) {
  const canon = normalize(group[0])
  for (const name of group) canonicalByNorm.set(normalize(name), canon)
}

function normalize(name: string): string {
  return name.trim().toLowerCase()
}

// Returns a stable key for the club named by any known alias; unknown names fold
// to their normalized form. Two names are the same club iff their keys are equal.
export function canonicalClub(name: string): string {
  const n = normalize(name)
  return canonicalByNorm.get(n) ?? n
}
