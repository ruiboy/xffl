package afltables

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleCSV = `"Player","ID","Team","Opponent","Round","Kicks","Marks","Hand Balls","Disp","Goals","Behinds","Hit Outs","Tackles","Rebounds","Inside 50","Clearances","Clangers","Frees For","Frees Against","Brownlow","Contested Possessions","Uncontested Possessions","Contested Marks","Marks Inside 50","One Percenters","Bounces","Goal Assists","% Time Played"
"Joel Amartey",12844,"SY","CA","1",6,4,1,7,3,1,2,1,0,2,0,2,2,0,0,3,4,0,3,3,0,0,75
"Nick Blakey",12699,"SY","CA","1",12,3,9,21,0,0,0,1,4,5,0,6,0,2,0,5,15,0,0,4,4,1,88
"Jordan Dawson",9876,"AD","CW","0",8,2,5,13,1,2,0,4,3,3,1,3,1,1,0,4,9,0,1,2,1,0,92`

func TestParseCSV_stats(t *testing.T) {
	rows, err := parseCSV(strings.NewReader(sampleCSV), 2026)
	require.NoError(t, err)
	require.Len(t, rows, 3)

	t.Run("external player ID is captured", func(t *testing.T) {
		assert.Equal(t, "12844", rows[0].ExternalPlayerID)
	})
	t.Run("player name is captured", func(t *testing.T) {
		assert.Equal(t, "Joel Amartey", rows[0].PlayerName)
	})
	t.Run("club code is resolved to canonical name", func(t *testing.T) {
		assert.Equal(t, "Sydney Swans", rows[0].ClubName)
	})
	t.Run("round number is mapped to round name", func(t *testing.T) {
		assert.Equal(t, "Round 1", rows[0].RoundName)
	})
	t.Run("season year is set from argument", func(t *testing.T) {
		assert.Equal(t, 2026, rows[0].SeasonYear)
	})
	t.Run("stats fields are parsed correctly", func(t *testing.T) {
		assert.Equal(t, 6, rows[0].Kicks)
		assert.Equal(t, 4, rows[0].Marks)
		assert.Equal(t, 1, rows[0].Handballs)
		assert.Equal(t, 3, rows[0].Goals)
		assert.Equal(t, 1, rows[0].Behinds)
		assert.Equal(t, 2, rows[0].Hitouts)
		assert.Equal(t, 1, rows[0].Tackles)
	})
	t.Run("round 0 maps to Opening Round", func(t *testing.T) {
		assert.Equal(t, "Opening Round", rows[2].RoundName)
	})
}

func TestParseCSV_skipsUnknownClubCode(t *testing.T) {
	csv := header() + "\n" + `"Unknown Player",99999,"XX","SY","1",5,2,3,8,0,0,0,1,0,1,0,2,0,1,0,3,5,0,0,1,0,0,80`
	rows, err := parseCSV(strings.NewReader(csv), 2026)
	require.NoError(t, err)
	assert.Empty(t, rows, "rows with unknown club codes should be skipped")
}

func TestParseCSV_skipsUnknownRound(t *testing.T) {
	csv := header() + "\n" + `"Some Player",11111,"SY","CA","EF",5,2,3,8,0,0,0,1,0,1,0,2,0,1,0,3,5,0,0,1,0,0,80`
	rows, err := parseCSV(strings.NewReader(csv), 2026)
	require.NoError(t, err)
	assert.Empty(t, rows, "rows with unrecognised round codes should be skipped")
}

func TestParseCSV_emptyInput(t *testing.T) {
	rows, err := parseCSV(strings.NewReader(header()), 2026)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func header() string {
	return `"Player","ID","Team","Opponent","Round","Kicks","Marks","Hand Balls","Disp","Goals","Behinds","Hit Outs","Tackles","Rebounds","Inside 50","Clearances","Clangers","Frees For","Frees Against","Brownlow","Contested Possessions","Uncontested Possessions","Contested Marks","Marks Inside 50","One Percenters","Bounces","Goal Assists","% Time Played"`
}
