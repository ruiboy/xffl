package graphql

import (
	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/application/dataops"
)

// Resolver is the dependency injection container for GraphQL resolvers.
type Resolver struct {
	Queries  *application.Queries
	Commands *application.Commands
	DataOps  *dataops.DataOpsCommands
	Captures *dataops.ForumCaptureBuffer
	Builder  *dataops.Builder
}
