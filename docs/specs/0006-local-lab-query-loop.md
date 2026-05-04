# Spec 0006: Local Lab Query Loop

## Purpose

Expose a minimal read path for local task and usage inspection after M5 writes append-only usage records.

## Inputs

| Input | Source |
|---|---|
| usage log path | `NANOBOT_USAGE_LOG_PATH` |
| limit | `GET /usage/recent?limit=N` |
| task id | `GET /usage/tasks/{id}` |

## Outputs

### `GET /usage/recent`

```json
{
  "count": 1,
  "records": []
}
```

### `GET /usage/tasks/{id}`

```json
{
  "task_id": "task_abc",
  "count": 1,
  "records": []
}
```

Records use the UsageLogger record shape from Spec 0005.

## Data Structures

- Keep JSONL as the source of truth.
- Read records in append order.
- Return recent records in chronological order.

## Error Cases

| Case | Expected |
|---|---|
| usage file missing | return empty list |
| invalid `limit` | 400 |
| malformed JSONL line | 500 |
| empty task id | 404 |

## Config

No new config. M6 reuses `NANOBOT_USAGE_LOG_PATH`.

## Test Matrix

| Test | Expected |
|---|---|
| read recent usage | returns newest records |
| read usage by task id | returns matching records |
| missing usage file | returns empty list |
| invalid limit | 400 |
| smoke script can submit task and inspect usage | pass |

## Non-goals

- No database.
- No dashboard.
- No analytics aggregation.
- No auth.
- No cross-process task persistence.
