//go:build integration

package graphql_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
	pg "xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	gql "xffl/services/ffl/internal/interface/graphql"
	memevents "xffl/shared/events/memory"
)

// db setup

func connectDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testPool
}

func setupTestServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
	t.Helper()

	q := sqlcgen.New(pool)
	queries := application.NewQueries(
		pg.NewClubRepository(q),
		pg.NewSeasonRepository(q),
		pg.NewRoundRepository(q),
		pg.NewMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewPlayerRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)

	db := pg.NewDB(pool)
	commands := application.NewCommands(db, memevents.New(), application.CommandsDeps{
		EventRepos: application.EventRepos{
			Rounds:        pg.NewRoundRepository(q),
			PlayerSeasons: pg.NewPlayerSeasonRepository(q),
			PlayerMatches: pg.NewPlayerMatchRepository(q),
		},
		PlayerLookup: &stubPlayerLookup{pool: pool},
	})

	resolver := &gql.Resolver{Queries: queries, Commands: commands}
	srv := gqlhandler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))

	// inject per-request loaders via HTTP middleware
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gql.InjectLoaders(r.Context(), gql.NewLoaders(queries))
		srv.ServeHTTP(w, r.WithContext(ctx))
	})
	return httptest.NewServer(h)
}

// fixture

// testIDs holds IDs of rows inserted by seedTestData, used by tests to query known data.
type testIDs struct {
	leagueID        int
	seasonID        int
	roundID         int
	homeClubID      int
	awayClubID      int
	homeClubSeaID   int
	awayClubSeaID   int
	matchID         int
	homeClubMatchID int
	awayClubMatchID int
	aflPlayerID     int
	playerID        int
	playerSeasonID  int
	playerMatchID   int
}

func seedTestData(t *testing.T, pool *pgxpool.Pool) testIDs {
	t.Helper()
	ctx := context.Background()
	var ids testIDs

	cleanupTestData(ctx, t, pool)

	// League
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.league (name) VALUES ('Test FFL') RETURNING id").Scan(&ids.leagueID))

	// Season
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.season (name, league_id) VALUES ('Test 2025', $1) RETURNING id",
		ids.leagueID).Scan(&ids.seasonID))

	// Round
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		ids.seasonID).Scan(&ids.roundID))

	// Two clubs
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club (name) VALUES ('Test Eagles') RETURNING id").Scan(&ids.homeClubID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club (name) VALUES ('Test Lions') RETURNING id").Scan(&ids.awayClubID))

	// Club seasons (home team higher on ladder)
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO ffl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 5, 4, 1, 0, 500, 400, 16) RETURNING id`,
		ids.homeClubID, ids.seasonID).Scan(&ids.homeClubSeaID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO ffl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 5, 3, 2, 0, 450, 420, 12) RETURNING id`,
		ids.awayClubID, ids.seasonID).Scan(&ids.awayClubSeaID))

	// Match
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES ($1, 'versus', 'Test Ground', '2025-06-15 14:00:00') RETURNING id",
		ids.roundID).Scan(&ids.matchID))

	// Club matches
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club_match (match_id, club_season_id, drv_score) VALUES ($1, $2, 85) RETURNING id",
		ids.matchID, ids.homeClubSeaID).Scan(&ids.homeClubMatchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club_match (match_id, club_season_id, drv_score) VALUES ($1, $2, 72) RETURNING id",
		ids.matchID, ids.awayClubSeaID).Scan(&ids.awayClubMatchID))

	// Link match to club matches
	_, err := pool.Exec(ctx,
		"UPDATE ffl.match SET home_club_match_id = $1, away_club_match_id = $2 WHERE id = $3",
		ids.homeClubMatchID, ids.awayClubMatchID, ids.matchID)
	require.NoError(t, err)

	// AFL player reference (needed for FFL player FK)
	// Name deliberately does not contain "Test" to avoid polluting AFL player search tests
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Seeded AFL Player') RETURNING id").Scan(&ids.aflPlayerID))

	// Player (linked to AFL player)
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player (afl_player_id) VALUES ($1) RETURNING id",
		ids.aflPlayerID).Scan(&ids.playerID))

	// Player season
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		ids.playerID, ids.homeClubSeaID).Scan(&ids.playerSeasonID))

	// Player match: goals position, played status, score 15
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_score)
		 VALUES ($1, $2, 'goals', 'played', 15) RETURNING id`,
		ids.homeClubMatchID, ids.playerSeasonID).Scan(&ids.playerMatchID))

	t.Cleanup(func() { cleanupTestData(context.Background(), t, pool) })

	return ids
}

func cleanupTestData(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	tables := []string{
		"ffl.player_match",
		"ffl.player_season", "ffl.player",
		"ffl.club_match", "ffl.match", "ffl.club_season",
		"ffl.club", "ffl.round", "ffl.season", "ffl.league",
		"afl.player",
	}
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
		require.NoError(t, err)
	}
}

// graphql helpers

type graphqlRequest struct {
	Query string `json:"query"`
}

type graphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func execQuery(t *testing.T, server *httptest.Server, query string) graphqlResponse {
	t.Helper()

	body, _ := json.Marshal(graphqlRequest{Query: query})
	resp, err := http.Post(server.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result graphqlResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

// tests

func TestFflClubs(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ fflClubs { id name } }`)

	require.Empty(t, result.Errors)

	var data struct {
		FflClubs []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"fflClubs"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns both seeded clubs", func(t *testing.T) {
		assert.Len(t, data.FflClubs, 2)
	})
	t.Run("clubs ordered alphabetically", func(t *testing.T) {
		require.Len(t, data.FflClubs, 2)
		assert.Equal(t, "Test Eagles", data.FflClubs[0].Name)
		assert.Equal(t, "Test Lions", data.FflClubs[1].Name)
	})
}

