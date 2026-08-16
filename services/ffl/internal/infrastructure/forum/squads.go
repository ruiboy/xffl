package forum

import (
	"context"
	"math"
	"regexp"
	"strconv"
	"strings"

	"xffl/services/ffl/internal/application"
)

// SquadParser parses a squads thread. It implements application.SquadThreadParser.
type SquadParser struct{}

func NewSquadParser() *SquadParser { return &SquadParser{} }

// squadMemberRE matches a member line: rank, name, price, club code.
// The name group is non-greedy so the price (second-to-last token) and club
// (last token) bind to the trailing fields even for multi-word names.
var squadMemberRE = regexp.MustCompile(`^(\d+)\s+(.+?)\s+(\d+(?:\.\d+)?)\s+(\S+)$`)

// ParseSquads groups member lines under the club header that precedes them. A
// non-member, non-blank line starts a new squad; blank lines are separators.
func (p *SquadParser) ParseSquads(_ context.Context, text string) ([]application.ParsedSquad, error) {
	var squads []application.ParsedSquad
	cur := -1 // index of the squad currently being filled, -1 before the first header

	for _, raw := range splitLines(text) {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if m := squadMemberRE.FindStringSubmatch(line); m != nil && cur >= 0 {
			rank, _ := strconv.Atoi(m[1])
			squads[cur].Members = append(squads[cur].Members, application.ParsedSquadMember{
				Rank:      rank,
				Name:      strings.TrimSpace(m[2]),
				ClubHint:  m[4],
				CostCents: parseCents(m[3]),
			})
			continue
		}
		// Header line — starts a new club's squad.
		squads = append(squads, application.ParsedSquad{ClubName: line})
		cur = len(squads) - 1
	}
	return squads, nil
}

// parseCents turns a decimal price token ("0.6", "1", "7.8") into whole cents.
func parseCents(s string) *int {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	cents := int(math.Round(f * 100))
	return &cents
}
