package forum

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
)

func findSquadMember(s application.ParsedSquad, name string) *application.ParsedSquadMember {
	for i := range s.Members {
		if s.Members[i].Name == name {
			return &s.Members[i]
		}
	}
	return nil
}

func TestParseSquads(t *testing.T) {
	b, err := os.ReadFile("testdata/squads_2025.txt")
	require.NoError(t, err)

	squads, err := NewSquadParser().ParseSquads(context.Background(), string(b))
	require.NoError(t, err)

	require.Len(t, squads, 4)
	assert.Equal(t, "Cheetahs", squads[0].ClubName)
	assert.Equal(t, "THC", squads[1].ClubName)
	assert.Equal(t, "RUIBOYS", squads[2].ClubName)
	assert.Equal(t, "SLASHERS", squads[3].ClubName)

	for _, s := range squads {
		assert.Len(t, s.Members, 30, "squad %s member count", s.ClubName)
	}

	t.Run("first member, decimal cost", func(t *testing.T) {
		m := squads[0].Members[0]
		assert.Equal(t, 1, m.Rank)
		assert.Equal(t, "Darcy Fogarty", m.Name)
		assert.Equal(t, "Adel", m.ClubHint)
		require.NotNil(t, m.CostCents)
		assert.Equal(t, 60, *m.CostCents)
	})

	t.Run("apostrophe in name", func(t *testing.T) {
		ob := findSquadMember(squads[2], "Reilly O'Brien")
		require.NotNil(t, ob)
		assert.Equal(t, "Adel", ob.ClubHint)
		require.NotNil(t, ob.CostCents)
		assert.Equal(t, 780, *ob.CostCents)
	})

	t.Run("hyphenated name", func(t *testing.T) {
		ldu := findSquadMember(squads[1], "Luke Davies-Uniacke")
		require.NotNil(t, ldu)
		assert.Equal(t, "NM", ldu.ClubHint)
	})

	t.Run("integer cost", func(t *testing.T) {
		neale := findSquadMember(squads[1], "Lachie Neale")
		require.NotNil(t, neale)
		assert.Equal(t, "Bris", neale.ClubHint)
		require.NotNil(t, neale.CostCents)
		assert.Equal(t, 100, *neale.CostCents)
	})

	t.Run("last member of last squad", func(t *testing.T) {
		m := squads[3].Members[29]
		assert.Equal(t, 30, m.Rank)
		assert.Equal(t, "Adam Treloar", m.Name)
		assert.Equal(t, "WB", m.ClubHint)
		require.NotNil(t, m.CostCents)
		assert.Equal(t, 120, *m.CostCents)
	})
}
