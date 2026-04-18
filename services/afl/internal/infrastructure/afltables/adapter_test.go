package afltables

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xffl/shared/clock"
)

const fixtureCSV = `"Player","ID","Team","Opponent","Round","Kicks","Marks","Hand Balls","Disp","Goals","Behinds","Hit Outs","Tackles","Rebounds","Inside 50","Clearances","Clangers","Frees For","Frees Against","Brownlow","Contested Possessions","Uncontested Possessions","Contested Marks","Marks Inside 50","One Percenters","Bounces","Goal Assists","% Time Played"
"Joel Amartey",12844,"SY","CA","1",6,4,1,7,3,1,2,1,0,2,0,2,2,0,0,3,4,0,3,3,0,0,75
"Nick Blakey",12699,"SY","CA","1",12,3,9,21,0,0,0,1,4,5,0,6,0,2,0,5,15,0,0,4,4,1,88`

func TestAdapter_FetchSeasonStats_fetchesAndParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureCSV)) //nolint:errcheck
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv)
	rows, err := a.FetchSeasonStats(context.Background(), 2026)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "Joel Amartey", rows[0].PlayerName)
	assert.Equal(t, "Sydney Swans", rows[0].ClubName)
	assert.Equal(t, "Round 1", rows[0].RoundName)
}

func TestAdapter_FetchSeasonStats_usesCache(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureCSV)) //nolint:errcheck
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv)
	ctx := context.Background()

	_, err := a.FetchSeasonStats(ctx, 2026)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "first call should fetch from server")

	_, err = a.FetchSeasonStats(ctx, 2026)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "second call should use cache, not re-fetch")
}

func TestAdapter_FetchSeasonStats_serverError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv)
	_, err := a.FetchSeasonStats(context.Background(), 2026)
	assert.ErrorContains(t, err, "unexpected status 500")
}

// newTestAdapter creates an Adapter pointed at the given test server,
// with a fresh temp cache dir and a fixed Tuesday clock (cache always starts cold).
func newTestAdapter(t *testing.T, srv *httptest.Server) *Adapter {
	t.Helper()
	// Rewrite the URL template so the adapter hits the test server.
	a := &Adapter{
		cache:      &fileCache{dir: t.TempDir(), clock: clock.FixedClock{T: tuesday()}},
		httpClient: srv.Client(),
		urlTemplate: srv.URL + "/%d",
	}
	return a
}
