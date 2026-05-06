# CI Status

## Last Run

Date: 2026-05-06 11:33:54 CST
Commit: working tree on base `8611cbd`

## Commands

| Command | Status | Notes |
|---|---|---|
| `bash -n scripts/smoke-real-provider.sh scripts/smoke-failure-cases.sh` | pass | scripts parse successfully; shell emitted a locale warning only |
| `go test ./...` | pass | all packages passed, including `internal/invocation` |
| `docker compose -f deploy/docker-compose.yml config` | pass | compose renders `NANOBOT_INVOCATION_LOG_PATH=/data/invocations.jsonl` |
| `docker compose -f deploy/docker-compose.yml build nanobot` | pass | multi-stage Dockerfile builds nanobot from source; no `deploy/bin/nanobot` dependency |
| `docker compose -f deploy/docker-compose.yml up -d --no-deps nanobot` | pass | rebuilt nanobot container started against existing LiteLLM container |
| `scripts/smoke-real-provider.sh` | pass | Go -> LiteLLM -> Xiaomi MiMo returned `200`; `/invocations/tasks`, `/invocations/runs`, `/invocations/recent`, and legacy `/usage/tasks` all returned expected records |
| `scripts/smoke-failure-cases.sh` | pass | empty diff and streaming wrote rejected records; timeout wrote `error_kind=timeout`; missing model wrote `error_kind=downstream`; recovery succeeded with token usage |

## Known Failures

| Failure | Owner | Next action |
|---|---|---|
| none | | |
