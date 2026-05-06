# Test Plan: Local Lab Query Loop

## Unit Tests

| Test | Expected |
|---|---|
| JSONL ledger returns recent records | newest N records returned in chronological order |
| JSONL ledger filters by task id | only matching records returned |
| JSONL ledger filters by run id | only matching records returned |
| missing JSONL file | empty records, no error |
| malformed JSONL file | error returned |
| invalid recent limit | HTTP 400 |
| concurrent append and read | no malformed JSONL error |

## API Tests

| Test | Setup | Expected |
|---|---|---|
| `GET /invocations/recent` | temp ledger | records returned |
| `GET /invocations/tasks/{id}` | temp ledger | matching task records returned |
| `GET /invocations/runs/{run_id}` | temp ledger | matching run records returned |
| `GET /usage/recent` | temp ledger | projected records returned |
| `GET /usage/tasks/{id}` | temp ledger | projected task records returned |

## Real Provider Smoke

| Test | Setup | Expected |
|---|---|---|
| review then recent invocations | Go + LiteLLM + Xiaomi MiMo | `200`, task id, route reason, usage tokens |
| compose review then invocation lookup | Docker Compose + Xiaomi MiMo | task id appears in `/invocations/tasks/{id}` |

## Non-goals

- No dashboard.
- No database.
- No auth.
