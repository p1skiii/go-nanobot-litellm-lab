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
  "latency_ms": 1234,
  "input_tokens": 100,
  "output_tokens": 50,
  "estimated_cost_usd": 0.001,
  "route_reason": "string"
}
```

## Error Cases

| Case | Expected |
|---|---|
| usage missing from provider | store task record with usage fields empty |
| write failure | task succeeds but logs usage failure |
| malformed usage values | reject usage record, not task result |

## Config

Cost mapping may come from LiteLLM first. Custom cost tables should not be added until M5 proves what fields LiteLLM returns.

## Test Matrix

| Test | Expected |
|---|---|
| records task id and latency | pass |
| records returned model | pass |
| handles missing usage | pass |
| usage write failure does not fail task | pass |

## Non-goals

- No billing product.
- No dashboard in M5.
- No database until append-only local records prove insufficient.
