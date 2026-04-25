package graphql

import (
	"context"
	"fmt"

	"github.com/vikstrous/dataloadgen"

	"xffl/services/afl/internal/application"
	"xffl/services/afl/internal/domain"
)

type loadersKey struct{}

type Loaders struct {
	PlayerByPlayerSeasonID *dataloadgen.Loader[int, *domain.Player]
	ClubByID               *dataloadgen.Loader[int, *domain.Club]
	MatchByID              *dataloadgen.Loader[int, *domain.Match]
}

func NewLoaders(q *application.Queries) *Loaders {
	return &Loaders{
		PlayerByPlayerSeasonID: dataloadgen.NewLoader(func(ctx context.Context, ids []int) ([]*domain.Player, []error) {
			m, err := q.GetPlayersForPlayerSeasonIDs(ctx, ids)
			return mapToSlice(ids, m, err)
		}),
		ClubByID: dataloadgen.NewLoader(func(ctx context.Context, ids []int) ([]*domain.Club, []error) {
			m, err := q.GetClubsByIDs(ctx, ids)
			return mapToSlice(ids, m, err)
		}),
		MatchByID: dataloadgen.NewLoader(func(ctx context.Context, ids []int) ([]*domain.Match, []error) {
			m, err := q.GetMatchesByIDs(ctx, ids)
			return mapToSlice(ids, m, err)
		}),
	}
}

func InjectLoaders(ctx context.Context, l *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey{}, l)
}

func LoadersFromCtx(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey{}).(*Loaders)
}

// mapToSlice converts a map[int]V result to a positionally-matched []*V slice,
// one entry per id.
func mapToSlice[V any](ids []int, m map[int]V, batchErr error) ([]*V, []error) {
	out := make([]*V, len(ids))
	errs := make([]error, len(ids))
	if batchErr != nil {
		for i := range errs {
			errs[i] = batchErr
		}
		return out, errs
	}
	for i, id := range ids {
		if v, ok := m[id]; ok {
			cp := v
			out[i] = &cp
		} else {
			errs[i] = fmt.Errorf("id %d: not found", id)
		}
	}
	return out, errs
}
