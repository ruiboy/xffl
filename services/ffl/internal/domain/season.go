package domain

import "context"

type Season struct {
	ID          int
	Name        string
	LeagueID    int
	AFLSeasonID int
	RulesID     string
}

type SeasonRepository interface {
	FindAll(ctx context.Context) ([]Season, error)
	FindByID(ctx context.Context, id int) (Season, error)
	Create(ctx context.Context, leagueID int, name string, aflSeasonID int, rulesID string) (Season, error)
}
