---
status: accepted
date: 2026-03-14
scope: infra
enforceable: true
rules:
  - "no cross-schema joins"
  - "each service accesses only its own schema"
  - "no cross-service database imports"
---

# ADR-003: Shared Database with Schema Isolation

## Context

SOA principles say each service owns its data. But this is a hobby project where operational simplicity matters.

## Decision

Single PostgreSQL database (`xffl`) with schema isolation (`afl.*`, `ffl.*`). Each service only accesses its own schema. No cross-schema joins — services communicate via events.

## Rationale

- **Why for Hobby:** Simple setup, easy backup/restore, single connection string, no cross-schema queries by discipline
- **Scale Path:** Split to separate databases when independently scaling services becomes necessary — schema isolation already enforces the discipline
