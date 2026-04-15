package graphql

import "xffl/services/search/internal/domain"

// Resolver is the dependency injection container for GraphQL resolvers.
type Resolver struct {
	Repo domain.DocumentRepository
}
