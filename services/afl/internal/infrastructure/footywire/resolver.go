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
	lev := levenshteinSimilarity(a, b)
	tok := tokenSimilarity(a, b)
	if tok > lev {
		return tok
	}
	return lev
}

func levenshteinSimilarity(a, b string) float64 {
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

// tokenSimilarity handles names where a token is a single-letter abbreviation for
// the corresponding token in the other name (e.g. "u" for "ugle", "h" for "horne").
// FootyWire abbreviates hyphenated surnames this way: "Ugle-Hagan" → "U-Hagan".
// Only activates when at least one token is a single letter; returns 0 otherwise.
func tokenSimilarity(a, b string) float64 {
	ta := strings.Fields(a)
	tb := strings.Fields(b)
	if len(ta) != len(tb) {
		return 0
	}
	hasInitial := false
	for _, t := range append(ta, tb...) {
		if len(t) == 1 {
			hasInitial = true
			break
		}
	}
	if !hasInitial {
		return 0
	}
	matched := 0
	for i := range ta {
		switch {
		case ta[i] == tb[i]:
			matched++
		case len(ta[i]) == 1 && strings.HasPrefix(tb[i], ta[i]):
			matched++ // "u" matches "ugle"
		case len(tb[i]) == 1 && strings.HasPrefix(ta[i], tb[i]):
			matched++ // reverse
		}
	}
	return float64(matched) / float64(len(ta))
}

func sortByConfidence(matches []application.PlayerMatch) {
	for i := 1; i < len(matches); i++ {
		for j := i; j > 0 && matches[j].Confidence > matches[j-1].Confidence; j-- {
			matches[j], matches[j-1] = matches[j-1], matches[j]
		}
	}
}
