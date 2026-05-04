# Spec 0005: UsageLogger

## Purpose

Record task-level model, latency, token, cost, and route metadata so the project can compare LiteLLM behavior with custom infra modules.

## Inputs

| Input | Source |
|---|---|
| task id | task layer |
| request id | API middleware |
| selected model alias | PolicyRouter |
| returned model | LiteLLM response |
| usage fields | LiteLLM response |
| latency | client timing |
| route reason | PolicyRouter |

## Outputs

M5 should expose a reviewable task-level usage record. Storage can begin as append-only local state before any database is introduced.

```json
{
  "task_id": "string",
  "request_id": "string",
  "model_alias": "code-cheap",
  "returned_model": "provider/model",
  "route_reason": "string",
  "latency_ms": 1234,
  "prompt_tokens": 100,
  "completion_tokens": 50,
  "total_tokens": 150,
  "estimated_cost_usd": 0.001,
  "status": "success|failed",
  "error": "string optional",
  "created_at": "RFC3339 timestamp"
}
```

Storage:

- Use append-only local JSONL first, for example `data/usage.jsonl`.
- The directory should be configurable.
- The usage log must be reviewable with standard shell tools.

## Current Implementation

- Usage records are written as append-only JSONL.
- Default path is `data/usage.jsonl`.
- `NANOBOT_USAGE_LOG_PATH` overrides the path.
- Success records include task id, request id, selected model alias, returned model, route reason, latency, and token usage when LiteLLM returns it.
- Downstream failure records include task id, request id, selected model alias, route reason, latency, failed status, and an error summary.
- Usage write failure is logged but does not fail the task response.
- Cost is not calculated locally in M5.

## Error Cases

| Case | Expected |
|---|---|
| usage missing from provider | store task record with usage fields empty |
| write failure | task succeeds but logs usage failure |
| malformed usage values | reject usage record, not task result |

## Config

Cost mapping may come from LiteLLM first. Custom cost tables should not be added until M5 proves what fields LiteLLM returns.

Recommended config:

| Name | Default | Purpose |
|---|---|---|
| `NANOBOT_USAGE_LOG_PATH` | `data/usage.jsonl` | append-only usage record path |

## Test Matrix

| Test | Expected |
|---|---|
| records task id and latency | pass |
| records returned model | pass |
| records selected model alias and route reason | pass |
| handles missing usage | pass |
| usage write failure does not fail task | pass |
| failed LiteLLM call records failure state | pass |
| real provider smoke writes a usage record | pass |

## Non-goals

- No billing product.
- No dashboard in M5.
- No database until append-only local records prove insufficient.
- No custom gateway accounting replacement yet.
