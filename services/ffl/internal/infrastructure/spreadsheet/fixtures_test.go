package spreadsheet

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
)

func roundByNumber(rounds []application.ParsedFixtureRound, n int) *application.ParsedFixtureRound {
	for i := range rounds {
		if rounds[i].Round == n {
			return &rounds[i]
		}
	}
	return nil
}

func roundByLabel(rounds []application.ParsedFixtureRound, label string) *application.ParsedFixtureRound {
	for i := range rounds {
		if rounds[i].Label == label {
			return &rounds[i]
		}
	}
	return nil
}

func TestParseFixtures(t *testing.T) {
	b, err := os.ReadFile("testdata/fixtures_2025.txt")
	require.NoError(t, err)

	rounds, err := NewFixtureParser().ParseFixtures(context.Background(), string(b))
	require.NoError(t, err)

	t.Run("all 22 home-and-away rounds parsed", func(t *testing.T) {
		for n := 1; n <= 22; n++ {
			r := roundByNumber(rounds, n)
			require.NotNilf(t, r, "round %d missing", n)
			assert.Lenf(t, r.Fixtures, 2, "round %d fixture count", n)
		}
	})

	t.Run("round 1 fixtures and reference scores", func(t *testing.T) {
		r := roundByNumber(rounds, 1)
		require.NotNil(t, r)
		f := r.Fixtures[0]
		assert.Equal(t, "THC", f.HomeClub)
		assert.Equal(t, "Ruiboys", f.AwayClub)
		require.NotNil(t, f.HomeScore)
		require.NotNil(t, f.AwayScore)
		assert.Equal(t, 452, *f.HomeScore)
		assert.Equal(t, 372, *f.AwayScore)

		assert.Equal(t, "Cheetahs", r.Fixtures[1].HomeClub)
		assert.Equal(t, "Slashers", r.Fixtures[1].AwayClub)
	})

	t.Run("n/a placeholder lines are skipped", func(t *testing.T) {
		for _, r := range rounds {
			for _, f := range r.Fixtures {
				assert.NotContains(t, f.HomeClub, "n/a")
				assert.NotContains(t, f.AwayClub, "n/a")
			}
		}
	})

	t.Run("super bye label attaches to round 19", func(t *testing.T) {
		r := roundByNumber(rounds, 19)
		require.NotNil(t, r)
		assert.Equal(t, "Super Bye", r.Label)
	})

	t.Run("grand final fixture", func(t *testing.T) {
		gf := roundByLabel(rounds, "Grand Final")
		require.NotNil(t, gf)
		require.Len(t, gf.Fixtures, 1)
		assert.Equal(t, "Ruiboys", gf.Fixtures[0].HomeClub)
		assert.Equal(t, "THC", gf.Fixtures[0].AwayClub)
		require.NotNil(t, gf.Fixtures[0].HomeScore)
		assert.Equal(t, 366, *gf.Fixtures[0].HomeScore)
	})
}
