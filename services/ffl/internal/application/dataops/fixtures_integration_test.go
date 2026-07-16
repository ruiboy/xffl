//go:build integration

package dataops

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/domain"
	"xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
)

// buildOddSeason creates a 3-club season (odd → one bye) and returns its id and
// the three club_season ids.
func buildOddSeason(ctx context.Context, t *testing.T, name string, aflSeasonID int) (seasonID, csA, csB, csC int) {
	t.Helper()
	builder := NewBuilder(postgres.NewDB(testPool))
	a := createClub(ctx, t, name+" A")
	b := createClub(ctx, t, name+" B")
	c := createClub(ctx, t, name+" C")
	built, err := builder.BuildSeason(ctx, BuildSeasonParams{
		SeasonName: name, RulesID: "2011", AFLSeasonID: aflSeasonID, ClubIDs: []int{a, b, c},
	})
	require.NoError(t, err)
	return built.SeasonID, built.ClubSeasons[0].ClubSeasonID, built.ClubSeasons[1].ClubSeasonID, built.ClubSeasons[2].ClubSeasonID
}

// roundClubMatches groups a round's club_matches by their match, so tests can
// assert fixture (2 sides) vs bye (1 side) structure.
func roundClubMatches(ctx context.Context, t *testing.T, roundID int) [][]domain.ClubMatch {
	t.Helper()
	q := sqlcgen.New(testPool)
	matches, err := postgres.NewMatchRepository(q).FindByRoundID(ctx, roundID)
	require.NoError(t, err)
	cmRepo := postgres.NewClubMatchRepository(q)
	out := make([][]domain.ClubMatch, 0, len(matches))
	for _, m := range matches {
		cms, err := cmRepo.FindByMatchID(ctx, m.ID)
		require.NoError(t, err)
		out = append(out, cms)
	}
	return out
}

// clubSeasonSet flattens grouped club_matches to the set of club_season ids present.
func clubSeasonSet(groups [][]domain.ClubMatch) map[int]bool {
	set := map[int]bool{}
	for _, g := range groups {
		for _, cm := range g {
			set[cm.ClubSeasonID] = true
		}
	}
	return set
}

func onlyRoundID(ctx context.Context, t *testing.T, seasonID int) (int, int) {
	t.Helper()
	rounds, err := postgres.NewRoundRepository(sqlcgen.New(testPool)).FindBySeasonID(ctx, seasonID)
	require.NoError(t, err)
	require.Len(t, rounds, 1)
	return rounds[0].ID, len(rounds)
}

func TestSaveFixtures_CreateWithBye(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFcreate", 1)

	err := builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Fixtures: []FixtureSpec{{HomeClubSeasonID: csA, AwayClubSeasonID: csB}},
		Byes:     []int{csC},
	}})
	require.NoError(t, err)

	roundID, _ := onlyRoundID(ctx, t, seasonID)
	groups := roundClubMatches(ctx, t, roundID)
	require.Len(t, groups, 2, "one fixture + one bye = two matches")

	// One match is a fixture (2 club_matches), one is a bye (1 club_match).
	var fixture, bye []domain.ClubMatch
	for _, g := range groups {
		switch len(g) {
		case 2:
			fixture = g
		case 1:
			bye = g
		}
	}
	require.Len(t, fixture, 2)
	require.Len(t, bye, 1)
	assert.Equal(t, csC, bye[0].ClubSeasonID, "the odd club is on the bye")
	assert.Equal(t, map[int]bool{csA: true, csB: true, csC: true}, clubSeasonSet(groups))

	// The bye is persisted as a single-sided bye match/club_match.
	var byeSides, byeStyles int
	require.NoError(t, testPool.QueryRow(ctx,
		`SELECT count(*) FROM ffl.club_match WHERE side = 'bye' AND club_season_id = $1 AND deleted_at IS NULL`, csC,
	).Scan(&byeSides))
	assert.Equal(t, 1, byeSides)
	require.NoError(t, testPool.QueryRow(ctx,
		`SELECT count(*) FROM ffl.match m JOIN ffl.round r ON r.id = m.round_id WHERE r.season_id = $1 AND m.match_style = 'bye' AND m.deleted_at IS NULL`, seasonID,
	).Scan(&byeStyles))
	assert.Equal(t, 1, byeStyles)
}

