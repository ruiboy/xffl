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

// MatchDataStatus tracks how complete the stats data is for a match.
type MatchDataStatus string

const (
	MatchDataNoData  MatchDataStatus = "no_data"
	MatchDataPartial MatchDataStatus = "partial"
	MatchDataFinal   MatchDataStatus = "final"
)

// PremiershipPoints awarded per match result.
const (
	PremiershipPointsWin  = 4
	PremiershipPointsDraw = 2
	PremiershipPointsLoss = 0
)

type Match struct {
	ID         int
	RoundID    int
	Home       ClubMatch
	Away       ClubMatch
	Venue      string
	StartTime  time.Time
	Result     MatchResult
	DataStatus MatchDataStatus
}

// DeriveResult computes the match result from StoredScore on each ClubMatch.
// Use when the match is loaded without full player match details.
func (m *Match) DeriveResult() MatchResult {
	if m.Home.StoredScore > m.Away.StoredScore {
		return MatchResultHomeWin
	}
	if m.Away.StoredScore > m.Home.StoredScore {
		return MatchResultAwayWin
	}
	return MatchResultDraw
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
	FindByIDs(ctx context.Context, ids []int) (map[int]Match, error)
	FindFinalBySeasonID(ctx context.Context, seasonID int) ([]Match, error)
	UpdateDataStatus(ctx context.Context, matchID int, status MatchDataStatus) error
	UpdateResult(ctx context.Context, matchID int, result MatchResult) error
}
