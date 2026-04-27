# go-nanobot-litellm-lab

Go nanobot service for learning how an upstream agent backend consumes LiteLLM as a downstream LLM gateway.

The project starts with LiteLLM instead of a custom gateway. After the local loop works, selected responsibilities move into custom modules:

- ContextManager
- PolicyRouter
- UsageLogger

## Current Milestone

M0 Harness Ready.

## Local Commands

```bash
go test ./...
go run ./cmd/server
curl http://localhost:8080/health
docker compose -f deploy/docker-compose.yml config
```

## Project Harness

- `CLAUDE.md` gives agents the stable project operating rules.
- `AGENTS.md` defines PM, Architect, Engineer, and QA ownership.
- `docs/adr/` records accepted architecture decisions.
- `docs/specs/` describes implementation contracts.
- `docs/test-plans/` describes how specs are verified.
- `docs/current-state.md` and `docs/ci-status.md` keep short-lived project state out of chat context.