func TestFflSeasons(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ fflSeasons { id name } }`)

	require.Empty(t, result.Errors)

	var data struct {
		FflSeasons []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"fflSeasons"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns the seeded season", func(t *testing.T) {
		require.Len(t, data.FflSeasons, 1)
		assert.Equal(t, "Test 2025", data.FflSeasons[0].Name)
	})
}

func TestFflSeasonWithLadder(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	result := execQuery(t, server, `{ fflSeason(id: "`+seasonID+`") { name ladder { club { name } season { name } played won lost percentage } } }`)

	require.Empty(t, result.Errors)

	var data struct {
		FflSeason struct {
			Name   string `json:"name"`
			Ladder []struct {
				Club struct {
					Name string `json:"name"`
				} `json:"club"`
				Season struct {
					Name string `json:"name"`
				} `json:"season"`
				Played     int     `json:"played"`
				Won        int     `json:"won"`
				Lost       int     `json:"lost"`
				Percentage float64 `json:"percentage"`
			} `json:"ladder"`
		} `json:"fflSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	require.Len(t, data.FflSeason.Ladder, 2)

	t.Run("home team is first on ladder with correct percentage", func(t *testing.T) {
		assert.Equal(t, "Test Eagles", data.FflSeason.Ladder[0].Club.Name)
		assert.Equal(t, 125.0, data.FflSeason.Ladder[0].Percentage)
	})
	t.Run("away team is second on ladder", func(t *testing.T) {
		assert.Equal(t, "Test Lions", data.FflSeason.Ladder[1].Club.Name)
	})
	t.Run("season is populated on ladder entries", func(t *testing.T) {
		assert.Equal(t, "Test 2025", data.FflSeason.Ladder[0].Season.Name)
	})
}

