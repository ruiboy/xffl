//go:build integration

package graphql_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/afl/internal/application"
	pg "xffl/services/afl/internal/infrastructure/postgres"
	"xffl/services/afl/internal/infrastructure/postgres/sqlcgen"
	gql "xffl/services/afl/internal/interface/graphql"
	"xffl/shared/clock"
	memevents "xffl/shared/events/memory"
)

// db setup

func connectDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testPool
}

func setupTestServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
	t.Helper()
	return setupTestServerWithClock(t, pool, clock.RealClock{})
}

func setupTestServerWithClock(t *testing.T, pool *pgxpool.Pool, clk clock.Clock) *httptest.Server {
	t.Helper()

	q := sqlcgen.New(pool)
	queries := application.NewQueries(
		clk,
		pg.NewClubRepository(q),
		pg.NewSeasonRepository(q),
		pg.NewRoundRepository(q, pool),
		pg.NewMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewPlayerRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)

	db := pg.NewDB(pool)
	commands := application.NewCommands(db, memevents.New())

	resolver := &gql.Resolver{Queries: queries, Commands: commands}
	srv := gqlhandler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))

	return httptest.NewServer(srv)
}

// test fixtures

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
	playerID        int
	playerSeasonID  int
	playerMatchID   int
}

func seedTestData(t *testing.T, pool *pgxpool.Pool) testIDs {
	t.Helper()
	ctx := context.Background()
	var ids testIDs

	cleanupTestData(ctx, t, pool)

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.league (name) VALUES ('Test AFL') RETURNING id").Scan(&ids.leagueID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ('Test 2025', $1) RETURNING id",
		ids.leagueID).Scan(&ids.seasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		ids.seasonID).Scan(&ids.roundID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Sky Pilots') RETURNING id").Scan(&ids.homeClubID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Mountain Goats') RETURNING id").Scan(&ids.awayClubID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 5, 4, 1, 0, 500, 400, 16) RETURNING id`,
		ids.homeClubID, ids.seasonID).Scan(&ids.homeClubSeaID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 5, 3, 2, 0, 450, 420, 12) RETURNING id`,
		ids.awayClubID, ids.seasonID).Scan(&ids.awayClubSeaID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt) VALUES ($1, 'Test Ground', '2025-06-15 14:00:00') RETURNING id",
		ids.roundID).Scan(&ids.matchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds) VALUES ($1, $2, 85, 2) RETURNING id",
		ids.matchID, ids.homeClubSeaID).Scan(&ids.homeClubMatchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds) VALUES ($1, $2, 72, 1) RETURNING id",
		ids.matchID, ids.awayClubSeaID).Scan(&ids.awayClubMatchID))
	_, err := pool.Exec(ctx,
		"UPDATE afl.match SET home_club_match_id = $1, away_club_match_id = $2 WHERE id = $3",
		ids.homeClubMatchID, ids.awayClubMatchID, ids.matchID)
	require.NoError(t, err, "link match to club matches")
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Test Player') RETURNING id").Scan(&ids.playerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		ids.playerID, ids.homeClubSeaID).Scan(&ids.playerSeasonID))
	// 10 kicks, 5 handballs, 3 marks, 0 hitouts, 2 tackles, 2 goals, 1 behind
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.player_match (club_match_id, player_season_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
		 VALUES ($1, $2, 'played', 10, 5, 3, 0, 2, 2, 1) RETURNING id`,
		ids.homeClubMatchID, ids.playerSeasonID).Scan(&ids.playerMatchID))

	t.Cleanup(func() {
		cleanupTestData(context.Background(), t, pool)
	})

	return ids
}

func cleanupTestData(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	tables := []string{
		"afl.player_match",
		"afl.player_season",
		"afl.player",
		"afl.club_match",
		"afl.match",
		"afl.club_season",
		"afl.club",
		"afl.round",
		"afl.season",
		"afl.league",
	}
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
		require.NoError(t, err, "truncate %s", table)
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

func TestAflClubs(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ aflClubs { id name } }`)
	require.Empty(t, result.Errors)

	var data struct {
		AflClubs []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"aflClubs"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns both seeded clubs", func(t *testing.T) {
		assert.Len(t, data.AflClubs, 2)
	})
	t.Run("clubs ordered alphabetically", func(t *testing.T) {
		if assert.Len(t, data.AflClubs, 2) {
			assert.Equal(t, "Mountain Goats", data.AflClubs[0].Name)
			assert.Equal(t, "Sky Pilots", data.AflClubs[1].Name)
		}
	})
}

func TestAflSeasons(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ aflSeasons { id name } }`)
	require.Empty(t, result.Errors)

	var data struct {
		AflSeasons []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"aflSeasons"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns the seeded season", func(t *testing.T) {
		if assert.Len(t, data.AflSeasons, 1) {
			assert.Equal(t, "Test 2025", data.AflSeasons[0].Name)
		}
	})
}

