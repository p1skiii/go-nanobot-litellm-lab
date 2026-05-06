# CLAUDE.md

## Project

This repo builds a Go nanobot backend that calls LiteLLM Proxy as a downstream LLM gateway.

The purpose is to learn LLM gateway usage first, then gradually implement selected infra modules:

- ContextManager
- PolicyRouter
- UsageLogger

Execution facts belong in the Invocation Ledger; UsageLogger is a compatibility projection, not a separate source of truth.

## Current Architecture

```text
CLI / HTTP client
  -> Go Nanobot Backend
  -> ContextManager
  -> PolicyRouter
  -> LiteLLM Proxy
  -> LLM provider or mock provider

Invocation Ledger records execution facts around this path.
```

## Current Milestone

Read `docs/current-state.md` before making changes.

## Non-goals

- Do not build a full custom LLM Gateway before M3.
- Do not split into microservices.
- Do not start with Kubernetes.
- Do not implement ML-based routing.
- Do not clone Claude Code UI.

## Build Commands

```bash
go test ./...
go run ./cmd/server
docker compose -f deploy/docker-compose.yml up
```

## Key Extension Points

### Add a new task

1. Add task handler in `internal/tasks/`.
2. Register route in `internal/api/`.
3. Add test plan under `docs/test-plans/`.

### Add a new context rule

1. Update `internal/contextmgr/`.
2. Update `docs/specs/0003-context-manager.md`.
3. Add tests for keep/compress/drop behavior.

### Add a new routing rule

1. Update `internal/router/`.
2. Update `configs/policies.yaml`.
3. Add `route_reason` assertions in tests.

### Add a new execution fact

1. Update `internal/invocation/`.
2. Update `docs/specs/0008-unified-invocation-ledger.md`.
3. Read from `/invocations/*`; use `/usage/*` only for compatibility projection.

## Required After Code Changes

- Run `go test ./...`.
- Run a real provider end-to-end smoke test after completing each milestone target.
- Update `docs/current-state.md`.
- Update `docs/ci-status.md` if tests change.