func TestFflSeasonGraphTraversal(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	result := execQuery(t, server, `{
		fflSeason(id: "`+seasonID+`") {
			name
			rounds {
				name
				matches {
					venue
					homeClubMatch {
						club { name }
						score
						playerMatches {
							player { aflPlayer { id } }
							position
							status
							score
						}
					}
					awayClubMatch {
						club { name }
						score
					}
				}
			}
		}
	}`)

	require.Empty(t, result.Errors)

	var data struct {
		FflSeason struct {
			Name   string `json:"name"`
			Rounds []struct {
				Name    string `json:"name"`
				Matches []struct {
					Venue         string `json:"venue"`
					HomeClubMatch *struct {
						Club struct {
							Name string `json:"name"`
						} `json:"club"`
						Score         int `json:"score"`
						PlayerMatches []struct {
							Player struct {
								AflPlayer struct {
									ID string `json:"id"`
								} `json:"aflPlayer"`
							} `json:"player"`
							Position *string `json:"position"`
							Status   *string `json:"status"`
							Score    int     `json:"score"`
						} `json:"playerMatches"`
					} `json:"homeClubMatch"`
					AwayClubMatch *struct {
						Club struct {
							Name string `json:"name"`
						} `json:"club"`
						Score int `json:"score"`
					} `json:"awayClubMatch"`
				} `json:"matches"`
			} `json:"rounds"`
		} `json:"fflSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	require.Len(t, data.FflSeason.Rounds, 1)

	round := data.FflSeason.Rounds[0]
	require.Len(t, round.Matches, 1)

	match := round.Matches[0]
	require.NotNil(t, match.HomeClubMatch)
	require.NotNil(t, match.AwayClubMatch)

	t.Run("round and venue are correct", func(t *testing.T) {
		assert.Equal(t, "Round 1", round.Name)
		assert.Equal(t, "Test Ground", match.Venue)
	})
	t.Run("home club match has correct club and score", func(t *testing.T) {
		assert.Equal(t, "Test Eagles", match.HomeClubMatch.Club.Name)
		assert.Equal(t, 85, match.HomeClubMatch.Score)
	})
	t.Run("away club match has correct club and score", func(t *testing.T) {
		assert.Equal(t, "Test Lions", match.AwayClubMatch.Club.Name)
		assert.Equal(t, 72, match.AwayClubMatch.Score)
	})
	t.Run("player match has correct player, position and score", func(t *testing.T) {
		require.Len(t, match.HomeClubMatch.PlayerMatches, 1)
		pm := match.HomeClubMatch.PlayerMatches[0]
		// aflPlayer.name is resolved via federation; the FFL subgraph in
		// isolation only exposes the AFLPlayer stub by id. End-to-end name
		// resolution is covered by the e2e suite.
		assert.NotEmpty(t, pm.Player.AflPlayer.ID)
		require.NotNil(t, pm.Position)
		assert.Equal(t, "goals", *pm.Position)
		assert.Equal(t, 15, pm.Score)
	})
}

func TestFflRound(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	roundID := fmt.Sprintf("%d", ids.roundID)
	result := execQuery(t, server, `{
		fflRound(id: "`+roundID+`") {
			id
			name
			season { id name }
			matches { venue }
		}
	}`)
	require.Empty(t, result.Errors)

	var data struct {
		FflRound struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Season struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"season"`
			Matches []struct {
				Venue string `json:"venue"`
			} `json:"matches"`
		} `json:"fflRound"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns the correct round", func(t *testing.T) {
		assert.Equal(t, roundID, data.FflRound.ID)
		assert.Equal(t, "Round 1", data.FflRound.Name)
	})
	t.Run("round includes season breadcrumb", func(t *testing.T) {
		assert.Equal(t, "Test 2025", data.FflRound.Season.Name)
	})
	t.Run("round includes its matches", func(t *testing.T) {
		require.Len(t, data.FflRound.Matches, 1)
		assert.Equal(t, "Test Ground", data.FflRound.Matches[0].Venue)
	})
}

func TestFflMatch(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	matchID := fmt.Sprintf("%d", ids.matchID)
	result := execQuery(t, server, `{
		fflMatch(id: "`+matchID+`") {
			id
			venue
			round {
				id
				name
				season { id name }
			}
			homeClubMatch { club { name } score }
			awayClubMatch { club { name } score }
		}
	}`)
	require.Empty(t, result.Errors)

	var data struct {
		FflMatch struct {
			ID    string `json:"id"`
			Venue string `json:"venue"`
			Round struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Season struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"season"`
			} `json:"round"`
			HomeClubMatch *struct {
				Club  struct{ Name string } `json:"club"`
				Score int                   `json:"score"`
			} `json:"homeClubMatch"`
			AwayClubMatch *struct {
				Club  struct{ Name string } `json:"club"`
				Score int                   `json:"score"`
			} `json:"awayClubMatch"`
		} `json:"fflMatch"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns the correct match", func(t *testing.T) {
		assert.Equal(t, matchID, data.FflMatch.ID)
		assert.Equal(t, "Test Ground", data.FflMatch.Venue)
	})
	t.Run("round breadcrumb is correct", func(t *testing.T) {
		assert.Equal(t, "Round 1", data.FflMatch.Round.Name)
		assert.Equal(t, "Test 2025", data.FflMatch.Round.Season.Name)
	})
	t.Run("club matches are resolved via loaders", func(t *testing.T) {
		require.NotNil(t, data.FflMatch.HomeClubMatch)
		assert.Equal(t, "Test Eagles", data.FflMatch.HomeClubMatch.Club.Name)
		assert.Equal(t, 85, data.FflMatch.HomeClubMatch.Score)
		require.NotNil(t, data.FflMatch.AwayClubMatch)
		assert.Equal(t, "Test Lions", data.FflMatch.AwayClubMatch.Club.Name)
		assert.Equal(t, 72, data.FflMatch.AwayClubMatch.Score)
	})
}

func TestFflClubSeason(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	clubID := fmt.Sprintf("%d", ids.homeClubID)
	result := execQuery(t, server, `{
		fflClubSeason(seasonId: "`+seasonID+`", clubId: "`+clubID+`") {
			id
			club { name }
			season { id name }
			played
			won
		}
	}`)

	require.Empty(t, result.Errors)

	var data struct {
		FflClubSeason struct {
			ID     string                `json:"id"`
			Club   struct{ Name string } `json:"club"`
			Season struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"season"`
			Played int `json:"played"`
			Won    int `json:"won"`
		} `json:"fflClubSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	cs := data.FflClubSeason
	t.Run("returns correct club and season", func(t *testing.T) {
		assert.Equal(t, "Test Eagles", cs.Club.Name)
		assert.Equal(t, "Test 2025", cs.Season.Name)
		assert.Equal(t, seasonID, cs.Season.ID)
	})
	t.Run("returns correct win/loss record", func(t *testing.T) {
		assert.Equal(t, 5, cs.Played)
		assert.Equal(t, 4, cs.Won)
	})
}

// createFFLPlayer / updateFFLPlayer / deleteFFLPlayer mutations were retired
// alongside ffl.player.drv_name — FFL players are now created implicitly by
// addFFLPlayerToSeason and identified by their afl_player_id link.

func TestCalculateFFLFantasyScore(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Player has position "goals" → score = goals * 5
	// Sending goals=3 → expected score = 15
	pmID := fmt.Sprintf("%d", ids.playerMatchID)
	result := execQuery(t, server, `mutation {
		calculateFFLFantasyScore(input: {
			playerMatchId: "`+pmID+`"
			goals: 3
			kicks: 10
			handballs: 5
			marks: 4
			tackles: 2
			hitouts: 0
		}) {
			id
			player { aflPlayer { id } }
			position
			score
		}
	}`)

	require.Empty(t, result.Errors)

	var data struct {
		CalculateFFLFantasyScore struct {
			ID     string `json:"id"`
			Player struct {
				AflPlayer struct {
					ID string `json:"id"`
				} `json:"aflPlayer"`
			} `json:"player"`
			Position *string `json:"position"`
			Score    int     `json:"score"`
		} `json:"calculateFFLFantasyScore"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("player match score is goals multiplied by 5", func(t *testing.T) {
		assert.NotEmpty(t, data.CalculateFFLFantasyScore.Player.AflPlayer.ID)
		assert.Equal(t, 15, data.CalculateFFLFantasyScore.Score)
	})
}

