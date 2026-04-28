# CI Status

## Last Run

Date: 2026-04-28 17:43:37 CST
Commit: `dae5fc7`

## Commands

| Command | Status | Notes |
|---|---|---|
| `go version` | pass | `go version go1.26.2 darwin/arm64` |
| `go test ./...` | pass | API, config, LiteLLM client, and task store tests passed |
| `docker compose -f deploy/docker-compose.yml config` | pass | compose file rendered successfully |
| `LITELLM_MASTER_KEY=sk-local-dev docker compose -f deploy/docker-compose.yml up litellm` | pass | LiteLLM Proxy started and initialized `code-cheap`, `code-smart` |
| `NANOBOT_ADDR=:18080 LITELLM_BASE_URL=http://127.0.0.1:4000 LITELLM_API_KEY=sk-local-dev LITELLM_MODEL=code-cheap LITELLM_TIMEOUT=10s go run ./cmd/server` | pass | server logged `server listening addr=:18080` |
| `curl -X POST http://127.0.0.1:18080/tasks/review-diff` | pass | returned 200 with `task_id`, `request_id`, `status=success`, `model=code-cheap` |
| `curl http://127.0.0.1:18080/tasks/task_71b3ea8ea55a7fca` | pass | returned stored task response from memory |

## Known Failures

| Failure | Owner | Next action |
|---|---|---|
| none | | |