func TestSaveFixtures_ReplaceRound(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFreplace", 2)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Fixtures: []FixtureSpec{{HomeClubSeasonID: csA, AwayClubSeasonID: csB}}, Byes: []int{csC},
	}}))
	roundID, _ := onlyRoundID(ctx, t, seasonID)

	// Re-save the same round with a different pairing and bye.
	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		RoundID: &roundID, Name: "Round 1 (edited)", AFLRoundID: 11, Type: domain.RoundTypeMinor,
		Fixtures: []FixtureSpec{{HomeClubSeasonID: csB, AwayClubSeasonID: csC}}, Byes: []int{csA},
	}}))

	// Still one round (updated in place), now B v C with A on the bye.
	stillOne, _ := onlyRoundID(ctx, t, seasonID)
	assert.Equal(t, roundID, stillOne)
	groups := roundClubMatches(ctx, t, roundID)
	require.Len(t, groups, 2)
	for _, g := range groups {
		if len(g) == 1 {
			assert.Equal(t, csA, g[0].ClubSeasonID, "A now on the bye")
		}
	}
	// The replaced (soft-deleted) rows are gone from the live set.
	var live int
	require.NoError(t, testPool.QueryRow(ctx,
		`SELECT count(*) FROM ffl.club_match cm JOIN ffl.match m ON m.id = cm.match_id
		 WHERE m.round_id = $1 AND cm.deleted_at IS NULL`, roundID).Scan(&live))
	assert.Equal(t, 3, live, "2 fixture sides + 1 bye")
}

func TestSaveFixtures_DeleteEmptyRound(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFdelete", 3)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Fixtures: []FixtureSpec{{HomeClubSeasonID: csA, AwayClubSeasonID: csB}}, Byes: []int{csC},
	}}))

	// Saving an empty set removes the (team-less) round.
	require.NoError(t, builder.SaveFixtures(ctx, seasonID, nil))
	rounds, err := postgres.NewRoundRepository(sqlcgen.New(testPool)).FindBySeasonID(ctx, seasonID)
	require.NoError(t, err)
	assert.Empty(t, rounds)
}

func TestSaveFixtures_HasTeamsGuard(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFguard", 4)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Fixtures: []FixtureSpec{{HomeClubSeasonID: csA, AwayClubSeasonID: csB}}, Byes: []int{csC},
	}}))
	roundID, _ := onlyRoundID(ctx, t, seasonID)

	// Enter a team into one of the round's club_matches.
	groups := roundClubMatches(ctx, t, roundID)
	var clubMatchID, clubSeasonID int
	for _, g := range groups {
		if len(g) == 2 {
			clubMatchID, clubSeasonID = g[0].ID, g[0].ClubSeasonID
		}
	}
	seedTeam(ctx, t, testPool, clubMatchID, clubSeasonID)

	// Deleting the round now fails.
	err := builder.SaveFixtures(ctx, seasonID, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "submitted teams")

	// Re-saving it with changed fixtures is a no-op (immutable), not an error.
	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		RoundID: &roundID, Name: "hacked", AFLRoundID: 99, Type: domain.RoundTypeMinor,
		Fixtures: []FixtureSpec{{HomeClubSeasonID: csB, AwayClubSeasonID: csC}}, Byes: []int{csA},
	}}))
	after, err := postgres.NewRoundRepository(sqlcgen.New(testPool)).FindByID(ctx, roundID)
	require.NoError(t, err)
	assert.Equal(t, "Round 1", after.Name, "immutable round unchanged")
}

func TestFindFinalByesBySeasonID(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFbyeladder", 5)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Fixtures: []FixtureSpec{{HomeClubSeasonID: csA, AwayClubSeasonID: csB}}, Byes: []int{csC},
	}}))
	roundID, _ := onlyRoundID(ctx, t, seasonID)

	// Find the bye club_match and finalise it with a score.
	groups := roundClubMatches(ctx, t, roundID)
	var byeClubMatchID int
	for _, g := range groups {
		if len(g) == 1 {
			byeClubMatchID = g[0].ID
		}
	}
	require.NotZero(t, byeClubMatchID)
	_, err := testPool.Exec(ctx,
		`UPDATE ffl.club_match SET data_status = 'final', drv_score = 850 WHERE id = $1`, byeClubMatchID)
	require.NoError(t, err)

	byes, err := postgres.NewClubMatchRepository(sqlcgen.New(testPool)).FindFinalByesBySeasonID(ctx, seasonID)
	require.NoError(t, err)
	require.Len(t, byes, 1)
	assert.Equal(t, domain.ByeResult{ClubSeasonID: csC, Score: 850, RoundType: domain.RoundTypeMinor}, byes[0])
}

// seedTeam inserts a minimal player_match so a club_match counts as having a team.
func seedTeam(ctx context.Context, t *testing.T, pool *pgxpool.Pool, clubMatchID, clubSeasonID int) {
	t.Helper()
	var playerID int
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO ffl.player (afl_player_id) VALUES (0) RETURNING id`).Scan(&playerID))
	var psID int
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO ffl.player_season (player_id, club_season_id, afl_player_season_id) VALUES ($1, $2, 0) RETURNING id`,
		playerID, clubSeasonID).Scan(&psID))
	_, err := pool.Exec(ctx,
		`INSERT INTO ffl.player_match (club_match_id, player_season_id) VALUES ($1, $2)`, clubMatchID, psID)
	require.NoError(t, err)
}
