package afltables_test

import (
	"testing"

	"xffl/services/afl/internal/infrastructure/afltables"
)

func TestClubNameForCode_knownCodes(t *testing.T) {
	cases := []struct {
		code string
		want string
	}{
		{"AD", "Adelaide Crows"},
		{"BL", "Brisbane Lions"},
		{"CA", "Carlton Blues"},
		{"CW", "Collingwood Magpies"},
		{"ES", "Essendon Bombers"},
		{"FR", "Fremantle Dockers"},
		{"GE", "Geelong Cats"},
		{"GC", "Gold Coast Suns"},
		{"GW", "Greater Western Sydney Giants"},
		{"HW", "Hawthorn Hawks"},
		{"ME", "Melbourne Demons"},
		{"NM", "North Melbourne Kangaroos"},
		{"PA", "Port Adelaide Power"},
		{"RI", "Richmond Tigers"},
		{"SK", "St Kilda Saints"},
		{"SY", "Sydney Swans"},
		{"WC", "West Coast Eagles"},
		{"WB", "Western Bulldogs"},
	}

	for _, tc := range cases {
		name, ok := afltables.ClubNameForCode(tc.code)
		if !ok {
			t.Errorf("ClubNameForCode(%q): not found", tc.code)
			continue
		}
		if name != tc.want {
			t.Errorf("ClubNameForCode(%q) = %q, want %q", tc.code, name, tc.want)
		}
	}
}

func TestClubNameForCode_unknownCode(t *testing.T) {
	_, ok := afltables.ClubNameForCode("XX")
	if ok {
		t.Error("ClubNameForCode(\"XX\"): expected not found")
	}
}

func TestClubNameForCode_coverage(t *testing.T) {
	// Ensure exactly 18 clubs are mapped (one per AFL team).
	if got := afltables.ClubCodeCount(); got != 18 {
		t.Errorf("ClubCodeCount() = %d, want 18", got)
	}
}
