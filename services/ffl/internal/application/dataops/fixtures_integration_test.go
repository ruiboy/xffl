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

func versusMatch(home, away int) MatchSpec {
	return MatchSpec{Style: domain.MatchStyleVersus, ClubSeasonIDs: []int{home, away}}
}
func byeMatch(cs int) MatchSpec {
	return MatchSpec{Style: domain.MatchStyleBye, ClubSeasonIDs: []int{cs}}
}
func superbyeMatch(csIDs ...int) MatchSpec {
	return MatchSpec{Style: domain.MatchStyleSuperbye, ClubSeasonIDs: csIDs}
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

func onlyRoundID(ctx context.Context, t *testing.T, seasonID int) int {
	t.Helper()
	rounds, err := postgres.NewRoundRepository(sqlcgen.New(testPool)).FindBySeasonID(ctx, seasonID)
	require.NoError(t, err)
	require.Len(t, rounds, 1)
	return rounds[0].ID
}

func TestSaveFixtures_CreateWithBye(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFcreate", 1)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{versusMatch(csA, csB), byeMatch(csC)},
	}}))

	roundID := onlyRoundID(ctx, t, seasonID)
	groups := roundClubMatches(ctx, t, roundID)
	require.Len(t, groups, 2, "one fixture + one bye = two matches")

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
	assert.Equal(t, "bye", bye[0].Side)
	assert.Equal(t, "home", fixture[0].Side)
	assert.Equal(t, "away", fixture[1].Side)
	assert.Equal(t, map[int]bool{csA: true, csB: true, csC: true}, clubSeasonSet(groups))

	// The versus match carries the 'versus' style, and the bye its own.
	var versusStyles, byeStyles int
	require.NoError(t, testPool.QueryRow(ctx,
		`SELECT count(*) FROM ffl.match m JOIN ffl.round r ON r.id = m.round_id WHERE r.season_id = $1 AND m.match_style = 'versus' AND m.deleted_at IS NULL`, seasonID).Scan(&versusStyles))
	assert.Equal(t, 1, versusStyles)
	require.NoError(t, testPool.QueryRow(ctx,
		`SELECT count(*) FROM ffl.match m JOIN ffl.round r ON r.id = m.round_id WHERE r.season_id = $1 AND m.match_style = 'bye' AND m.deleted_at IS NULL`, seasonID).Scan(&byeStyles))
	assert.Equal(t, 1, byeStyles)
}

func TestSaveFixtures_LoadRoundTrip(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFroundtrip", 8)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{versusMatch(csA, csB), byeMatch(csC)},
	}}))

	loaded, err := builder.LoadFixtures(ctx, seasonID)
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Len(t, loaded[0].Matches, 2)
	byStyle := map[domain.MatchStyle][]int{}
	for _, m := range loaded[0].Matches {
		byStyle[m.Style] = m.ClubSeasonIDs
	}
	assert.Equal(t, []int{csA, csB}, byStyle[domain.MatchStyleVersus], "versus loads [home, away]")
	assert.Equal(t, []int{csC}, byStyle[domain.MatchStyleBye])
}

func TestSaveFixtures_ReplaceRound(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFreplace", 2)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{versusMatch(csA, csB), byeMatch(csC)},
	}}))
	roundID := onlyRoundID(ctx, t, seasonID)

	// Re-save the same round with a different pairing and bye.
	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		RoundID: &roundID, Name: "Round 1 (edited)", AFLRoundID: 11, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{versusMatch(csB, csC), byeMatch(csA)},
	}}))

	stillOne := onlyRoundID(ctx, t, seasonID)
	assert.Equal(t, roundID, stillOne)
	groups := roundClubMatches(ctx, t, roundID)
	require.Len(t, groups, 2)
	for _, g := range groups {
		if len(g) == 1 {
			assert.Equal(t, csA, g[0].ClubSeasonID, "A now on the bye")
		}
	}
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
		Matches: []MatchSpec{versusMatch(csA, csB), byeMatch(csC)},
	}}))

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
		Matches: []MatchSpec{versusMatch(csA, csB), byeMatch(csC)},
	}}))
	roundID := onlyRoundID(ctx, t, seasonID)

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
		Matches: []MatchSpec{versusMatch(csB, csC), byeMatch(csA)},
	}}))
	after, err := postgres.NewRoundRepository(sqlcgen.New(testPool)).FindByID(ctx, roundID)
	require.NoError(t, err)
	assert.Equal(t, "Round 1", after.Name, "immutable round unchanged")
}

