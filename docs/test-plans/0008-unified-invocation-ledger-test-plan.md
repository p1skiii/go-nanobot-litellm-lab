# Test Plan: Unified Invocation Ledger

## Unit Tests

| Test | Expected |
|---|---|
| append/read invocation records | recent records returned in append order |
| filter by task id | only matching records returned |
| filter by run id | only matching records returned |
| concurrent append and read | no malformed JSONL error |
| usage projection | `usage.Record` fields are filled from invocation record |
| rejected records without task id | skipped by legacy usage projection |

## API Tests

| Test | Expected |
|---|---|
| success review task | response includes `run_id` and `attempt_id` |
| validation failure | writes rejected invocation record without task id |
| timeout failure | writes failed invocation with `error_kind=timeout` |
| downstream failure | writes failed invocation with `error_kind=downstream` |
| `/invocations/recent` | returns invocation records |
| `/invocations/tasks/{id}` | returns records for task |
| `/invocations/runs/{run_id}` | returns records for run |
| legacy `/usage/recent` | returns projected usage records |
| legacy `/usage/tasks/{id}` | returns projected usage records |

## Smoke Tests

| Command | Expected |
|---|---|
| `go test ./...` | pass |
| `docker compose -f deploy/docker-compose.yml config` | pass |
| `docker compose -f deploy/docker-compose.yml up --build -d` | pass |
| `scripts/smoke-real-provider.sh` | real provider request writes invocation record |
| `scripts/smoke-failure-cases.sh` | validation, timeout, downstream, and recovery records are queryable |

## Non-goals

- Do not validate streaming.
- Do not validate fallback.
- Do not compare router quality yet.
- Do not change ContextManager rules.