func TestCalculateFFLFantasyScore_RecalculatesClubMatchScore(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Player has position "goals", sending goals=6 → score = 6*5 = 30
	pmID := fmt.Sprintf("%d", ids.playerMatchID)
	calcResult := execQuery(t, server, `mutation {
		calculateFFLFantasyScore(input: {
			playerMatchId: "`+pmID+`"
			goals: 6
			kicks: 0
			handballs: 0
			marks: 0
			tackles: 0
			hitouts: 0
		}) {
			score
		}
	}`)
	require.Empty(t, calcResult.Errors)

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	queryResult := execQuery(t, server, `{
		fflSeason(id: "`+seasonID+`") {
			rounds {
				matches {
					homeClubMatch {
						score
					}
				}
			}
		}
	}`)

	require.Empty(t, queryResult.Errors)

	var data struct {
		FflSeason struct {
			Rounds []struct {
				Matches []struct {
					HomeClubMatch struct {
						Score int `json:"score"`
					} `json:"homeClubMatch"`
				} `json:"matches"`
			} `json:"rounds"`
		} `json:"fflSeason"`
	}
	require.NoError(t, json.Unmarshal(queryResult.Data, &data))
	require.Len(t, data.FflSeason.Rounds, 1)
	require.Len(t, data.FflSeason.Rounds[0].Matches, 1)

	t.Run("club match score is recalculated after player score update", func(t *testing.T) {
		assert.Equal(t, 30, data.FflSeason.Rounds[0].Matches[0].HomeClubMatch.Score)
	})
}

func TestCalculateFFLFantasyScore_InvalidPlayerMatchID(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `mutation {
		calculateFFLFantasyScore(input: {
			playerMatchId: "999999"
			goals: 1
			kicks: 1
			handballs: 1
			marks: 1
			tackles: 1
			hitouts: 1
		}) {
			id
		}
	}`)

	t.Run("returns an error for unknown player match id", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

// ── SetFFLTeam: team composition rule tests ─────────────────────────────────

type teamPlayer struct {
	playerSeasonID      string
	position            string
	backupPositions     *string
	interchangePosition *string
}

func buildSetTeamMutation(clubMatchID string, players []teamPlayer) string {
	playersStr := ""
	for _, p := range players {
		bp := "null"
		if p.backupPositions != nil {
			bp = `"` + *p.backupPositions + `"`
		}
		ic := "null"
		if p.interchangePosition != nil {
			ic = `"` + *p.interchangePosition + `"`
		}
		playersStr += fmt.Sprintf(`
			{
				playerSeasonId: "%s"
				position: "%s"
				backupPositions: %s
				interchangePosition: %s
			}`, p.playerSeasonID, p.position, bp, ic)
	}
	return fmt.Sprintf(`mutation {
		setFFLTeam(input: {
			clubMatchId: "%s"
			players: [%s]
		}) { id position backupPositions interchangePosition }
	}`, clubMatchID, playersStr)
}

// seedExtraPlayers inserts n additional players into the home club season and returns their playerSeasonIDs.
func seedExtraPlayers(t *testing.T, pool *pgxpool.Pool, ids testIDs, count int) []string {
	t.Helper()
	ctx := context.Background()
	psIDs := make([]string, count)
	for i := range count {
		var aflID, playerID, psID int
		require.NoError(t, pool.QueryRow(ctx,
			fmt.Sprintf("INSERT INTO afl.player (name) VALUES ('Extra Player %d') RETURNING id", i)).Scan(&aflID))
		require.NoError(t, pool.QueryRow(ctx,
			"INSERT INTO ffl.player (afl_player_id) VALUES ($1) RETURNING id",
			aflID).Scan(&playerID))
		require.NoError(t, pool.QueryRow(ctx,
			"INSERT INTO ffl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
			playerID, ids.homeClubSeaID).Scan(&psID))
		psIDs[i] = fmt.Sprintf("%d", psID)
	}
	return psIDs
}

func strp(s string) *string { return &s }

func TestSetFFLTeam_ValidTeamSaves(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: psID, position: "goals"},
	}))

	t.Run("single starter saves without errors", func(t *testing.T) {
		assert.Empty(t, result.Errors)
	})
}

func TestSetFFLTeam_TooManyStartersForPosition(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 3)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: psID, position: "goals"},
		{playerSeasonID: extras[0], position: "goals"},
		{playerSeasonID: extras[1], position: "goals"},
		{playerSeasonID: extras[2], position: "goals"}, // 4th goals kicker — invalid
	}))

	t.Run("returns an error for four goal kickers", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

func TestSetFFLTeam_TooManyBenchPlayers(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 5)
	server := setupTestServer(t, pool)
	defer server.Close()

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	players := make([]teamPlayer, 0, 5)
	for _, id := range extras {
		players = append(players, teamPlayer{
			playerSeasonID:  id,
			position:        "goals",
			backupPositions: strp("goals,kicks"),
		})
	}

	result := execQuery(t, server, buildSetTeamMutation(cmID, players))

	t.Run("returns an error for five bench players", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

func TestSetFFLTeam_TwoBenchStars(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 1)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: psID, position: "star", backupPositions: strp("star")},
		{playerSeasonID: extras[0], position: "star", backupPositions: strp("star")},
	}))

	t.Run("returns an error for two backup star players", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

func TestSetFFLTeam_SamePositionCoveredByTwoBenchPlayers(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 1)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: psID, position: "goals", backupPositions: strp("goals,kicks")},
		{playerSeasonID: extras[0], position: "goals", backupPositions: strp("goals,marks")}, // goals covered twice
	}))

	t.Run("returns an error when two bench players cover the same position", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

func TestSetFFLTeam_TwoInterchangePositions(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 1)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: psID, position: "goals", backupPositions: strp("goals,kicks"), interchangePosition: strp("goals")},
		{playerSeasonID: extras[0], position: "marks", backupPositions: strp("marks,tackles"), interchangePosition: strp("marks")},
	}))

	t.Run("returns an error for two interchange positions", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

func TestSetFFLTeam_MultipleStartersScoreCorrectly(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 2)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()
	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	// Set 3 goal kickers
	setResult := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: psID, position: "goals"},
		{playerSeasonID: extras[0], position: "goals"},
		{playerSeasonID: extras[1], position: "goals"},
	}))
	require.Empty(t, setResult.Errors)

	// Set scores for all 3 goal kickers directly in the DB
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_score = 10 WHERE club_match_id = $1",
		ids.homeClubMatchID)
	require.NoError(t, err)

	// Sum starter scores (no backup_positions and no interchange_position)
	rows, err := pool.Query(ctx,
		"SELECT backup_positions, interchange_position, drv_score FROM ffl.player_match WHERE club_match_id = $1",
		ids.homeClubMatchID)
	require.NoError(t, err)
	defer rows.Close()

	totalScore := 0
	for rows.Next() {
		var score int
		var bp, ic *string
		require.NoError(t, rows.Scan(&bp, &ic, &score))
		if bp == nil && ic == nil {
			totalScore += score
		}
	}

	t.Run("three goal kicker starters each score 10", func(t *testing.T) {
		assert.Equal(t, 30, totalScore)
	})
}