func TestSaveFixtures_Superbye(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFsuper", 6)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Superbye round", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{superbyeMatch(csA, csB, csC)},
	}}))

	// Loads back as a single superbye match holding all three clubs.
	loaded, err := builder.LoadFixtures(ctx, seasonID)
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Len(t, loaded[0].Matches, 1)
	assert.Equal(t, domain.MatchStyleSuperbye, loaded[0].Matches[0].Style)
	assert.ElementsMatch(t, []int{csA, csB, csC}, loaded[0].Matches[0].ClubSeasonIDs)

	var sides int
	require.NoError(t, testPool.QueryRow(ctx,
		`SELECT count(*) FROM ffl.club_match cm JOIN ffl.match m ON m.id = cm.match_id
		 JOIN ffl.round r ON r.id = m.round_id
		 WHERE r.season_id = $1 AND cm.side = 'superbye' AND cm.deleted_at IS NULL`, seasonID).Scan(&sides))
	assert.Equal(t, 3, sides)
}

// The unified loader returns every final club_match tagged with its match style.
func TestFindFinalClubMatchesBySeasonID(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFfinal", 5)

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{versusMatch(csA, csB), byeMatch(csC)},
	}}))

	final := func(cs, score int) {
		_, err := testPool.Exec(ctx,
			`UPDATE ffl.club_match SET data_status = 'final', drv_score = $2 WHERE club_season_id = $1 AND deleted_at IS NULL`, cs, score)
		require.NoError(t, err)
	}
	final(csA, 1000)
	final(csB, 900)
	final(csC, 850)

	rows, err := postgres.NewClubMatchRepository(sqlcgen.New(testPool)).FindFinalClubMatchesBySeasonID(ctx, seasonID)
	require.NoError(t, err)
	require.Len(t, rows, 3)

	byCS := map[int]domain.ScoredClubMatch{}
	for _, r := range rows {
		byCS[r.ClubSeasonID] = r
	}
	assert.Equal(t, domain.MatchStyleVersus, byCS[csA].Style)
	assert.Equal(t, 1000, byCS[csA].Score)
	assert.Equal(t, domain.MatchStyleVersus, byCS[csB].Style)
	assert.Equal(t, domain.MatchStyleBye, byCS[csC].Style)
	assert.Equal(t, 850, byCS[csC].Score)
	assert.Equal(t, domain.RoundTypeMinor, byCS[csC].RoundType)
	assert.Equal(t, byCS[csA].MatchID, byCS[csB].MatchID, "A and B share the versus match")
	assert.NotEqual(t, byCS[csA].MatchID, byCS[csC].MatchID, "the bye is a separate match")
}

// clubMatchIDsByClubSeason maps each live club_season in a round to its
// club_match row id, so tests can assert which rows survive a re-save.
func clubMatchIDsByClubSeason(ctx context.Context, t *testing.T, roundID int) map[int]int {
	t.Helper()
	out := map[int]int{}
	for _, g := range roundClubMatches(ctx, t, roundID) {
		for _, cm := range g {
			out[cm.ClubSeasonID] = cm.ID
		}
	}
	return out
}

