package afltables

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openFixture(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })
	return f
}

func TestParseSeasonIndex(t *testing.T) {
	paths, err := ParseSeasonIndex(openFixture(t, "season_index.html"))
	require.NoError(t, err)
	require.Len(t, paths, 4)
	assert.Equal(t, "stats/games/2023/031420230316.html", paths[0])
	assert.Equal(t, "stats/games/2023/131920230318.html", paths[3])
}

func TestParseGamePage_Metadata(t *testing.T) {
	g, err := ParseGamePage(openFixture(t, "game_stats.html"))
	require.NoError(t, err)

	assert.Equal(t, "1", g.Round)
	assert.Equal(t, "M.C.G.", g.Venue)
	assert.Equal(t, "2023-03-16", g.Date)
	assert.Equal(t, "Richmond", g.HomeClub)
	assert.Equal(t, "Carlton", g.AwayClub)
}

func TestParseGamePage_Players(t *testing.T) {
	g, err := ParseGamePage(openFixture(t, "game_stats.html"))
	require.NoError(t, err)

	// Fixture keeps 3 players per club (6 total); real pages have ~22 each.
	require.Len(t, g.Players, 6)

	// First Richmond player: name flipped, empty stats are zero.
	baker := g.Players[0]
	assert.Equal(t, "Richmond", baker.Club)
	assert.Equal(t, "Liam Baker", baker.Name)
	assert.Equal(t, 10, baker.Kicks)
	assert.Equal(t, 5, baker.Marks)
	assert.Equal(t, 5, baker.Handballs)
	assert.Equal(t, 0, baker.Goals)   // blank cell
	assert.Equal(t, 0, baker.Behinds) // blank cell
	assert.Equal(t, 0, baker.Hitouts) // blank cell
	assert.Equal(t, 1, baker.Tackles)

	// A scoring player with goals/behinds populated.
	bolton := g.Players[2]
	assert.Equal(t, "Shai Bolton", bolton.Name)
	assert.Equal(t, 15, bolton.Kicks)
	assert.Equal(t, 1, bolton.Goals)
	assert.Equal(t, 1, bolton.Behinds)
	assert.Equal(t, 4, bolton.Tackles)

	// Second club's players carry the away club name.
	assert.Equal(t, "Carlton", g.Players[3].Club)
}

func TestParseGamePage_SkipsTotalsRow(t *testing.T) {
	g, err := ParseGamePage(openFixture(t, "game_stats.html"))
	require.NoError(t, err)
	for _, p := range g.Players {
		assert.NotEqual(t, "Totals", p.Name)
		assert.NotEmpty(t, p.Name)
	}
}

func TestFlipName(t *testing.T) {
	assert.Equal(t, "Liam Baker", flipName("Baker, Liam"))
	assert.Equal(t, "Nic Naitanui", flipName("Naitanui, Nic"))
	assert.Equal(t, "Cher", flipName("Cher")) // no comma → unchanged
}

func TestParseDate(t *testing.T) {
	assert.Equal(t, "2023-03-16", parseDate("Thu, 16-Mar-2023 7:20 PM (6:20 PM)"))
	assert.Equal(t, "1998-08-01", parseDate("Sat, 1-Aug-1998 2:10 PM"))
}
