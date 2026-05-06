# Spec 0006: Local Lab Query Loop

## Purpose

Expose a minimal read path for local task, invocation, and usage inspection.

After ADR-0003, the query loop reads Invocation Ledger records first.
Usage APIs remain as compatibility projections.

## Inputs

| Input | Source |
|---|---|
| invocation log path | `NANOBOT_INVOCATION_LOG_PATH` |
| limit | `GET /invocations/recent?limit=N` or `GET /usage/recent?limit=N` |
| task id | `GET /invocations/tasks/{id}` or `GET /usage/tasks/{id}` |
| run id | `GET /invocations/runs/{run_id}` |

## Outputs

### `GET /invocations/recent`

```json
{
  "count": 1,
  "records": []
}
```

### `GET /invocations/tasks/{id}`

```json
{
  "task_id": "task_abc",
  "count": 1,
  "records": []
}
```

### `GET /invocations/runs/{run_id}`

```json
{
  "run_id": "run_abc",
  "count": 1,
  "records": []
}
```

Legacy `/usage/*` endpoints return `usage.Record` projections from these invocation records.

## Data Structures

- Invocation Ledger is the source of truth.
- Records are returned in append order.
- Recent records are returned in chronological order.
- JSONL read and write share one mutex.

## Error Cases

| Case | Expected |
|---|---|
| invocation file missing | return empty list |
| invalid `limit` | 400 |
| malformed JSONL line | 500 |
| empty task id | 404 |
| empty run id | 404 |

## Config

No query-specific config.
Use `NANOBOT_INVOCATION_LOG_PATH`.

## Test Matrix

| Test | Expected |
|---|---|
| read recent invocations | returns newest records |
| read invocations by task id | returns matching records |
| read invocations by run id | returns matching records |
| read legacy usage projection | returns projected records |
| missing invocation file | returns empty list |
| invalid limit | 400 |
| smoke script can submit task and inspect invocation | pass |

## Non-goals

- No database.
- No dashboard.
- No analytics aggregation.
- No auth.
- No cross-process task persistence.
