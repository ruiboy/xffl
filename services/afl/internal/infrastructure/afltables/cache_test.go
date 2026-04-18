package afltables

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xffl/shared/clock"
)

func TestFileCache_isFresh_missingFile(t *testing.T) {
	c := &fileCache{dir: t.TempDir(), clock: clock.RealClock{}}
	assert.False(t, c.isFresh(2026), "missing file should not be fresh")
}

func TestFileCache_isFresh_recentFile(t *testing.T) {
	dir := t.TempDir()
	c := &fileCache{dir: dir, clock: clock.FixedClock{T: tuesday()}}
	require.NoError(t, os.WriteFile(c.filePath(2026), []byte("data"), 0600))
	// Set mtime to 3 days ago (still within 7-day window, not Monday)
	mtime := tuesday().Add(-3 * 24 * time.Hour)
	require.NoError(t, os.Chtimes(c.filePath(2026), mtime, mtime))
	assert.True(t, c.isFresh(2026), "file fetched within 7 days on a non-Monday should be fresh")
}

func TestFileCache_isFresh_staleFile(t *testing.T) {
	dir := t.TempDir()
	c := &fileCache{dir: dir, clock: clock.FixedClock{T: tuesday()}}
	require.NoError(t, os.WriteFile(c.filePath(2026), []byte("data"), 0600))
	mtime := tuesday().Add(-8 * 24 * time.Hour)
	require.NoError(t, os.Chtimes(c.filePath(2026), mtime, mtime))
	assert.False(t, c.isFresh(2026), "file older than 7 days should not be fresh")
}

func TestFileCache_isFresh_mondayBusts(t *testing.T) {
	dir := t.TempDir()
	c := &fileCache{dir: dir, clock: clock.FixedClock{T: monday()}}
	require.NoError(t, os.WriteFile(c.filePath(2026), []byte("data"), 0600))
	// File was fetched last Wednesday — should be busted on Monday
	mtime := monday().Add(-5 * 24 * time.Hour)
	require.NoError(t, os.Chtimes(c.filePath(2026), mtime, mtime))
	assert.False(t, c.isFresh(2026), "cache fetched before today should be stale on Monday")
}

func TestFileCache_isFresh_mondayAlreadyFetchedToday(t *testing.T) {
	dir := t.TempDir()
	now := monday()
	c := &fileCache{dir: dir, clock: clock.FixedClock{T: now}}
	require.NoError(t, os.WriteFile(c.filePath(2026), []byte("data"), 0600))
	// File was fetched earlier today (Monday)
	mtime := now.Add(-1 * time.Hour)
	require.NoError(t, os.Chtimes(c.filePath(2026), mtime, mtime))
	assert.True(t, c.isFresh(2026), "cache fetched today on Monday should still be fresh")
}

func TestFileCache_roundtrip(t *testing.T) {
	c := &fileCache{dir: t.TempDir(), clock: clock.RealClock{}}
	data := []byte("some,csv,data\n1,2,3\n")
	require.NoError(t, c.put(2026, data))
	got, err := c.get(2026)
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

// monday returns a fixed Monday at 10:00 UTC.
func monday() time.Time {
	return time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC) // April 20 2026 is a Monday
}

// tuesday returns a fixed Tuesday at 10:00 UTC.
func tuesday() time.Time {
	return time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
}
