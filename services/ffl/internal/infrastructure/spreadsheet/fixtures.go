// Package spreadsheet holds the import-only parsers for pasted FFL spreadsheet
// data — currently the season fixture sheet. It is a source-named infrastructure
// package (ADR-021): a paste from the league spreadsheet is the source; formats
// it produces are files here. Parser ports and DTOs live in the application layer.
package spreadsheet

import (
	"context"
	"strconv"
	"strings"

	"xffl/services/ffl/internal/application"
)

// FixtureParser parses a pasted season fixture sheet. It implements
// application.FixtureSheetParser.
type FixtureParser struct{}

func NewFixtureParser() *FixtureParser { return &FixtureParser{} }

// ParseFixtures reads the sheet top to bottom. A line whose first tab-field is a
// number opens a home-and-away round (any trailing text becomes its Label, e.g.
// "Super Bye"); a non-numeric label line with no fixture opens a finals round
// (Round 0, Label the name, e.g. "Semi-Final", "Grand Final"). Fixture lines —
// "Home<TAB>Score<TAB>vs<TAB>Away<TAB>Score" — attach to the open round. Blank
// lines, the "Fixtures" header, and "n/a" placeholder lines are ignored.
func (p *FixtureParser) ParseFixtures(_ context.Context, text string) ([]application.ParsedFixtureRound, error) {
	var rounds []application.ParsedFixtureRound
	cur := -1 // index of the open round, -1 before the first round line

	for _, raw := range splitLines(text) {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(line), "Fixtures") {
			continue
		}

		fields := strings.Split(line, "\t")

		// Fixture line: contains a "vs" separator.
		if fx, ok := parseFixtureLine(fields); ok {
			if cur >= 0 {
				rounds[cur].Fixtures = append(rounds[cur].Fixtures, fx)
			}
			continue
		}

		first := strings.TrimSpace(fields[0])
		// Skip filler with no opening token: empty leads (e.g. an unplayed "  vs")
		// and "n/a" placeholder pairings are neither rounds nor fixtures.
		if first == "" || isEmptyOrNA(first) {
			continue
		}
		if n, err := strconv.Atoi(first); err == nil {
			// Numbered home-and-away round; trailing text is a label (e.g. "Super Bye").
			rounds = append(rounds, application.ParsedFixtureRound{Round: n, Label: joinLabel(fields[1:])})
			cur = len(rounds) - 1
			continue
		}
		// Non-numeric, non-fixture line: a finals round label ("Semi-Final", "Grand Final").
		rounds = append(rounds, application.ParsedFixtureRound{Label: joinLabel(fields)})
		cur = len(rounds) - 1
	}
	return rounds, nil
}

// parseFixtureLine reads "Home Score vs Away Score" from tab fields. It reports
// false for any line without a "vs" token, or whose named clubs are empty or
// "n/a" placeholders — those are not fixtures.
func parseFixtureLine(fields []string) (application.ParsedFixture, bool) {
	vs := -1
	for i, f := range fields {
		if strings.EqualFold(strings.TrimSpace(f), "vs") {
			vs = i
			break
		}
	}
	if vs < 0 {
		return application.ParsedFixture{}, false
	}
	homeClub, homeScore := clubAndScore(fields[:vs])
	awayClub, awayScore := clubAndScore(fields[vs+1:])
	if isEmptyOrNA(homeClub) || isEmptyOrNA(awayClub) {
		return application.ParsedFixture{}, false
	}
	return application.ParsedFixture{
		HomeClub:  homeClub,
		HomeScore: homeScore,
		AwayClub:  awayClub,
		AwayScore: awayScore,
	}, true
}

// clubAndScore reads a club name and optional score from one side of a fixture
// line. The club is the first non-empty non-numeric field; the score is the first
// numeric field (nil if none).
func clubAndScore(side []string) (string, *int) {
	var club string
	var score *int
	for _, f := range side {
		t := strings.TrimSpace(f)
		if t == "" {
			continue
		}
		if n, err := strconv.Atoi(t); err == nil {
			if score == nil {
				v := n
				score = &v
			}
			continue
		}
		if club == "" {
			club = t
		}
	}
	return club, score
}

func isEmptyOrNA(club string) bool {
	return club == "" || strings.HasPrefix(strings.ToLower(club), "n/a")
}

// joinLabel trims and joins non-empty fields into a single label string.
func joinLabel(fields []string) string {
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		if t := strings.TrimSpace(f); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}
