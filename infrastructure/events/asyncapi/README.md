# XFFL AsyncAPI Event Documentation

This directory contains AsyncAPI specifications for the XFFL event architecture.

## Files

- `afl-events.yaml` - AFL service event specifications
- `ffl-events.yaml` - FFL service event specifications  
- `xffl-events.yaml` - Unified event architecture overview

## Event Architecture

```
AFL Service → PostgreSQL → FFL Service → PostgreSQL → Search Service
           (PlayerMatchUpdated)      (FantasyScoreCalculated)
```

### Event Flow
1. **AFL Service** publishes `AFL.PlayerMatchUpdated` when player statistics change
2. **FFL Service** subscribes to AFL events and calculates fantasy scores, then publishes `FFL.FantasyScoreCalculated`
3. **Search Service** subscribes to both AFL and FFL events for search indexing

### Channel Naming
PostgreSQL channels are derived from event types:
- `AFL.PlayerMatchUpdated` → `afl_player_match_updated`
- `FFL.FantasyScoreCalculated` → `ffl_fantasy_score_calculated`

## Usage

### Generate Documentation

```bash
# Install AsyncAPI CLI
npm install -g @asyncapi/cli

# Generate HTML documentation
asyncapi generate docs infrastructure/events/asyncapi/xffl-events.yaml --output docs/events/

# Start documentation server
asyncapi start studio infrastructure/events/asyncapi/xffl-events.yaml
```

### Validation

Event validation is built into the PostgreSQL event dispatcher:

```go
// Enable validation in your service
eventDispatcher, err := postgres.NewPostgresDispatcher(config.EventDBURL, logger)
if err != nil {
    log.Fatal(err)
}

// Enable AsyncAPI validation
if err := eventDispatcher.EnableValidation("infrastructure/events/asyncapi"); err != nil {
    logger.Printf("Warning: Could not enable event validation: %v", err)
}
```

### Code Generation

```bash
# Generate Go types from AsyncAPI schemas
asyncapi generate fromTemplate infrastructure/events/asyncapi/ @asyncapi/go-template --output pkg/events/generated/

# Generate TypeScript types for frontend
asyncapi generate fromTemplate infrastructure/events/asyncapi/ @asyncapi/typescript-template --output frontend/web/src/types/events/
```

## Schema Evolution

When adding new events or modifying existing ones:

1. Update the appropriate `.yaml` file
2. Run validation tests
3. Generate new documentation
4. Update consuming services

## Event Types

### AFL.PlayerMatchUpdated
- **Source**: AFL Service
- **Trigger**: Player match statistics updated via GraphQL
- **Consumers**: FFL Service, Search Service
- **Payload**: Old and new player statistics

### FFL.FantasyScoreCalculated  
- **Source**: FFL Service
- **Trigger**: AFL player statistics processed
- **Consumers**: Search Service
- **Payload**: AFL score, fantasy score, calculation source

## Development

### Adding New Events

1. Define the event in the appropriate service's AsyncAPI spec
2. Add validation schema to `pkg/events/validation/validator.go`
3. Update the unified `xffl-events.yaml`
4. Regenerate documentation

### Testing

```bash
# Validate AsyncAPI specs
asyncapi validate infrastructure/events/asyncapi/afl-events.yaml
asyncapi validate infrastructure/events/asyncapi/ffl-events.yaml
asyncapi validate infrastructure/events/asyncapi/xffl-events.yaml
```

## Tools and Integrations

- **AsyncAPI Studio**: Interactive design and documentation
- **AsyncAPI CLI**: Code generation and validation
- **JSON Schema**: Runtime event validation
- **PostgreSQL**: Event transport layer