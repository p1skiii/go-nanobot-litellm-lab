# ADR-0002: Modular monolith before microservices

## Status

Accepted

## Context

The project is still in learning and exploration phase.
Splitting ContextManager, Router, UsageLogger, and API into separate services would increase coordination cost.

## Decision

Use one Go service with clear internal packages:

- api
- litellm
- tasks
- contextmgr
- router
- usage
- telemetry
- config

## Consequences

- Easier local development
- Easier refactoring
- Future service split remains possible
