// Package afltables provides the secondary adapter for fetching AFL stats
// from afltables.com. This file maps the 2-letter team codes used in the
// afltables CSV format to the canonical club names used in the afl.club table.
package afltables

// teamCodes maps AFLTables 2-letter team codes to canonical afl.club names.
// These names must exactly match the values inserted by dev/postgres/seed/01_afl_seed.sql.
var teamCodes = map[string]string{
	"AD": "Adelaide Crows",
	"BL": "Brisbane Lions",
	"CA": "Carlton Blues",
	"CW": "Collingwood Magpies",
	"ES": "Essendon Bombers",
	"FR": "Fremantle Dockers",
	"GE": "Geelong Cats",
	"GC": "Gold Coast Suns",
	"GW": "Greater Western Sydney Giants",
	"HW": "Hawthorn Hawks",
	"ME": "Melbourne Demons",
	"NM": "North Melbourne Kangaroos",
	"PA": "Port Adelaide Power",
	"RI": "Richmond Tigers",
	"SK": "St Kilda Saints",
	"SY": "Sydney Swans",
	"WC": "West Coast Eagles",
	"WB": "Western Bulldogs",
}

// ClubNameForCode returns the canonical club name for a given AFLTables
// 2-letter team code. Returns ("", false) if the code is not recognised.
func ClubNameForCode(code string) (string, bool) {
	name, ok := teamCodes[code]
	return name, ok
}

// ClubCodeCount returns the number of team codes in the mapping.
// Used in tests to verify completeness.
func ClubCodeCount() int {
	return len(teamCodes)
}
