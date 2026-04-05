package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayerMatch_Disposals(t *testing.T) {
	tests := []struct {
		name      string
		kicks     int
		handballs int
		want      int
	}{
		{"disposals are zero with no kicks or handballs", 0, 0, 0},
		{"kicks alone count as disposals", 10, 0, 10},
		{"handballs alone count as disposals", 0, 7, 7},
		{"kicks and handballs are summed into disposals", 12, 8, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := PlayerMatch{Kicks: tt.kicks, Handballs: tt.handballs}
			assert.Equal(t, tt.want, pm.Disposals())
		})
	}
}

func TestPlayerMatch_Score(t *testing.T) {
	tests := []struct {
		name    string
		goals   int
		behinds int
		want    int
	}{
		{"score is zero with no goals or behinds", 0, 0, 0},
		{"goals score six points each", 3, 0, 18},
		{"behinds score one point each", 0, 5, 5},
		{"goals and behinds are combined into total score", 2, 3, 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := PlayerMatch{Goals: tt.goals, Behinds: tt.behinds}
			assert.Equal(t, tt.want, pm.Score())
		})
	}
}
