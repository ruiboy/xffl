package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// starter is a named starter at pos.
func starter(pos Position) PlayerMatch {
	return PlayerMatch{Position: PositionPtr(pos)}
}

// interchangeBench is a non-star bench player covering two positions, with an
// interchange slot on the first.
func interchangeBench() PlayerMatch {
	return PlayerMatch{
		BackupPositions:     strPtr("goals,kicks"),
		InterchangePosition: strPtr("goals"),
	}
}

// Team composition is validated by the season's Rules, not hardcoded constants:
// slot counts, bench size, and interchange availability all come from the rules.
func TestRulesValidate_DrivenByRules(t *testing.T) {
	t.Run("per-position slots come from the rules", func(t *testing.T) {
		// 2011 allows 3 goals starters; a 4th is rejected.
		team := []PlayerMatch{starter(PositionGoals), starter(PositionGoals), starter(PositionGoals)}
		require.NoError(t, rules2011.Validate(team))

		team = append(team, starter(PositionGoals))
		err := rules2011.Validate(team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum is 3")
	})

	t.Run("bench size comes from the rules", func(t *testing.T) {
		bench := []PlayerMatch{
			{BackupPositions: strPtr("goals,kicks")},
			{BackupPositions: strPtr("marks,tackles")},
			{BackupPositions: strPtr("handballs,hitouts")},
			{BackupPositions: strPtr("star")},
		}
		require.NoError(t, rules2011.Validate(bench), "4 bench players fit BenchSize 4")

		tight := rules2011
		tight.Composition.BenchSize = 2
		err := tight.Validate(bench)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum is 2")
	})

	t.Run("interchange is allowed only when the rules enable it", func(t *testing.T) {
		team := []PlayerMatch{starter(PositionGoals), interchangeBench()}

		require.NoError(t, rules2011.Validate(team), "2011 has interchange")

		// Same rules with interchange turned off (bench still allowed, to isolate
		// the interchange rule from the bench-size rule).
		noInterchange := rules2011
		noInterchange.Composition.Interchange = false
		err := noInterchange.Validate(team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "interchange is not permitted")
	})

	t.Run("a no-bench era rejects any bench player", func(t *testing.T) {
		// 1998–2000 had no bench (BenchSize 0).
		team := []PlayerMatch{starter(PositionGoals), {BackupPositions: strPtr("goals,kicks")}}
		err := rules1998.Validate(team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum is 0")
	})
}

// The structural bench/interchange composition rules (independent of the era's
// scalar parameters).
func TestRulesValidate_BenchAndInterchangeShape(t *testing.T) {
	t.Run("non-star bench player must have exactly 2 backup positions", func(t *testing.T) {
		err := rules2011.Validate([]PlayerMatch{{BackupPositions: strPtr("goals")}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly 2 backup positions")
	})

	t.Run("non-star bench player cannot list star as a backup", func(t *testing.T) {
		err := rules2011.Validate([]PlayerMatch{{BackupPositions: strPtr("goals,star")}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot list star as a backup position")
	})

	t.Run("a non-star position may be covered by at most one bench player", func(t *testing.T) {
		team := []PlayerMatch{
			{BackupPositions: strPtr("goals,kicks")},
			{BackupPositions: strPtr("goals,marks")}, // goals already covered
		}
		err := rules2011.Validate(team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already covered")
	})

	t.Run("interchange position must be a recognised position", func(t *testing.T) {
		err := rules2011.Validate([]PlayerMatch{
			{BackupPositions: strPtr("goals,kicks"), InterchangePosition: strPtr("bogus")},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a valid position")
	})

	t.Run("interchange position must be one of the player's own backups", func(t *testing.T) {
		err := rules2011.Validate([]PlayerMatch{
			{BackupPositions: strPtr("goals,kicks"), InterchangePosition: strPtr("marks")},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not one of this player's backup positions")
	})
}
