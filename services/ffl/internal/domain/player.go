package domain

import "context"

type Player struct {
	ID          int
	AFLPlayerID int
}

type PlayerRepository interface {
	FindAll(ctx context.Context) ([]Player, error)
	FindByID(ctx context.Context, id int) (Player, error)
	FindByAFLPlayerID(ctx context.Context, aflPlayerID int) (Player, error)
	Create(ctx context.Context, aflPlayerID int) (Player, error)
	Delete(ctx context.Context, id int) error
}
