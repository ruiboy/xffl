package zinc_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"xffl/services/search/internal/domain"
	zincinfra "xffl/services/search/internal/infrastructure/zinc"
)

var testBaseURL string

func TestMain(m *testing.M) {
	ctx := context.Background()

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "public.ecr.aws/zinclabs/zincsearch:latest",
			ExposedPorts: []string{"4080/tcp"},
			Env: map[string]string{
				"ZINC_FIRST_ADMIN_USER":     "admin",
				"ZINC_FIRST_ADMIN_PASSWORD": "admin",
				"ZINC_DATA_PATH":            "/data",
			},
			WaitingFor: wait.ForHTTP("/healthz").
				WithPort("4080/tcp").
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "start zinc container: %v\n", err)
		os.Exit(1)
	}

	host, err := ctr.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "zinc host: %v\n", err)
		ctr.Terminate(ctx) //nolint:errcheck
		os.Exit(1)
	}
	port, err := ctr.MappedPort(ctx, "4080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "zinc port: %v\n", err)
		ctr.Terminate(ctx) //nolint:errcheck
		os.Exit(1)
	}

	testBaseURL = fmt.Sprintf("http://%s:%s", host, port.Port())

	// Create index with keyword mappings for source/type before any test runs.
	repo := zincinfra.NewRepository(zincinfra.NewClient(testBaseURL, "admin", "admin"))
	if err := repo.EnsureIndex(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ensure zinc index: %v\n", err)
		ctr.Terminate(ctx) //nolint:errcheck
		os.Exit(1)
	}

	code := m.Run()
	ctr.Terminate(ctx) //nolint:errcheck
	os.Exit(code)
}

func newRepo(t *testing.T) *zincinfra.Repository {
	t.Helper()
	return zincinfra.NewRepository(zincinfra.NewClient(testBaseURL, "admin", "admin"))
}

// refreshWait gives Zinc time to make freshly-indexed docs searchable.
func refreshWait() { time.Sleep(1 * time.Second) }

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
	refreshWait()

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
	refreshWait()

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
	refreshWait()

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

	// Re-index same doc with updated data — should not error.
	doc.Data["kicks"] = 10
	require.NoError(t, repo.Index(ctx, doc))
	refreshWait()

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
