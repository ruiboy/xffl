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

type Match struct {
	ID        int
	RoundID   int
	Home      ClubMatch
	Away      ClubMatch
	Venue     string
	StartTime time.Time
	Result    MatchResult
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

type MatchRepository interface {
	FindByRoundID(ctx context.Context, roundID int) ([]Match, error)
	FindByID(ctx context.Context, id int) (Match, error)
	FindByIDWithDetails(ctx context.Context, id int) (Match, error)
}