func TestSetFFLTeam_ReplacesStaleEntries(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 1)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()
	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	extraID := extras[0]
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	// First team: original player at goals.
	firstResult := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: psID, position: "goals"},
	}))
	require.Empty(t, firstResult.Errors)

	// Second team: different player at kicks — replaces the first.
	secondResult := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: extraID, position: "kicks"},
	}))
	require.Empty(t, secondResult.Errors)

	// Only the new player_match should exist in the DB.
	rows, err := pool.Query(ctx,
		"SELECT player_season_id, position FROM ffl.player_match WHERE club_match_id = $1",
		ids.homeClubMatchID)
	require.NoError(t, err)
	defer rows.Close()

	type row struct {
		playerSeasonID int
		position       string
	}
	var found []row
	for rows.Next() {
		var r row
		require.NoError(t, rows.Scan(&r.playerSeasonID, &r.position))
		found = append(found, r)
	}

	require.Len(t, found, 1)
	extraIDInt, _ := strconv.Atoi(extraID)

	t.Run("only the replacement player remains after second set", func(t *testing.T) {
		assert.Equal(t, extraIDInt, found[0].playerSeasonID)
		assert.Equal(t, "kicks", found[0].position)
	})
}

func TestSetFFLTeam_EmptyTeamClearsAll(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	// Set a one-player team first.
	firstResult := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{
		{playerSeasonID: psID, position: "goals"},
	}))
	require.Empty(t, firstResult.Errors)

	// Now set an empty team.
	emptyResult := execQuery(t, server, buildSetTeamMutation(cmID, []teamPlayer{}))
	require.Empty(t, emptyResult.Errors)

	// No player_match rows should remain for this club match.
	ctx := context.Background()
	var count int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM ffl.player_match WHERE club_match_id = $1",
		ids.homeClubMatchID).Scan(&count))

	t.Run("all player match entries are cleared after empty team submission", func(t *testing.T) {
		assert.Equal(t, 0, count)
	})
}

