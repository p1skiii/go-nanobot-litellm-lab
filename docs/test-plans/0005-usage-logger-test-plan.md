# Test Plan: Usage Projection

## Unit Tests

| Test | Expected |
|---|---|
| projection from nested usage | `usage.Record` contains token fields |
| projection from invocation metadata | `usage.Record` contains task/model/status fields |
| rejected invocation without task id | omitted from usage projection |
| invocation write failure | task response still follows API behavior and error is logged |

## API Tests

| Test | Setup | Expected |
|---|---|---|
| `GET /usage/recent` | Invocation Ledger records | projected records returned |
| `GET /usage/tasks/{id}` | Invocation Ledger records | matching projected records returned |
| review then usage lookup | fake reviewer | task id appears in usage projection |

## Real Provider Smoke

| Test | Setup | Expected |
|---|---|---|
| real provider success writes invocation with usage | Go service + LiteLLM + Xiaomi MiMo | `200`, task result, invocation record with nested usage |
| legacy usage projection still works | same run | `/usage/tasks/{id}` returns projected usage |

## Non-goals

- No separate usage JSONL.
- No dashboard.
- No database.
- No billing product.
