package dataops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalClub(t *testing.T) {
	t.Run("aliases of one club share a key", func(t *testing.T) {
		assert.Equal(t, canonicalClub("THC"), canonicalClub("The Howling Cows"))
	})

	t.Run("matching is case- and whitespace-insensitive", func(t *testing.T) {
		assert.Equal(t, canonicalClub("THC"), canonicalClub("  the howling cows "))
	})

	t.Run("an unknown club folds to itself, still matching its own spelling variants", func(t *testing.T) {
		assert.Equal(t, canonicalClub("Ruiboys"), canonicalClub(" ruiboys "))
	})

	t.Run("different clubs do not collide", func(t *testing.T) {
		assert.NotEqual(t, canonicalClub("THC"), canonicalClub("Cheetahs"))
	})
}
