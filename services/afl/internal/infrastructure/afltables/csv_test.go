package afltables

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleCSV = `season,round,date,venue,home_club,away_club,club,player,kicks,marks,handballs,goals,behinds,hitouts,tackles
2023,1,2023-03-16,M.C.G.,Richmond,Carlton,Richmond,Liam Baker,10,5,5,0,0,0,1
2023,1,2023-03-16,M.C.G.,Richmond,Carlton,Carlton,Adam Cerra,14,6,7,0,0,0,3
2023,Grand Final,2023-09-30,M.C.G.,Collingwood,Brisbane Lions,Collingwood,Nick Daicos,20,4,10,1,2,0,5
`

func TestReadSeasonCSV(t *testing.T) {
	rows, err := ReadSeasonCSV(strings.NewReader(sampleCSV))
	require.NoError(t, err)
	require.Len(t, rows, 3)

	assert.Equal(t, Row{
		Season: 2023, Round: "1", Date: "2023-03-16", Venue: "M.C.G.",
		HomeClub: "Richmond", AwayClub: "Carlton", Club: "Richmond", Player: "Liam Baker",
		Kicks: 10, Marks: 5, Handballs: 5, Goals: 0, Behinds: 0, Hitouts: 0, Tackles: 1,
	}, rows[0])

	// Finals round names carry through as strings.
	assert.Equal(t, "Grand Final", rows[2].Round)
	assert.Equal(t, "Nick Daicos", rows[2].Player)
	assert.Equal(t, 1, rows[2].Goals)
}

func TestReadSeasonCSV_RejectsBadHeader(t *testing.T) {
	_, err := ReadSeasonCSV(strings.NewReader("wrong,header\n1,2\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "header")
}

func TestReadSeasonCSV_RejectsNonIntStat(t *testing.T) {
	bad := strings.Replace(sampleCSV, "10,5,5,0,0,0,1", "xx,5,5,0,0,0,1", 1)
	_, err := ReadSeasonCSV(strings.NewReader(bad))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kicks")
}
