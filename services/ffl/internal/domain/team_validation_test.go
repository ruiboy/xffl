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
		// 2015 allows 3 goals starters; a 4th is rejected.
		team := []PlayerMatch{starter(PositionGoals), starter(PositionGoals), starter(PositionGoals)}
		require.NoError(t, rules2015.Validate(team))

		team = append(team, starter(PositionGoals))
		err := rules2015.Validate(team)
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
		require.NoError(t, rules2015.Validate(bench), "4 bench players fit BenchSize 4")

		tight := rules2015
		tight.Composition.BenchSize = 2
		err := tight.Validate(bench)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum is 2")
	})

	t.Run("interchange is allowed only when the season enables it", func(t *testing.T) {
		team := []PlayerMatch{starter(PositionGoals), interchangeBench()}

		require.NoError(t, rules2015.Validate(team), "2015 has interchange")

		err := rules1998.Validate(team)
		require.Error(t, err, "1998 has no interchange")
		assert.Contains(t, err.Error(), "interchange is not permitted")
	})
}
