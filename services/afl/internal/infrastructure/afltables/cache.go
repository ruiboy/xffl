package afltables

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"xffl/shared/clock"
)

// fileCache caches afltables CSV responses on disk.
// Cache policy: fresh if fetched within 7 days, except on Monday where the
// cache is busted if it was fetched before today (to pick up weekend results).
type fileCache struct {
	dir   string
	clock clock.Clock
}

func (c *fileCache) filePath(year int) string {
	return filepath.Join(c.dir, fmt.Sprintf("afltables_%d_stats.txt", year))
}

// isFresh returns true if a valid cached file exists and does not need re-fetching.
func (c *fileCache) isFresh(year int) bool {
	info, err := os.Stat(c.filePath(year))
	if err != nil {
		return false
	}
	now := c.clock.Now()
	modTime := info.ModTime()

	if now.Weekday() == time.Monday {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if modTime.Before(today) {
			return false // Monday: bust if not fetched today
		}
	}

	return now.Sub(modTime) < 7*24*time.Hour
}

// get returns the cached bytes for the given year.
func (c *fileCache) get(year int) ([]byte, error) {
	data, err := os.ReadFile(c.filePath(year))
	if err != nil {
		return nil, fmt.Errorf("cache read: %w", err)
	}
	return data, nil
}

// put writes data to the cache for the given year.
func (c *fileCache) put(year int, data []byte) error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("cache mkdir: %w", err)
	}
	if err := os.WriteFile(c.filePath(year), data, 0600); err != nil {
		return fmt.Errorf("cache write: %w", err)
	}
	return nil
}
