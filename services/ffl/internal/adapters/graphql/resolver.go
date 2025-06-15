package graphql

import (
	"xffl/services/ffl/internal/ports/in"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver holds the use cases for GraphQL resolvers
type Resolver struct {
	clubUseCase       in.ClubUseCase
	playerUseCase     in.PlayerUseCase
	clubSeasonUseCase in.ClubSeasonUseCase
}

// NewResolver creates a new GraphQL resolver
func NewResolver(clubUseCase in.ClubUseCase, playerUseCase in.PlayerUseCase, clubSeasonUseCase in.ClubSeasonUseCase) *Resolver {
	return &Resolver{
		clubUseCase:       clubUseCase,
		playerUseCase:     playerUseCase,
		clubSeasonUseCase: clubSeasonUseCase,
	}
}
