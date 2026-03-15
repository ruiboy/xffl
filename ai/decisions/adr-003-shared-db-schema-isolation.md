# ADR-003: Shared Database with Schema Isolation

**Status:** Accepted
**Date:** 2026-03-14

## Context

SOA principles say each service owns its data. But this is a hobby project where operational simplicity matters.

## Decision

Single PostgreSQL database (`xffl`) with schema isolation (`afl.*`, `ffl.*`). Each service only accesses its own schema. No cross-schema joins — services communicate via events.

## Rationale

- **Why for Hobby:** Simple setup, easy backup/restore, single connection string, no cross-schema queries by discipline
- **Scale Path:** Split to separate databases when independently scaling services becomes necessary — schema isolation already enforces the discipline
