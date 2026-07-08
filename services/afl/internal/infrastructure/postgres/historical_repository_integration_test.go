//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/afl/internal/application"
	"xffl/services/afl/internal/infrastructure/footywire"
	pg "xffl/services/afl/internal/infrastructure/postgres"
)

// autoCreatePrompter always chooses "create new" (returns 0). The integration
// test avoids ambiguity so this is never actually called; it's here to satisfy
// the port.
type autoCreatePrompter struct{ calls int }

func (p *autoCreatePrompter) Choose(context.Context, string, string, string, []application.PlayerChoice) (int, error) {
	p.calls++
	return 0, nil
}

type noopReviewLog struct{ newPlayers int }

func (l *noopReviewLog) NewPlayer(string, string, string, int)            { l.newPlayers++ }
func (l *noopReviewLog) NearMiss(string, string, string, string, float64) {}
func (l *noopReviewLog) Gap(string, int, int, string, int)               {}

func truncateHistorical(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	pool := connectDB(t)
	_, _ = pool.Exec(ctx, "TRUNCATE afl.dataops_player_source CASCADE")
	for _, table := range []string{
		"afl.player_match", "afl.player_season", "afl.player",
		"afl.club_match", "afl.match", "afl.club_season", "afl.club",
		"afl.round", "afl.season", "afl.league",
	} {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
		require.NoError(t, err, "truncate %s", table)
	}
}

func countRows(t *testing.T, table string) int {
	t.Helper()
	var n int
	require.NoError(t, connectDB(t).QueryRow(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&n))
	return n
}

func historicalRows() []application.HistoricalRow {
	// One round, two matches. "Existing Star" is pre-inserted to test exact-link.
	return []application.HistoricalRow{
		{Round: "1", Date: "1998-03-27", Venue: "MCG", HomeClub: "Richmond", AwayClub: "Carlton", Club: "Richmond", Player: "Existing Star", Kicks: 20, Marks: 5, Handballs: 8, Goals: 2, Behinds: 1, Hitouts: 0, Tackles: 4},
		{Round: "1", Date: "1998-03-27", Venue: "MCG", HomeClub: "Richmond", AwayClub: "Carlton", Club: "Richmond", Player: "Wayne Campbell", Kicks: 25, Marks: 6, Handballs: 10, Goals: 0, Behinds: 0, Hitouts: 0, Tackles: 3},
		{Round: "1", Date: "1998-03-27", Venue: "MCG", HomeClub: "Richmond", AwayClub: "Carlton", Club: "Carlton", Player: "Anthony Koutoufides", Kicks: 22, Marks: 8, Handballs: 6, Goals: 3, Behinds: 1, Hitouts: 5, Tackles: 2},
		{Round: "1", Date: "1998-03-29", Venue: "Gabba", HomeClub: "Brisbane Lions", AwayClub: "Geelong", Club: "Brisbane Lions", Player: "Michael Voss", Kicks: 30, Marks: 4, Handballs: 12, Goals: 1, Behinds: 2, Hitouts: 0, Tackles: 6},
		{Round: "Grand Final", Date: "1998-09-26", Venue: "MCG", HomeClub: "Adelaide", AwayClub: "North Melbourne", Club: "Adelaide", Player: "Andrew McLeod", Kicks: 28, Marks: 7, Handballs: 9, Goals: 2, Behinds: 0, Hitouts: 0, Tackles: 5},
	}
}

func TestHistoricalImport_CreatesScaffoldAndStats(t *testing.T) {
	ctx := context.Background()
	pool := connectDB(t)
	truncateHistorical(t)
	t.Cleanup(func() { truncateHistorical(t) })

	// Pre-insert one player to verify exact-name auto-link (no duplicate created).
	var existingID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Existing Star') RETURNING id").Scan(&existingID))

	prompter := &autoCreatePrompter{}
	log := &noopReviewLog{}
	imp := application.NewHistoricalImporter(pg.NewHistoricalRepository(pool), footywire.NewLevenshteinResolver(), prompter, log)

	sum, err := imp.ImportSeason(ctx, 1998, historicalRows())
	require.NoError(t, err)

	assert.Equal(t, 3, sum.Matches, "3 distinct fixtures")
	assert.Equal(t, 5, sum.PlayerMatches)
	assert.Equal(t, 4, sum.NewPlayers, "5 players, 1 pre-existing")
	assert.Equal(t, 0, sum.Prompted, "no ambiguity")
	assert.Equal(t, 0, prompter.calls)

	// Scaffold rows.
	assert.Equal(t, 1, countRows(t, "afl.season"))
	assert.Equal(t, 2, countRows(t, "afl.round"), "Round 1 + Grand Final")
	assert.Equal(t, 3, countRows(t, "afl.match"))
	assert.Equal(t, 6, countRows(t, "afl.club_match"), "3 matches x 2")
	assert.Equal(t, 6, countRows(t, "afl.club_season"), "6 distinct clubs")
	assert.Equal(t, 5, countRows(t, "afl.player"), "4 new + 1 pre-existing")
	assert.Equal(t, 5, countRows(t, "afl.player_season"))
	assert.Equal(t, 5, countRows(t, "afl.player_match"))

	// Season name convention and finals round name.
	var seasonName string
	require.NoError(t, pool.QueryRow(ctx, "SELECT name FROM afl.season LIMIT 1").Scan(&seasonName))
	assert.Equal(t, "AFL 1998", seasonName)
	var gfCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM afl.round WHERE name = 'Grand Final'").Scan(&gfCount))
	assert.Equal(t, 1, gfCount)

	// Exact-name link reused the pre-existing player (no duplicate "Existing Star").
	var existingStarCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM afl.player WHERE name = 'Existing Star'").Scan(&existingStarCount))
	assert.Equal(t, 1, existingStarCount)

	// Matches are 'final' so their stats count in averages.
	var nonFinal int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM afl.match WHERE data_status <> 'final'").Scan(&nonFinal))
	assert.Equal(t, 0, nonFinal)

	// Stat line landed correctly for Koutoufides.
	var goals, hitouts int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT pm.goals, pm.hitouts FROM afl.player_match pm
		JOIN afl.player_season ps ON ps.id = pm.player_season_id
		JOIN afl.player p ON p.id = ps.player_id
		WHERE p.name = 'Anthony Koutoufides'`).Scan(&goals, &hitouts))
	assert.Equal(t, 3, goals)
	assert.Equal(t, 5, hitouts)
}

func TestHistoricalImport_Idempotent(t *testing.T) {
	ctx := context.Background()
	pool := connectDB(t)
	truncateHistorical(t)
	t.Cleanup(func() { truncateHistorical(t) })

	imp := application.NewHistoricalImporter(pg.NewHistoricalRepository(pool), footywire.NewLevenshteinResolver(), &autoCreatePrompter{}, &noopReviewLog{})

	_, err := imp.ImportSeason(ctx, 1998, historicalRows())
	require.NoError(t, err)
	pmAfterFirst := countRows(t, "afl.player_match")

	// Second run must not duplicate any rows (xref + upserts).
	sum2, err := imp.ImportSeason(ctx, 1998, historicalRows())
	require.NoError(t, err)

	assert.Equal(t, pmAfterFirst, countRows(t, "afl.player_match"), "player_match rows stable on re-run")
	assert.Equal(t, 5, countRows(t, "afl.player"), "no duplicate players")
	assert.Equal(t, 3, countRows(t, "afl.match"), "no duplicate matches")
	assert.Equal(t, 0, sum2.NewPlayers, "xref makes re-run create nothing")
}
