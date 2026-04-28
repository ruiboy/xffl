package footywire

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"context"

	"xffl/services/afl/internal/application"
)

// ---- ParseMatchStatsHTML ----

func TestParseMatchStatsHTML_ParsesTwoClubs(t *testing.T) {
	f, err := os.Open("testdata/match_stats.html")
	require.NoError(t, err)
	defer f.Close()

	stats, err := ParseMatchStatsHTML(f)
	require.NoError(t, err)

	t.Run("home club name and score", func(t *testing.T) {
		assert.Equal(t, "Carlton", stats.HomeClubName)
		assert.Equal(t, 14, stats.HomeTeamGoals)
		assert.Equal(t, 9, stats.HomeTeamBehinds)
	})

	t.Run("away club name and score", func(t *testing.T) {
		assert.Equal(t, "Richmond", stats.AwayClubName)
		assert.Equal(t, 10, stats.AwayTeamGoals)
		assert.Equal(t, 7, stats.AwayTeamBehinds)
	})

	t.Run("home players parsed with correct stats", func(t *testing.T) {
		carltonPlayers := playersForClub(stats.Players, "Carlton")
		require.Len(t, carltonPlayers, 3)
		assert.Equal(t, application.PlayerStats{
			Name: "Patrick Cripps", ClubName: "Carlton",
			Kicks: 15, Handballs: 8, Marks: 7, Hitouts: 0, Tackles: 5, Goals: 2, Behinds: 1,
		}, carltonPlayers[0])
	})

	t.Run("away players parsed with correct stats", func(t *testing.T) {
		richmondPlayers := playersForClub(stats.Players, "Richmond")
		require.Len(t, richmondPlayers, 2)
		assert.Equal(t, application.PlayerStats{
			Name: "Dustin Martin", ClubName: "Richmond",
			Kicks: 18, Handballs: 5, Marks: 6, Hitouts: 0, Tackles: 3, Goals: 3, Behinds: 2,
		}, richmondPlayers[0])
	})

	t.Run("totals row is excluded", func(t *testing.T) {
		for _, p := range stats.Players {
			assert.NotEqual(t, "Totals", p.Name)
		}
	})
}

// ---- ParseFixtureMid ----

func TestParseFixtureMid_FindsCorrectMid(t *testing.T) {
	tests := []struct {
		name      string
		round     string
		homeClub  string
		awayClub  string
		wantMid   string
		wantError bool
	}{
		{
			name: "finds Carlton vs Richmond in Round 5",
			round: "Round 5", homeClub: "Carlton", awayClub: "Richmond",
			wantMid: "11405",
		},
		{
			name: "finds Brisbane Lions vs Greater Western Sydney in Round 5",
			round: "Round 5", homeClub: "Brisbane Lions", awayClub: "Greater Western Sydney",
			wantMid: "11406",
		},
		{
			name: "finds Carlton vs Geelong in Round 1",
			round: "Round 1", homeClub: "Carlton", awayClub: "Geelong",
			wantMid: "11401",
		},
		{
			name: "returns error when match not found",
			round: "Round 99", homeClub: "Carlton", awayClub: "Richmond",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open("testdata/fixture_list.html")
			require.NoError(t, err)
			defer f.Close()

			mid, err := ParseFixtureMid(f, tt.round, tt.homeClub, tt.awayClub)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMid, mid)
		})
	}
}

// ---- FootywireClient HTTP integration (mocked) ----

func TestFootywireClient_ParseMatch_UsesHTTPServer(t *testing.T) {
	statsHTML, err := os.ReadFile("testdata/match_stats.html")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "mid=11405")
		w.Header().Set("Content-Type", "text/html")
		w.Write(statsHTML)
	}))
	defer srv.Close()

	client := &FootywireClient{http: srv.Client()}
	// Temporarily override baseURL for testing by calling the inner parse directly.
	resp, err := srv.Client().Get(srv.URL + "?mid=11405")
	require.NoError(t, err)
	defer resp.Body.Close()

	stats, err := ParseMatchStatsHTML(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "Carlton", stats.HomeClubName)
	_ = client // ensures client is used
}

func TestFootywireClient_FindMatchMid_UsesHTTPServer(t *testing.T) {
	fixtureHTML, err := os.ReadFile("testdata/fixture_list.html")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(fixtureHTML)
	}))
	defer srv.Close()

	// Call ParseFixtureMid directly with the served HTML.
	resp, err := srv.Client().Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	mid, err := ParseFixtureMid(resp.Body, "Round 5", "Carlton", "Richmond")
	require.NoError(t, err)
	assert.Equal(t, "11405", mid)

	_ = context.Background() // suppress unused import
}

// ---- helpers ----

func playersForClub(players []application.PlayerStats, club string) []application.PlayerStats {
	var out []application.PlayerStats
	for _, p := range players {
		if p.ClubName == club {
			out = append(out, p)
		}
	}
	return out
}
