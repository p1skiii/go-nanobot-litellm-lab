# Roadmap

## M0 Harness Ready

Status: done

Exit:

- repo skeleton exists
- `CLAUDE.md` exists
- vision/gap/roadmap exists
- ADR-0001 accepted
- ADR-0002 accepted
- empty Go server can start

## M1 Go Calls LiteLLM

Status: done

Exit:

- Go service exposes `/tasks/review-diff`
- Go service calls LiteLLM non-stream
- `task_id`, `request_id`, latency, and model are logged
- minimal unit tests exist

## M2 LiteLLM Behavior Study

Status: done

Exit:

- fallback tested with real provider-backed LiteLLM
- usage tested with real provider-backed LiteLLM
- streaming tested with real provider-backed LiteLLM
- rate limit/budget behavior tested and documented
- notes written in `docs/research/litellm-behavior-notes.md`

## M3 Custom ContextManager

Status: done

Exit:

- context blocks exist
- keep/compress/drop rules exist
- review-diff task uses ContextManager before LiteLLM
- `context_report` returned on review task submission
- real provider smoke test passed

## M4 Custom PolicyRouter

Status: done

Exit:

- model profiles exist in `configs/models.yaml`
- score-based router exists
- `route_reason` logged and returned
- review-diff task uses routed LiteLLM model alias
- real provider smoke test passed for `code-smart` and `code-cheap`

## M5 UsageLogger + Docker Compose

Status: done

Current:

- Usage visibility is preserved as a projection over Invocation Ledger.
- LiteLLM non-stream usage fields are captured when returned and nested under invocation records.
- API writes invocation records for success and downstream failure paths.
- Invocation write failure is logged and does not fail task responses.
- Docker Compose config renders successfully.
- Full `docker compose up` runtime verification passes with real Xiaomi MiMo provider.

Exit:

- task-level usage is queryable through Invocation Ledger projection
- usage record includes task id, request id, model alias, returned model, route reason, latency, token usage when available, and error state
- invocation write failure does not fail the task
- docker-compose starts Go service + LiteLLM
- real provider end-to-end smoke test passes after implementation

## M6 Local Lab Query Loop

Status: done

Current:

- `GET /invocations/recent` returns recent invocation records from JSONL.
- `GET /invocations/tasks/{id}` returns invocation records for one task id.
- `GET /invocations/runs/{run_id}` returns invocation records for one run id.
- Legacy `/usage/*` returns compatibility projections over Invocation Ledger.
- `scripts/smoke-real-provider.sh` verifies review task, task lookup, invocation lookup, run lookup, and legacy usage projection.
- Docker Compose runtime E2E passes with real Xiaomi MiMo provider.

Exit:

- invocation records can be queried from the Go service
- task and invocation records can be correlated by task id and run id
- local smoke script demonstrates review task -> task lookup -> invocation lookup
- Docker Compose runtime E2E passes
- real provider E2E passes after implementation

## M7 Failure Replay Lab

Status: done

Current:

- `scripts/smoke-failure-cases.sh` replays expected failure cases.
- Empty diff maps to `400` and writes a rejected invocation record.
- Streaming request maps to `400` and writes a rejected invocation record.
- Tiny LiteLLM timeout maps to `504` and writes failed invocation.
- Missing LiteLLM model maps to `502` and writes failed invocation.
- Normal real-provider request succeeds after failure cases.

Exit:

- empty diff failure is replayable and maps to `400`
- streaming request failure is replayable and maps to `400`
- tiny LiteLLM timeout is replayable and maps to `504`
- missing LiteLLM model error is replayable and maps to `502`
- failed invocation records are written for timeout and downstream error cases
- normal real-provider request succeeds after failure cases

## M8 Invocation Ledger Consolidation

Status: in progress

Current:

- ADR-0003 defines Invocation Ledger as the observability boundary.
- `internal/invocation` owns append-only JSONL ledger reads and writes.
- `usage.Record` remains as a compatibility projection/subset.
- `/invocations/recent`, `/invocations/tasks/{id}`, and `/invocations/runs/{run_id}` are implemented.
- `/usage/*` reads Invocation Ledger and projects `usage.Record`.
- `POST /tasks/review-diff` returns `run_id` and `attempt_id`.
- Validation failures write rejected invocation records without creating tasks.
- Compose build no longer depends on gitignored `deploy/bin/nanobot`.

Exit:

- `go test ./...` passes.
- Docker Compose config renders.
- Docker Compose can build from clean clone path.
- real provider smoke writes and queries invocation records.
- failure replay smoke validates rejected, failed, and recovery invocation records.

## M9 PolicyRouter Evaluation

Status: planned

Exit:

- Same diff is run with `budget_hint=low` and `budget_hint=high_quality`.
- Both attempts share one `run_id`.
- Invocation Ledger comparison covers model, route reason, latency, tokens, status, and result length.
- Findings are written to research notes without changing Router weights.

## M10 ContextManager Evaluation

Status: planned

Exit:

- Same diff/model strategy is run as diff-only, repo-summary, and full-context scenarios.
- All attempts share one `run_id`.
- Invocation Ledger comparison covers `context_report`, `context_chars`, tokens, latency, status, and result length.
- ContextManager rule changes are only planned if evaluation evidence supports them.
