package typesense_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"xffl/services/search/internal/domain"
	ts "xffl/services/search/internal/infrastructure/typesense"
)

var testAPIURL string

func TestMain(m *testing.M) {
	ctx := context.Background()

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "typesense/typesense:27.1",
			ExposedPorts: []string{"8108/tcp"},
			Cmd:          []string{"--data-dir=/data", "--api-key=xyz"},
			Tmpfs:        map[string]string{"/data": "rw"},
			WaitingFor: wait.ForLog("Peer refresh succeeded").
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "start typesense container: %v\n", err)
		os.Exit(1)
	}

	port, err := ctr.MappedPort(ctx, "8108")
	if err != nil {
		fmt.Fprintf(os.Stderr, "typesense port: %v\n", err)
		ctr.Terminate(ctx) //nolint:errcheck
		os.Exit(1)
	}

	testAPIURL = fmt.Sprintf("http://127.0.0.1:%s", port.Port())

	// Typesense needs a moment after raft initialization before the API accepts requests.
	repo := ts.NewRepository(ts.NewClient(testAPIURL, "xyz"))
	var collErr error
	for range 5 {
		if collErr = repo.EnsureCollection(ctx); collErr == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if collErr != nil {
		fmt.Fprintf(os.Stderr, "ensure typesense collection: %v\n", collErr)
		ctr.Terminate(ctx) //nolint:errcheck
		os.Exit(1)
	}

	code := m.Run()
	ctr.Terminate(ctx) //nolint:errcheck
	os.Exit(code)
}

func newRepo(t *testing.T) *ts.Repository {
	t.Helper()
	return ts.NewRepository(ts.NewClient(testAPIURL, "xyz"))
}

func TestRepository_IndexAndSearch_RoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	doc := domain.SearchDocument{
		ID:     "afl_player_match_100",
		Source: domain.SourceAFL,
		Type:   domain.TypePlayerMatch,
		Data:   map[string]any{"player_match_id": 100, "kicks": 15, "goals": 3},
	}
	require.NoError(t, repo.Index(ctx, doc))

	result, err := repo.Search(ctx, domain.SearchQuery{Source: domain.SourceAFL})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, 1)

	var found bool
	for _, d := range result.Documents {
		if d.ID == doc.ID {
			found = true
			assert.Equal(t, domain.SourceAFL, d.Source)
			assert.Equal(t, domain.TypePlayerMatch, d.Type)
		}
	}
	assert.True(t, found, "indexed document not found in search results")
}

func TestRepository_SourceFiltering(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	aflDoc := domain.SearchDocument{
		ID:     "afl_player_match_200",
		Source: domain.SourceAFL,
		Type:   domain.TypePlayerMatch,
		Data:   map[string]any{"player_match_id": 200},
	}
	fflDoc := domain.SearchDocument{
		ID:     "ffl_fantasy_score_200",
		Source: domain.SourceFFL,
		Type:   domain.TypeFantasyScore,
		Data:   map[string]any{"player_match_id": 200, "score": 88},
	}
	require.NoError(t, repo.Index(ctx, aflDoc))
	require.NoError(t, repo.Index(ctx, fflDoc))

	aflResult, err := repo.Search(ctx, domain.SearchQuery{Source: domain.SourceAFL})
	require.NoError(t, err)
	for _, d := range aflResult.Documents {
		assert.Equal(t, domain.SourceAFL, d.Source, "expected only AFL docs")
	}

	fflResult, err := repo.Search(ctx, domain.SearchQuery{Source: domain.SourceFFL})
	require.NoError(t, err)
	for _, d := range fflResult.Documents {
		assert.Equal(t, domain.SourceFFL, d.Source, "expected only FFL docs")
	}
}

func TestRepository_TypeFiltering(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	pmDoc := domain.SearchDocument{
		ID:     "afl_player_match_300",
		Source: domain.SourceAFL,
		Type:   domain.TypePlayerMatch,
		Data:   map[string]any{"player_match_id": 300},
	}
	fsDoc := domain.SearchDocument{
		ID:     "ffl_fantasy_score_300",
		Source: domain.SourceFFL,
		Type:   domain.TypeFantasyScore,
		Data:   map[string]any{"player_match_id": 300, "score": 72},
	}
	require.NoError(t, repo.Index(ctx, pmDoc))
	require.NoError(t, repo.Index(ctx, fsDoc))

	result, err := repo.Search(ctx, domain.SearchQuery{Type: domain.TypeFantasyScore})
	require.NoError(t, err)
	for _, d := range result.Documents {
		assert.Equal(t, domain.TypeFantasyScore, d.Type, "expected only fantasy_score docs")
	}
}

func TestRepository_IdempotentReindex(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	doc := domain.SearchDocument{
		ID:     "afl_player_match_999",
		Source: domain.SourceAFL,
		Type:   domain.TypePlayerMatch,
		Data:   map[string]any{"player_match_id": 999, "kicks": 5},
	}
	require.NoError(t, repo.Index(ctx, doc))

	// Re-index same doc with updated data — upsert should not create a duplicate.
	doc.Data["kicks"] = 10
	require.NoError(t, repo.Index(ctx, doc))

	result, err := repo.Search(ctx, domain.SearchQuery{Source: domain.SourceAFL})
	require.NoError(t, err)

	var count int
	for _, d := range result.Documents {
		if d.ID == doc.ID {
			count++
		}
	}
	assert.Equal(t, 1, count, "re-indexing should not create duplicate documents")
}
