package domain

import (
	"context"
	"time"
)

// MatchResult represents the outcome of a match.
type MatchResult string

const (
	MatchResultHomeWin  MatchResult = "home_win"
	MatchResultAwayWin  MatchResult = "away_win"
	MatchResultDraw     MatchResult = "draw"
	MatchResultNoResult MatchResult = "no_result"
)

// MatchStyle classifies a match by shape — a regular home-vs-away contest, a
// single-club scoring bye, or an all-club superbye. It is the single discriminator
// the ladder and fixture builder dispatch on, so it is always one of these values.
type MatchStyle string

const (
	MatchStyleVersus   MatchStyle = "versus"
	MatchStyleBye      MatchStyle = "bye"
	MatchStyleSuperbye MatchStyle = "superbye"
)

// ParseMatchStyle normalises a stored match_style into a known style. Empty or
// unrecognised values are treated as versus, so callers never see a blank style.
func ParseMatchStyle(s string) MatchStyle {
	switch MatchStyle(s) {
	case MatchStyleBye:
		return MatchStyleBye
	case MatchStyleSuperbye:
		return MatchStyleSuperbye
	default:
		return MatchStyleVersus
	}
}

type Match struct {
	ID         int
	RoundID    int
	RoundType  RoundType
	MatchStyle MatchStyle
	Home       ClubMatch
	Away       ClubMatch
	Venue      string
	StartTime  time.Time
	Result     MatchResult
}

// Winner returns a pointer to the winning ClubMatch, or nil for a draw.
func (m *Match) Winner() *ClubMatch {
	homeScore := m.Home.Score()
	awayScore := m.Away.Score()
	if homeScore > awayScore {
		return &m.Home
	}
	if awayScore > homeScore {
		return &m.Away
	}
	return nil
}

// DeriveResult derives the match result from the stored (denormalised) scores on each ClubMatch.
func (m *Match) DeriveResult() MatchResult {
	if m.Home.StoredScore > m.Away.StoredScore {
		return MatchResultHomeWin
	}
	if m.Away.StoredScore > m.Home.StoredScore {
		return MatchResultAwayWin
	}
	return MatchResultDraw
}

type MatchRepository interface {
	FindByRoundID(ctx context.Context, roundID int) ([]Match, error)
	FindByID(ctx context.Context, id int) (Match, error)
	FindByIDWithDetails(ctx context.Context, id int) (Match, error)
	FindByIDs(ctx context.Context, ids []int) (map[int]Match, error)
	UpdateResult(ctx context.Context, matchID int, result MatchResult) error
	// Create inserts a match in a round. matchStyle is nil for a regular
	// home-vs-away match
	Create(ctx context.Context, roundID int, matchStyle *string) (Match, error)
	// Matches dropped by the fixture builder are removed outright — they only
	// ever belong to rounds with no submitted teams. Their club_matches follow
	// via ON DELETE CASCADE.
	DeleteByRoundID(ctx context.Context, roundID int) error
	DeleteByID(ctx context.Context, id int) error
}