func TestAddFFLPlayerToSeason_FromAFLPlayerSeason(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()

	// Seed an AFL player + player_season not yet linked to any ffl.player.
	aflSeasonID := insertAFLSeason(t, pool)
	var aflClubID, aflClubSeasonID, aflPlayerID, aflPlayerSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Richmond Tigers (Test)') RETURNING id").Scan(&aflClubID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0) RETURNING id`,
		aflClubID, aflSeasonID).Scan(&aflClubSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Dustin Martin') RETURNING id").Scan(&aflPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		aflPlayerID, aflClubSeasonID).Scan(&aflPlayerSeasonID))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.player_season WHERE id = $1", aflPlayerSeasonID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.player WHERE id = $1", aflPlayerID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.club_season WHERE id = $1", aflClubSeasonID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.club WHERE id = $1", aflClubID)
	})

	apsIDStr := fmt.Sprintf("%d", aflPlayerSeasonID)
	clubSeasonID := fmt.Sprintf("%d", ids.homeClubSeaID)
	roundIDStr := fmt.Sprintf("%d", ids.roundID)

	t.Run("creates a new FFL player and player_season linked to the AFL player_season", func(t *testing.T) {
		result := execQuery(t, server, `mutation {
			addFFLPlayerToSeason(input: {
				clubSeasonId: "`+clubSeasonID+`"
				aflPlayerSeasonId: "`+apsIDStr+`"
				fromRoundId: "`+roundIDStr+`"
			}) {
				id
				clubSeasonId
				aflPlayerSeasonId
				fromRoundId
			}
		}`)
		require.Empty(t, result.Errors)

		var data struct {
			AddFFLPlayerToSeason struct {
				ID                string  `json:"id"`
				ClubSeasonID      string  `json:"clubSeasonId"`
				AflPlayerSeasonID *string `json:"aflPlayerSeasonId"`
				FromRoundID       *string `json:"fromRoundId"`
			} `json:"addFFLPlayerToSeason"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))

		assert.NotEmpty(t, data.AddFFLPlayerToSeason.ID)
		assert.Equal(t, clubSeasonID, data.AddFFLPlayerToSeason.ClubSeasonID)
		require.NotNil(t, data.AddFFLPlayerToSeason.AflPlayerSeasonID)
		assert.Equal(t, apsIDStr, *data.AddFFLPlayerToSeason.AflPlayerSeasonID)
		require.NotNil(t, data.AddFFLPlayerToSeason.FromRoundID)
		assert.Equal(t, roundIDStr, *data.AddFFLPlayerToSeason.FromRoundID)

		// drv_name has been dropped from ffl.player — verify the row exists
		// linked to the AFL player and nothing more.
		var fflPlayerCount int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM ffl.player WHERE afl_player_id = $1", aflPlayerID).Scan(&fflPlayerCount))
		assert.Equal(t, 1, fflPlayerCount)
	})

	t.Run("reuses the existing FFL player when called again for a different club season", func(t *testing.T) {
		result := execQuery(t, server, `mutation {
			addFFLPlayerToSeason(input: {
				clubSeasonId: "`+fmt.Sprintf("%d", ids.awayClubSeaID)+`"
				aflPlayerSeasonId: "`+apsIDStr+`"
			}) {
				id
				player { id }
			}
		}`)
		require.Empty(t, result.Errors)

		// Only one ffl.player row should exist for this AFL player ID.
		var count int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM ffl.player WHERE afl_player_id = $1", aflPlayerID).Scan(&count))
		assert.Equal(t, 1, count)
	})
}

func TestCalculateFFLFantasyScore_StarPosition(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Update the seeded player match to star position.
	ctx := context.Background()
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET position = 'star' WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	pmID := fmt.Sprintf("%d", ids.playerMatchID)
	result := execQuery(t, server, `mutation {
		calculateFFLFantasyScore(input: {
			playerMatchId: "`+pmID+`"
			goals: 2
			kicks: 10
			handballs: 5
			marks: 4
			tackles: 3
			hitouts: 0
		}) {
			score
			position
		}
	}`)

	require.Empty(t, result.Errors)

	var data struct {
		CalculateFFLFantasyScore struct {
			Score    int    `json:"score"`
			Position string `json:"position"`
		} `json:"calculateFFLFantasyScore"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	// star: 2*5 + 10*1 + 5*1 + 4*2 + 3*4 = 10+10+5+8+12 = 45
	t.Run("star player score uses all stat multipliers", func(t *testing.T) {
		assert.Equal(t, 45, data.CalculateFFLFantasyScore.Score)
	})
}

// ---------------------------------------------------------------------------
// Phase 20 — graph traversal, in-season trades, notes, pageInfo.totalCount.
// ---------------------------------------------------------------------------

// insertAFLSeason creates a minimal afl.league + afl.season pair and registers
// cleanup. Returns the afl.season id. Used by FFL tests that exercise the
// fflSeason.aflSeason / fflRound.aflRound traversals — afl rows aren't part of
// the standard FFL seedTestData fixture.
func insertAFLSeason(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	ctx := context.Background()
	var leagueID, seasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.league (name) VALUES ('Test AFL League') RETURNING id").Scan(&leagueID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ('Test AFL 2025', $1) RETURNING id",
		leagueID).Scan(&seasonID))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.season WHERE id = $1", seasonID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.league WHERE id = $1", leagueID)
	})
	return seasonID
}

// insertAFLRound creates a minimal afl.round under the given afl.season.
// Returns the afl.round id with cleanup registered.
func insertAFLRound(t *testing.T, pool *pgxpool.Pool, aflSeasonID int) int {
	t.Helper()
	ctx := context.Background()
	var id int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('AFL R1', $1) RETURNING id",
		aflSeasonID).Scan(&id))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.round WHERE id = $1", id)
	})
	return id
}

// insertAFLPlayerSeason creates a minimal afl.player + club_season + player_season
// chain so the test can pass a real afl.player_season.id to addFFLPlayerToSeason.
func insertAFLPlayerSeason(t *testing.T, pool *pgxpool.Pool, aflSeasonID int) (aflPlayerSeasonID int) {
	t.Helper()
	ctx := context.Background()
	var clubID, clubSeasonID, playerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Test AFL Club') RETURNING id").Scan(&clubID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0) RETURNING id`,
		clubID, aflSeasonID).Scan(&clubSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Linked AFL Player') RETURNING id").Scan(&playerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		playerID, clubSeasonID).Scan(&aflPlayerSeasonID))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.player_season WHERE id = $1", aflPlayerSeasonID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.player WHERE id = $1", playerID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.club_season WHERE id = $1", clubSeasonID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM afl.club WHERE id = $1", clubID)
	})
	return aflPlayerSeasonID
}