// rowCounts returns how many match and club_match rows a round holds in total,
// tombstones included. Fixture edits remove rows outright, so these counts track
// the live fixture exactly — any drift is churn.
func rowCounts(ctx context.Context, t *testing.T, roundID int) (matches, clubMatches int) {
	t.Helper()
	require.NoError(t, testPool.QueryRow(ctx,
		`SELECT count(*) FROM ffl.match WHERE round_id = $1`, roundID).Scan(&matches))
	require.NoError(t, testPool.QueryRow(ctx,
		`SELECT count(*) FROM ffl.club_match cm JOIN ffl.match m ON m.id = cm.match_id
		 WHERE m.round_id = $1`, roundID).Scan(&clubMatches))
	return matches, clubMatches
}

// Re-saving an unchanged round reuses every row — no soft-delete/reinsert churn.
func TestSaveFixtures_ResaveIdenticalNoChurn(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "SFnochurn", 7)

	spec := []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{versusMatch(csA, csB), byeMatch(csC)},
	}}
	require.NoError(t, builder.SaveFixtures(ctx, seasonID, spec))
	roundID := onlyRoundID(ctx, t, seasonID)
	before := clubMatchIDsByClubSeason(ctx, t, roundID)

	// Re-save the exact same fixtures (now targeting the existing round).
	spec[0].RoundID = &roundID
	require.NoError(t, builder.SaveFixtures(ctx, seasonID, spec))

	after := clubMatchIDsByClubSeason(ctx, t, roundID)
	assert.Equal(t, before, after, "identical re-save reuses every club_match row")
	m, cm := rowCounts(ctx, t, roundID)
	assert.Equal(t, 2, m, "still just the versus + bye match, no churn")
	assert.Equal(t, 3, cm, "still just the 3 club_match rows, no churn")
}

// Editing one match leaves the other match's rows untouched, and the edited
// match is repointed in place rather than rebuilt.
func TestSaveFixtures_EditPreservesUnchangedMatch(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))
	a := createClub(ctx, t, "SFpreserve A")
	b := createClub(ctx, t, "SFpreserve B")
	c := createClub(ctx, t, "SFpreserve C")
	d := createClub(ctx, t, "SFpreserve D")
	built, err := builder.BuildSeason(ctx, BuildSeasonParams{
		SeasonName: "SFpreserve", RulesID: "2011", AFLSeasonID: 9, ClubIDs: []int{a, b, c, d},
	})
	require.NoError(t, err)
	seasonID := built.SeasonID
	csA := built.ClubSeasons[0].ClubSeasonID
	csB := built.ClubSeasons[1].ClubSeasonID
	csC := built.ClubSeasons[2].ClubSeasonID
	csD := built.ClubSeasons[3].ClubSeasonID

	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{versusMatch(csA, csB), versusMatch(csC, csD)},
	}}))
	roundID := onlyRoundID(ctx, t, seasonID)
	before := clubMatchIDsByClubSeason(ctx, t, roundID)

	// Swap home/away in the first match only; leave the second untouched.
	require.NoError(t, builder.SaveFixtures(ctx, seasonID, []RoundSpec{{
		RoundID: &roundID, Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
		Matches: []MatchSpec{versusMatch(csB, csA), versusMatch(csC, csD)},
	}}))
	after := clubMatchIDsByClubSeason(ctx, t, roundID)

	// The untouched match keeps both club_match rows exactly.
	assert.Equal(t, before[csC], after[csC], "unchanged match's C row reused")
	assert.Equal(t, before[csD], after[csD], "unchanged match's D row reused")
	// The swapped match keeps each club on its own row — only the sides moved.
	assert.Equal(t, before[csA], after[csA], "A's row reused, side flipped")
	assert.Equal(t, before[csB], after[csB], "B's row reused, side flipped")
	sides := map[int]string{}
	for _, g := range roundClubMatches(ctx, t, roundID) {
		for _, cm := range g {
			sides[cm.ClubSeasonID] = cm.Side
		}
	}
	assert.Equal(t, "home", sides[csB], "B is now home")
	assert.Equal(t, "away", sides[csA], "A is now away")

	m, cm := rowCounts(ctx, t, roundID)
	assert.Equal(t, 2, m, "still two matches, no churn")
	assert.Equal(t, 4, cm, "still four club_match rows, no churn")
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
