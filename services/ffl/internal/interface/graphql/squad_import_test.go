package graphql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/application/dataops"
	"xffl/services/ffl/internal/infrastructure/forum"
)

// exactResolver matches parsed names against the pool by exact name: confidence
// 1.0 on a hit, empty on a miss.
type exactResolver struct{}

func (exactResolver) Resolve(_ context.Context, name, _ string, candidates []application.PlayerCandidate) ([]application.PlayerNameMatch, error) {
	for _, cand := range candidates {
		if cand.Name == name {
			return []application.PlayerNameMatch{{Candidate: cand, Confidence: 1.0}}, nil
		}
	}
	return nil, nil
}

// seasonLookup implements just the LookupPlayerSeasonsBySeasonID slice of
// PlayerLookup that ParseSquadThreadForSeason uses; other methods stay nil.
type seasonLookup struct {
	application.PlayerLookup
	candidates []application.PlayerCandidate
}

func (f seasonLookup) LookupPlayerSeasonsBySeasonID(context.Context, int) ([]application.PlayerCandidate, error) {
	return f.candidates, nil
}

// End-to-end (no DB) proof of the squad-import parse path: a pasted thread flows
// through the resolver → real squad parser → resolution, and the GraphQL result
// carries resolved handles, confidence, and the review flags.
func TestParseFFLSquadThread_ResolvesAndFlags(t *testing.T) {
	candidates := []application.PlayerCandidate{
		{AFLPlayerSeasonID: 101, Name: "Darcy Fogarty", Club: "Adelaide"},
	}
	lookup := seasonLookup{candidates: candidates}
	ops := dataops.NewDataOpsCommands(nil, lookup, exactResolver{}, nil, nil, nil, forum.NewSquadParser())
	r := &Resolver{DataOps: ops}

	thread := "Cheetahs\n1 Darcy Fogarty 0.6 ADE\n2 Unmatched Guy 0.5 XXX\n"
	res, err := r.Mutation().ParseFFLSquadThread(context.Background(), ParseFFLSquadThreadInput{
		AflSeasonID: "1",
		Thread:      thread,
	})
	require.NoError(t, err)
	require.Len(t, res.Squads, 1)

	sq := res.Squads[0]
	assert.Equal(t, "Cheetahs", sq.ClubName)
	require.Len(t, sq.Members, 2)
	assert.Equal(t, []int{1}, sq.NeedsReview, "the unmatched member is flagged for review")

	resolved := sq.Members[0]
	require.NotNil(t, resolved.AflPlayerSeasonID)
	assert.Equal(t, "101", *resolved.AflPlayerSeasonID)
	require.NotNil(t, resolved.ResolvedName)
	assert.Equal(t, "Darcy Fogarty", *resolved.ResolvedName)
	assert.Equal(t, 60, *resolved.CostCents)
	assert.Equal(t, 1.0, resolved.Confidence)

	unresolved := sq.Members[1]
	assert.Nil(t, unresolved.AflPlayerSeasonID, "an unmatched member carries no handle")
	assert.Nil(t, unresolved.ResolvedName)
}
