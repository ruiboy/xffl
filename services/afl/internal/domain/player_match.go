package domain

import "context"

// PointsPerGoal is the number of points a goal is worth in AFL.
const PointsPerGoal = 6

type PlayerMatch struct {
	ID             int
	ClubMatchID    int
	PlayerSeasonID int
	Status         string
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
	Status         *string
	Kicks          *int
	Handballs      *int
	Marks          *int
	Hitouts        *int
	Tackles        *int
	Goals          *int
	Behinds        *int
}

// PlayerSeasonStats holds aggregated match statistics for a player across a season.
type PlayerSeasonStats struct {
	PlayerSeasonID int
	GamesPlayed    int
	AvgKicks       float64
	AvgHandballs   float64
	AvgMarks       float64
	AvgHitouts     float64
	AvgTackles     float64
	AvgGoals       float64
	AvgBehinds     float64
}

type PlayerMatchRepository interface {
	FindByClubMatchID(ctx context.Context, clubMatchID int) ([]PlayerMatch, error)
	FindByID(ctx context.Context, id int) (PlayerMatch, error)
	Upsert(ctx context.Context, params UpsertPlayerMatchParams) (PlayerMatch, error)
	FindStatsByPlayerSeasonIDs(ctx context.Context, ids []int) ([]PlayerSeasonStats, error)
}
