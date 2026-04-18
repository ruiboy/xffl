package afltables

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"xffl/services/afl/internal/application"
	"xffl/shared/clock"
)

const defaultURLTemplate = "https://afltables.com/afl/stats/%d_stats.txt"

// Adapter fetches AFL player match statistics from afltables.com and
// implements the application.StatsProvider outbound port.
type Adapter struct {
	cache       *fileCache
	httpClient  *http.Client
	urlTemplate string // fmt pattern, receives year as sole argument
}

// NewAdapter creates an Adapter that caches fetched CSV files in cacheDir.
func NewAdapter(cacheDir string) *Adapter {
	return &Adapter{
		cache:       &fileCache{dir: cacheDir, clock: clock.RealClock{}},
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		urlTemplate: defaultURLTemplate,
	}
}

// FetchSeasonStats returns all player match stats for the given year.
// Returns cached data if fresh; otherwise fetches from afltables.com and
// updates the cache.
func (a *Adapter) FetchSeasonStats(ctx context.Context, year int) ([]application.PlayerMatchStats, error) {
	data, err := a.load(ctx, year)
	if err != nil {
		return nil, err
	}
	return parseCSV(bytes.NewReader(data), year)
}

// load returns the raw CSV bytes, using the cache when fresh.
func (a *Adapter) load(ctx context.Context, year int) ([]byte, error) {
	if a.cache.isFresh(year) {
		data, err := a.cache.get(year)
		if err == nil {
			return data, nil
		}
		slog.WarnContext(ctx, "afltables: cache read failed, re-fetching", "year", year, "error", err)
	}

	data, err := a.fetch(ctx, year)
	if err != nil {
		return nil, err
	}

	if err := a.cache.put(year, data); err != nil {
		slog.WarnContext(ctx, "afltables: failed to write cache", "year", year, "error", err)
	}

	return data, nil
}

// fetch downloads the stats CSV from afltables.com.
func (a *Adapter) fetch(ctx context.Context, year int) ([]byte, error) {
	url := fmt.Sprintf(a.urlTemplate, year)
	slog.InfoContext(ctx, "afltables: fetching stats", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: unexpected status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return data, nil
}
