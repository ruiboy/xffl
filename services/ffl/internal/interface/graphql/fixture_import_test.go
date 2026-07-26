package graphql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application/dataops"
	"xffl/services/ffl/internal/infrastructure/spreadsheet"
)

// No-DB proof of the fixture-import parse path: a pasted sheet flows through the
// resolver → real spreadsheet parser, and the GraphQL result carries the rounds,
// pairings, and reference scores for review.
func TestParseFFLFixtureSheet_ParsesRoundsAndScores(t *testing.T) {
	b := dataops.NewBuilder(nil, spreadsheet.NewFixtureParser())
	r := &Resolver{Builder: b}

	sheet := "1\nTHC\t452\tvs\tRuiboys\t372\nCheetahs\t300\tvs\tSlashers\t280\n"
	res, err := r.Mutation().ParseFFLFixtureSheet(context.Background(), ParseFFLFixtureSheetInput{Sheet: sheet})
	require.NoError(t, err)
	require.Len(t, res.Rounds, 1)

	rd := res.Rounds[0]
	assert.Equal(t, 1, rd.Round)
	require.Len(t, rd.Fixtures, 2)

	f := rd.Fixtures[0]
	assert.Equal(t, "THC", f.HomeClub)
	assert.Equal(t, "Ruiboys", f.AwayClub)
	require.NotNil(t, f.HomeScore)
	assert.Equal(t, 452, *f.HomeScore)
	require.NotNil(t, f.AwayScore)
	assert.Equal(t, 372, *f.AwayScore)
}
