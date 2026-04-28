package footywire

import (
	"context"
	"strings"
	"unicode"

	"github.com/agnivade/levenshtein"

	"xffl/services/afl/internal/application"
)

const highConfidenceThreshold = 0.85

// LevenshteinResolver implements application.PlayerResolver using normalised
// Levenshtein distance for the AFL service.
type LevenshteinResolver struct{}

func NewLevenshteinResolver() *LevenshteinResolver {
	return &LevenshteinResolver{}
}

func (r *LevenshteinResolver) Resolve(_ context.Context, name, _ string, candidates []application.PlayerCandidate) ([]application.PlayerMatch, error) {
	norm := normaliseName(name)
	results := make([]application.PlayerMatch, 0, len(candidates))
	for _, c := range candidates {
		conf := similarity(norm, normaliseName(c.Name))
		results = append(results, application.PlayerMatch{
			Candidate:  c,
			Confidence: conf,
		})
	}
	sortByConfidence(results)
	return results, nil
}

func normaliseName(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSpace = false
		} else if !prevSpace {
			b.WriteRune(' ')
			prevSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	maxLen := len([]rune(a))
	if lb := len([]rune(b)); lb > maxLen {
		maxLen = lb
	}
	if maxLen == 0 {
		return 1.0
	}
	dist := levenshtein.ComputeDistance(a, b)
	score := 1.0 - float64(dist)/float64(maxLen)
	if score < 0 {
		return 0
	}
	return score
}

func sortByConfidence(matches []application.PlayerMatch) {
	for i := 1; i < len(matches); i++ {
		for j := i; j > 0 && matches[j].Confidence > matches[j-1].Confidence; j-- {
			matches[j], matches[j-1] = matches[j-1], matches[j]
		}
	}
}
