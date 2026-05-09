package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateLadder(t *testing.T) {
	t.Run("home win accumulates correctly", func(t *testing.T) {
		matches := []Match{
			{
				Home: ClubMatch{ClubSeasonID: 1, StoredScore: 100},
				Away: ClubMatch{ClubSeasonID: 2, StoredScore: 80},
			},
		}
		got := CalculateLadder(matches)

		assert.Equal(t, ClubSeason{ID: 1, Played: 1, Won: 1, For: 100, Against: 80, PremiershipPoints: 4}, got[1])
		assert.Equal(t, ClubSeason{ID: 2, Played: 1, Lost: 1, For: 80, Against: 100}, got[2])
	})

	t.Run("away win accumulates correctly", func(t *testing.T) {
		matches := []Match{
			{
				Home: ClubMatch{ClubSeasonID: 1, StoredScore: 60},
				Away: ClubMatch{ClubSeasonID: 2, StoredScore: 90},
			},
		}
		got := CalculateLadder(matches)

		assert.Equal(t, ClubSeason{ID: 1, Played: 1, Lost: 1, For: 60, Against: 90}, got[1])
		assert.Equal(t, ClubSeason{ID: 2, Played: 1, Won: 1, For: 90, Against: 60, PremiershipPoints: 4}, got[2])
	})

	t.Run("draw awards 2 premiership points each", func(t *testing.T) {
		matches := []Match{
			{
				Home: ClubMatch{ClubSeasonID: 1, StoredScore: 75},
				Away: ClubMatch{ClubSeasonID: 2, StoredScore: 75},
			},
		}
		got := CalculateLadder(matches)

		assert.Equal(t, ClubSeason{ID: 1, Played: 1, Drawn: 1, For: 75, Against: 75, PremiershipPoints: 2}, got[1])
		assert.Equal(t, ClubSeason{ID: 2, Played: 1, Drawn: 1, For: 75, Against: 75, PremiershipPoints: 2}, got[2])
	})

	t.Run("multiple matches accumulate across rounds", func(t *testing.T) {
		matches := []Match{
			{Home: ClubMatch{ClubSeasonID: 1, StoredScore: 100}, Away: ClubMatch{ClubSeasonID: 2, StoredScore: 80}},
			{Home: ClubMatch{ClubSeasonID: 1, StoredScore: 60}, Away: ClubMatch{ClubSeasonID: 3, StoredScore: 90}},
		}
		got := CalculateLadder(matches)

		assert.Equal(t, ClubSeason{ID: 1, Played: 2, Won: 1, Lost: 1, For: 160, Against: 170, PremiershipPoints: 4}, got[1])
		assert.Equal(t, ClubSeason{ID: 2, Played: 1, Lost: 1, For: 80, Against: 100}, got[2])
		assert.Equal(t, ClubSeason{ID: 3, Played: 1, Won: 1, For: 90, Against: 60, PremiershipPoints: 4}, got[3])
	})

	t.Run("skips entries with missing club season IDs", func(t *testing.T) {
		matches := []Match{
			{Home: ClubMatch{ClubSeasonID: 0, StoredScore: 100}, Away: ClubMatch{ClubSeasonID: 2, StoredScore: 80}},
		}
		got := CalculateLadder(matches)

		assert.Empty(t, got)
	})

	t.Run("empty matches returns empty map", func(t *testing.T) {
		got := CalculateLadder(nil)
		assert.Empty(t, got)
	})
}
