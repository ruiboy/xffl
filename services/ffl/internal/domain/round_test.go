package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundType_IsFinal(t *testing.T) {
	tests := []struct {
		name string
		rt   RoundType
		want bool
	}{
		{"minor round is not a final", RoundTypeMinor, false},
		{"empty/unknown counts as non-final", RoundType(""), false},
		{"semi-final is a final", RoundTypeSemiFinal, true},
		{"grand final is a final", RoundTypeGrandFinal, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rt.IsFinal())
		})
	}
}
