# Spec 0008: Unified Invocation Ledger

## Purpose

Unify task execution facts into one append-only ledger that Router, ContextManager, Usage, Replay, Debug, and Evaluation can all read.

## Inputs

| Input | Source |
|---|---|
| run id | `X-Run-ID` header or generated id |
| attempt id | `X-Attempt-ID` header or generated id |
| scenario | `X-Scenario` header or default API scenario |
| task id | task layer when a task is created |
| request id | `X-Request-ID` header or generated id |
| HTTP status | API response |
| task status | task execution result |
| error kind | validation, timeout, downstream, bad response, routing |
| latency | API/LiteLLM timing |
| context report | ContextManager |
| model alias and route reason | PolicyRouter |
| returned model and usage | LiteLLM response |

## Outputs

Source of truth:

- `data/invocations.jsonl`
- configurable with `NANOBOT_INVOCATION_LOG_PATH`

Record shape:

```json
{
  "run_id": "run_x",
  "attempt_id": "attempt_x",
  "scenario": "smoke.real_provider",
  "operation": "review_diff",
  "task_id": "task_x",
  "request_id": "req_x",
  "http_status": 200,
  "task_status": "success|failed|rejected",
  "error_kind": "none|validation|timeout|downstream|bad_response|routing",
  "latency_ms": 1234,
  "context_chars": 123,
  "context_report": {},
  "model_alias": "code-cheap",
  "returned_model": "code-cheap",
  "route_reason": "string",
  "usage": {},
  "started_at": "RFC3339",
  "finished_at": "RFC3339"
}
```

The `usage` field is the existing `usage.Record` projection and compatibility subset.

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| GET | `/invocations/recent?limit=N` | read recent invocation records |
| GET | `/invocations/tasks/{id}` | read records for a task |
| GET | `/invocations/runs/{run_id}` | read records for a run |
| GET | `/usage/recent?limit=N` | compatibility usage projection |
| GET | `/usage/tasks/{id}` | compatibility usage projection |

`POST /tasks/review-diff` returns `run_id` and `attempt_id`.

Validation failures write rejected invocation records but do not create tasks.

## Data Structures

`invocation.Record` wraps execution metadata and optional `usage.Record`.
`usage.Record` remains in `internal/usage` and should not become a separate source of truth.

JSONL read and write operations must share the same mutex so a read cannot observe a partial append.

## Error Cases

| Case | Expected |
|---|---|
| invalid JSON | `400`, rejected invocation record |
| empty diff | `400`, rejected invocation record |
| streaming requested | `400`, rejected invocation record |
| no routable model | `400`, rejected invocation record with routing error |
| LiteLLM timeout | `504`, failed invocation record |
| LiteLLM downstream error | `502`, failed invocation record |
| malformed JSONL | `500` on read |

## Config

| Name | Default | Purpose |
|---|---|---|
| `NANOBOT_INVOCATION_LOG_PATH` | `data/invocations.jsonl` | append-only invocation ledger path |

## Test Matrix

| Test | Expected |
|---|---|
| append/read invocation records | pass |
| concurrent append and recent read | no malformed JSONL |
| usage projection from invocation | pass |
| validation failure writes rejected record | pass |
| timeout writes failed record with `error_kind=timeout` | pass |
| downstream error writes failed record with `error_kind=downstream` | pass |
| `/invocations/recent` | pass |
| `/invocations/tasks/{id}` | pass |
| `/invocations/runs/{run_id}` | pass |
| legacy `/usage/*` projection | pass |
| real provider E2E writes invocation record | pass |

## Non-goals

- No database.
- No streaming.
- No fallback implementation.
- No advanced Router changes.
- No ContextManager rule changes.
- No K8s, OTel, auth, or dashboard.
