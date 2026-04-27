package domain

import "context"

type Player struct {
	ID int
	// To be deleted - this is a denromalised value that will be retired.
	Name        string
	AFLPlayerID int
}

type PlayerRepository interface {
	FindAll(ctx context.Context) ([]Player, error)
	FindByID(ctx context.Context, id int) (Player, error)
	FindByAFLPlayerID(ctx context.Context, aflPlayerID int) (Player, error)
	Create(ctx context.Context, name string, aflPlayerID int) (Player, error)
	Update(ctx context.Context, id int, name string) (Player, error)
	Delete(ctx context.Context, id int) error
}
