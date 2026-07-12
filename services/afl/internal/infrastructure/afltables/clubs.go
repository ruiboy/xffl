package afltables

// afltables uses short club names ("Essendon", and "Kangaroos" for North
// Melbourne's 1999–2007 branding). The rest of the system uses the fuller
// canonical names seeded from the live source ("Essendon Bombers"). This map
// translates afltables → canonical so the historical import attaches to the
// existing club records rather than creating duplicates.
var canonicalClubs = map[string]string{
	"Adelaide":               "Adelaide Crows",
	"Carlton":                "Carlton Blues",
	"Collingwood":            "Collingwood Magpies",
	"Essendon":               "Essendon Bombers",
	"Fremantle":              "Fremantle Dockers",
	"Geelong":                "Geelong Cats",
	"Gold Coast":             "Gold Coast Suns",
	"Greater Western Sydney": "Greater Western Sydney Giants",
	"Hawthorn":               "Hawthorn Hawks",
	"Melbourne":              "Melbourne Demons",
	"Kangaroos":              "North Melbourne Kangaroos",
	"North Melbourne":        "North Melbourne Kangaroos",
	"Port Adelaide":          "Port Adelaide Power",
	"Richmond":               "Richmond Tigers",
	"St Kilda":               "St Kilda Saints",
	"Sydney":                 "Sydney Swans",
	"West Coast":             "West Coast Eagles",
	// "Brisbane Lions" and "Western Bulldogs" already match — no entry needed.
}

// CanonicalClub returns the canonical club name for an afltables club name,
// or the input unchanged if it already matches.
func CanonicalClub(name string) string {
	if c, ok := canonicalClubs[name]; ok {
		return c
	}
	return name
}
