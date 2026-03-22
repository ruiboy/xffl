package domain

import "context"

type Season struct {
	ID       int
	Name     string
	LeagueID int
}

type SeasonRepository interface {
	FindAll(ctx context.Context) ([]Season, error)
	FindByID(ctx context.Context, id int) (Season, error)
}
