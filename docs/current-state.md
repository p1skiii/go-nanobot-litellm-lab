# Current State

## Current Milestone

M8 Invocation Ledger Consolidation

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
- M5 committed as `a5f0422`
- M6 proposal/spec/test plan accepted
- M6 UsageReader implemented for append-only JSONL
- M6 `GET /usage/recent` implemented
- M6 `GET /usage/tasks/{id}` implemented
- M6 smoke script added
- M6 `go test ./...` passed
- M6 Docker Compose real provider smoke passed
- M6 committed as `1bb5658`
- M6 pushed to `origin/main`
- M7 proposal/spec/test plan accepted
- M7 failure replay script added
- M7 `go test ./...` passed
- M7 failure replay smoke passed against Docker Compose LiteLLM and real Xiaomi MiMo recovery path
- M7 behavior notes updated with timeout, missing model, validation, and recovery observations
- M8 proposal accepted in `docs/proposals/0005-unify-m5-m8-invocation-ledger.md`
- ADR-0003 added for Invocation Ledger as the observability boundary
- Spec 0008 and test plan added for Unified Invocation Ledger
- Specs 0005-0007 updated so Usage, query APIs, and failure replay are projections over Invocation Ledger
- M9 PolicyRouter Evaluation and M10 ContextManager Evaluation specs added
- `internal/invocation` added with append-only JSONL ledger and shared read/write mutex
- API now writes invocation records for success, failed, and rejected review requests
- API now exposes `/invocations/recent`, `/invocations/tasks/{id}`, and `/invocations/runs/{run_id}`
- Legacy `/usage/*` now reads Invocation Ledger and returns usage projections
- `POST /tasks/review-diff` now returns `run_id` and `attempt_id`
- Compose Dockerfile updated to build from source instead of gitignored `deploy/bin/nanobot`
- M8 `go test ./...` passed
- M8 Docker Compose config passed
- M8 Docker Compose nanobot image built from source with multi-stage Dockerfile
- M8 real provider smoke passed through Go -> LiteLLM -> Xiaomi MiMo and wrote invocation records
- M8 failure replay smoke passed with rejected, failed, and recovery invocation records

## In Progress

- none

## Blocked

- none

## Next Step

Start M9 PolicyRouter Evaluation using shared-run Invocation Ledger comparisons.
