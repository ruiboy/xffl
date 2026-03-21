package domain

import "time"

// MatchResult represents the outcome of a match.
type MatchResult string

const (
	MatchResultHomeWin  MatchResult = "home_win"
	MatchResultAwayWin  MatchResult = "away_win"
	MatchResultDraw     MatchResult = "draw"
	MatchResultNoResult MatchResult = "no_result"
)

// PremiershipPoints awarded per match result.
const (
	PremiershipPointsWin  = 4
	PremiershipPointsDraw = 2
	PremiershipPointsLoss = 0
)

// PointsPerGoal is the number of points a goal is worth in AFL.
const PointsPerGoal = 6

type League struct {
	ID   int
	Name string
}

type Season struct {
	ID       int
	Name     string
	LeagueID int
}

type Round struct {
	ID       int
	Name     string
	SeasonID int
}

type Club struct {
	ID   int
	Name string
}

type Match struct {
	ID              int
	RoundID         int
	HomeClubMatchID int
	AwayClubMatchID int
	Venue           string
	StartTime       time.Time
	Result          MatchResult
}

type ClubSeason struct {
	ID                  int
	ClubID              int
	SeasonID            int
	Played              int
	Won                 int
	Lost                int
	Drawn               int
	For                 int
	Against             int
	PremiershipPoints   int
}

type ClubMatch struct {
	ID            int
	MatchID       int
	ClubSeasonID  int
	RushedBehinds int
	Score         int
}

type Player struct {
	ID     int
	Name   string
	ClubID int
}

type PlayerSeason struct {
	ID           int
	PlayerID     int
	ClubSeasonID int
}

type PlayerMatch struct {
	ID             int
	ClubMatchID    int
	PlayerSeasonID int
	Kicks          int
	Handballs      int
	Marks          int
	Hitouts        int
	Tackles        int
	Goals          int
	Behinds        int
}

// Disposals returns the total disposals (kicks + handballs).
func (pm PlayerMatch) Disposals() int {
	return pm.Kicks + pm.Handballs
}

// Score returns the player's scoring contribution in points (goals * 6 + behinds).
func (pm PlayerMatch) Score() int {
	return pm.Goals*PointsPerGoal + pm.Behinds
}

// UpsertPlayerMatchParams holds optional fields for creating or updating a PlayerMatch.
// Nil fields are left unchanged on update.
type UpsertPlayerMatchParams struct {
	ClubMatchID    int
	PlayerSeasonID int
	Kicks          *int
	Handballs      *int
	Marks          *int
	Hitouts        *int
	Tackles        *int
	Goals          *int
	Behinds        *int
}
