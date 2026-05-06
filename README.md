# go-nanobot-litellm-lab

Go nanobot service for learning how an upstream agent backend consumes LiteLLM as a downstream LLM gateway.

The project starts with LiteLLM instead of a custom gateway. After the local loop works, selected responsibilities move into custom modules:

- ContextManager
- PolicyRouter
- UsageLogger
- Invocation Ledger

## Current Milestone

M8 Invocation Ledger consolidation is in progress. The next planned milestones are M9 PolicyRouter Evaluation and M10 ContextManager Evaluation.

## Local Commands

```bash
go test ./...
go run ./cmd/server
curl http://localhost:8080/health
docker compose -f deploy/docker-compose.yml config
```

## Current Runtime Chain

```text
HTTP client
  -> Go Nanobot Backend
  -> ContextManager
  -> PolicyRouter
  -> LiteLLM Proxy
  -> LLM provider or mock provider

Invocation Ledger records execution facts around this path.
```

## Implemented Milestones

| Milestone | Status | Summary |
|---|---|---|
| M0 | done | filesystem harness and minimal server |
| M1 | done | `POST /tasks/review-diff`, task store, non-stream LiteLLM client |
| M2 | done | real provider-backed LiteLLM behavior study |
| M3 | done | deterministic ContextManager with `context_report` |
| M4 | done | score-based PolicyRouter with `route_reason` |
| M5 | done | usage projection over execution records |
| M6 | done | local lab query loop |
| M7 | done | failure replay lab |
| M8 | in progress | Invocation Ledger consolidation |

## Project Harness

- `CLAUDE.md` gives agents the stable project operating rules.
- `AGENTS.md` defines PM, Architect, Engineer, and QA ownership.
- `docs/adr/` records accepted architecture decisions.
- `docs/specs/` describes implementation contracts.
- `docs/test-plans/` describes how specs are verified.
- `docs/current-state.md` and `docs/ci-status.md` keep short-lived project state out of chat context.
