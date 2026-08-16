package dataops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
)

// fakeSquadParser returns a fixed set of parsed squads regardless of input.
type fakeSquadParser struct{ squads []application.ParsedSquad }

func (f fakeSquadParser) ParseSquads(context.Context, string) ([]application.ParsedSquad, error) {
	return f.squads, nil
}

// fakeResolver matches by exact name against the candidate pool: confidence 1.0
// for a hit, empty result for a miss.
type fakeResolver struct{}

func (fakeResolver) Resolve(_ context.Context, name, _ string, candidates []application.PlayerCandidate) ([]application.PlayerNameMatch, error) {
	for _, cand := range candidates {
		if cand.Name == name {
			return []application.PlayerNameMatch{{Candidate: cand, Confidence: 1.0}}, nil
		}
	}
	return nil, nil
}

// fakeSeasonLookup implements just the LookupPlayerSeasonsBySeasonID slice of
// PlayerLookup used by ParseSquadThreadForSeason; the embedded interface leaves the
// other methods nil (they panic if ever called).
type fakeSeasonLookup struct {
	application.PlayerLookup
	candidates []application.PlayerCandidate
}

func (f fakeSeasonLookup) LookupPlayerSeasonsBySeasonID(context.Context, int) ([]application.PlayerCandidate, error) {
	return f.candidates, nil
}

func TestParseSquadThread_ResolvesAndFlags(t *testing.T) {
	candidates := []application.PlayerCandidate{
		{AFLPlayerSeasonID: 101, Name: "Darcy Fogarty", Club: "Adelaide"},
		{AFLPlayerSeasonID: 102, Name: "Reilly O'Brien", Club: "Adelaide"},
	}
	parser := fakeSquadParser{squads: []application.ParsedSquad{{
		ClubName: "Cheetahs",
		Members: []application.ParsedSquadMember{
			{Rank: 1, Name: "Darcy Fogarty", ClubHint: "Adel"},
			{Rank: 2, Name: "Unmatched Guy", ClubHint: "XXX"},
		},
	}}}

	// Only squadParser + playerResolver are exercised by ParseSquadThread.
	c := NewDataOpsCommands(nil, nil, fakeResolver{}, nil, nil, nil, parser)

	res, err := c.ParseSquadThread(context.Background(), "irrelevant raw text", candidates)
	require.NoError(t, err)

	require.Len(t, res.Squads, 1)
	sq := res.Squads[0]
	assert.Equal(t, "Cheetahs", sq.ClubName)
	require.Len(t, sq.Members, 2)

	t.Run("confident match carries the AFL player_season handle", func(t *testing.T) {
		m := sq.Members[0]
		assert.True(t, m.Confident)
		assert.Equal(t, 101, m.AFLPlayerSeasonID)
	})

	t.Run("unmatched member is flagged for review", func(t *testing.T) {
		m := sq.Members[1]
		assert.False(t, m.Confident)
		assert.Zero(t, m.AFLPlayerSeasonID)
		assert.Equal(t, []int{1}, sq.NeedsReview)
	})
}

func TestParseSquadThreadForSeason_SourcesCandidates(t *testing.T) {
	lookup := fakeSeasonLookup{candidates: []application.PlayerCandidate{
		{AFLPlayerSeasonID: 101, Name: "Darcy Fogarty"},
	}}
	parser := fakeSquadParser{squads: []application.ParsedSquad{{
		ClubName: "Cheetahs",
		Members:  []application.ParsedSquadMember{{Rank: 1, Name: "Darcy Fogarty"}},
	}}}

	c := NewDataOpsCommands(nil, lookup, fakeResolver{}, nil, nil, nil, parser)

	res, err := c.ParseSquadThreadForSeason(context.Background(), 2025, "irrelevant")
	require.NoError(t, err)
	require.Len(t, res.Squads, 1)
	require.Len(t, res.Squads[0].Members, 1)
	// The candidate pool came from LookupPlayerSeasonsBySeasonID, so the member resolved.
	assert.True(t, res.Squads[0].Members[0].Confident)
	assert.Equal(t, 101, res.Squads[0].Members[0].AFLPlayerSeasonID)
}
