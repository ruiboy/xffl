package dataops

import "strings"

// Club-name aliasing: fold the different spellings a club is written under — in
// forum threads, spreadsheets, and elsewhere — onto one identity, so importers
// resolve "THC" and "The Howling Cows" to the same club.

// clubAliasGroups lists club-name aliases. Every spelling in a row refers to the
// same club; the first entry is canonical. Add rows as new spellings turn up in
// the source data — matching is case- and whitespace-insensitive.
var clubAliasGroups = [][]string{
	{"THC", "The Howling Cows"},
}

// canonicalClubByNorm maps a normalized alias → its group's canonical normalized form.
var canonicalClubByNorm = func() map[string]string {
	m := make(map[string]string)
	for _, g := range clubAliasGroups {
		canon := normalizeClub(g[0])
		for _, name := range g {
			m[normalizeClub(name)] = canon
		}
	}
	return m
}()

// normalizeClub folds case and surrounding whitespace so spellings line up.
func normalizeClub(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// canonicalClub returns a stable key identifying the club named by any known
// alias. Unknown names fold to their normalized form (identity), so a club with
// no alias still matches itself. Two names are the same club iff their
// canonicalClub values are equal — canonicalise both sides of a lookup.
func canonicalClub(name string) string {
	n := normalizeClub(name)
	if c, ok := canonicalClubByNorm[n]; ok {
		return c
	}
	return n
}
