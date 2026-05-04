# Test Plan: UsageLogger

## Unit Tests

| Test | Expected |
|---|---|
| writes usage record | append-only record is created |
| records task id and request id | fields are present |
| records selected model alias | `model_alias` is present |
| records returned model | `returned_model` is present when LiteLLM returns it |
| records route reason | `route_reason` is present |
| records latency | `latency_ms` is present |
| handles missing usage | token fields may be empty without failing task |
| write failure does not fail task | task response still succeeds, logger error is logged |
| failed LiteLLM call records failure | record has `status=failed` and error summary |

## Integration Tests

| Test | Setup | Expected |
|---|---|---|
| successful review writes JSONL record | fake reviewer / temp file | one valid JSON record |
| downstream error writes JSONL record | fake LiteLLM error | failed usage record exists |
| usage log path configurable | temp path env/config | record written to configured path |

## Real Provider Smoke

| Test | Setup | Expected |
|---|---|---|
| real provider success writes usage | Go service + LiteLLM + Xiaomi MiMo | `200`, task result, usage record with route/model/latency |
| real provider routed request writes usage | `budget_hint=high_quality` | usage record has `model_alias=code-smart` |

## Non-goals

- No dashboard.
- No database.
- No billing product.
- No custom cost table unless LiteLLM fields are insufficient and documented first.
