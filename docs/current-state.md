# Current State

## Current Milestone

M3 Custom ContextManager

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

## In Progress

- none

## Blocked

- none

## Next Step

Run a real-provider M3 smoke test through LiteLLM and decide whether to start M4 PolicyRouter or first add lightweight LiteLLM metadata capture.
