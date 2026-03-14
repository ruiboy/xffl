# ADR-010: Denormalized Read Models

**Status:** Accepted
**Date:** 2026-03-14

## Context

Ladder standings, match scores, and premiership points could be calculated on read or stored pre-calculated.

## Decision

Store denormalized aggregates in the database. Maintain consistency through domain logic on writes.

## Rationale

- **Why for Hobby:** Reads (ladder, results) far more frequent than writes (stat updates), avoids complex aggregation queries on every page load
- **Scale Path:** Add CQRS with separate read/write models, event-sourced projections, materialized views
