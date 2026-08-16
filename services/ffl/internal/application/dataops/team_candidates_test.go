package dataops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
)

// LookupSquadCandidates must take each member's club from the season being
// imported — not a former club from another year the player also has on record.
func TestLookupSquadCandidates_UsesSeasonClub(t *testing.T) {
	// The AFL season's players: Jordon Sweet at Port (this season, aps 101).
	// His 2021 Bulldogs season (aps 99) is not in this season's list at all.
	lookup := fakeSeasonLookup{candidates: []application.PlayerCandidate{
		{AFLPlayerSeasonID: 101, AFLPlayerID: 7, Name: "Jordon Sweet", Club: "Port Adelaide"},
		{AFLPlayerSeasonID: 200, AFLPlayerID: 8, Name: "Someone Else", Club: "Carlton"},
	}}
	c := NewDataOpsCommands(nil, lookup, nil, nil, nil, nil, nil)

	// The Cheetahs squad member links to Sweet's Port season (aps 101).
	squadByAPS := map[int]int{101: 5}

	got, err := c.LookupSquadCandidates(context.Background(), 2025, squadByAPS)
	require.NoError(t, err)
	require.Len(t, got, 1, "only squad members are candidates")
	assert.Equal(t, 5, got[0].PlayerID, "carries the ffl player_season id")
	assert.Equal(t, "Jordon Sweet", got[0].Name)
	assert.Equal(t, "Port Adelaide", got[0].Club, "club is this season's, not a former one")
}
