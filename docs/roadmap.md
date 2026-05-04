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

- UsageLogger implemented as append-only local JSONL.
- LiteLLM non-stream usage fields are captured when returned.
- API writes usage records for success and downstream failure paths.
- Usage write failure is logged and does not fail task responses.
- Docker Compose config renders successfully.
- Full `docker compose up` runtime verification passes with real Xiaomi MiMo provider.

Exit:

- task-level usage stored
- usage record includes task id, request id, model alias, returned model, route reason, latency, token usage when available, and error state
- usage write failure does not fail the task
- docker-compose starts Go service + LiteLLM
- real provider end-to-end smoke test passes after implementation

## M6 Local Lab Query Loop

Status: next

Exit:

- usage records can be queried from the Go service
- task and usage records can be correlated by task id
- local smoke script demonstrates review task -> task lookup -> usage lookup
- Docker Compose runtime E2E passes
- real provider E2E passes after implementation
