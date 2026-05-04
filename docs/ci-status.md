# CI Status

## Last Run

Date: 2026-05-05 00:15:52 CST
Commit: working tree on base `729ec36`

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
| `docker run --rm --name litellm-m2-study ... --config /app/config.yaml` | pass | LiteLLM `1.82.6` started on `127.0.0.1:4100` with M2 temp config |
| `curl http://127.0.0.1:4100/health/readiness` | pass | returned healthy, DB not connected, version `1.82.6` |
| M2 fallback curl | pass | `fallback-primary` returned 200 via `fallback-ok`; header `x-litellm-attempted-fallbacks: 1` |
| M2 usage curl | pass | non-stream mock returned `usage` tokens plus cost/duration headers |
| M2 streaming curl | pass | returned `text/event-stream` chunks ending with `data: [DONE]` |
| M2 rate limit curl loop | partial | `rpm: 1` emitted rate-limit headers but did not enforce 429 in local mock/no-DB setup |
| M2 budget key generation curl | expected fail | returned 500 because LiteLLM DB was not connected |
| M2 error mapping curl | pass | broken downstream returned LiteLLM 500; missing model returned LiteLLM 400 |
| Go service missing-model check | pass | Go service mapped LiteLLM 400 to task failure with HTTP 502 |
| `docker stop litellm-m2-study` | pass | temporary proxy stopped |
| Xiaomi MiMo direct OpenAI-compatible curl | pass | real provider returned 200, content `provider-ok`, and real token usage |
| `docker run --rm --name litellm-m2-real ... --config /app/config.yaml` | pass | LiteLLM `1.82.6` started on `127.0.0.1:4101` with real MiMo-backed aliases |
| `curl http://127.0.0.1:4101/health/readiness` | pass | returned healthy, DB not connected, version `1.82.6` |
| M2 real fallback curl | pass | broken primary returned 200 through `mimo-fallback-ok`; header `x-litellm-attempted-fallbacks: 1` |
| M2 real usage curl | pass | non-stream real provider response returned `usage` with token detail fields |
| M2 real streaming curl | pass | returned real `text/event-stream` chunks ending with `data: [DONE]`; no stream usage observed |
| M2 real rate limit curl loop | partial | `rpm: 1` emitted rate-limit headers but did not enforce 429 in local no-DB setup |
| M2 real budget key generation curl | expected fail | returned 500 because LiteLLM DB was not connected |
| M2 real error mapping curl | pass | invalid upstream API key returned LiteLLM 401; missing model returned LiteLLM 400 |
| Go service real provider check | pass | `POST /tasks/review-diff` returned 200 through Go -> LiteLLM -> real MiMo |
| Go service real provider error check | pass | Go service mapped LiteLLM 401 to task failure with HTTP 502 |
| `docker stop litellm-m2-real` | pass | temporary real-provider proxy stopped |
| `gofmt -w internal/contextmgr/manager.go internal/contextmgr/manager_test.go internal/api/handler.go internal/api/handler_test.go internal/litellm/client.go internal/litellm/client_test.go` | pass | M3 ContextManager and API/LiteLLM integration formatted |
| `go test ./...` | pass | M3 tests passed, including `internal/contextmgr` and updated API/LiteLLM tests |
| `uvx --from 'litellm[proxy]' litellm --port 4101 --config /tmp/litellm-m3-real/config.yaml` | pass | started temporary LiteLLM proxy with real `mimo-real-flash` provider mapping |
| `NANOBOT_ADDR=:18084 LITELLM_BASE_URL=http://127.0.0.1:4101 LITELLM_API_KEY=sk-local-dev LITELLM_MODEL=mimo-real-flash LITELLM_TIMEOUT=40s go run ./cmd/server` | pass | Go server started for M3 smoke |
| `curl -X POST http://127.0.0.1:18084/tasks/review-diff` with `prior_plan/logs/notes` | pass | returned `200`, real model output, and `context_report` with keep/compress/drop decisions |
| stop temporary LiteLLM + Go processes | pass | test processes terminated cleanly |
| `go get gopkg.in/yaml.v3` | pass | added YAML parser for policy/model config loading |
| `gofmt -w cmd/server/main.go internal/api/handler.go internal/api/handler_test.go internal/config/config.go internal/config/config_test.go internal/litellm/client.go internal/litellm/client_test.go internal/router/router.go internal/router/router_test.go internal/tasks/store.go internal/tasks/store_test.go` | pass | formatted M4 router and integration changes |
| `go test ./...` | pass | M4 router tests and updated API/LiteLLM/config/task tests passed |
| M4 real provider smoke: `budget_hint=high_quality` | pass | Go -> LiteLLM -> Xiaomi MiMo returned `200`, `model=code-smart`, and `route_reason` selected `code-smart` |
| M4 real provider smoke: `budget_hint=low` | pass | Go -> LiteLLM -> Xiaomi MiMo returned `200`, `model=code-cheap`, and `route_reason` selected `code-cheap` |
| M1-M4 documentation sync | pass | proposal, milestone summary, specs, test plans, roadmap, README, and harness rule updated |
| `go test ./...` after documentation sync | pass | all packages passed |
| documentation sync real provider smoke | pass | Go -> LiteLLM -> Xiaomi MiMo returned `200`, `model=code-smart`, `route_reason`, and `context_report` |
| `gofmt -w ...` for M5 UsageLogger | pass | formatted usage logger, API, config, LiteLLM, and server changes |
| `go test ./...` for M5 | pass | all packages passed, including `internal/usage` |
| M5 real provider E2E | pass | Go -> LiteLLM -> Xiaomi MiMo returned `200`, `model=code-smart`, `context_report`, and `route_reason` |
| M5 usage JSONL verification | pass | `/tmp/nanobot-m5/usage.jsonl` recorded task id, request id, model alias, returned model, route reason, latency, token usage, and success status |
| `docker compose -f deploy/docker-compose.yml config` | pass | compose config renders Go build service, LiteLLM service, and usage volume |
| `docker info` | blocked | Docker daemon unavailable, so full `docker compose up` runtime verification was not possible in this run |
| `go test ./...` after M5 docs status sync | pass | all packages passed |
| M5 real provider E2E rerun after docs status sync | pass | Go -> LiteLLM -> Xiaomi MiMo returned `200`, `model=code-smart`, `route_reason`, `context_report`, and usage JSONL token fields |
| `docker info` after opening Docker Desktop | pass | Docker Desktop daemon available, server version `28.0.4` |
| initial M5 Docker Compose runtime E2E | expected fail | real provider request reached Compose stack but Go timed out at 30s; failed usage record was written |
| local Linux binary build for `scratch` nanobot image | pass | built `deploy/bin/nanobot` as linux/arm64 static binary for Compose image |
| `docker compose -f deploy/docker-compose.yml up --build -d` | pass | Compose started `litellm` and `nanobot` containers |
| M5 Docker Compose runtime E2E | pass | Go container -> LiteLLM container -> Xiaomi MiMo returned `200`, `model=code-cheap`, `context_report`, and `route_reason` |
| M5 Docker Compose usage JSONL verification | pass | `data/usage.jsonl` recorded model alias, returned model, route reason, latency, prompt/completion/total tokens, and success status |
| `go test ./...` after M5 Compose verification | pass | all packages passed |
| `gofmt -w internal/api/handler.go internal/api/handler_test.go internal/usage/logger.go internal/usage/logger_test.go` | pass | formatted M6 API and usage reader changes |
| `go test ./...` for M6 | pass | all packages passed, including usage reader and usage API tests |
| Compose rebuild with M6 binary | pass | local linux/arm64 nanobot binary copied into scratch image and container restarted |
| `scripts/smoke-real-provider.sh` | pass | Docker Compose Go -> LiteLLM -> Xiaomi MiMo returned `200`; `/tasks/{id}`, `/usage/tasks/{id}`, and `/usage/recent` returned expected records |

## Known Failures

| Failure | Owner | Next action |
|---|---|---|
| none | | |
