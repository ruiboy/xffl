---
status: accepted
date: 2026-03-14
scope: repo
enforceable: true
rules:
  - "go.work exists at repo root"
  - "go.work references all modules under services/, shared/, contracts/"
---

# ADR-007: Go Workspace

## Context

Monorepo has multiple Go modules (services, shared packages, contracts) that need to work together during development.

## Decision

Use `go.work` at repo root referencing all modules.

## Rationale

- **Why for Hobby:** Cross-module imports work immediately, changes to shared packages available without publishing, single workspace keeps everything in sync
- **Scale Path:** Extract shared packages to separate repository/module, or inline into services if needed