func TestAflSeasonWithLadder(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ aflSeason(id: "`+fmt.Sprintf("%d", ids.seasonID)+`") { name ladder { club { name } played won lost premiershipPoints } } }`)
	require.Empty(t, result.Errors)

	var data struct {
		AflSeason struct {
			Name   string `json:"name"`
			Ladder []struct {
				Club struct {
					Name string `json:"name"`
				} `json:"club"`
				Played            int `json:"played"`
				Won               int `json:"won"`
				Lost              int `json:"lost"`
				PremiershipPoints int `json:"premiershipPoints"`
			} `json:"ladder"`
		} `json:"aflSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	require.Len(t, data.AflSeason.Ladder, 2)

	t.Run("ladder ordered by premiership points descending", func(t *testing.T) {
		assert.Equal(t, "Sky Pilots", data.AflSeason.Ladder[0].Club.Name)
		assert.Equal(t, 16, data.AflSeason.Ladder[0].PremiershipPoints)
		assert.Equal(t, "Mountain Goats", data.AflSeason.Ladder[1].Club.Name)
		assert.Equal(t, 12, data.AflSeason.Ladder[1].PremiershipPoints)
	})
}

func TestAflSeasonGraphTraversal(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{
		aflSeason(id: "`+fmt.Sprintf("%d", ids.seasonID)+`") {
			name
			rounds {
				name
				matches {
					venue
					homeClubMatch {
						club { name }
						score
						playerMatches {
							player { name }
							status
							kicks
							handballs
							disposals
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
		AflSeason struct {
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
								Name string `json:"name"`
							} `json:"player"`
							Status    *string `json:"status"`
							Kicks     int     `json:"kicks"`
							Handballs int     `json:"handballs"`
							Disposals int     `json:"disposals"`
							Score     int     `json:"score"`
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
		} `json:"aflSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	require.Len(t, data.AflSeason.Rounds, 1)
	require.Len(t, data.AflSeason.Rounds[0].Matches, 1)

	match := data.AflSeason.Rounds[0].Matches[0]

	t.Run("match venue and round name are correct", func(t *testing.T) {
		assert.Equal(t, "Round 1", data.AflSeason.Rounds[0].Name)
		assert.Equal(t, "Test Ground", match.Venue)
	})
	t.Run("home club match has correct club and score", func(t *testing.T) {
		require.NotNil(t, match.HomeClubMatch)
		assert.Equal(t, "Sky Pilots", match.HomeClubMatch.Club.Name)
		assert.Equal(t, 85, match.HomeClubMatch.Score)
	})
	t.Run("away club match has correct club and score", func(t *testing.T) {
		require.NotNil(t, match.AwayClubMatch)
		assert.Equal(t, "Mountain Goats", match.AwayClubMatch.Club.Name)
		assert.Equal(t, 72, match.AwayClubMatch.Score)
	})
	t.Run("player match disposals and score are derived correctly", func(t *testing.T) {
		require.NotNil(t, match.HomeClubMatch)
		require.Len(t, match.HomeClubMatch.PlayerMatches, 1)
		pm := match.HomeClubMatch.PlayerMatches[0]
		assert.Equal(t, "Test Player", pm.Player.Name)
		assert.Equal(t, 10, pm.Kicks)
		assert.Equal(t, 15, pm.Disposals) // 10 kicks + 5 handballs
		assert.Equal(t, "played", *pm.Status)
		assert.Equal(t, 13, pm.Score) // 2*6 + 1
	})
}

func TestUpdateAFLPlayerMatch_Update(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Change kicks from 10 to 20; other fields unchanged
	mutation := fmt.Sprintf(`mutation {
		updateAFLPlayerMatch(input: {
			playerSeasonId: "%d"
			clubMatchId: "%d"
			kicks: 20
		}) {
			id player { name } kicks handballs disposals goals behinds score
		}
	}`, ids.playerSeasonID, ids.homeClubMatchID)

	result := execQuery(t, server, mutation)
	require.Empty(t, result.Errors)

	var data struct {
		UpdateAFLPlayerMatch struct {
			ID     string `json:"id"`
			Player struct {
				Name string `json:"name"`
			} `json:"player"`
			Kicks     int `json:"kicks"`
			Handballs int `json:"handballs"`
			Disposals int `json:"disposals"`
			Goals     int `json:"goals"`
			Behinds   int `json:"behinds"`
			Score     int `json:"score"`
		} `json:"updateAFLPlayerMatch"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	pm := data.UpdateAFLPlayerMatch

	t.Run("updated field reflects new value", func(t *testing.T) {
		assert.Equal(t, 20, pm.Kicks)
	})
	t.Run("unspecified fields remain unchanged", func(t *testing.T) {
		assert.Equal(t, 5, pm.Handballs)
	})
	t.Run("derived fields recalculate from new values", func(t *testing.T) {
		assert.Equal(t, 25, pm.Disposals) // 20 kicks + 5 handballs
		assert.Equal(t, 13, pm.Score)     // 2 goals * 6 + 1 behind, unchanged
	})
}

func TestUpdateAFLPlayerMatch_Create(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Insert a second player + player_season for the away team
	ctx := context.Background()
	var player2ID, playerSeason2ID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Test Player 2') RETURNING id").Scan(&player2ID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		player2ID, ids.awayClubSeaID).Scan(&playerSeason2ID))

	mutation := fmt.Sprintf(`mutation {
		updateAFLPlayerMatch(input: {
			playerSeasonId: "%d"
			clubMatchId: "%d"
			kicks: 8
			handballs: 4
			marks: 6
			hitouts: 0
			tackles: 3
			goals: 1
			behinds: 2
		}) {
			id player { name } kicks handballs disposals goals behinds score
		}
	}`, playerSeason2ID, ids.awayClubMatchID)

	result := execQuery(t, server, mutation)
	require.Empty(t, result.Errors)

	var data struct {
		UpdateAFLPlayerMatch struct {
			ID     string `json:"id"`
			Player struct {
				Name string `json:"name"`
			} `json:"player"`
			Kicks     int `json:"kicks"`
			Handballs int `json:"handballs"`
			Disposals int `json:"disposals"`
			Goals     int `json:"goals"`
			Behinds   int `json:"behinds"`
			Score     int `json:"score"`
		} `json:"updateAFLPlayerMatch"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))
	pm := data.UpdateAFLPlayerMatch

	t.Run("new record is created with correct player", func(t *testing.T) {
		assert.Equal(t, "Test Player 2", pm.Player.Name)
	})
	t.Run("derived fields are calculated from provided values", func(t *testing.T) {
		assert.Equal(t, 8, pm.Kicks)
		assert.Equal(t, 12, pm.Disposals) // 8 kicks + 4 handballs
		assert.Equal(t, 8, pm.Score)      // 1 goal * 6 + 2 behinds
	})
}

func TestUpdateAFLPlayerMatch_RecalculatesClubMatchScore(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Change goals 2→5, behinds 1→3: player score = 5*6+3 = 33, club score = 33+2 rushed = 35
	mutation := fmt.Sprintf(`mutation {
		updateAFLPlayerMatch(input: {
			playerSeasonId: "%d"
			clubMatchId: "%d"
			goals: 5
			behinds: 3
		}) {
			id goals behinds score
		}
	}`, ids.playerSeasonID, ids.homeClubMatchID)

	result := execQuery(t, server, mutation)
	require.Empty(t, result.Errors)

	queryResult := execQuery(t, server, `{
		aflSeason(id: "`+fmt.Sprintf("%d", ids.seasonID)+`") {
			rounds { matches { homeClubMatch { score } } }
		}
	}`)
	require.Empty(t, queryResult.Errors)

	var data struct {
		AflSeason struct {
			Rounds []struct {
				Matches []struct {
					HomeClubMatch struct {
						Score int `json:"score"`
					} `json:"homeClubMatch"`
				} `json:"matches"`
			} `json:"rounds"`
		} `json:"aflSeason"`
	}
	require.NoError(t, json.Unmarshal(queryResult.Data, &data))

	t.Run("club match score recalculates from updated player goals and behinds", func(t *testing.T) {
		homeScore := data.AflSeason.Rounds[0].Matches[0].HomeClubMatch.Score
		assert.Equal(t, 35, homeScore) // 5*6 + 3 + 2 rushed behinds
	})
}

func TestAflPlayerSearch(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	t.Run("matching query returns the player", func(t *testing.T) {
		result := execQuery(t, server, `{ aflPlayerSearch(query: "Test") { id name } }`)
		require.Empty(t, result.Errors)

		var data struct {
			AflPlayerSearch []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"aflPlayerSearch"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		if assert.Len(t, data.AflPlayerSearch, 1) {
			assert.Equal(t, "Test Player", data.AflPlayerSearch[0].Name)
		}
	})

	t.Run("non-matching query returns empty results", func(t *testing.T) {
		result := execQuery(t, server, `{ aflPlayerSearch(query: "Nonexistent") { id name } }`)
		require.Empty(t, result.Errors)

		var data struct {
			AflPlayerSearch []struct {
				ID string `json:"id"`
			} `json:"aflPlayerSearch"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		assert.Empty(t, data.AflPlayerSearch)
	})
}

func TestUpdateAFLPlayerMatch_InvalidPlayerSeasonID(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	mutation := fmt.Sprintf(`mutation {
		updateAFLPlayerMatch(input: {
			playerSeasonId: "999999"
			clubMatchId: "%d"
			kicks: 5
		}) { id }
	}`, ids.homeClubMatchID)

	result := execQuery(t, server, mutation)

	t.Run("returns a graphql error for unknown player season", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

func TestAflLiveRound(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool) // Round 1, single match at 2025-06-15 14:00:00 UTC = 2025-06-15 23:30 ACST

	// Round 2: first match 2025-06-22 14:00:00 UTC = 2025-06-22 23:30 ACST
	// midnight Adelaide before that = 2025-06-22 00:00 ACST = 2025-06-21 14:30:00 UTC
	var round2ID int
	require.NoError(t, pool.QueryRow(context.Background(),
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 2', $1) RETURNING id",
		ids.seasonID).Scan(&round2ID))
	_, err := pool.Exec(context.Background(),
		"INSERT INTO afl.match (round_id, venue, start_dt) VALUES ($1, 'Test Ground', '2025-06-22 14:00:00')",
		round2ID)
	require.NoError(t, err)

	mustParseRFC3339 := func(s string) time.Time {
		t.Helper()
		tm, err := time.Parse(time.RFC3339, s)
		require.NoError(t, err)
		return tm
	}

	t.Run("returns Round 1 when clock is after the match start", func(t *testing.T) {
		// 2025-06-16 00:00 UTC — match has started (2025-06-15 14:00 UTC), no upcoming round
		clk := clock.FixedClock{T: mustParseRFC3339("2025-06-16T00:00:00Z")}
		server := setupTestServerWithClock(t, pool, clk)
		defer server.Close()

		result := execQuery(t, server, `{ aflLiveRound { round { name } startDate } }`)
		require.Empty(t, result.Errors)

		var data struct {
			AflLiveRound *struct {
				Round     struct{ Name string } `json:"round"`
				StartDate string                `json:"startDate"`
			} `json:"aflLiveRound"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		require.NotNil(t, data.AflLiveRound)
		assert.Equal(t, "Round 1", data.AflLiveRound.Round.Name)
		assert.Equal(t, "2025-06-15T14:00:00Z", data.AflLiveRound.StartDate)
	})

	t.Run("returns nil when clock is before any round has started", func(t *testing.T) {
		// 2025-01-01 00:00 UTC — before Round 1's match on 2025-06-15
		clk := clock.FixedClock{T: mustParseRFC3339("2025-01-01T00:00:00Z")}
		server := setupTestServerWithClock(t, pool, clk)
		defer server.Close()

		result := execQuery(t, server, `{ aflLiveRound { round { name } startDate } }`)
		require.Empty(t, result.Errors)

		var data struct {
			AflLiveRound *struct {
				Round struct{ Name string } `json:"round"`
			} `json:"aflLiveRound"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		assert.Nil(t, data.AflLiveRound)
	})

	t.Run("returns previous round when between rounds and not yet on next round's Adelaide day", func(t *testing.T) {
		// 2025-06-18 00:00 UTC — R1 has started, R2 hasn't; midnight Adelaide before R2
		// is 2025-06-21 14:30 UTC, so we are before the transition window
		clk := clock.FixedClock{T: mustParseRFC3339("2025-06-18T00:00:00Z")}
		server := setupTestServerWithClock(t, pool, clk)
		defer server.Close()

		result := execQuery(t, server, `{ aflLiveRound { round { name } startDate } }`)
		require.Empty(t, result.Errors)

		var data struct {
			AflLiveRound *struct {
				Round     struct{ Name string } `json:"round"`
				StartDate string                `json:"startDate"`
			} `json:"aflLiveRound"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		require.NotNil(t, data.AflLiveRound)
		assert.Equal(t, "Round 1", data.AflLiveRound.Round.Name)
	})

	t.Run("returns upcoming round once past midnight Adelaide on its first match day", func(t *testing.T) {
		// 2025-06-22 00:00 UTC = 2025-06-22 09:30 ACST — past midnight Adelaide of R2's match day
		// (midnight ACST crossed at 2025-06-21 14:30 UTC), but R2's first match is still 14 h away
		clk := clock.FixedClock{T: mustParseRFC3339("2025-06-22T00:00:00Z")}
		server := setupTestServerWithClock(t, pool, clk)
		defer server.Close()

		result := execQuery(t, server, `{ aflLiveRound { round { name } startDate } }`)
		require.Empty(t, result.Errors)

		var data struct {
			AflLiveRound *struct {
				Round     struct{ Name string } `json:"round"`
				StartDate string                `json:"startDate"`
			} `json:"aflLiveRound"`
		}
		require.NoError(t, json.Unmarshal(result.Data, &data))
		require.NotNil(t, data.AflLiveRound)
		assert.Equal(t, "Round 2", data.AflLiveRound.Round.Name)
		assert.Equal(t, "2025-06-22T14:00:00Z", data.AflLiveRound.StartDate)
	})
}
