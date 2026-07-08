package afltables

import "testing"

func TestCanonicalClub(t *testing.T) {
	cases := map[string]string{
		"Essendon":         "Essendon Bombers",
		"Kangaroos":        "North Melbourne Kangaroos",
		"North Melbourne":  "North Melbourne Kangaroos",
		"West Coast":       "West Coast Eagles",
		"Brisbane Lions":   "Brisbane Lions",   // already canonical
		"Western Bulldogs": "Western Bulldogs", // already canonical
		"Unknown FC":       "Unknown FC",       // passthrough
	}
	for in, want := range cases {
		if got := CanonicalClub(in); got != want {
			t.Errorf("CanonicalClub(%q) = %q, want %q", in, got, want)
		}
	}
}
