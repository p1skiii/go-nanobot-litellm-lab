# Spec 0005: Usage Projection

## Purpose

Preserve task-level usage visibility while making Invocation Ledger the source of truth.

M5 originally introduced a standalone UsageLogger. After ADR-0003, Usage is no longer an independent architecture boundary.
`usage.Record` remains as a compatibility projection/subset of `invocation.Record`.

## Inputs

| Input | Source |
|---|---|
| task id | Invocation Ledger |
| request id | Invocation Ledger |
| selected model alias | Invocation Ledger |
| returned model | Invocation Ledger |
| route reason | Invocation Ledger |
| latency | Invocation Ledger |
| usage fields | nested `usage` field in Invocation Ledger |
| status and error | Invocation Ledger |

## Outputs

Compatibility usage record:

```json
{
  "task_id": "task_x",
  "request_id": "req_x",
  "model_alias": "code-cheap",
  "returned_model": "provider/model",
  "route_reason": "string",
  "latency_ms": 1234,
  "prompt_tokens": 100,
  "completion_tokens": 50,
  "total_tokens": 150,
  "status": "success|failed",
  "error": "string optional",
  "created_at": "RFC3339 timestamp"
}
```

Storage source of truth:

- `data/invocations.jsonl`
- `NANOBOT_INVOCATION_LOG_PATH`

Legacy `data/usage.jsonl` is lab state and is not migrated.

## Data Structures

- `internal/usage.Record` stays as the compatibility shape.
- `internal/invocation.Record.Usage` may contain a `usage.Record`.
- If a ledger record has task-level metadata but no nested usage object, the usage projection fills compatible fields from the invocation record.

## Error Cases

| Case | Expected |
|---|---|
| provider omits usage | projection returns metadata with token fields empty |
| invocation write failure | task response still follows API behavior and logs the write failure |
| rejected request without task id | omitted from legacy usage projection |

## Config

Do not add new UsageLogger config.
Use Spec 0008 config: `NANOBOT_INVOCATION_LOG_PATH`.

## Test Matrix

| Test | Expected |
|---|---|
| usage projection from nested usage | pass |
| usage projection from invocation metadata | pass |
| rejected record without task id is skipped | pass |
| legacy `/usage/*` reads Invocation Ledger | pass |

## Non-goals

- No separate usage JSONL source of truth.
- No billing product.
- No dashboard.
- No database.