func TestFflClubSeason_PageInfoTotalCount(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	clubID := fmt.Sprintf("%d", ids.homeClubID)
	result := execQuery(t, server, `{
		fflClubSeason(seasonId: "`+seasonID+`", clubId: "`+clubID+`") {
			players {
				nodes {
					id
					player { aflPlayer { id } }
					fromRoundId
					toRoundId
					notes
					costCents
				}
				pageInfo { hasNextPage endCursor totalCount }
			}
		}
	}`)
	require.Empty(t, result.Errors)

	var data struct {
		FflClubSeason struct {
			Players struct {
				Nodes []struct {
					ID          string  `json:"id"`
					Player      struct{ Name string } `json:"player"`
					FromRoundID *string `json:"fromRoundId"`
					ToRoundID   *string `json:"toRoundId"`
					Notes       *string `json:"notes"`
					CostCents   *int    `json:"costCents"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool    `json:"hasNextPage"`
					EndCursor   *string `json:"endCursor"`
					TotalCount  *int    `json:"totalCount"`
				} `json:"pageInfo"`
			} `json:"players"`
		} `json:"fflClubSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("totalCount lives on pageInfo (not on connection)", func(t *testing.T) {
		require.NotNil(t, data.FflClubSeason.Players.PageInfo.TotalCount)
		assert.Equal(t, 1, *data.FflClubSeason.Players.PageInfo.TotalCount)
		assert.Equal(t, len(data.FflClubSeason.Players.Nodes), *data.FflClubSeason.Players.PageInfo.TotalCount)
	})
	t.Run("nodes expose new player season fields (null when unset)", func(t *testing.T) {
		require.Len(t, data.FflClubSeason.Players.Nodes, 1)
		n := data.FflClubSeason.Players.Nodes[0]
		assert.Nil(t, n.FromRoundID)
		assert.Nil(t, n.ToRoundID)
		assert.Nil(t, n.Notes)
		assert.Nil(t, n.CostCents)
	})
}

func TestUpdateFFLPlayerSeason(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)

	t.Run("setting notes persists and is returned in the response", func(t *testing.T) {
		result := execQuery(t, server, `mutation {
			updateFFLPlayerSeason(input: { id: "`+psID+`", notes: "First two rounds quiet — backline rotation" }) {
				id
				notes
			}
		}`)
		require.Empty(t, result.Errors)

		var data struct {
			UpdateFFLPlayerSeason struct {
				ID    string  `json:"id"`
				Notes *string `json:"notes"`
			} `json:"updateFFLPlayerSeason"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		assert.Equal(t, psID, data.UpdateFFLPlayerSeason.ID)
		require.NotNil(t, data.UpdateFFLPlayerSeason.Notes)
		assert.Equal(t, "First two rounds quiet — backline rotation", *data.UpdateFFLPlayerSeason.Notes)

		var stored *string
		require.NoError(t, pool.QueryRow(context.Background(),
			"SELECT notes FROM ffl.player_season WHERE id = $1", ids.playerSeasonID).Scan(&stored))
		require.NotNil(t, stored)
		assert.Equal(t, "First two rounds quiet — backline rotation", *stored)
	})

	t.Run("clearing notes by passing null nulls the column", func(t *testing.T) {
		result := execQuery(t, server, `mutation {
			updateFFLPlayerSeason(input: { id: "`+psID+`", notes: null }) {
				notes
			}
		}`)
		require.Empty(t, result.Errors)

		var stored *string
		require.NoError(t, pool.QueryRow(context.Background(),
			"SELECT notes FROM ffl.player_season WHERE id = $1", ids.playerSeasonID).Scan(&stored))
		assert.Nil(t, stored)
	})
}

func TestRemoveFFLPlayerFromSeason_HistoryAndReAdd(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Seed an AFL player_season we can pass through addFFLPlayerToSeason.
	aflSeasonID := insertAFLSeason(t, pool)
	aflPlayerSeasonID := insertAFLPlayerSeason(t, pool, aflSeasonID)
	apsIDStr := fmt.Sprintf("%d", aflPlayerSeasonID)
	clubSeasonID := fmt.Sprintf("%d", ids.awayClubSeaID)

	addResult := execQuery(t, server, `mutation {
		addFFLPlayerToSeason(input: {
			clubSeasonId: "`+clubSeasonID+`"
			aflPlayerSeasonId: "`+apsIDStr+`"
		}) { id }
	}`)
	require.Empty(t, addResult.Errors)
	var added struct {
		AddFFLPlayerToSeason struct {
			ID string `json:"id"`
		} `json:"addFFLPlayerToSeason"`
	}
	require.NoError(t, json.Unmarshal(addResult.Data, &added))
	psID, err := strconv.Atoi(added.AddFFLPlayerToSeason.ID)
	require.NoError(t, err)

	roundID := fmt.Sprintf("%d", ids.roundID)
	removeResult := execQuery(t, server, `mutation {
		removeFFLPlayerFromSeason(input: { id: "`+added.AddFFLPlayerToSeason.ID+`", toRoundId: "`+roundID+`" })
	}`)
	require.Empty(t, removeResult.Errors)

	t.Run("row is preserved (not deleted) and to_round_id is set", func(t *testing.T) {
		var deletedAt *string
		var toRoundID *int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT to_round_id, deleted_at FROM ffl.player_season WHERE id = $1", psID).Scan(&toRoundID, &deletedAt))
		assert.Nil(t, deletedAt, "row should not be soft-deleted")
		require.NotNil(t, toRoundID)
		assert.Equal(t, ids.roundID, *toRoundID)
	})

	t.Run("traded row is still queryable via fflClubSeason.players with toRoundId set", func(t *testing.T) {
		seasonID := fmt.Sprintf("%d", ids.seasonID)
		awayClubID := fmt.Sprintf("%d", ids.awayClubID)
		result := execQuery(t, server, `{
			fflClubSeason(seasonId: "`+seasonID+`", clubId: "`+awayClubID+`") {
				players { nodes { id toRoundId } }
			}
		}`)
		require.Empty(t, result.Errors)

		var data struct {
			FflClubSeason struct {
				Players struct {
					Nodes []struct {
						ID        string  `json:"id"`
						ToRoundID *string `json:"toRoundId"`
					} `json:"nodes"`
				} `json:"players"`
			} `json:"fflClubSeason"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))

		var found bool
		for _, n := range data.FflClubSeason.Players.Nodes {
			if n.ID == added.AddFFLPlayerToSeason.ID {
				found = true
				require.NotNil(t, n.ToRoundID)
				assert.Equal(t, roundID, *n.ToRoundID)
			}
		}
		assert.True(t, found, "traded player season should still be returned")
	})

	t.Run("re-adding to same club season clears toRoundId (un-trade)", func(t *testing.T) {
		// addFFLPlayerToSeason reuses the existing ffl.player via afl_player_id
		// resolved from aflPlayerSeasonId, then upserts the player_season —
		// same (player_id, club_season_id) pair already exists from the call
		// above, so ON CONFLICT should clear to_round_id.
		result := execQuery(t, server, `mutation {
			addFFLPlayerToSeason(input: {
				clubSeasonId: "`+clubSeasonID+`"
				aflPlayerSeasonId: "`+apsIDStr+`"
			}) { id toRoundId }
		}`)
		require.Empty(t, result.Errors)

		var data struct {
			AddFFLPlayerToSeason struct {
				ID        string  `json:"id"`
				ToRoundID *string `json:"toRoundId"`
			} `json:"addFFLPlayerToSeason"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		assert.Equal(t, added.AddFFLPlayerToSeason.ID, data.AddFFLPlayerToSeason.ID, "should reuse the existing player_season row")
		assert.Nil(t, data.AddFFLPlayerToSeason.ToRoundID, "to_round_id should be cleared")
	})
}

func TestFFLSeason_AflSeasonTraversal(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	seasonIDStr := fmt.Sprintf("%d", ids.seasonID)

	t.Run("returns null when ffl.season.afl_season_id is unset", func(t *testing.T) {
		result := execQuery(t, server, `{
			fflSeason(id: "`+seasonIDStr+`") { aflSeason { id } }
		}`)
		require.Empty(t, result.Errors)

		var data struct {
			FflSeason struct {
				AflSeason *struct {
					ID string `json:"id"`
				} `json:"aflSeason"`
			} `json:"fflSeason"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		assert.Nil(t, data.FflSeason.AflSeason)
	})

	t.Run("returns AFL season id when linked", func(t *testing.T) {
		aflSeasonID := insertAFLSeason(t, pool)
		_, err := pool.Exec(context.Background(),
			"UPDATE ffl.season SET afl_season_id = $1 WHERE id = $2", aflSeasonID, ids.seasonID)
		require.NoError(t, err)
		// Restore so other subtests / cleanup behave.
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(),
				"UPDATE ffl.season SET afl_season_id = NULL WHERE id = $1", ids.seasonID)
		})

		result := execQuery(t, server, `{
			fflSeason(id: "`+seasonIDStr+`") { aflSeason { id } }
		}`)
		require.Empty(t, result.Errors)

		var data struct {
			FflSeason struct {
				AflSeason *struct {
					ID string `json:"id"`
				} `json:"aflSeason"`
			} `json:"fflSeason"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		require.NotNil(t, data.FflSeason.AflSeason)
		assert.Equal(t, fmt.Sprintf("%d", aflSeasonID), data.FflSeason.AflSeason.ID)
	})
}

func TestFFLRound_AflRoundTraversal(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	roundIDStr := fmt.Sprintf("%d", ids.roundID)

	t.Run("returns null when ffl.round.afl_round_id is unset", func(t *testing.T) {
		result := execQuery(t, server, `{
			fflRound(id: "`+roundIDStr+`") { aflRoundId aflRound { id } }
		}`)
		require.Empty(t, result.Errors)

		var data struct {
			FflRound struct {
				AflRoundID *string `json:"aflRoundId"`
				AflRound   *struct {
					ID string `json:"id"`
				} `json:"aflRound"`
			} `json:"fflRound"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		assert.Nil(t, data.FflRound.AflRoundID)
		assert.Nil(t, data.FflRound.AflRound)
	})

	t.Run("returns AFL round id when linked", func(t *testing.T) {
		aflSeasonID := insertAFLSeason(t, pool)
		aflRoundID := insertAFLRound(t, pool, aflSeasonID)
		_, err := pool.Exec(context.Background(),
			"UPDATE ffl.round SET afl_round_id = $1 WHERE id = $2", aflRoundID, ids.roundID)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(),
				"UPDATE ffl.round SET afl_round_id = NULL WHERE id = $1", ids.roundID)
		})

		result := execQuery(t, server, `{
			fflRound(id: "`+roundIDStr+`") { aflRoundId aflRound { id } }
		}`)
		require.Empty(t, result.Errors)

		var data struct {
			FflRound struct {
				AflRoundID *string `json:"aflRoundId"`
				AflRound   *struct {
					ID string `json:"id"`
				} `json:"aflRound"`
			} `json:"fflRound"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		require.NotNil(t, data.FflRound.AflRoundID)
		assert.Equal(t, fmt.Sprintf("%d", aflRoundID), *data.FflRound.AflRoundID)
		require.NotNil(t, data.FflRound.AflRound)
		assert.Equal(t, fmt.Sprintf("%d", aflRoundID), data.FflRound.AflRound.ID)
	})
}
