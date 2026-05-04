# Current State

## Current Milestone

M6 Local Lab Query Loop

## Completed

- Remote repository created at `https://github.com/p1skiii/go-nanobot-litellm-lab`
- Go toolchain installed with Homebrew
- Repo skeleton created
- Harness docs, ADRs, specs, and test plans created
- Minimal `/health` server implemented
- `go test ./...` passed
- `/health` verified with local server on `:18080`
- `docker compose -f deploy/docker-compose.yml config` passed
- M0 committed as `865ac58`
- `POST /tasks/review-diff` implemented
- task id and request id returned
- in-memory task store implemented
- `GET /tasks/{id}` reads stored task result
- Go service calls LiteLLM Proxy non-stream
- timeout maps to 504 and downstream errors map to 502
- Docker LiteLLM Proxy mock response verified
- M1 committed as `dae5fc7`
- `main` pushed to `origin/main`
- Previous M2 mock-only observations reclassified as synthetic/local baseline
- M2 real provider-backed Xiaomi MiMo direct API sanity check completed
- M2 real provider-backed LiteLLM fallback experiment completed
- M2 real provider-backed LiteLLM usage fields experiment completed
- M2 real provider-backed LiteLLM streaming experiment completed
- M2 real provider-backed LiteLLM rate limit / budget experiment completed
- M2 real provider-backed LiteLLM downstream error mapping experiment completed
- Existing Go service verified through LiteLLM to real Xiaomi MiMo provider
- M3 ContextManager implemented with deterministic keep/compress/drop rules
- M3 `review-diff` now builds governed `final_context` before LiteLLM call
- M3 API response now includes `context_report` block decisions
- M3 tests added for ContextManager and API/LiteLLM context wiring
- M3 `go test ./...` passed
- M3 committed as `fe19623`
- M4 PolicyRouter implemented with deterministic score-based model selection
- M4 router reads model profiles from `configs/models.yaml`
- M4 router reads weights/default fallback from `configs/policies.yaml`
- M4 `review-diff` now routes model alias before LiteLLM call
- M4 response now includes `route_reason`
- M4 tests added for router scoring, stream filtering, and context-limit rejection
- M4 `go test ./...` passed
- M4 real provider smoke passed for `budget_hint=high_quality` -> `code-smart`
- M4 real provider smoke passed for `budget_hint=low` -> `code-cheap`
- M4 committed as `4931ac4`
- M1-M4 documentation sync completed in proposal `docs/proposals/0001-sync-m1-m4-docs-and-prepare-m5.md`
- M1-M4 documentation sync committed as `729ec36`
- M5 UsageLogger implemented as append-only JSONL
- M5 LiteLLM client extracts non-stream usage token fields
- M5 API writes usage records for success and downstream failure paths
- M5 usage write failure does not fail task responses
- M5 default usage log path is `data/usage.jsonl`
- M5 Dockerfile and compose build config added
- M5 `go test ./...` passed
- M5 real provider E2E passed and wrote a usage record with model, route, latency, and tokens
- M5 real provider E2E rerun passed after documentation status sync
- Docker Desktop started and Docker daemon verified
- M5 Docker Compose runtime E2E passed with real Xiaomi MiMo provider
- M5 Compose runtime request wrote usage JSONL with model, route, latency, and token usage

## In Progress

- M6 planning

## Blocked

- none

## Next Step

Commit M5 UsageLogger, then create M6 proposal/spec/test plan for local usage query loop.
