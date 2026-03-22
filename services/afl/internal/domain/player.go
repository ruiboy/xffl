package domain

import "context"

type Player struct {
	ID     int
	Name   string
	ClubID int
}

type PlayerRepository interface {
	FindByClubID(ctx context.Context, clubID int) ([]Player, error)
	FindByID(ctx context.Context, id int) (Player, error)
}
