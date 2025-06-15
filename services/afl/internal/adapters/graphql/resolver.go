package graphql

import (
	"xffl/services/afl/internal/ports/in"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	clubUseCase in.ClubUseCase
}

// NewResolver creates a new GraphQL resolver with injected dependencies
func NewResolver(clubUseCase in.ClubUseCase) *Resolver {
	return &Resolver{
		clubUseCase: clubUseCase,
	}
}
