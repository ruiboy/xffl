# ADR-007: Go Workspace

**Status:** Accepted
**Date:** 2026-03-14

## Context

Monorepo has multiple Go modules (services, shared packages, contracts) that need to work together during development.

## Decision

Use `go.work` at repo root referencing all modules.

## Rationale

- **Why for Hobby:** Cross-module imports work immediately, changes to shared packages available without publishing, single workspace keeps everything in sync
- **Scale Path:** Extract shared packages to separate repository/module, or inline into services if needed
